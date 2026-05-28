package context

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
)

var supportedContextFiles = []string{
	"AGENTS.md",
	"CLAUDE.md",
	".windsurfrules",
	".github/copilot-instructions.md",
	"opencode.md",
	"codex.md",
	"DESIGN.md",
	"SKILL.md",
}

func RunCTXRules(root string, inv repository.Inventory) []findings.Finding {
	var out []findings.Finding

	if finding, ok := checkCTX001(root, inv); ok {
		out = append(out, finding)
	}
	if finding, ok := checkCTX002(root, inv); ok {
		out = append(out, finding)
	}
	if finding, ok := checkCTX004(root, inv); ok {
		out = append(out, finding)
	}

	return out
}

func checkCTX001(root string, inv repository.Inventory) (findings.Finding, bool) {
	var firstFailure *findings.Finding

	for _, candidate := range supportedContextLocations(inv) {
		data, err := readContextCandidate(root, candidate)
		if err != nil {
			finding := findings.Finding{
				RuleID:      "AE-CTX-001",
				Severity:    findings.SeverityBlocker,
				Category:    "Context",
				Summary:     "Agent context file could not be read",
				Remediation: "Restore a readable root context file with project guidance and a verification command.",
				Evidence:    []string{"context location: " + candidate},
			}
			if firstFailure == nil {
				firstFailure = &finding
			}
			continue
		}

		content := string(data)
		if isMeaningfulContext(content) {
			return findings.Finding{}, false
		}

		evidence := []string{"context location: " + candidate}
		evidence = append(evidence, contextShapeEvidence(content)...)
		evidence = append(evidence, missingContextSignals(content)...)

		finding := findings.Finding{
			RuleID:      "AE-CTX-001",
			Severity:    findings.SeverityBlocker,
			Category:    "Context",
			Summary:     "Agent context file is present but not meaningful enough for agent use",
			Remediation: "Add concrete project guidance and a verification command to the root context file.",
			Evidence:    evidence,
		}
		if firstFailure == nil {
			firstFailure = &finding
		}
	}

	if firstFailure != nil {
		return *firstFailure, true
	}

	evidence := append([]string(nil), supportedContextFiles...)
	evidence = append(evidence, ".cursor/rules")
	sort.Strings(evidence)
	return findings.Finding{
		RuleID:      "AE-CTX-001",
		Severity:    findings.SeverityBlocker,
		Category:    "Context",
		Summary:     "No supported agent context file exists at the repository root",
		Remediation: "Create a root context file such as AGENTS.md with project guidance and a verification command.",
		Evidence:    evidence,
	}, true
}

func supportedContextLocations(inv repository.Inventory) []string {
	var locations []string
	for _, candidate := range supportedContextFiles {
		if inv.Has(candidate) {
			locations = append(locations, candidate)
		}
	}
	if hasCursorRules(inv) {
		locations = append(locations, ".cursor/rules")
	}
	sort.Strings(locations)
	return locations
}

func hasCursorRules(inv repository.Inventory) bool {
	for _, path := range inv.Paths {
		if strings.HasPrefix(path, ".cursor/rules/") {
			return true
		}
	}
	return false
}

func readContextCandidate(root string, candidate string) ([]byte, error) {
	if candidate != ".cursor/rules" {
		// #nosec G304 -- candidate is constrained to the supported root context file set.
		return os.ReadFile(filepath.Join(root, filepath.FromSlash(candidate)))
	}

	var chunks []string
	entries, err := os.ReadDir(filepath.Join(root, ".cursor", "rules"))
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// #nosec G304 -- file names come from the tracked .cursor/rules directory entries.
		data, err := os.ReadFile(filepath.Join(root, ".cursor", "rules", entry.Name()))
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, string(data))
	}

	return []byte(strings.Join(chunks, "\n")), nil
}

func isMeaningfulContext(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}

	if countNonEmptyLines(trimmed) < 5 {
		return false
	}

	if estimatedTokenCount(trimmed) > 600 {
		return false
	}

	return len(missingContextSignals(trimmed)) == 0
}

func countNonEmptyLines(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func missingContextSignals(content string) []string {
	lower := strings.ToLower(content)
	var missing []string

	if !hasProjectSummarySignal(lower) {
		missing = append(missing, "missing content signal: project summary")
	}
	if !hasAnySignal(lower, "tech stack", "stack", "uses go", "uses bun", "uses python", "javascript", "typescript", "go ", "bun", "python", "node") {
		missing = append(missing, "missing content signal: tech stack")
	}
	if !hasAnySignal(lower, "off-limits", "edit scope", "edit boundaries", "safe for agent edits", "do not edit", "protected path") {
		missing = append(missing, "missing content signal: edit boundaries")
	}
	if !hasAnySignal(lower, "moon run :check", "charter doctor", "verification command", "verify with") {
		missing = append(missing, "missing content signal: verification command")
	}
	if countNonEmptyLines(content) < 5 || len(strings.TrimSpace(content)) < 120 {
		missing = append(missing, "missing content signal: reasonable content")
	}
	if estimatedTokenCount(content) > 600 {
		missing = append(missing, "context appears over budget: ~"+strconv.Itoa(estimatedTokenCount(content))+" tokens")
	}

	return missing
}

func hasProjectSummarySignal(content string) bool {
	if hasAnySignal(content, "project summary", "current state", "charter is") {
		return true
	}

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "`") {
			continue
		}
		if len(strings.Fields(trimmed)) >= 4 && len(trimmed) >= 20 {
			return true
		}
	}

	return false
}

func hasAnySignal(content string, signals ...string) bool {
	for _, signal := range signals {
		if strings.Contains(content, signal) {
			return true
		}
	}
	return false
}

func contextShapeEvidence(content string) []string {
	return []string{
		"non-empty lines: " + strconv.Itoa(countNonEmptyLines(content)),
		"estimated tokens: ~" + strconv.Itoa(estimatedTokenCount(content)),
	}
}

func estimatedTokenCount(content string) int {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return 0
	}
	return (len(trimmed) + 3) / 4
}
