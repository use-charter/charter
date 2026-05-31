package repository

import (
	"bytes"
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
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return Inventory{}, fmt.Errorf("list repository files: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return parseGitLSFilesOutput(output), nil
}

func (i Inventory) Has(path string) bool {
	_, ok := i.set[filepath.ToSlash(path)]
	return ok
}

// NewInventoryForTest builds an Inventory from a fixed path list. Test-only seam.
func NewInventoryForTest(paths []string) Inventory {
	inv := Inventory{set: map[string]struct{}{}}
	for _, p := range paths {
		s := filepath.ToSlash(p)
		inv.Paths = append(inv.Paths, s)
		inv.set[s] = struct{}{}
	}
	sort.Strings(inv.Paths)
	return inv
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
