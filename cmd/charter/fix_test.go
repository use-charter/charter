package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runFix(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(append([]string{"fix"}, args...))
	cmd.SetContext(context.Background())
	err := cmd.Execute()
	return out.String(), err
}

// initFixRepo creates a throwaway git repo, writes the given repo-relative
// files, and stages them so BuildInventory (git ls-files) and doctor see them.
func initFixRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "charter@example.com")
	runGit(t, dir, "config", "user.name", "Charter Test")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	for rel, contents := range files {
		writeTempFile(t, dir, rel, contents)
	}
	runGit(t, dir, "add", "-A")
	return dir
}

// findBackupFile returns the first file named name found anywhere under root, or
// "" if none exists.
func findBackupFile(t *testing.T, root, name string) string {
	t.Helper()
	var found string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Base(path) == name {
			found = path
		}
		return nil
	})
	return found
}

func TestFixDryRunWritesNothing(t *testing.T) {
	// Partial .gitignore: missing .hk/ and .env* so AE-CTX-004 fires.
	repo := initFixRepo(t, map[string]string{
		"go.mod":     "module example.com/x\n\ngo 1.26\n",
		".gitignore": "# project ignores\n.charter/\n*.charter-session\n.claude/local/\n.cursor/cache/\n",
	})

	before, rerr := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if rerr != nil {
		t.Fatalf("read .gitignore: %v", rerr)
	}

	out, err := runFix(t, "--path", repo, "--rule", "AE-CTX-004", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"+.hk/", "+.env*", "(dry run"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected dry-run output to contain %q, got:\n%s", want, out)
		}
	}

	after, rerr := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if rerr != nil {
		t.Fatalf("read .gitignore after: %v", rerr)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("dry run mutated .gitignore on disk:\nbefore=%q\nafter=%q", before, after)
	}
}

func TestFixApplyAppendsAndBacksUp(t *testing.T) {
	repo := initFixRepo(t, map[string]string{
		"go.mod":     "module example.com/x\n\ngo 1.26\n",
		".gitignore": "# project ignores\n.charter/\n*.charter-session\n.claude/local/\n.cursor/cache/\n",
	})
	original, rerr := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if rerr != nil {
		t.Fatalf("read .gitignore: %v", rerr)
	}

	out, err := runFix(t, "--path", repo, "--rule", "AE-CTX-004")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "fixed") {
		t.Fatalf("expected a 'fixed' summary line, got:\n%s", out)
	}
	if !strings.Contains(out, "backups:") {
		t.Fatalf("expected a backups path in output, got:\n%s", out)
	}

	live, rerr := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if rerr != nil {
		t.Fatalf("read .gitignore after: %v", rerr)
	}
	for _, want := range []string{".hk/", ".env*"} {
		if !strings.Contains(string(live), want) {
			t.Fatalf("expected applied .gitignore to gain %q, got:\n%s", want, live)
		}
	}

	backupsRoot := filepath.Join(repo, ".charter", "backups")
	if _, statErr := os.Stat(backupsRoot); statErr != nil {
		t.Fatalf("expected %s to exist: %v", backupsRoot, statErr)
	}
	backup := findBackupFile(t, backupsRoot, ".gitignore")
	if backup == "" {
		t.Fatalf("no .gitignore backup found under %s", backupsRoot)
	}
	got, rerr := os.ReadFile(backup)
	if rerr != nil {
		t.Fatalf("read backup %s: %v", backup, rerr)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("backup contents = %q, want original %q", got, original)
	}
}

func TestFixSecretRuleIsManualOnly(t *testing.T) {
	repo := initFixRepo(t, map[string]string{
		"go.mod":    "module example.com/x\n\ngo 1.26\n",
		"README.md": "# temp\n",
	})

	out, err := runFix(t, "--path", repo, "--rule", "AE-SEC-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "AE-SEC-001") || !strings.Contains(out, "not auto-fixable") {
		t.Fatalf("expected a manual-only note for AE-SEC-001, got:\n%s", out)
	}
	if _, statErr := os.Stat(filepath.Join(repo, ".charter", "backups")); !os.IsNotExist(statErr) {
		t.Fatalf("a manual-only no-op must not create .charter/backups (stat err = %v)", statErr)
	}
}

func TestFixRejectsInvalidRule(t *testing.T) {
	_, err := runFix(t, "--rule", "bogus")
	if err == nil {
		t.Fatal("expected an invalid rule id error")
	}

	var signal interface{ ExitCode() int }
	if !errors.As(err, &signal) {
		t.Fatalf("expected command exit error, got %T", err)
	}
	if signal.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", signal.ExitCode())
	}
}

func TestFixCreateDiffForMissingAGENTS(t *testing.T) {
	// A Go repo with no agent-context file at all: AE-CTX-001 fires.
	repo := initFixRepo(t, map[string]string{
		"go.mod": "module example.com/x\n\ngo 1.26\n",
	})

	out, err := runFix(t, "--path", repo, "--rule", "AE-CTX-001", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--- /dev/null") {
		t.Fatalf("expected a /dev/null create header, got:\n%s", out)
	}
	if !strings.Contains(out, "+++ b/AGENTS.md") {
		t.Fatalf("expected a create diff for AGENTS.md, got:\n%s", out)
	}
	if _, statErr := os.Stat(filepath.Join(repo, "AGENTS.md")); !os.IsNotExist(statErr) {
		t.Fatalf("dry run must not create AGENTS.md (stat err = %v)", statErr)
	}
}
