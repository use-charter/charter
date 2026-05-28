package repository

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBuildInventoryRespectsGitignore(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, ".gitignore"), ".charter/\n")
	writeTestFile(t, filepath.Join(root, "AGENTS.md"), "# Fixture Repo\n")
	writeTestFile(t, filepath.Join(root, "notes.txt"), "keep me\n")
	writeTestFile(t, filepath.Join(root, ".charter", "local.txt"), "ignored local artifact\n")
	initTestRepository(t, root)
	gitInTestRepo(t, root, "add", ".gitignore", "AGENTS.md")

	inv, err := BuildInventory(root)
	if err != nil {
		t.Fatalf("expected inventory build to succeed: %v", err)
	}

	if inv.Has(".charter/local.txt") {
		t.Fatalf("expected ignored local artifact to be excluded")
	}

	if !inv.Has("AGENTS.md") {
		t.Fatalf("expected AGENTS.md to be present in inventory")
	}

	if !inv.Has("notes.txt") {
		t.Fatalf("expected untracked non-ignored file to be present in inventory")
	}
}

func initTestRepository(t *testing.T, root string) {
	t.Helper()
	gitInTestRepo(t, root, "init")
}

func gitInTestRepo(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("expected parent directory for %q to be created: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("expected test file %q to be written: %v", path, err)
	}
}
