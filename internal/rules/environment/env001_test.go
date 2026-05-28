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
		"package.json": "{\n  \"name\": \"pass-env\",\n  \"private\": true,\n  \"engines\": { \"bun\": \"1.3.14\" }\n}\n",
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

func TestRunPassesWithEquivalentReproducibilitySignals(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".nvmrc":                    "22\n",
		"package-lock.json":         "{}\n",
		"lefthook.yml":              "pre-commit:\n  commands: {}\n",
		"package.json":              "{\n  \"name\": \"pass-env\",\n  \"private\": true,\n  \"engines\": { \"node\": \"22.x\" }\n}\n",
		"pyproject.toml":            "[project]\nname = \"pass-env\"\nrequires-python = \">=3.12\"\n",
		"uv.lock":                   "version = 1\n",
		"Gemfile":                   "source \"https://rubygems.org\"\nruby \"3.3.1\"\n",
		"Gemfile.lock":              "GEM\n",
		"gradle/libs.versions.toml": "[versions]\ngradle = \"9.0\"\n",
		"gradle/wrapper/gradle-wrapper.properties": "distributionUrl=https\\://services.gradle.org/distributions/gradle-9.0-bin.zip\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunFindsManifestWithoutRuntimePin(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"package.json":      "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
		"package-lock.json": "{}\n",
		"hk.pkl":            "hooks {}\n",
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
	if findings[0].Evidence[0] != "missing toolchain signal for active language: javascript" {
		t.Fatalf("expected missing javascript toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunPassesWithoutGoSumWhenGoHasNoDependencyState(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"go.mod":    "module example.com/passenv\n\ngo 1.26.0\n",
		"mise.toml": "[tools]\ngo = \"1.26.3\"\n",
		"hk.pkl":    "hooks {}\n",
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
	if findings[0].Evidence[0] != "missing lockfile signal for active language: javascript" {
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
