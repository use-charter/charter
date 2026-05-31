package suppress

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.use-charter.dev/charter/internal/repository"
	"gopkg.in/yaml.v3"
)

// LoadFile reads .charter-suppress.yml from the repo root when it is tracked.
// A missing file yields a nil slice and no error; malformed YAML fails fast.
func LoadFile(root string, inv repository.Inventory) ([]Entry, error) {
	if !inv.Has(File) {
		return nil, nil
	}
	// #nosec G304 -- File is the fixed .charter-suppress.yml path under the scan root.
	data, err := os.ReadFile(filepath.Join(root, File))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", File, err)
	}
	entries, err := parseSuppressFile(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", File, err)
	}
	return entries, nil
}

func parseSuppressFile(data []byte) ([]Entry, error) {
	var doc suppressDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(doc.Suppressions))
	for i, e := range doc.Suppressions {
		rule := strings.TrimSpace(e.Rule)
		if rule == "" {
			return nil, fmt.Errorf("suppression entry %d: missing required %q field", i+1, "rule")
		}
		entries = append(entries, Entry{
			Rule:     rule,
			Reason:   strings.TrimSpace(e.Reason),
			Approver: strings.TrimSpace(e.Approver),
			Expires:  strings.TrimSpace(e.Expires),
			Path:     filepath.ToSlash(strings.TrimSpace(e.Path)),
			Source:   SourceExternal,
		})
	}
	return entries, nil
}
