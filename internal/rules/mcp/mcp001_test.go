package mcp

import (
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/catalog"
	"go.use-charter.dev/charter/internal/scoring"
)

func inlineCat(t *testing.T, yaml string) *catalog.Catalog {
	t.Helper()
	c, err := catalog.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse inline catalog: %v", err)
	}
	return c
}

func TestClassifyPackageSpec(t *testing.T) {
	cases := []struct {
		token      string
		wantPinned bool
	}{
		{"@modelcontextprotocol/server-filesystem@1.0.4", true},
		{"gumroad-mcp@latest", false},
		{"@scope/pkg@^1.2.3", false},
		{"server-perplexity-ask", false}, // no version
		{"pkg@~1.0", false},
		{"pkg@>=2.0.0", false},
		{"pkg@1.2.3", true},
		{"github:owner/repo", false}, // floating git ref
		{"pkg@${VER}", false},        // dynamic version cannot be verified
	}
	for _, c := range cases {
		_, _, pinned := classifyPackageSpec(c.token)
		if pinned != c.wantPinned {
			t.Errorf("classifyPackageSpec(%q) pinned=%v want %v", c.token, pinned, c.wantPinned)
		}
	}
}

func TestCheckPinningFlagsLatest(t *testing.T) {
	cf := ConfigFile{Path: ".mcp.json", Servers: []Server{
		{Name: "g", Command: "npx", Args: []string{"-y", "gumroad-mcp@latest"}, Line: 4},
		{Name: "fs", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem@1.0.4"}, Line: 3},
	}}
	fs := checkPinning([]ConfigFile{cf}, catalog.Default())
	if len(fs) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(fs))
	}
	f := fs[0]
	if f.RuleID != "AE-MCP-001" || f.Severity != "HIGH" {
		t.Fatalf("wrong finding: %+v", f)
	}
	if len(f.Locations) != 1 || f.Locations[0].Path != ".mcp.json" || f.Locations[0].Line != 4 {
		t.Fatalf("wrong location: %+v", f.Locations)
	}
	if !strings.Contains(strings.Join(f.Evidence, " "), "gumroad-mcp@latest") {
		t.Fatalf("evidence missing offending spec: %v", f.Evidence)
	}
}

func TestPackageTokenFromArgs(t *testing.T) {
	cases := []struct {
		name    string
		command string
		args    []string
		want    string
		wantOK  bool
	}{
		{"npx flag then pkg", "npx", []string{"-y", "pkg@1.2.3"}, "pkg@1.2.3", true},
		{"pnpm dlx pkg", "pnpm", []string{"dlx", "pkg@latest"}, "pkg@latest", true},
		{"yarn dlx pkg", "yarn", []string{"dlx", "pkg@1.0.0"}, "pkg@1.0.0", true},
		{"pnpm exec local binary", "pnpm", []string{"exec", "my-local-server"}, "", false},
		{"npx local path", "npx", []string{"-y", "./dist/server.js"}, "", false},
		{"non-runner node", "node", []string{"server.js"}, "", false},
		{"npx empty args", "npx", []string{}, "", false},
		{"absolute path runner", "/usr/local/bin/npx", []string{"pkg@2.0.0"}, "pkg@2.0.0", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := packageTokenFromArgs(c.command, c.args)
			if got != c.want || ok != c.wantOK {
				t.Fatalf("packageTokenFromArgs(%q,%v) = (%q,%v), want (%q,%v)", c.command, c.args, got, ok, c.want, c.wantOK)
			}
		})
	}
}

