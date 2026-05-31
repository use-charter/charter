package mcp

import (
	"fmt"
	"os"
	"path/filepath"

	"go.use-charter.dev/charter/internal/config"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

// isMCPConfigPath reports whether p is an MCP config file Charter scans. Keep
// this list consistent with AE-SEC-002's MCP targets (future drift-guard candidate).
func isMCPConfigPath(p string) bool {
	switch p {
	case ".mcp.json", "mcp.json", ".cursor/mcp.json", ".vscode/mcp.json":
		return true
	}
	return false
}

// Run evaluates AE-MCP-001/002/003 across all MCP config files in the repo. It
// fails fast (returns a wrapped error) when a discovered MCP config or
// charter.yaml is unreadable or malformed, mirroring gosecrets.RunSecretRules.
func Run(root string, inv repository.Inventory) ([]findings.Finding, error) {
	var paths []string
	for _, p := range inv.Paths {
		if isMCPConfigPath(p) {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		return nil, nil
	}

	files := make([]ConfigFile, 0, len(paths))
	for _, rel := range paths {
		// #nosec G304 -- rel is a fixed MCP config path from the tracked inventory.
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", rel, err)
		}
		cf, err := parseConfigFile(rel, data)
		if err != nil {
			return nil, err
		}
		files = append(files, cf)
	}

	allow, err := config.LoadTrustedRemotes(root, inv)
	if err != nil {
		return nil, err
	}

	var all []findings.Finding
	all = append(all, checkPinning(files)...)
	all = append(all, checkTrustedRemotes(files, allow)...)
	all = append(all, checkRemoteAuth(files)...)
	return all, nil
}
