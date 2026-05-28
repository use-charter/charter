package ci

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
)

var pinnedActionRefPattern = regexp.MustCompile(`@[a-f0-9]{40}(\s|$)`)

func Run(root string, inv repository.Inventory) []findings.Finding {
	workflowPaths := collectWorkflowPaths(inv)
	if len(workflowPaths) == 0 {
		return []findings.Finding{{
			RuleID:      "AE-CI-002",
			Severity:    findings.SeverityLow,
			Category:    "CI",
			Summary:     "Repository does not define any GitHub workflow coverage for Charter-related quality gates",
			Remediation: "Add workflow coverage for repo quality, workflow linting, and security checks.",
			Evidence:    []string{".github/workflows/"},
		}}
	}

	coverage := map[string]bool{}
	var evidence []string

	for _, rel := range workflowPaths {
		// #nosec G304 -- rel comes from the inventory's tracked workflow paths under .github/workflows/.
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			evidence = append(evidence, "workflow unreadable: "+rel)
			continue
		}

		executable := extractWorkflowExecutables(string(data))
		markCoverage(executable, coverage)
		evidence = append(evidence, unpinnedActionEvidence(executable.Uses, rel)...)
	}

	if !coverage["repo-quality"] {
		evidence = append(evidence, "missing repo quality workflow coverage")
	}
	if !coverage["charter-product-gate"] && !hasDeferredProductGateDocumentation(root, inv) {
		evidence = append(evidence, "missing charter doctor CI gate or documented bootstrap deferment")
	}
	if !coverage["workflow-lint"] {
		evidence = append(evidence, "missing workflow lint coverage")
	}
	if !coverage["security"] {
		evidence = append(evidence, "missing security workflow coverage")
	}

	if len(evidence) == 0 {
		return nil
	}

	sort.Strings(evidence)
	return []findings.Finding{{
		RuleID:      "AE-CI-002",
		Severity:    findings.SeverityLow,
		Category:    "CI",
		Summary:     "GitHub workflow coverage is incomplete for Charter-related quality gates",
		Remediation: "Add the missing workflow checks and pin third-party actions to immutable SHAs.",
		Evidence:    evidence,
	}}
}

func collectWorkflowPaths(inv repository.Inventory) []string {
	var paths []string
	for _, path := range inv.Paths {
		if strings.HasPrefix(path, ".github/workflows/") && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

type workflowExecutables struct {
	Runs []string
	Uses []string
}

func markCoverage(executable workflowExecutables, coverage map[string]bool) {
	if hasRunContaining(executable.Runs, "moon run :check") || hasAllRuns(executable.Runs, "moon run :lint", "moon run :vet", "moon run :test", "moon run :build", "moon run :docs", "moon run :eval") {
		coverage["repo-quality"] = true
	}
	if hasRunContaining(executable.Runs, "charter doctor") || hasUsePrefix(executable.Uses, "use-charter/charter-action@") {
		coverage["charter-product-gate"] = true
	}
	if hasRunContaining(executable.Runs, "moon run :actionlint") && hasRunContaining(executable.Runs, "moon run :zizmor") {
		coverage["workflow-lint"] = true
	}
	if hasRunContaining(executable.Runs, "moon run :security") {
		coverage["security"] = true
	}
}

func unpinnedActionEvidence(uses []string, rel string) []string {
	var evidence []string
	for _, ref := range uses {
		if strings.HasPrefix(ref, "./") {
			continue
		}
		if !strings.Contains(ref, "@") {
			evidence = append(evidence, "unpinned action: "+rel+" -> "+ref)
			continue
		}
		if !pinnedActionRefPattern.MatchString(ref) {
			evidence = append(evidence, "unpinned action: "+rel+" -> "+ref)
		}
	}
	return evidence
}

func extractWorkflowExecutables(text string) workflowExecutables {
	var result workflowExecutables
	lines := strings.Split(text, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		normalized := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		switch {
		case strings.HasPrefix(normalized, "run:"):
			value := strings.TrimSpace(strings.TrimPrefix(normalized, "run:"))
			if value == "|" || value == ">" || value == "|-" || value == ">-" {
				block, next := collectIndentedBlock(lines, i+1, leadingIndent(line))
				if block != "" {
					result.Runs = append(result.Runs, block)
				}
				i = next - 1
				continue
			}
			if value != "" {
				result.Runs = append(result.Runs, value)
			}
		case strings.HasPrefix(normalized, "uses:"):
			value := strings.TrimSpace(strings.TrimPrefix(normalized, "uses:"))
			if value != "" {
				result.Uses = append(result.Uses, value)
			}
		}
	}

	return result
}

func collectIndentedBlock(lines []string, start int, parentIndent int) (string, int) {
	var block []string
	i := start
	for ; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if leadingIndent(line) <= parentIndent {
			break
		}
		block = append(block, trimmed)
	}
	return strings.Join(block, "\n"), i
}

func leadingIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func hasRunContaining(runs []string, want string) bool {
	for _, run := range runs {
		if strings.Contains(run, want) {
			return true
		}
	}
	return false
}

func hasAllRuns(runs []string, wants ...string) bool {
	for _, want := range wants {
		if !hasRunContaining(runs, want) {
			return false
		}
	}
	return true
}

func hasUsePrefix(uses []string, prefix string) bool {
	for _, use := range uses {
		if strings.HasPrefix(use, prefix) {
			return true
		}
	}
	return false
}

func hasDeferredProductGateDocumentation(root string, inv repository.Inventory) bool {
	if hasDocMarker(root, inv, "docs/internal/audit/charter-v1-audit-checklist.md", "deferred until phase 1 scanner implementation exists") {
		return true
	}

	phaseNotStarted := hasDocMarker(root, inv, "README.md", "phase 1 implementation not started") || hasDocMarker(root, inv, "AGENTS.md", "phase 1 implementation not started")
	bootstrapCLI := hasDocMarker(root, inv, "AGENTS.md", "current cli: bootstrap placeholder only")
	return phaseNotStarted && bootstrapCLI
}

func hasDocMarker(root string, inv repository.Inventory, path string, marker string) bool {
	if !inv.Has(path) {
		return false
	}
	// #nosec G304 -- path is constrained to a fixed documentation file checked by this rule.
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), marker)
}
