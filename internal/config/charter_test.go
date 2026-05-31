package config

import (
	"os"
	"path/filepath"
	"testing"

	"go.charter.dev/charter/internal/repository"
)

func writeCharterRepo(t *testing.T, contents string) (string, repository.Inventory) {
	t.Helper()
	root := t.TempDir()
	if contents != "" {
		if err := os.WriteFile(filepath.Join(root, "charter.yaml"), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	inv := repository.NewInventoryForTest(nil)
	if contents != "" {
		inv = repository.NewInventoryForTest([]string{"charter.yaml"})
	}
	return root, inv
}

// writeCharterFile always writes charter.yaml (even when contents is empty) and
// marks it present in the inventory, so tests can exercise empty-but-present and
// malformed files that writeCharterRepo's "" sentinel cannot express.
func writeCharterFile(t *testing.T, contents string) (string, repository.Inventory) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "charter.yaml"), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, repository.NewInventoryForTest([]string{"charter.yaml"})
}

func TestLoadTrustedRemotes(t *testing.T) {
	root, inv := writeCharterRepo(t, "mcp:\n  trustedRemotes:\n    - api.example.com\n    - mcp.asana.com\n")
	hosts, err := LoadTrustedRemotes(root, inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 2 || hosts[0] != "api.example.com" || hosts[1] != "mcp.asana.com" {
		t.Fatalf("got %v, want [api.example.com mcp.asana.com]", hosts)
	}
}

func TestLoadTrustedRemotesMissingFile(t *testing.T) {
	root, inv := writeCharterRepo(t, "")
	hosts, err := LoadTrustedRemotes(root, inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 0 {
		t.Fatalf("expected empty allowlist, got %v", hosts)
	}
}

func TestLoadTrustedRemotesPresentFile(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		wantErr  bool
	}{
		{name: "malformed yaml", contents: "mcp: [unterminated", wantErr: true},
		{name: "empty mcp block", contents: "mcp:\n"},
		{name: "empty but present file", contents: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, inv := writeCharterFile(t, tc.contents)
			hosts, err := LoadTrustedRemotes(root, inv)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(hosts) != 0 {
				t.Fatalf("expected empty allowlist, got %v", hosts)
			}
		})
	}
}
