package context

import (
	"os"
	"path/filepath"
	"sort"
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

		return findings.Finding{
			RuleID:      "AE-CTX-001",
			Severity:    findings.SeverityBlocker,
			Category:    "Context",
			Summary:     "Agent context file is present but not meaningful enough for agent use",
			Remediation: "Add concrete project guidance and a verification command to the root context file.",
			Evidence:    []string{candidate},
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

	if countNonEmptyLines(trimmed) < 3 {
		return false
	}

	return strings.Contains(trimmed, "moon run :check") || strings.Contains(trimmed, "charter doctor")
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
