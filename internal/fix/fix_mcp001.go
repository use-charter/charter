package fix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.use-charter.dev/charter/internal/catalog"
	"go.use-charter.dev/charter/internal/repository"
	"go.use-charter.dev/charter/internal/rules/mcp"
)

// fixMCP001 bumps MCP server packages to a catalog-known safe version: an
// advisory-affected pin to its fixed version, and an unpinned or behind-stable
// cataloged package to the catalog stable version. It NEVER rewrites a
// deprecated/archived package (the successor is a different package/transport,
// so migration is manual — ADR-0021) and never changes anything but the version
// token. Returns one Replace plan per affected config file. Pure: no writes.
func fixMCP001(root string, inv repository.Inventory) ([]FilePlan, error) {
	rewrites, err := mcp.PlanPins(root, inv, catalog.Default())
	if err != nil {
		return nil, err
	}
	if len(rewrites) == 0 {
		return nil, nil
	}

	byFile := map[string][]mcp.PinRewrite{}
	var order []string
	for _, r := range rewrites {
		if _, seen := byFile[r.Path]; !seen {
			order = append(order, r.Path)
		}
		byFile[r.Path] = append(byFile[r.Path], r)
	}

	var plans []FilePlan
	for _, path := range order {
		// #nosec G304 -- path is a fixed MCP config path from the tracked inventory.
		existing, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if readErr != nil {
			return nil, fmt.Errorf("fix: read %s: %w", path, readErr)
		}
		updated := string(existing)
		for _, r := range byFile[path] {
			// Replace the exact quoted JSON arg so the match is bounded to the
			// package-spec token and cannot collide with a substring.
			oldQ := `"` + r.OldToken + `"`
			newQ := `"` + r.NewToken + `"`
			if !strings.Contains(updated, oldQ) {
				continue
			}
			updated = strings.ReplaceAll(updated, oldQ, newQ)
		}
		if updated == string(existing) {
			continue
		}
		plans = append(plans, FilePlan{
			RuleID:   "AE-MCP-001",
			Path:     path,
			Action:   Replace,
			Contents: []byte(updated),
			Diff:     buildReplaceDiff(path, existing, []byte(updated)),
		})
	}
	return plans, nil
}
