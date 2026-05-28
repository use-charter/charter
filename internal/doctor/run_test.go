package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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
		{"git", "add", "."},
		{"git", "commit", "-m", "fixture"},
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

func TestRunAgainstFixtureRepo(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "pass-slice1")
	repo, err := makeTempGitRepoFromFixture(t, root)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("expected doctor run to succeed: %v", err)
	}

	if result.Score.Final != 100 {
		t.Fatalf("expected passing score 100 for the fixture and rule set, got %d", result.Score.Final)
	}
}

func TestRunSetsThresholdAndPassed(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "pass-slice1")
	repo, err := makeTempGitRepoFromFixture(t, root)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("expected doctor run to succeed: %v", err)
	}

	if result.Threshold != 80 {
		t.Fatalf("expected threshold 80, got %d", result.Threshold)
	}

	if !result.Passed {
		t.Fatalf("expected run to pass")
	}
}
