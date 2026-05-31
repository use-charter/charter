package agentconfig

import (
	"os"
	"path/filepath"
	"testing"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
)

func TestDeclaresOffLimits(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"env and secrets", "## Edit Scope\n- Off-limits: .env*, secrets/, signing keys", true},
		{"permissions ref", "See PERMISSIONS.md for edit boundaries.", true},
		{"workflows", "Do not edit .github/workflows/ or terraform/", true},
		{"no declaration", "# Project\nCharter is a Go CLI. Tech stack: Go.", false},
		{"generic only", "Be careful editing files.", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := declaresOffLimits(c.text); got != c.want {
				t.Errorf("declaresOffLimits(%q) = %v, want %v", c.text, got, c.want)
			}
		})
	}
}

func TestCheckEditScope(t *testing.T) {
	t.Run("no context file present returns nil", func(t *testing.T) {
		dir := t.TempDir()
		if got := checkEditScope(dir, repository.New(nil)); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("context with off-limits declaration returns nil", func(t *testing.T) {
		dir := t.TempDir()
		content := "## Edit Scope\n- Off-limits: .env*, secrets/, .github/workflows/\n"
		if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if got := checkEditScope(dir, repository.New([]string{"AGENTS.md"})); got != nil {
			t.Fatalf("expected nil (declared), got %v", got)
		}
	})

	t.Run("context without off-limits returns one AE-CC-002 finding", func(t *testing.T) {
		dir := t.TempDir()
		content := "# Charter\nThis is a Go CLI tool. Tech stack: Go.\n"
		if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		got := checkEditScope(dir, repository.New([]string{"AGENTS.md"}))
		if len(got) != 1 || got[0].RuleID != "AE-CC-002" || got[0].Severity != findings.SeverityHigh {
			t.Fatalf("expected one AE-CC-002 HIGH finding, got %+v", got)
		}
		if len(got[0].Locations) != 1 || got[0].Locations[0].Path != "AGENTS.md" {
			t.Fatalf("wrong location: %+v", got[0].Locations)
		}
	})
}
