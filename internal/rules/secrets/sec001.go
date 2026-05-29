package secrets

import (
	"os"
	"path/filepath"
	"strings"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
	sharedsecrets "go.charter.dev/charter/internal/secrets"
)

var agentVisibleFileTargets = []string{
	"AGENTS.md",
	"CLAUDE.md",
	".cursor/rules",
	".windsurfrules",
	".github/copilot-instructions.md",
	"opencode.md",
	"codex.md",
	"DESIGN.md",
	"SKILL.md",
}

func RunSecretRules(root string, inv repository.Inventory) []findings.Finding {
	var out []findings.Finding
	if finding, ok := checkSEC001(root, inv); ok {
		out = append(out, finding)
	}
	return out
}

func checkSEC001(root string, inv repository.Inventory) (findings.Finding, bool) {
	for _, target := range sec001Targets(inv) {
		// #nosec G304 -- target is constrained to agent-visible inventory paths.
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target)))
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(data), "\n") {
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
			}, true
		}
	}

	return findings.Finding{}, false
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
	return targets
}
