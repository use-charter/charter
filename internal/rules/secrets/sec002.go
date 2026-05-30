package secrets

import (
	"os"
	"path/filepath"
	"strings"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
	sharedsecrets "go.charter.dev/charter/internal/secrets"
)

var sec002Targets = []string{
	".mcp.json",
	".mcp.yml",
	".cursor/mcp.json",
	".claude/settings.json",
	"claude_desktop_config.json",
	"cline_mcp_settings.json",
}

func checkSEC002(root string, inv repository.Inventory) (findings.Finding, bool) {
	for _, target := range sec002Targets {
		if !inv.Has(target) {
			continue
		}

		if finding, ok := scanSEC002File(root, target); ok {
			return finding, true
		}
	}

	for _, path := range inv.Paths {
		if !isConfigAdjacentPKL(path) {
			continue
		}

		if finding, ok := scanSEC002File(root, path); ok {
			return finding, true
		}
	}

	return findings.Finding{}, false
}

func scanSEC002File(root, rel string) (findings.Finding, bool) {
	// #nosec G304 -- rel is constrained to inventory-backed MCP/config paths.
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return findings.Finding{}, false
	}

	for i, line := range strings.Split(string(data), "\n") {
		match := sharedsecrets.DetectLine(line)
		if !match.Found {
			continue
		}

		return findings.Finding{
			RuleID:      "AE-SEC-002",
			Severity:    findings.SeverityBlocker,
			Category:    "Secrets",
			Summary:     "Literal secret detected in MCP or adjacent config",
			Remediation: "Replace the literal secret with an environment variable reference",
			Evidence:    []string{rel + ": " + sharedsecrets.RedactValue(match.Secret)},
			Locations:   []findings.Location{{Path: rel, Line: i + 1}},
			Cap:         secretScoreCap,
		}, true
	}

	return findings.Finding{}, false
}

func isConfigAdjacentPKL(path string) bool {
	if filepath.Ext(path) != ".pkl" {
		return false
	}

	for _, part := range strings.Split(filepath.ToSlash(strings.ToLower(path)), "/") {
		if strings.Contains(part, "mcp") || strings.Contains(part, "config") {
			return true
		}
	}

	return false
}
