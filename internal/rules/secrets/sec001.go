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
		data, err := readSEC001Target(root, target)
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
	targets := make([]string, 0, len(agentVisibleFileTargets)+1)
	for _, candidate := range agentVisibleFileTargets {
		if inv.Has(candidate) {
			targets = append(targets, candidate)
		}
	}
	if hasCursorRules(inv) {
		targets = append(targets, ".cursor/rules")
	}
	return targets
}

func hasCursorRules(inv repository.Inventory) bool {
	for _, path := range inv.Paths {
		if strings.HasPrefix(path, ".cursor/rules/") {
			return true
		}
	}
	return false
}

func readSEC001Target(root, target string) ([]byte, error) {
	if target != ".cursor/rules" {
		// #nosec G304 -- target is constrained to known agent-visible surfaces.
		return os.ReadFile(filepath.Join(root, filepath.FromSlash(target)))
	}

	entries, err := os.ReadDir(filepath.Join(root, ".cursor", "rules"))
	if err != nil {
		return nil, err
	}

	chunks := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// #nosec G304 -- file names come from tracked .cursor/rules directory entries.
		data, err := os.ReadFile(filepath.Join(root, ".cursor", "rules", entry.Name()))
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, string(data))
	}

	return []byte(strings.Join(chunks, "\n")), nil
}
