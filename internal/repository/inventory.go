package repository

import (
	"os"
	"path/filepath"
	"strings"
)

type Inventory struct {
	Paths           []string
	set             map[string]struct{}
	ignoredPrefixes []string
}

func BuildInventory(root string) (Inventory, error) {
	inv := Inventory{set: map[string]struct{}{}}

	gitignore := filepath.Join(root, ".gitignore")
	// #nosec G304 -- root is the resolved repository root for the active scan target.
	if data, err := os.ReadFile(gitignore); err == nil {
		for _, raw := range strings.Split(string(data), "\n") {
			line := strings.TrimSpace(raw)
			if strings.HasSuffix(line, "/") && !strings.HasPrefix(line, "#") {
				inv.ignoredPrefixes = append(inv.ignoredPrefixes, strings.TrimSuffix(line, "/"))
			}
		}
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		rel = filepath.ToSlash(rel)
		if rel == ".git" || strings.HasPrefix(rel, ".git/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		for _, ignored := range inv.ignoredPrefixes {
			if rel == ignored || strings.HasPrefix(rel, ignored+"/") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if d.IsDir() {
			return nil
		}

		inv.Paths = append(inv.Paths, rel)
		inv.set[rel] = struct{}{}
		return nil
	})

	return inv, err
}

func (i Inventory) Has(path string) bool {
	_, ok := i.set[path]
	return ok
}
