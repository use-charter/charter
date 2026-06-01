package catalog

import "testing"

// TestCatalogValid is the curation contract: the embedded seed must be
// well-formed, or the build fails. This is what makes Default() safe to panic on
// a malformed embed (it can never ship).
func TestCatalogValid(t *testing.T) {
	c := Default()

	if c.Version == "" || c.Generated == "" {
		t.Fatalf("catalog must set version and generated; got %q / %q", c.Version, c.Generated)
	}
	if len(c.Servers) == 0 {
		t.Fatal("catalog has no servers")
	}
	if len(c.TrustedHosts) == 0 {
		t.Fatal("catalog has no trusted hosts")
	}

	seenHost := map[string]bool{}
	for _, h := range c.TrustedHosts {
		if h != lower(h) {
			t.Errorf("trusted host %q is not lowercase/trimmed", h)
		}
		if seenHost[h] {
			t.Errorf("duplicate trusted host %q", h)
		}
		seenHost[h] = true
	}

	seenPkg := map[string]bool{}
	for _, e := range c.Servers {
		if e.Package == "" {
			t.Fatal("server entry with empty package")
		}
		if seenPkg[e.Package] {
			t.Errorf("duplicate package %q", e.Package)
		}
		seenPkg[e.Package] = true

		switch e.Ecosystem {
		case "npm", "pypi":
		default:
			t.Errorf("%s: invalid ecosystem %q", e.Package, e.Ecosystem)
		}

		switch e.Status {
		case "deprecated":
			if e.Successor == "" {
				t.Errorf("%s: deprecated entry must name a successor", e.Package)
			}
			if e.StableVersion != "" || len(e.KnownVersions) != 0 {
				t.Errorf("%s: deprecated entry must not carry version data", e.Package)
			}
		case "active":
			if len(e.KnownVersions) > 0 {
				assertAscendingUnique(t, e.Package, e.KnownVersions)
				if last := e.KnownVersions[len(e.KnownVersions)-1]; e.StableVersion != last {
					t.Errorf("%s: stableVersion %q must equal last knownVersion %q", e.Package, e.StableVersion, last)
				}
			} else if e.StableVersion != "" {
				t.Errorf("%s: stableVersion set without knownVersions", e.Package)
			}
		default:
			t.Errorf("%s: invalid status %q", e.Package, e.Status)
		}

		for _, a := range e.Advisories {
			if a.ID == "" || a.FixedIn == "" || (len(a.Affected) == 0 && a.AffectedBelow == "") {
				t.Errorf("%s: advisory must set id, fixedIn, and affected or affectedBelow; got %+v", e.Package, a)
			}
			if a.Severity != "high" {
				t.Errorf("%s: advisory %s severity must be \"high\", got %q", e.Package, a.ID, a.Severity)
			}
		}
	}
}

// TestCatalogCoversMajorServers guards against accidental shrinkage of the
// curated coverage: a representative set of major vendor hosts and known
// packages must stay present (the catalog is meant to be broad — Slice 13).
func TestCatalogCoversMajorServers(t *testing.T) {
	c := Default()

	hosts := c.TrustedHostSet()
	for _, h := range []string{
		"api.githubcopilot.com", "mcp.supabase.com", "mcp.neon.tech", "mcp.figma.com",
		"mcp.stripe.com", "mcp.linear.app", "mcp.notion.com", "mcp.sentry.dev",
		"mcp.context7.com", "hf.co", "mcp.cloudflare.com", "mcp.vercel.com",
	} {
		if _, ok := hosts[h]; !ok {
			t.Errorf("expected trusted host %q in catalog", h)
		}
	}
	if len(hosts) < 50 {
		t.Errorf("expected a broad trusted-host baseline (>=50), got %d", len(hosts))
	}

	for _, pkg := range []string{
		"shadcn",
		"@modelcontextprotocol/server-filesystem",
		"@modelcontextprotocol/server-github", // deprecated
	} {
		if _, ok := c.Lookup(pkg); !ok {
			t.Errorf("expected package %q in catalog", pkg)
		}
	}
}

