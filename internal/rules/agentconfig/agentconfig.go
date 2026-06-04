package agentconfig

import (
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

func isHookConfigPath(p string) bool {
	switch p {
	case ".claude/settings.json", ".claude/settings.local.json", ".cursor/hooks.json":
		return true
	}
	return false
}

// Run evaluates AE-CC-001 (dangerous hook commands) and AE-CC-002 (edit scope).
// Each hook config is read through repository.ReadTrackedFile (inventory-gated,
// symlink-contained, size-capped); a tracked config that fails a safety gate is
// skipped, while a malformed one still fails fast with a wrapped error,
// mirroring the MCP scanner.
func Run(root string, inv repository.Inventory) ([]findings.Finding, error) {
	var files []ConfigFile
	for _, p := range inv.Paths {
		if !isHookConfigPath(p) {
			continue
		}
		content, ok := repository.ReadTrackedFile(root, inv, p)
		if !ok {
			continue
		}
		cf, err := parseHookConfig(p, []byte(content))
		if err != nil {
			return nil, err
		}
		files = append(files, cf)
	}

	var all []findings.Finding
	all = append(all, checkDangerousCommands(files)...)

	cc, err := checkEditScope(root, inv)
	if err != nil {
		return nil, err
	}
	all = append(all, cc...)
	return all, nil
}
