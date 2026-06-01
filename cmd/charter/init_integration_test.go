package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.use-charter.dev/charter/internal/doctor"
)

// TestInitThenDoctorScoresAtLeast80 proves the architecture's headline promise:
// a blank Go repo scaffolded by `charter init` scores at or above the default
// Go threshold (80) when immediately scored by `charter doctor`. It is fully
// hermetic — a throwaway git repo under t.TempDir() with no network access.
func TestInitThenDoctorScoresAtLeast80(t *testing.T) {
	repo := t.TempDir()

	// Blank Go repo: a git repo with only a minimal module + entrypoint and no
	// agent-context scaffolding (no AGENTS.md, charter.yaml, or .gitignore yet).
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.email", "charter@example.com")
	runGit(t, repo, "config", "user.name", "Charter Test")
	runGit(t, repo, "config", "commit.gpgsign", "false")

	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/demo\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	runGit(t, repo, "add", "-A")

	// Scaffold the agent-context files via the real `charter init` path.
	cmd := newRootCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"init", "--path", repo, "--yes"})
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Stage the freshly-created scaffold so it lands in the inventory: doctor's
	// BuildInventory reads `git ls-files`, so unstaged files would be invisible.
	runGit(t, repo, "add", "-A")

	result, err := doctor.Run(repo, 80, true)
	if err != nil {
		t.Fatalf("doctor.Run: %v", err)
	}

	ruleIDs := make([]string, 0, len(result.Findings))
	for _, f := range result.Findings {
		ruleIDs = append(ruleIDs, f.RuleID)
	}
	t.Logf("init→doctor out-of-the-box score: %d (passed=%v); findings: %v", result.Score.Final, result.Passed, ruleIDs)

	// Assert the floor (>= 80), not an exact number: a few points off for an
	// unconfigured hook (AE-ENV-001) or missing CI workflow (AE-CI-002) is
	// acceptable as long as the promise holds.
	if result.Score.Final < 80 {
		t.Errorf("expected out-of-the-box score >= 80 after init, got %d; findings: %v", result.Score.Final, ruleIDs)
	}
	if !result.Passed {
		t.Errorf("expected result.Passed=true at threshold 80, got false (score=%d); findings: %v", result.Score.Final, ruleIDs)
	}
}

// runGit runs a git command in dir and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