func assertAscendingUnique(t *testing.T, pkg string, vs []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, v := range vs {
		if seen[v] {
			t.Errorf("%s: duplicate knownVersion %q", pkg, v)
		}
		seen[v] = true
	}
	// Lexicographic ascending is sufficient as a determinism guard; exact
	// ordering semantics are not relied on (comparison is exact-match only).
	for i := 1; i < len(vs); i++ {
		if vs[i-1] > vs[i] {
			t.Errorf("%s: knownVersions not ascending near %q,%q", pkg, vs[i-1], vs[i])
		}
	}
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

const inlineCatalog = `
version: "test"
generated: "2026-06-02"
trustedHosts:
  - Api.GitHubCopilot.com
  - mcp.example.com
servers:
  - package: "@scope/active"
    ecosystem: npm
    status: active
    stableVersion: "1.0.2"
    knownVersions: ["1.0.0", "1.0.1", "1.0.2"]
  - package: "@scope/advised"
    ecosystem: npm
    status: active
    stableVersion: "2.0.0"
    knownVersions: ["1.9.0", "2.0.0"]
    advisories:
      - id: "GHSA-test-0001"
        affected: ["1.9.0"]
        fixedIn: "2.0.0"
        severity: high
        summary: "test advisory"
  - package: "@scope/gone"
    ecosystem: npm
    status: deprecated
    successor: "@scope/new"
`

func parseInline(t *testing.T) *Catalog {
	t.Helper()
	c, err := Parse([]byte(inlineCatalog))
	if err != nil {
		t.Fatalf("parse inline: %v", err)
	}
	return c
}

func TestLookup(t *testing.T) {
	c := parseInline(t)
	if e, ok := c.Lookup("@scope/active"); !ok || e.Status != "active" {
		t.Fatalf("Lookup active = %+v, %v", e, ok)
	}
	if _, ok := c.Lookup("@scope/missing"); ok {
		t.Fatal("Lookup of unknown package should miss")
	}
}

func TestAdvisoryFor(t *testing.T) {
	c := parseInline(t)
	e, _ := c.Lookup("@scope/advised")
	if a, ok := e.AdvisoryFor("1.9.0"); !ok || a.ID != "GHSA-test-0001" || a.FixedIn != "2.0.0" {
		t.Fatalf("AdvisoryFor(1.9.0) = %+v, %v", a, ok)
	}
	if _, ok := e.AdvisoryFor("2.0.0"); ok {
		t.Fatal("fixed version must not match an advisory")
	}
	if _, ok := e.AdvisoryFor("9.9.9"); ok {
		t.Fatal("unknown version must not match an advisory")
	}
}

func TestKnownBehind(t *testing.T) {
	c := parseInline(t)
	active, _ := c.Lookup("@scope/active")
	if stable, behind := active.KnownBehind("1.0.0"); !behind || stable != "1.0.2" {
		t.Fatalf("KnownBehind(1.0.0) = %q,%v want 1.0.2,true", stable, behind)
	}
	if _, behind := active.KnownBehind("1.0.2"); behind {
		t.Fatal("stable version is not behind")
	}
	if _, behind := active.KnownBehind("9.9.9"); behind {
		t.Fatal("version absent from knownVersions must be silent (staleness-safe)")
	}

	// An advisory-covered version is reported by AdvisoryFor, never as behind.
	advised, _ := c.Lookup("@scope/advised")
	if _, behind := advised.KnownBehind("1.9.0"); behind {
		t.Fatal("advisory-covered version must not also be reported as behind")
	}
}

func TestVersionLess(t *testing.T) {
	cases := []struct {
		a, b       string
		less, okay bool
	}{
		{"2025.9.25", "2026.1.14", true, true},      // CalVer year
		{"2025.12.18", "2026.1.14", true, true},     // lexical would be wrong (12>1); numeric is right
		{"2026.1.14", "2026.1.14", false, true},     // equal (the fixed version)
		{"2026.2.1", "2026.1.14", false, true},      // later month
		{"1.0.0", "1.0.1", true, true},              // semver
		{"v1.2.3", "1.2.4", true, true},             // leading v ignored
		{"latest", "2026.1.14", false, false},       // non-numeric -> not orderable
		{"2025.8.0-rc1", "2026.1.14", false, false}, // prerelease tag -> not orderable
	}
	for _, c := range cases {
		less, ok := versionLess(c.a, c.b)
		if less != c.less || ok != c.okay {
			t.Errorf("versionLess(%q,%q) = (%v,%v), want (%v,%v)", c.a, c.b, less, ok, c.less, c.okay)
		}
	}
}

func TestAdvisoryForRange(t *testing.T) {
	c := Default()
	git, ok := c.Lookup("mcp-server-git")
	if !ok {
		t.Fatal("mcp-server-git missing from catalog")
	}
	// An old pinned version is caught by the affectedBelow range; the fixed
	// version and an unparseable version are not.
	if a, hit := git.AdvisoryFor("2025.8.0"); !hit || a.FixedIn != "2026.1.14" {
		t.Fatalf("expected 2025.8.0 affected, fix 2026.1.14; got %+v hit=%v", a, hit)
	}
	if _, hit := git.AdvisoryFor("2026.1.14"); hit {
		t.Fatal("the fixed version 2026.1.14 must not be flagged")
	}
	if _, hit := git.AdvisoryFor("latest"); hit {
		t.Fatal("an unparseable version must not be flagged by a range advisory")
	}
}

func TestTrustedHostSet(t *testing.T) {
	c := parseInline(t)
	set := c.TrustedHostSet()
	if _, ok := set["api.githubcopilot.com"]; !ok {
		t.Fatalf("TrustedHostSet should lowercase hosts; got %v", set)
	}
	if _, ok := set["mcp.example.com"]; !ok {
		t.Fatal("missing mcp.example.com")
	}
}
