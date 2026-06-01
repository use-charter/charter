package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func TestDoctorCommandTextOutputShowsLocation(t *testing.T) {
	repo := t.TempDir()
	writeTempFile(t, repo, "AGENTS.md", "# weak context\n")
	gitInRepo(t, repo, "init", "-q")
	gitInRepo(t, repo, "config", "user.name", "Charter Test")
	gitInRepo(t, repo, "config", "user.email", "charter@example.com")
	gitInRepo(t, repo, "config", "commit.gpgsign", "false")
	gitInRepo(t, repo, "add", ".")
	gitInRepo(t, repo, "commit", "-q", "-m", "fixture")

	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"doctor", "--path", repo, "--threshold", "80"})
	cmd.SetContext(context.Background())

	// Below threshold returns a (silent) error; text output is still written.
	_ = cmd.Execute()

	if !bytes.Contains(out.Bytes(), []byte("location: AGENTS.md")) {
		t.Fatalf("expected text output to show the finding location, got:\n%s", out.String())
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

func TestDoctorCommandRejectsInvalidFormat(t *testing.T) {
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"doctor", "--format", "yaml"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected invalid format error")
	}

	var signal interface{ ExitCode() int }
	if !errors.As(err, &signal) {
		t.Fatalf("expected command exit error, got %T", err)
	}

	if signal.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", signal.ExitCode())
	}
}

func TestDoctorCommandJSONOutput(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	cmd := newRootCommand()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"doctor", "--path", repo, "--threshold", "80", "--format", "json"})
	cmd.SetContext(context.Background())

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected json output to succeed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid json output: %v", err)
	}

	if payload["threshold"] != float64(80) {
		t.Fatalf("unexpected threshold: %#v", payload["threshold"])
	}
	if payload["passed"] != true {
		t.Fatalf("expected passed=true, got %#v", payload["passed"])
	}
}

func TestDoctorCommandQuietJSONStillOutputsPayload(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	cmd := newRootCommand()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"doctor", "--path", repo, "--threshold", "80", "--format", "json", "--quiet"})
	cmd.SetContext(context.Background())

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected quiet json output to succeed: %v", err)
	}

	if len(bytes.TrimSpace(out.Bytes())) == 0 {
		t.Fatalf("expected full json payload even when quiet is set")
	}
}

func TestDoctorCommandSARIFOutput(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"doctor", "--path", repo, "--format", "sarif"})
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sarif run: %v", err)
	}
	var log map[string]any
	if err := json.Unmarshal(out.Bytes(), &log); err != nil {
		t.Fatalf("expected valid SARIF JSON: %v", err)
	}
	if log["version"] != "2.1.0" {
		t.Fatalf("expected SARIF 2.1.0, got %#v", log["version"])
	}
}

func TestDoctorCommandOutWritesFile(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "charter.sarif")
	cmd := newRootCommand()
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"doctor", "--path", repo, "--format", "sarif", "--out", outPath})
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected nothing on stdout when --out is set, got %q", stdout.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected %s written: %v", outPath, err)
	}
}

func makeTempDoctorRepo(t *testing.T) (string, error) {
	t.Helper()

	root := filepath.Join("..", "..", "testdata", "repos", "pass-slice1")
	repo := t.TempDir()
	if err := copyDir(root, repo); err != nil {
		return "", err
	}

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Charter Test"},
		{"git", "config", "user.email", "charter@example.com"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "add", "."},
		{"git", "commit", "-m", "fixture"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("%s failed: %w\n%s", args[0], err, out)
		}
	}

	return repo, nil
}

func initTempRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	writeTempFile(t, repo, "README.md", "# temp repo\n")
	gitInRepo(t, repo, "init", "-q")
	gitInRepo(t, repo, "config", "user.name", "Charter Test")
	gitInRepo(t, repo, "config", "user.email", "charter@example.com")
	gitInRepo(t, repo, "config", "commit.gpgsign", "false")
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

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}

		out, err := os.Create(target)
		if err != nil {
			_ = in.Close()
			return err
		}

		if _, err := io.Copy(out, in); err != nil {
			_ = in.Close()
			_ = out.Close()
			return err
		}

		if err := in.Close(); err != nil {
			_ = out.Close()
			return err
		}

		if err := out.Close(); err != nil {
			return err
		}

		return os.Chmod(target, info.Mode())
	})
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
