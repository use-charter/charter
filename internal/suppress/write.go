package suppress

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// UpsertFile reads the existing .charter-suppress.yml under root (if any),
// replaces an entry with the same rule+path or appends a new one, and returns the
// marshaled file bytes. It does not write to disk — the caller decides (honoring
// --dry-run). A missing file is treated as empty; malformed YAML fails fast.
func UpsertFile(root string, entry FileEntry) ([]byte, error) {
	var doc suppressDoc
	// #nosec G304 -- File is the fixed .charter-suppress.yml path under the repo root.
	data, err := os.ReadFile(filepath.Join(root, File))
	switch {
	case err == nil:
		if uerr := yaml.Unmarshal(data, &doc); uerr != nil {
			return nil, fmt.Errorf("parse %s: %w", File, uerr)
		}
	case !errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("read %s: %w", File, err)
	}

	replaced := false
	for i := range doc.Suppressions {
		if doc.Suppressions[i].Rule == entry.Rule && doc.Suppressions[i].Path == entry.Path {
			doc.Suppressions[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		doc.Suppressions = append(doc.Suppressions, entry)
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", File, err)
	}
	return out, nil
}
