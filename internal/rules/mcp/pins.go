package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"go.use-charter.dev/charter/internal/catalog"
	"go.use-charter.dev/charter/internal/repository"
)

// PinRewrite is a single safe AE-MCP-001 auto-fix: replace a package spec token
// with a catalog-known safe version in a specific MCP config file.
type PinRewrite struct {
	Path     string // repo-relative config path
	Server   string // server key (for the human summary)
	OldToken string // the package spec as written, e.g. "mcp-server-git@2025.8.0"
	NewToken string // the safe replacement, e.g. "mcp-server-git@2026.1.14"
	Reason   string // why (advisory id / "behind catalog stable" / "unpinned")
}

// PlanPins returns the safe version-bump rewrites for AE-MCP-001 across the
// repo's MCP configs. Only SAFE, drop-in version changes are offered (same
// package, newer version): an advisory-affected pin → its fixed version, and an
// unpinned or behind-stable cataloged package → the catalog stable version.
// Deprecated/archived packages are deliberately NOT rewritten — their successor
// is a different package/transport, so migration is manual (ADR-0021). Pure: no
// writes. Mirrors checkPinning's discovery so fixes only ever target what the
// rule flags.
func PlanPins(root string, inv repository.Inventory, cat *catalog.Catalog) ([]PinRewrite, error) {
	var rewrites []PinRewrite
	for _, rel := range inv.Paths {
		if !isMCPConfigPath(rel) {
			continue
		}
		// #nosec G304 -- rel is a fixed MCP config path from the tracked inventory.
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", rel, err)
		}
		cf, err := parseConfigFile(rel, data)
		if err != nil {
			return nil, err
		}
		for _, s := range cf.Servers {
			token, ok := packageTokenFromArgs(s.Command, s.Args)
			if !ok {
				continue
			}
			name, version, pinned := classifyPackageSpec(token)
			entry, known := cat.Lookup(name)
			if !known || entry.Status == "deprecated" {
				continue // unknown package or no safe drop-in
			}
			target, reason := safePinTarget(entry, version, pinned)
			if target == "" || target == version {
				continue
			}
			rewrites = append(rewrites, PinRewrite{
				Path:     rel,
				Server:   s.Name,
				OldToken: token,
				NewToken: name + "@" + target,
				Reason:   reason,
			})
		}
	}
	sort.Slice(rewrites, func(i, j int) bool {
		if rewrites[i].Path != rewrites[j].Path {
			return rewrites[i].Path < rewrites[j].Path
		}
		return rewrites[i].OldToken < rewrites[j].OldToken
	})
	return rewrites, nil
}

// safePinTarget returns the safe version to pin to and why, or "" when there is
// no safe target. Advisory (pinned-and-vulnerable) → the fixed version;
// unpinned or behind-stable → the catalog stable version (when one is recorded).
func safePinTarget(e catalog.ServerEntry, version string, pinned bool) (target, reason string) {
	if pinned {
		if adv, hit := e.AdvisoryFor(version); hit {
			return adv.FixedIn, "affected by " + adv.ID
		}
		if stable, behind := e.KnownBehind(version); behind {
			return stable, "behind catalog stable version"
		}
		return "", ""
	}
	if e.StableVersion != "" {
		return e.StableVersion, "unpinned — pin to catalog stable version"
	}
	return "", ""
}
