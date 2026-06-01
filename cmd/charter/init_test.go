package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func runInit(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(append([]string{"init"}, args...))
	cmd.SetContext(context.Background())
	err := cmd.Execute()
	return out.String(), err
}

// snapshotDir maps every file (repo-relative path -> contents) under root so a
// before/after comparison can prove init touched nothing.
func snapshotDir(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := make(map[string]string)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		snap[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return snap
}

func TestInitCreatesExpectedFiles(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "go.mod", "module example.com/x\n\ngo 1.26\n")

	out, err := runInit(t, "--path", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Default agents fall back to claude, so .claude/settings.json is expected.
	want := []string{
		"AGENTS.md",
		"charter.yaml",
		".gitignore",
		"ARCHITECTURE.md",
		".env.example",
		filepath.Join(".claude", "settings.json"),
	}
	for _, rel := range want {
		if _, statErr := os.Stat(filepath.Join(dir, rel)); statErr != nil {
			t.Errorf("expected %s to exist after init: %v", rel, statErr)
		}
	}
	if !strings.Contains(out, "created") {
		t.Fatalf("expected a created/skipped summary, got:\n%s", out)
	}
	if !strings.Contains(out, "charter doctor") {
		t.Fatalf("expected a next-step hint, got:\n%s", out)
	}
}

func TestInitNeverOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "go.mod", "module example.com/x\n\ngo 1.26\n")

	const sentinel = "PRE-EXISTING"
	writeTempFile(t, dir, "AGENTS.md", sentinel)

	out, err := runInit(t, "--path", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, rerr := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if rerr != nil {
		t.Fatalf("read AGENTS.md: %v", rerr)
	}
	if string(data) != sentinel {
		t.Fatalf("init must never overwrite an existing file; AGENTS.md = %q, want %q", string(data), sentinel)
	}
	if !strings.Contains(out, "skip AGENTS.md") {
		t.Fatalf("expected AGENTS.md reported as skip, got:\n%s", out)
	}

	// Second run: everything now exists, so it must be all-skip and byte-identical.
	before := snapshotDir(t, dir)
	out2, err := runInit(t, "--path", dir)
	if err != nil {
		t.Fatalf("unexpected error on second run: %v", err)
	}
	if strings.Contains(out2, "create ") {
		t.Fatalf("second run must create nothing, got:\n%s", out2)
	}
	if !strings.Contains(out2, "0 created") {
		t.Fatalf("expected second run to report 0 created, got:\n%s", out2)
	}
	after := snapshotDir(t, dir)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("second run changed bytes on disk:\nbefore=%v\nafter=%v", before, after)
	}
}

func TestInitDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()

	out, err := runInit(t, "--path", dir, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "would create") {
		t.Fatalf("expected 'would create' lines in dry-run, got:\n%s", out)
	}

	candidates := []string{
		"AGENTS.md",
		"charter.yaml",
		".gitignore",
		"ARCHITECTURE.md",
		".env.example",
		filepath.Join(".claude", "settings.json"),
	}
	for _, rel := range candidates {
		if _, statErr := os.Stat(filepath.Join(dir, rel)); !os.IsNotExist(statErr) {
			t.Fatalf("dry run must not create %s (stat err = %v)", rel, statErr)
		}
	}
}

func TestInitProfileStrictWritesProfile(t *testing.T) {
	dir := t.TempDir()

	if _, err := runInit(t, "--path", dir, "--profile", "strict"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, rerr := os.ReadFile(filepath.Join(dir, "charter.yaml"))
	if rerr != nil {
		t.Fatalf("read charter.yaml: %v", rerr)
	}
	if !strings.Contains(string(data), "profile: strict") {
		t.Fatalf("charter.yaml missing strict profile:\n%s", string(data))
	}
}

func TestInitRejectsInvalidProfile(t *testing.T) {
	dir := t.TempDir()

	_, err := runInit(t, "--path", dir, "--profile", "bogus")
	if err == nil {
		t.Fatal("expected invalid profile error")
	}

	var signal interface{ ExitCode() int }
	if !errors.As(err, &signal) {
		t.Fatalf("expected command exit error, got %T", err)
	}
	if signal.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", signal.ExitCode())
	}
}
