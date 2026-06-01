package config

import (
	"os"
	"path/filepath"
	"testing"

	"go.use-charter.dev/charter/internal/repository"
)

func writeCharterRepo(t *testing.T, contents string) (string, repository.Inventory) {
	t.Helper()
	root := t.TempDir()
	if contents != "" {
		if err := os.WriteFile(filepath.Join(root, "charter.yaml"), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	inv := repository.New(nil)
	if contents != "" {
		inv = repository.New([]string{"charter.yaml"})
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
	return root, repository.New([]string{"charter.yaml"})
}

func TestResolveThresholdPrecedence(t *testing.T) {
	cases := []struct {
		name   string
		policy Policy
		flag   int
		set    bool
		want   int
	}{
		{"flag wins", Policy{Profile: "strict"}, 55, true, 55},
		{"explicit threshold over profile", Policy{Profile: "strict", Threshold: 70, HasThreshold: true}, 80, false, 70},
		{"profile when no flag/threshold", Policy{Profile: "relaxed"}, 80, false, 60},
		{"strict profile", Policy{Profile: "strict"}, 80, false, 90},
		{"default when empty", Policy{}, 80, false, 80},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ResolveThreshold(c.policy, c.flag, c.set)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %d, want %d", got, c.want)
			}
		})
	}
}

func TestResolveThresholdErrors(t *testing.T) {
	if _, err := ResolveThreshold(Policy{Profile: "bogus"}, 80, false); err == nil {
		t.Fatal("expected unknown-profile error")
	}
	if _, err := ResolveThreshold(Policy{Threshold: 150, HasThreshold: true}, 80, false); err == nil {
		t.Fatal("expected out-of-range policy.threshold error")
	}
	if _, err := ResolveThreshold(Policy{}, 150, true); err == nil {
		t.Fatal("expected out-of-range flag error")
	}
}

func TestLoadPolicy(t *testing.T) {
	root, inv := writeCharterFile(t, "policy:\n  profile: strict\n")
	p, err := LoadPolicy(root, inv)
	if err != nil {
		t.Fatal(err)
	}
	if p.Profile != "strict" || p.HasThreshold {
		t.Fatalf("unexpected policy: %+v", p)
	}

	rootT, invT := writeCharterFile(t, "policy:\n  threshold: 70\n")
	pt, err := LoadPolicy(rootT, invT)
	if err != nil {
		t.Fatal(err)
	}
	if !pt.HasThreshold || pt.Threshold != 70 {
		t.Fatalf("expected threshold 70, got %+v", pt)
	}

	missingRoot, missingInv := writeCharterRepo(t, "")
	if p2, err := LoadPolicy(missingRoot, missingInv); err != nil || p2.Profile != "" || p2.HasThreshold {
		t.Fatalf("expected zero policy for missing file, got %+v err %v", p2, err)
	}

	badRoot, badInv := writeCharterFile(t, "policy: [unterminated")
	if _, err := LoadPolicy(badRoot, badInv); err == nil {
		t.Fatal("expected error for malformed YAML")
	}
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
