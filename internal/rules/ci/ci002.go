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

		text := string(data)
		markCoverage(text, coverage)
		evidence = append(evidence, unpinnedActionEvidence(text, rel)...)
	}

	if !coverage["repo-quality"] {
		evidence = append(evidence, "missing repo quality workflow coverage")
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

func markCoverage(text string, coverage map[string]bool) {
	if strings.Contains(text, "moon run :check") || (strings.Contains(text, "moon run :lint") && strings.Contains(text, "moon run :vet") && strings.Contains(text, "moon run :test") && strings.Contains(text, "moon run :build") && strings.Contains(text, "moon run :docs") && strings.Contains(text, "moon run :eval")) {
		coverage["repo-quality"] = true
	}
	if strings.Contains(text, "moon run :actionlint") && strings.Contains(text, "moon run :zizmor") {
		coverage["workflow-lint"] = true
	}
	if strings.Contains(text, "moon run :security") {
		coverage["security"] = true
	}
}

func unpinnedActionEvidence(text string, rel string) []string {
	var evidence []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "- ")
		if !strings.HasPrefix(trimmed, "uses:") {
			continue
		}

		ref := strings.TrimSpace(strings.TrimPrefix(trimmed, "uses:"))
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
