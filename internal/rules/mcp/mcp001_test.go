package mcp

import (
	"strings"
	"testing"
)

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
	fs := checkPinning([]ConfigFile{cf})
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
	if got := checkPinning([]ConfigFile{cf}); len(got) != 0 {
		t.Fatalf("expected no findings, got %d: %+v", len(got), got)
	}
}