func TestCheckPinningAllPinned(t *testing.T) {
	cf := ConfigFile{Path: ".mcp.json", Servers: []Server{
		{Name: "fs", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem@1.0.4"}, Line: 1},
	}}
	if got := checkPinning([]ConfigFile{cf}, catalog.Default()); len(got) != 0 {
		t.Fatalf("expected no findings, got %d: %+v", len(got), got)
	}
}

const facetCatalog = `
version: "test"
generated: "2026-06-02"
trustedHosts: ["mcp.example.com"]
servers:
  - package: "@modelcontextprotocol/server-github"
    ecosystem: npm
    status: deprecated
    successor: "github/github-mcp-server"
    reference: "https://github.com/github/github-mcp-server"
  - package: "foo"
    ecosystem: npm
    status: active
    stableVersion: "1.0.1"
    knownVersions: ["1.0.0", "1.0.1"]
    advisories:
      - id: "GHSA-foo-0001"
        affected: ["1.0.0"]
        fixedIn: "1.0.1"
        severity: high
        summary: "RCE in foo"
        reference: "https://example.com/ghsa-foo-0001"
  - package: "bar"
    ecosystem: npm
    status: active
    stableVersion: "1.0.1"
    knownVersions: ["1.0.0", "1.0.1"]
`

func npxServer(name, spec string) ConfigFile {
	return ConfigFile{Path: ".mcp.json", Servers: []Server{
		{Name: name, Command: "npx", Args: []string{"-y", spec}, Line: 2},
	}}
}

func TestCheckPinningDeprecated(t *testing.T) {
	cat := inlineCat(t, facetCatalog)
	// Deprecated package, pinned: HIGH with successor.
	fs := checkPinning([]ConfigFile{npxServer("gh", "@modelcontextprotocol/server-github@1.2.3")}, cat)
	if len(fs) != 1 || fs[0].Severity != "HIGH" {
		t.Fatalf("expected 1 HIGH, got %+v", fs)
	}
	joined := strings.Join(fs[0].Evidence, " ") + " " + fs[0].Summary
	if !strings.Contains(fs[0].Summary, "archived") || !strings.Contains(joined, "github/github-mcp-server") {
		t.Fatalf("expected archived + successor, got %q / %v", fs[0].Summary, fs[0].Evidence)
	}
	// Deprecated outranks unpinned: @latest of a deprecated package still HIGH-archived.
	fs = checkPinning([]ConfigFile{npxServer("gh", "@modelcontextprotocol/server-github@latest")}, cat)
	if len(fs) != 1 || !strings.Contains(fs[0].Summary, "archived") {
		t.Fatalf("expected archived to outrank unpinned, got %+v", fs)
	}
}

func TestCheckPinningAdvisory(t *testing.T) {
	cat := inlineCat(t, facetCatalog)
	fs := checkPinning([]ConfigFile{npxServer("foo", "foo@1.0.0")}, cat)
	if len(fs) != 1 || fs[0].Severity != "HIGH" {
		t.Fatalf("expected 1 HIGH, got %+v", fs)
	}
	if !strings.Contains(fs[0].Summary, "GHSA-foo-0001") || !strings.Contains(fs[0].Summary, "1.0.1") {
		t.Fatalf("expected advisory id + fixedIn, got %q", fs[0].Summary)
	}
	// Fixed version: clean.
	if fs := checkPinning([]ConfigFile{npxServer("foo", "foo@1.0.1")}, cat); len(fs) != 0 {
		t.Fatalf("fixed version should be clean, got %+v", fs)
	}
}

func TestCheckPinningBehindStable(t *testing.T) {
	cat := inlineCat(t, facetCatalog)
	fs := checkPinning([]ConfigFile{npxServer("bar", "bar@1.0.0")}, cat)
	if len(fs) != 1 || fs[0].RuleID != "AE-MCP-001" {
		t.Fatalf("expected 1 AE-MCP-001, got %+v", fs)
	}
	if !fs[0].Informational {
		t.Fatalf("behind-stable must be informational (non-deducting), got %+v", fs[0])
	}
	if !strings.Contains(fs[0].Summary, "1.0.1") {
		t.Fatalf("expected stable version in summary, got %q", fs[0].Summary)
	}
	// Stable pin: clean. Unknown version: clean (staleness-safe).
	if fs := checkPinning([]ConfigFile{npxServer("bar", "bar@1.0.1")}, cat); len(fs) != 0 {
		t.Fatalf("stable pin should be clean, got %+v", fs)
	}
	if fs := checkPinning([]ConfigFile{npxServer("bar", "bar@9.9.9")}, cat); len(fs) != 0 {
		t.Fatalf("unknown version should be silent, got %+v", fs)
	}
}

func TestBehindStableDoesNotDeduct(t *testing.T) {
	cat := inlineCat(t, facetCatalog)
	fs := checkPinning([]ConfigFile{npxServer("bar", "bar@1.0.0")}, cat)
	if got := scoring.Calculate(fs).Final; got != 100 {
		t.Fatalf("behind-stable informational finding must not deduct; score = %d, want 100", got)
	}
}
