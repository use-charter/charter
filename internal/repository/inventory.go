package repository

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type Inventory struct {
	Paths []string
	set   map[string]struct{}
}

func BuildInventory(root string) (Inventory, error) {
	// #nosec G204 -- root is the resolved repository root for the active scan target.
	cmd := exec.Command("git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--full-name")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return Inventory{}, fmt.Errorf("list repository files: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return parseGitLSFilesOutput(output), nil
}

func (i Inventory) Has(path string) bool {
	_, ok := i.set[filepath.ToSlash(path)]
	return ok
}

func parseGitLSFilesOutput(output []byte) Inventory {
	inv := Inventory{set: map[string]struct{}{}}

	for _, raw := range strings.Split(string(output), "\x00") {
		if raw == "" {
			continue
		}

		path := filepath.ToSlash(raw)
		if path == ".git" || strings.HasPrefix(path, ".git/") {
			continue
		}

		inv.Paths = append(inv.Paths, path)
		inv.set[path] = struct{}{}
	}

	sort.Strings(inv.Paths)
	inv.Paths = slices.Clip(inv.Paths)
	return inv
}
