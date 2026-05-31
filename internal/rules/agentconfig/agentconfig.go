package agentconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
)

func isHookConfigPath(p string) bool {
	switch p {
	case ".claude/settings.json", ".claude/settings.local.json", ".cursor/hooks.json":
		return true
	}
	return false
}

// Run evaluates AE-CC-001 (dangerous hook commands) and AE-CC-002 (edit scope).
// It fails fast (wrapped error) when a discovered JSON hook config is unreadable
// or malformed, mirroring the MCP scanner.
func Run(root string, inv repository.Inventory) ([]findings.Finding, error) {
	var files []ConfigFile
	for _, p := range inv.Paths {
		if !isHookConfigPath(p) {
			continue
		}
		// #nosec G304 -- p is a fixed hook-config path from the tracked inventory.
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		cf, err := parseHookConfig(p, data)
		if err != nil {
			return nil, err
		}
		files = append(files, cf)
	}

	var all []findings.Finding
	all = append(all, checkDangerousCommands(files)...)
	all = append(all, checkEditScope(root, inv)...)
	return all, nil
}
