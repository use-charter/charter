package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.use-charter.dev/charter/internal/repository"
	"gopkg.in/yaml.v3"
)

// File is the repo-relative charter config path.
const File = "charter.yaml"

type charterConfig struct {
	MCP struct {
		TrustedRemotes []string `yaml:"trustedRemotes"`
	} `yaml:"mcp"`
}

// LoadTrustedRemotes returns the host allowlist from charter.yaml's
// mcp.trustedRemotes. A missing file yields a nil slice and no error.
func LoadTrustedRemotes(root string, inv repository.Inventory) ([]string, error) {
	if !inv.Has(File) {
		return nil, nil
	}
	// #nosec G304 -- File is the fixed charter.yaml path under the scan root.
	data, err := os.ReadFile(filepath.Join(root, File))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", File, err)
	}
	var cfg charterConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", File, err)
	}
	return cfg.MCP.TrustedRemotes, nil
}
