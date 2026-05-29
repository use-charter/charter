package secrets

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func fakeOpenAIKey() string {
	return "sk-" + strings.Repeat("a", 24)
}

func makeTempGitRepoFromFixture(t *testing.T, fixtureRoot string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	if err := copyDir(fixtureRoot, dir); err != nil {
		return "", err
	}

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Charter Test"},
		{"git", "config", "user.email", "charter@example.com"},
		{"git", "-c", "commit.gpgsign=false", "add", "."},
		{"git", "-c", "commit.gpgsign=false", "commit", "-m", "fixture"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("%s failed: %w\n%s", args[0], err, out)
		}
	}

	return dir, nil
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

		_, copyErr := io.Copy(out, in)
		closeInErr := in.Close()
		closeOutErr := out.Close()

		if copyErr != nil {
			return copyErr
		}
		if closeInErr != nil {
			return closeInErr
		}
		if closeOutErr != nil {
			return closeOutErr
		}

		if err := os.Chmod(target, info.Mode()); err != nil {
			return err
		}

		return nil
	})
}

func writeFile(t *testing.T, root, rel, contents string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func stageAndCommitAll(t *testing.T, root, message string) {
	t.Helper()

	cmd := exec.Command("git", "-C", root, "-c", "commit.gpgsign=false", "add", ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}

	cmd = exec.Command("git", "-C", root, "-c", "commit.gpgsign=false", "commit", "-m", message)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}
}
