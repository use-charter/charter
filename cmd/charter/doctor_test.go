package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDoctorCommandRuns(t *testing.T) {
	repo := initTempRepo(t)

	cmd := newRootCommand()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"doctor", "--path", repo, "--threshold", "80", "--quiet"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected command to fail when score stays below threshold")
	}

	if err.Error() != "score below threshold" {
		t.Fatalf("expected threshold failure, got %q", err.Error())
	}

	if !bytes.Contains(out.Bytes(), []byte("charter: score ")) {
		t.Fatalf("expected quiet failure summary line")
	}

	if !bytes.Contains(out.Bytes(), []byte(" — FAIL")) {
		t.Fatalf("expected quiet failure summary line to use em dash contract, got %q", out.String())
	}
}

func TestDoctorCommandHelpRuns(t *testing.T) {
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"doctor", "--help"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected doctor help to run without error: %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("Scan a repository and compute a Charter score")) {
		t.Fatalf("expected help output to include doctor command description")
	}
}

func initTempRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	writeTempFile(t, repo, "README.md", "# temp repo\n")
	gitInRepo(t, repo, "init", "-q")
	gitInRepo(t, repo, "config", "user.name", "Charter Test")
	gitInRepo(t, repo, "config", "user.email", "charter@example.com")
	gitInRepo(t, repo, "add", ".")
	gitInRepo(t, repo, "commit", "-q", "-m", "fixture")

	return repo
}

func writeTempFile(t *testing.T, root string, relative string, content string) {
	t.Helper()

	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir temp fixture path: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp fixture file: %v", err)
	}
}

func gitInRepo(t *testing.T, root string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
