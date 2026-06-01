package mcp

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.use-charter.dev/charter/internal/repository"
)

func newMCPRepo(t *testing.T, files map[string]string) (string, repository.Inventory) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	return root, inv
}

func runIDs(t *testing.T, files map[string]string) []string {
	t.Helper()
	root, inv := newMCPRepo(t, files)
	fs, err := Run(root, inv)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	var ids []string
	for _, f := range fs {
		ids = append(ids, f.RuleID)
	}
	return ids
}

const (
	cleanMCP   = `{ "mcpServers": { "fs": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem@1.0.4"] }, "asana": { "type": "http", "url": "https://mcp.asana.com/mcp", "headers": { "Authorization": "Bearer ${ASANA_TOKEN}" } } } }`
	allowAsana = "mcp:\n  trustedRemotes:\n    - mcp.asana.com\n"
)

func TestRunCleanNoFindings(t *testing.T) {
	ids := runIDs(t, map[string]string{".mcp.json": cleanMCP, "charter.yaml": allowAsana})
	if len(ids) != 0 {
		t.Fatalf("expected no findings, got %v", ids)
	}
}

func TestRunUnpinned(t *testing.T) {
	ids := runIDs(t, map[string]string{".mcp.json": `{ "mcpServers": { "gum": { "command": "npx", "args": ["-y", "gumroad-mcp@latest"] } } }`})
	if len(ids) != 1 || ids[0] != "AE-MCP-001" {
		t.Fatalf("expected [AE-MCP-001], got %v", ids)
	}
}

func TestRunUntrustedRemote(t *testing.T) {
	ids := runIDs(t, map[string]string{
		".mcp.json":    `{ "mcpServers": { "shadow": { "type": "http", "url": "https://unknown.example.net/mcp", "headers": { "Authorization": "Bearer ${T}" } } } }`,
		"charter.yaml": allowAsana,
	})
	if len(ids) != 1 || ids[0] != "AE-MCP-002" {
		t.Fatalf("expected [AE-MCP-002], got %v", ids)
	}
}

func TestRunNoAuth(t *testing.T) {
	// A non-catalog host that the repo allowlists (so AE-MCP-002 passes) but that
	// declares no auth header still fires AE-MCP-003. (Catalog OAuth vendor hosts
	// are exempt from AE-MCP-003 — see TestRunCatalogTrustedHostPasses / CF-13.)
	ids := runIDs(t, map[string]string{
		".mcp.json":    `{ "mcpServers": { "open": { "type": "http", "url": "https://mcp.privatevendor.example/mcp" } } }`,
		"charter.yaml": "mcp:\n  trustedRemotes:\n    - mcp.privatevendor.example\n",
	})
	if len(ids) != 1 || ids[0] != "AE-MCP-003" {
		t.Fatalf("expected [AE-MCP-003], got %v", ids)
	}
}

func TestRunCatalogTrustedHostPasses(t *testing.T) {
	// A catalog vendor host (api.githubcopilot.com) with auth and NO charter.yaml
	// allowlist must pass AE-MCP-002 on the catalog baseline alone.
	ids := runIDs(t, map[string]string{
		".mcp.json": `{ "mcpServers": { "gh": { "type": "http", "url": "https://api.githubcopilot.com/mcp/", "headers": { "Authorization": "Bearer ${GH}" } } } }`,
	})
	if len(ids) != 0 {
		t.Fatalf("expected no findings for a catalog-trusted host, got %v", ids)
	}
}

func TestRunDeprecatedPackageFlags(t *testing.T) {
	ids := runIDs(t, map[string]string{
		".mcp.json": `{ "mcpServers": { "github": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-github@1.2.3"] } } }`,
	})
	if len(ids) != 1 || ids[0] != "AE-MCP-001" {
		t.Fatalf("expected [AE-MCP-001] for a deprecated package, got %v", ids)
	}
}

func TestRunNoConfigNoFindings(t *testing.T) {
	if ids := runIDs(t, map[string]string{"README.md": "# x\n"}); len(ids) != 0 {
		t.Fatalf("expected no findings without MCP config, got %v", ids)
	}
}

func TestRunMalformedConfigErrors(t *testing.T) {
	root, inv := newMCPRepo(t, map[string]string{".mcp.json": "{ not json"})
	if _, err := Run(root, inv); err == nil {
		t.Fatal("expected error for malformed .mcp.json")
	}
}

func TestRunUntrustedAndNoAuth(t *testing.T) {
	// Untrusted host + no auth header + no allowlist: the rules are independent,
	// so both AE-MCP-002 and AE-MCP-003 fire (in pipeline order).
	ids := runIDs(t, map[string]string{
		".mcp.json": `{ "mcpServers": { "x": { "type": "http", "url": "https://unknown.example.net/mcp" } } }`,
	})
	if len(ids) != 2 || ids[0] != "AE-MCP-002" || ids[1] != "AE-MCP-003" {
		t.Fatalf("expected [AE-MCP-002 AE-MCP-003], got %v", ids)
	}
}
