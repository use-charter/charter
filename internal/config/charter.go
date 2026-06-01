package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.use-charter.dev/charter/internal/repository"
	"gopkg.in/yaml.v3"
)

// File is the repo-relative charter config path.
const File = "charter.yaml"

// DefaultThreshold is the passing gate when no flag, policy.threshold, or
// policy.profile is set.
const DefaultThreshold = 80

// profileThresholds maps the built-in policy profiles to their passing
// thresholds, aligned to the score zones (80 = ship-ready boundary).
var profileThresholds = map[string]int{"strict": 90, "standard": 80, "relaxed": 60}

type charterConfig struct {
	MCP struct {
		TrustedRemotes []string `yaml:"trustedRemotes"`
	} `yaml:"mcp"`
	Policy struct {
		Profile   string `yaml:"profile"`
		Threshold *int   `yaml:"threshold"`
	} `yaml:"policy"`
}

// Policy is the resolved policy block from charter.yaml.
type Policy struct {
	Profile      string
	Threshold    int
	HasThreshold bool
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

// LoadPolicy returns the policy block from charter.yaml when tracked. A missing
// file yields a zero Policy and no error; malformed YAML fails fast.
func LoadPolicy(root string, inv repository.Inventory) (Policy, error) {
	if !inv.Has(File) {
		return Policy{}, nil
	}
	// #nosec G304 -- File is the fixed charter.yaml path under the scan root.
	data, err := os.ReadFile(filepath.Join(root, File))
	if err != nil {
		return Policy{}, fmt.Errorf("read %s: %w", File, err)
	}
	var cfg charterConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Policy{}, fmt.Errorf("parse %s: %w", File, err)
	}
	p := Policy{Profile: strings.TrimSpace(cfg.Policy.Profile)}
	if cfg.Policy.Threshold != nil {
		p.Threshold = *cfg.Policy.Threshold
		p.HasThreshold = true
	}
	return p, nil
}

// ResolveThreshold applies the precedence: explicit flag > policy.threshold >
// policy.profile > DefaultThreshold. Unknown profiles and out-of-range
// thresholds fail fast.
func ResolveThreshold(p Policy, flagThreshold int, flagSet bool) (int, error) {
	if flagSet {
		return inRange("threshold", flagThreshold)
	}
	if p.HasThreshold {
		return inRange("policy.threshold", p.Threshold)
	}
	if p.Profile != "" {
		t, ok := profileThresholds[p.Profile]
		if !ok {
			return 0, fmt.Errorf("unknown policy.profile %q (want strict, standard, or relaxed)", p.Profile)
		}
		return t, nil
	}
	return DefaultThreshold, nil
}

func inRange(label string, v int) (int, error) {
	if v < 0 || v > 100 {
		return 0, fmt.Errorf("%s %d out of range 0..100", label, v)
	}
	return v, nil
}
