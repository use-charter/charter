package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"go.use-charter.dev/charter/internal/doctor"
)

// fixFixtureAGENTS is a minimal AGENTS.md that satisfies AE-CTX-001 (project
// summary, "Go" tech stack, off-limits boundaries, a `charter doctor`
// verification command, >= 5 non-empty lines, under the ~600-token budget),
// AE-CTX-002 (the `.env*` / `secrets/` repo-truth markers plus the verification
// command), and AE-CC-002 (concrete off-limits paths). It mirrors the shape of
// the scaffold's generated context file so those three rules stay quiet and the
// fixture isolates AE-CTX-004 and AE-CI-002.
const fixFixtureAGENTS = `# AGENTS.md

## Project Overview

demo is a Go project. Describe what it does and who uses it (edit this line).

## Tech Stack

- Go 1.26

## Commands

- Verify: ` + "`charter doctor`" + `
- Build: ` + "`go build ./...`" + `
- Test: ` + "`go test ./...`" + `

## Edit Boundaries

- Safe for agents: application source, tests, docs
- Off-limits: ` + "`.github/workflows/`, `.env*`, `secrets/`, `db/migrations/`, `terraform/`" + `

## Verification

- Run ` + "`charter doctor`" + ` before committing.
`

// fixFixtureCI is a realistic NON-moon GitHub workflow. Its direct-form run
// steps satisfy the generalized AE-CI-002 coverage for repo quality
// (`go test ./...`), workflow linting (`actionlint` + `zizmor`), and security
// (`govulncheck ./...`), and `actions/checkout` is SHA-pinned (the v6.0.2 SHA
// reused from the repo's real .github/workflows/ci.yml). It deliberately does
// NOT run Charter, so the only AE-CI-002 gap at baseline is the missing
// Charter product gate.
const fixFixtureCI = `name: CI

on:
  push:
    branches: ["main"]
  pull_request:
    branches: ["main"]

permissions:
  contents: read

jobs:
  build:
    name: Build and verify
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
      - name: Run unit tests
        run: go test ./...
      - name: Lint workflows
        run: actionlint
      - name: Audit workflows
        run: zizmor .
      - name: Scan for vulnerabilities
        run: govulncheck ./...
`

// fixFixtureGitignore is a PARTIAL .gitignore: it carries four of the six agent
// artifact patterns AE-CTX-004 requires but omits `.hk/` and `.env*`, so the
// rule fires at baseline and the fixer has a real gap to repair.
const fixFixtureGitignore = `# project ignores
.charter/
*.charter-session
.claude/local/
.cursor/cache/
`

// TestFixClearsCTX004AndCI002 proves end-to-end that `charter fix` clears both
// AE-CTX-004 (a partial .gitignore) and AE-CI-002 (a missing Charter CI gate on
// a realistic non-moon repo) and raises the doctor score. The AE-CI-002
// generalization lets the repo's existing direct-form CI count for repo
// quality / workflow lint / security so only the Charter gate is missing. It is
// fully hermetic — a throwaway git repo under t.TempDir() with no network.
func TestFixClearsCTX004AndCI002(t *testing.T) {
	repo := t.TempDir()

	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.email", "charter@example.com")
	runGit(t, repo, "config", "user.name", "Charter Test")
	runGit(t, repo, "config", "commit.gpgsign", "false")

	writeTempFile(t, repo, "go.mod", "module example.com/demo\n\ngo 1.26\n")
	writeTempFile(t, repo, "main.go", "package main\n\nfunc main() {}\n")
	writeTempFile(t, repo, "AGENTS.md", fixFixtureAGENTS)
	writeTempFile(t, repo, ".gitignore", fixFixtureGitignore)
	writeTempFile(t, repo, ".github/workflows/ci.yml", fixFixtureCI)
	runGit(t, repo, "add", "-A")

	// Baseline: doctor must report both target findings before any repair.
	baseline, err := doctor.Run(repo, 80, true)
	if err != nil {
		t.Fatalf("baseline doctor.Run: %v", err)
	}
	baselineIDs := ruleIDSet(baseline)
	t.Logf("baseline score: %d (passed=%v); findings: %v", baseline.Score.Final, baseline.Passed, sortedRuleIDs(baselineIDs))
	if !baselineIDs["AE-CTX-004"] {
		t.Fatalf("expected AE-CTX-004 present at baseline; findings: %v", sortedRuleIDs(baselineIDs))
	}
	if !baselineIDs["AE-CI-002"] {
		t.Fatalf("expected AE-CI-002 present at baseline; findings: %v", sortedRuleIDs(baselineIDs))
	}

	// Repair diff-first via the real `charter fix` command path.
	cmd := newRootCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"fix", "--path", repo, "--yes"})
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("fix: %v", err)
	}

	// Stage the new charter.yaml workflow + the appended .gitignore so doctor's
	// git ls-files inventory sees them on the re-scan.
	runGit(t, repo, "add", "-A")

	after, err := doctor.Run(repo, 80, true)
	if err != nil {
		t.Fatalf("after doctor.Run: %v", err)
	}
	afterIDs := ruleIDSet(after)
	t.Logf("after score: %d (passed=%v); findings: %v", after.Score.Final, after.Passed, sortedRuleIDs(afterIDs))

	if afterIDs["AE-CTX-004"] {
		t.Errorf("expected AE-CTX-004 cleared after fix, still present; findings: %v", sortedRuleIDs(afterIDs))
	}
	if afterIDs["AE-CI-002"] {
		t.Errorf("expected AE-CI-002 cleared after fix, still present; findings: %v", sortedRuleIDs(afterIDs))
	}
	if after.Score.Final <= baseline.Score.Final {
		t.Errorf("expected score to rise after fix: baseline=%d, after=%d", baseline.Score.Final, after.Score.Final)
	}

	// The .gitignore was backed up before the AE-CTX-004 append, so the backups
	// directory must exist on disk.
	backupsRoot := filepath.Join(repo, ".charter", "backups")
	if _, statErr := os.Stat(backupsRoot); statErr != nil {
		t.Errorf("expected %s to exist (the .gitignore should have been backed up before append): %v", backupsRoot, statErr)
	}
}

// ruleIDSet collects the rule ids of a doctor result's active findings.
func ruleIDSet(result doctor.Result) map[string]bool {
	set := make(map[string]bool, len(result.Findings))
	for _, f := range result.Findings {
		set[f.RuleID] = true
	}
	return set
}

// sortedRuleIDs returns the set's rule ids sorted for stable log output.
func sortedRuleIDs(set map[string]bool) []string {
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
