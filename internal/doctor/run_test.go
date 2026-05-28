package doctor

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAgainstFixtureRepo(t *testing.T) {
	root := prepareFixtureRepo(t, filepath.Join("..", "..", "testdata", "repos", "pass-slice1"))

	result, err := Run(root)
	if err != nil {
		t.Fatalf("expected doctor run to succeed: %v", err)
	}

	if result.Score.Final != 100 {
		t.Fatalf("expected passing score 100 for the fixture and rule set, got %d", result.Score.Final)
	}
}

func prepareFixtureRepo(t *testing.T, source string) string {
	t.Helper()

	root := t.TempDir()
	copyFixtureTree(t, source, root)

	git(t, root, "init", "-q")
	git(t, root, "config", "user.name", "Charter Test")
	git(t, root, "config", "user.email", "charter@example.com")
	git(t, root, "add", ".")
	git(t, root, "commit", "-q", "-m", "fixture")

	return root
}

func copyFixtureTree(t *testing.T, source string, destination string) {
	t.Helper()

	err := filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		target := filepath.Join(destination, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		return copyFile(path, target)
	})
	if err != nil {
		t.Fatalf("copy fixture repo: %v", err)
	}
}

func copyFile(source string, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		_ = input.Close()
		return err
	}

	output, err := os.Create(destination)
	if err != nil {
		_ = input.Close()
		return err
	}

	if _, err := io.Copy(output, input); err != nil {
		_ = input.Close()
		_ = output.Close()
		return err
	}

	if err := input.Close(); err != nil {
		_ = output.Close()
		return err
	}

	return output.Close()
}

func git(t *testing.T, root string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
