// Package catalog holds Charter's static, founder-curated MCP catalog: a
// versioned list of known MCP server packages, their advisories, and a baseline
// of trusted vendor-operated remote hosts. It powers the catalog-aware facets of
// AE-MCP-001 (deprecated/advisory/behind-stable) and AE-MCP-002 (trusted hosts).
//
// Comparison is EXACT-MATCH only (ADR-0021): version ordering is never inferred
// across schemes (the official servers use CalVer, others semver). A pinned
// version that is absent from an entry's KnownVersions is intentionally SILENT,
// so a stale catalog under-reports rather than misreporting — protecting
// Charter's low-false-positive promise (Commitment #9).
//
// The catalog is embedded at build time and parsed once. A malformed embed is a
// programming error caught by the package's validity test (catalog_test.go), so
// Default never surfaces a parse error at `charter doctor` runtime — consistent
// with the offline-first, fail-fast, no-network, no-LLM core.
package catalog

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed data/catalog.yaml
var embedded []byte

// Advisory is a known security advisory (CVE/GHSA) against specific versions of
// a cataloged MCP server package. A pinned version matches when it is listed in
// Affected (exact) or, for the common "fixed in X, affected below X" CVE shape,
// when AffectedBelow is set and the version orders numerically below it. The
// numeric comparison is conservative (see versionLess) and applies ONLY to
// authoritative advisories — the behind-stable nudge stays exact-match only.
type Advisory struct {
	ID            string   `yaml:"id"`
	Affected      []string `yaml:"affected"`
	AffectedBelow string   `yaml:"affectedBelow"`
	FixedIn       string   `yaml:"fixedIn"`
	Severity      string   `yaml:"severity"`
	Summary       string   `yaml:"summary"`
	Reference     string   `yaml:"reference"`
}

// ServerEntry is one curated MCP server package.
type ServerEntry struct {
	// Package is the registry spec name, e.g. "@modelcontextprotocol/server-filesystem".
	Package   string `yaml:"package"`
	Ecosystem string `yaml:"ecosystem"` // "npm" | "pypi"
	Status    string `yaml:"status"`    // "active" | "deprecated"
	// StableVersion is the current recommended pin (exact). Empty for deprecated
	// entries and for active entries with no version tracking yet.
	StableVersion string `yaml:"stableVersion"`
	// KnownVersions is an ascending list of recognized versions; the last element
	// equals StableVersion. Used only for the exact-match behind-stable nudge.
	KnownVersions []string `yaml:"knownVersions"`
	// Successor names the migration target for a deprecated package (required when
	// Status == "deprecated").
	Successor  string     `yaml:"successor"`
	Reference  string     `yaml:"reference"`
	Advisories []Advisory `yaml:"advisories"`
}

// Catalog is the parsed MCP catalog.
type Catalog struct {
	Version      string        `yaml:"version"`
	Generated    string        `yaml:"generated"`
	TrustedHosts []string      `yaml:"trustedHosts"`
	Servers      []ServerEntry `yaml:"servers"`
}

var (
	defaultOnce sync.Once
	defaultCat  *Catalog
)

// Default returns the embedded, parsed catalog (parsed once). It panics only if
// the embedded data is malformed, which the validity test prevents from shipping.
func Default() *Catalog {
	defaultOnce.Do(func() {
		c, err := Parse(embedded)
		if err != nil {
			panic(fmt.Sprintf("catalog: embedded data is invalid: %v", err))
		}
		defaultCat = c
	})
	return defaultCat
}

// Parse decodes a catalog from YAML. Used by Default and by tests that build
// inline catalogs independent of the curated seed.
func Parse(data []byte) (*Catalog, error) {
	var c Catalog
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	return &c, nil
}

// Lookup returns the entry for an exact package name.
func (c *Catalog) Lookup(pkg string) (ServerEntry, bool) {
	for _, e := range c.Servers {
		if e.Package == pkg {
			return e, true
		}
	}
	return ServerEntry{}, false
}

// AdvisoryFor returns the first advisory affecting version — by exact membership
// in Affected, or by AffectedBelow (version orders numerically below the fix).
func (e ServerEntry) AdvisoryFor(version string) (Advisory, bool) {
	for _, a := range e.Advisories {
		for _, v := range a.Affected {
			if v == version {
				return a, true
			}
		}
		if a.AffectedBelow != "" {
			if less, ok := versionLess(version, a.AffectedBelow); ok && less {
				return a, true
			}
		}
	}
	return Advisory{}, false
}

// versionLess reports whether version a orders below b using component-wise
// numeric comparison of dot-separated integers (correct for CalVer like
// 2026.1.14 and semver like 1.2.3). A leading "v" is ignored. ok is false when
// either version contains a non-numeric component (prerelease/build tags, dist
// tags, "latest", git refs) — callers treat ok=false as "no match", so an
// ambiguous version is never flagged by a range advisory.
func versionLess(a, b string) (less, ok bool) {
	pa, ok := numericParts(a)
	if !ok {
		return false, false
	}
	pb, ok := numericParts(b)
	if !ok {
		return false, false
	}
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x != y {
			return x < y, true
		}
	}
	return false, true // equal
}

func numericParts(v string) ([]int, bool) {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if v == "" {
		return nil, false
	}
	fields := strings.Split(v, ".")
	parts := make([]int, 0, len(fields))
	for _, f := range fields {
		n := 0
		for _, r := range f {
			if r < '0' || r > '9' {
				return nil, false
			}
			n = n*10 + int(r-'0')
		}
		parts = append(parts, n)
	}
	return parts, true
}

// KnownBehind reports whether version is a recognized, non-stable version with
// no covering advisory — i.e. a safe "newer stable available" nudge. It returns
// the stable version to upgrade to. A version absent from KnownVersions yields
// (",", false): the catalog makes no claim, so nothing fires.
func (e ServerEntry) KnownBehind(version string) (string, bool) {
	if version == "" || version == e.StableVersion {
		return "", false
	}
	known := false
	for _, v := range e.KnownVersions {
		if v == version {
			known = true
			break
		}
	}
	if !known {
		return "", false
	}
	if _, hit := e.AdvisoryFor(version); hit {
		return "", false // an advisory is a stronger, separate signal
	}
	return e.StableVersion, true
}

// TrustedHostSet returns the lowercased trusted-host set for AE-MCP-002.
func (c *Catalog) TrustedHostSet() map[string]struct{} {
	set := make(map[string]struct{}, len(c.TrustedHosts))
	for _, h := range c.TrustedHosts {
		set[strings.ToLower(strings.TrimSpace(h))] = struct{}{}
	}
	return set
}
