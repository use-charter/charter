package secrets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	targets, err := sec001Targets(root, inv)
	if err != nil {
		return findings.Finding{}, false
	}

	for _, target := range targets {
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

func sec001Targets(root string, inv repository.Inventory) ([]string, error) {
	tracked, err := trackedSEC001Paths(root)
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(agentVisibleFileTargets))
	for _, candidate := range agentVisibleFileTargets {
		if inv.Has(candidate) && tracked[candidate] {
			targets = append(targets, candidate)
		}
	}
	for _, path := range inv.Paths {
		if strings.HasPrefix(path, ".cursor/rules/") && tracked[path] {
			targets = append(targets, path)
		}
	}
	sort.Strings(targets)
	return targets, nil
}

func trackedSEC001Paths(root string) (map[string]bool, error) {
	// #nosec G204 -- root is the resolved repository root for the active scan target.
	cmd := exec.Command("git", "-C", root, "ls-files", "-z", "--cached", "--full-name")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("list tracked agent-visible files: %w: %s", err, strings.TrimSpace(string(output)))
	}

	tracked := make(map[string]bool)
	for _, raw := range strings.Split(string(output), "\x00") {
		if raw == "" {
			continue
		}
		tracked[filepath.ToSlash(raw)] = true
	}

	return tracked, nil
}
