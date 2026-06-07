package secrets

import (
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/agentcontext"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
	sharedsecrets "go.use-charter.dev/charter/internal/secrets"
)

// secretScoreCap is the maximum final score a repository may earn while any
// raw secret finding is present.
const secretScoreCap = 49

// agentVisibleFileTargets is the canonical context-file set plus the Cursor
// rules directory, so AE-SEC-001 scans exactly the files the context rules
// recognize as agent context.
var agentVisibleFileTargets = append(append([]string{}, agentcontext.Files...), agentcontext.CursorRulesDir)

func RunSecretRules(root string, inv repository.Inventory) ([]findings.Finding, error) {
	var out []findings.Finding

	finding, ok, err := checkSEC001(root, inv)
	if err != nil {
		return nil, err
	}
	if ok {
		out = append(out, finding)
	}

	if finding, ok := checkSEC002(root, inv); ok {
		out = append(out, finding)
	}

	return out, nil
}

func checkSEC001(root string, inv repository.Inventory) (findings.Finding, bool, error) {
	for _, target := range sec001Targets(inv) {
		content, ok := repository.ReadTrackedFile(root, inv, target)
		if !ok {
			continue
		}

		for i, line := range strings.Split(content, "\n") {
			match := sharedsecrets.DetectLine(line)
			if !match.Found {
				continue
			}

			return findings.Finding{
				RuleID:      "AE-SEC-001",
				Severity:    findings.SeverityBlocker,
				Category:    "Secrets",
				Summary:     "Secret detected in agent-visible context file",
				Remediation: "Remove the literal secret and reference an environment variable instead",
				Evidence:    []string{target + ": " + sharedsecrets.RedactValue(match.Secret)},
				Locations:   []findings.Location{{Path: target, Line: i + 1}},
				Cap:         secretScoreCap,
			}, true, nil
		}
	}

	return findings.Finding{}, false, nil
}

func sec001Targets(inv repository.Inventory) []string {
	targets := make([]string, 0, len(agentVisibleFileTargets))
	for _, candidate := range agentVisibleFileTargets {
		if inv.Has(candidate) {
			targets = append(targets, candidate)
		}
	}
	for _, path := range inv.Paths {
		if strings.HasPrefix(path, ".cursor/rules/") {
			targets = append(targets, path)
		}
	}
	sort.Strings(targets)
	return targets
}
