package repository

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Inventory struct {
	Paths []string
	set   map[string]struct{}
}

func BuildInventory(root string) (Inventory, error) {
	inv := Inventory{set: map[string]struct{}{}}

	// #nosec G204 -- root is the resolved repository root for the active scan target.
	cmd := exec.Command("git", "-C", root, "ls-files", "--cached", "--others", "--exclude-standard", "--full-name")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return Inventory{}, fmt.Errorf("list repository files: %w: %s", err, strings.TrimSpace(string(output)))
	}

	for _, raw := range strings.Split(string(output), "\n") {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}

		path = filepath.ToSlash(path)
		if path == ".git" || strings.HasPrefix(path, ".git/") {
			continue
		}

		inv.Paths = append(inv.Paths, path)
		inv.set[path] = struct{}{}
	}

	sort.Strings(inv.Paths)
	return inv, nil
}

func (i Inventory) Has(path string) bool {
	_, ok := i.set[filepath.ToSlash(path)]
	return ok
}
