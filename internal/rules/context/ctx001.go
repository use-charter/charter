package context

import (
	"sort"
	"strconv"
	"strings"

	"go.use-charter.dev/charter/internal/agentcontext"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

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
	if finding, ok := checkCTX006(root, inv); ok {
		out = append(out, finding)
	}

	return out
}

func checkCTX001(root string, inv repository.Inventory) (findings.Finding, bool) {
	locations := supportedContextLocations(inv)
	if len(locations) > 0 {
		candidate := locations[0]
		content, ok := readContextCandidate(root, inv, candidate)
		if !ok {
			return findings.Finding{
				RuleID:      "AE-CTX-001",
				Severity:    findings.SeverityBlocker,
				Category:    "Context",
				Summary:     "Agent context file could not be read",
				Remediation: "Restore a readable root context file with project guidance and a verification command.",
				Evidence:    []string{"context location: " + candidate},
				Locations:   []findings.Location{{Path: candidate}},
			}, true
		}

		if estimatedTokenCount(content) > 600 {
			evidence := []string{"context location: " + candidate}
			evidence = append(evidence, contextBudgetEvidence(content)...)
			evidence = append(evidence, contextShapeEvidence(content)...)
			return findings.Finding{
				RuleID:      "AE-CTX-001",
				Severity:    findings.SeverityBlocker,
				Category:    "Context",
				Summary:     "Agent context file exceeds the 600-token budget",
				Remediation: "Trim the root context file to ≤600 tokens while keeping project guidance and a verification command.",
				Evidence:    evidence,
				Locations:   []findings.Location{{Path: candidate}},
			}, true
		}

		if isMeaningfulContext(content) {
			return findings.Finding{}, false
		}

		evidence := []string{"context location: " + candidate}
		evidence = append(evidence, contextShapeEvidence(content)...)
		evidence = append(evidence, firstSubstantiveLineEvidence(content))
		evidence = append(evidence, missingContextSignals(content)...)

		return findings.Finding{
			RuleID:      "AE-CTX-001",
			Severity:    findings.SeverityBlocker,
			Category:    "Context",
			Summary:     "Agent context file is present but not meaningful enough for agent use",
			Remediation: "Add concrete project guidance and a verification command to the root context file.",
			Evidence:    evidence,
			Locations:   []findings.Location{{Path: candidate}},
		}, true
	}

	evidence := append([]string(nil), agentcontext.Files...)
	evidence = append(evidence, agentcontext.CursorRulesDir)
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
	for _, candidate := range agentcontext.Files {
		if inv.Has(candidate) {
			locations = append(locations, candidate)
		}
	}
	if hasCursorRules(inv) {
		locations = append(locations, agentcontext.CursorRulesDir)
	}
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

// readContextCandidate returns the content of a single-file context candidate,
// or the concatenation of the tracked .cursor/rules files for the directory
// candidate. Every read is routed through repository.ReadTrackedFile so it is
// inventory-gated, symlink-contained, and size-capped. The .cursor/rules tree
// is enumerated from the inventory (mirroring cursorRuleFiles in
// internal/rules/agentconfig/cc002.go) rather than walked on disk, so a tracked
// rule file that fails a safety gate is simply skipped — matching the standard
// rule contract — instead of failing the whole check.
func readContextCandidate(root string, inv repository.Inventory, candidate string) (string, bool) {
	if candidate != agentcontext.CursorRulesDir {
		return repository.ReadTrackedFile(root, inv, candidate)
	}

	var rels []string
	prefix := agentcontext.CursorRulesDir + "/"
	for _, p := range inv.Paths {
		if strings.HasPrefix(p, prefix) {
			rels = append(rels, p)
		}
	}
	sort.Strings(rels)

	var chunks []string
	for _, rel := range rels {
		if content, ok := repository.ReadTrackedFile(root, inv, rel); ok {
			chunks = append(chunks, content)
		}
	}

	return strings.Join(chunks, "\n"), true
}

func isMeaningfulContext(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}

	if countNonEmptyLines(trimmed) < 5 {
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
	if !hasVerificationSignal(lower) {
		missing = append(missing, "missing content signal: verification command")
	}
	if countNonEmptyLines(content) < 5 || len(strings.TrimSpace(content)) < 120 {
		missing = append(missing, "missing content signal: reasonable content")
	}

	return missing
}

// hasVerificationSignal reports whether the context references a recognized
// verification command. Shared by AE-CTX-001 (must have one) and AE-CTX-002
// (the stated command must be recognized, not hardcoded to this repo's
// `moon run :check`).
func hasVerificationSignal(lower string) bool {
	return hasAnySignal(lower, "moon run :check", "charter doctor", "verification command", "verify with")
}

func contextBudgetEvidence(content string) []string {
	if estimatedTokenCount(content) <= 600 {
		return nil
	}

	return []string{"context appears over budget: ~" + strconv.Itoa(estimatedTokenCount(content)) + " tokens"}
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

func firstSubstantiveLineEvidence(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if lineNeedsRedaction(trimmed) {
			return "first substantive line: [redacted]"
		}
		return "first substantive line: " + safeEvidenceLine(trimmed)
	}

	return "first substantive line: [none]"
}

func lineNeedsRedaction(line string) bool {
	if strings.Contains(line, "-----BEGIN") && strings.Contains(line, "PRIVATE KEY-----") {
		return true
	}
	lower := strings.ToLower(line)
	for _, token := range []string{"api_key", "apikey", "token", "secret", "password", "passwd", "private key", "bearer ", "authorization:"} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	for _, token := range []string{"sk-", "ghp_", "github_pat_", "xoxb-", "akia"} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func safeEvidenceLine(line string) string {
	line = strings.Join(strings.Fields(line), " ")
	if len(line) > 120 {
		return line[:117] + "..."
	}
	return line
}

func estimatedTokenCount(content string) int {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return 0
	}
	return (len(trimmed) + 3) / 4
}
