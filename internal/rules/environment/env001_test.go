package environment

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.charter.dev/charter/internal/repository"
)

func TestRunPassesWhenReproducibilityFilesExist(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"mise.toml":    "[tools]\ngo = \"1.26.3\"\n",
		"mise.lock":    "lock-placeholder\n",
		"go.mod":       "module example.com/passenv\n\ngo 1.26.0\n",
		"go.sum":       "example.com v0.0.0 h1:abc\n",
		"hk.pkl":       "hooks {}\n",
		"package.json": "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
		"bun.lock":     "lock-placeholder\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunFindsMissingLockfile(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"mise.toml":    "[tools]\ngo = \"1.26.3\"\n",
		"mise.lock":    "lock-placeholder\n",
		"go.mod":       "module example.com/passenv\n\ngo 1.26.0\n",
		"go.sum":       "example.com v0.0.0 h1:abc\n",
		"hk.pkl":       "hooks {}\n",
		"package.json": "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if findings[0].RuleID != "AE-ENV-001" {
		t.Fatalf("expected AE-ENV-001, got %#v", findings[0])
	}
	if findings[0].Evidence[0] != "bun.lock" {
		t.Fatalf("expected missing bun.lock evidence, got %#v", findings[0].Evidence)
	}
}

func newEnvironmentRepo(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create dir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	gitInit := exec.Command("git", "init", "-q", root)
	if output, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	gitAdd := exec.Command("git", "-C", root, "add", ".")
	if output, err := gitAdd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v: %s", err, output)
	}

	return root
}
