package secrets

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
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
	targets, err := sec001Targets(root, inv)
	if err != nil {
		return findings.Finding{}, false, fmt.Errorf("AE-SEC-001: %w", err)
	}

	for _, target := range targets {
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
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list tracked agent-visible files: %w: %s", err, strings.TrimSpace(stderr.String()))
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
