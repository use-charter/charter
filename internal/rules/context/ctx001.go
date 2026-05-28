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
	for _, candidate := range supportedContextFiles {
		if !inv.Has(candidate) {
			continue
		}

		// #nosec G304 -- candidate is constrained to the supported root context file set.
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(candidate)))
		if err != nil {
			return findings.Finding{
				RuleID:      "AE-CTX-001",
				Severity:    findings.SeverityBlocker,
				Category:    "Context",
				Summary:     "Agent context file could not be read",
				Remediation: "Restore a readable root context file with project guidance and a verification command.",
				Evidence:    []string{candidate},
			}, true
		}

		if isMeaningfulContext(string(data)) {
			return findings.Finding{}, false
		}

		evidence := []string{"context file: " + candidate}
		evidence = append(evidence, contextExcerptEvidence(string(data))...)
		evidence = append(evidence, missingContextSignals(string(data))...)

		return findings.Finding{
			RuleID:      "AE-CTX-001",
			Severity:    findings.SeverityBlocker,
			Category:    "Context",
			Summary:     "Agent context file is present but not meaningful enough for agent use",
			Remediation: "Add concrete project guidance and a verification command to the root context file.",
			Evidence:    evidence,
		}, true
	}

	evidence := append([]string(nil), supportedContextFiles...)
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

func contextExcerptEvidence(content string) []string {
	var evidence []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		evidence = append(evidence, "context excerpt: "+trimmed)
		if len(evidence) == 3 {
			break
		}
	}
	return evidence
}

func estimatedTokenCount(content string) int {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return 0
	}
	return (len(trimmed) + 3) / 4
}
