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
		got, err := checkEditScope(dir, repository.New(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("context with off-limits declaration returns nil", func(t *testing.T) {
		dir := t.TempDir()
		content := "## Edit Scope\n- Off-limits: .env*, secrets/, .github/workflows/\n"
		if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := checkEditScope(dir, repository.New([]string{"AGENTS.md"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil (declared), got %v", got)
		}
	})

	t.Run("context without off-limits returns one AE-CC-002 finding", func(t *testing.T) {
		dir := t.TempDir()
		content := "# Charter\nThis is a Go CLI tool. Tech stack: Go.\n"
		if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := checkEditScope(dir, repository.New([]string{"AGENTS.md"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].RuleID != "AE-CC-002" || got[0].Severity != findings.SeverityHigh {
			t.Fatalf("expected one AE-CC-002 HIGH finding, got %+v", got)
		}
		if len(got[0].Locations) != 1 || got[0].Locations[0].Path != "AGENTS.md" {
			t.Fatalf("wrong location: %+v", got[0].Locations)
		}
	})

	t.Run("unreadable context file fails fast", func(t *testing.T) {
		dir := t.TempDir()
		// Make AGENTS.md a directory so os.ReadFile returns an error.
		if err := os.Mkdir(filepath.Join(dir, "AGENTS.md"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := checkEditScope(dir, repository.New([]string{"AGENTS.md"})); err == nil {
			t.Fatal("expected a read error for an unreadable context file")
		}
	})

	t.Run("unreadable .cursor/rules file fails fast", func(t *testing.T) {
		dir := t.TempDir()
		// Make the .mdc path a directory so os.ReadFile returns an error.
		if err := os.MkdirAll(filepath.Join(dir, ".cursor", "rules", "scope.mdc"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := checkEditScope(dir, repository.New([]string{".cursor/rules/scope.mdc"})); err == nil {
			t.Fatal("expected a read error for an unreadable .cursor/rules file")
		}
	})

	t.Run(".cursor/rules declaring off-limits returns nil", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".cursor", "rules", "scope.mdc"), []byte("Off-limits: secrets/ and .github/workflows/\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := checkEditScope(dir, repository.New([]string{".cursor/rules/scope.mdc"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil (declared via .cursor/rules), got %v", got)
		}
	})

	t.Run(".cursor/rules without off-limits returns one AE-CC-002 finding", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".cursor", "rules", "style.mdc"), []byte("Use tabs. Be concise.\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := checkEditScope(dir, repository.New([]string{".cursor/rules/style.mdc"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].RuleID != "AE-CC-002" {
			t.Fatalf("expected one AE-CC-002 finding, got %+v", got)
		}
	})
}
