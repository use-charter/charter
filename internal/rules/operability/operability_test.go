package operability

import (
	"os"
	"path/filepath"
	"testing"

	"go.use-charter.dev/charter/internal/repository"
	"go.use-charter.dev/charter/internal/scoring"
)

// repo writes files to a temp root and returns (root, inventory) built from the
// same paths, so detection that reads file contents works.
func repo(t *testing.T, files map[string]string) (string, repository.Inventory) {
	t.Helper()
	root := t.TempDir()
	paths := make([]string, 0, len(files))
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, name)
	}
	return root, repository.New(paths)
}

func ruleIDs(t *testing.T, files map[string]string) []string {
	t.Helper()
	root, inv := repo(t, files)
	var ids []string
	for _, f := range Run(root, inv) {
		ids = append(ids, f.RuleID)
	}
	return ids
}

func has(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func TestNonCodeRepoIsNA(t *testing.T) {
	if ids := ruleIDs(t, map[string]string{"README.md": "# docs\n", "AGENTS.md": "x\n"}); len(ids) != 0 {
		t.Fatalf("docs-only repo should yield no operability findings, got %v", ids)
	}
}

func TestGoWithTestsClean(t *testing.T) {
	ids := ruleIDs(t, map[string]string{
		"go.mod":                   "module x\n\ngo 1.26\n",
		"internal/app/app.go":      "package app\n",
		"internal/app/app_test.go": "package app\n",
	})
	if len(ids) != 0 {
		t.Fatalf("Go repo with tests + go.mod toolchain should be clean, got %v", ids)
	}
}

func TestGoWithoutTestsFlagsTest(t *testing.T) {
	ids := ruleIDs(t, map[string]string{
		"go.mod":              "module x\n\ngo 1.26\n",
		"internal/app/app.go": "package app\n",
	})
	if !has(ids, "AE-TEST-001") {
		t.Fatalf("Go repo without tests should flag AE-TEST-001, got %v", ids)
	}
	// go.mod is a conventional test toolchain -> AE-AUTO-001 must NOT fire.
	if has(ids, "AE-AUTO-001") {
		t.Fatalf("Go (conventional toolchain) should not flag AE-AUTO-001, got %v", ids)
	}
}

func TestToolingOnlyJSIsNotActive(t *testing.T) {
	// A Go repo whose only JS/TS lives in scripts/ (build tooling) must NOT be
	// treated as a JS surface — the core false-positive guard.
	ids := ruleIDs(t, map[string]string{
		"go.mod":                   "module x\n\ngo 1.26\n",
		"internal/app/app.go":      "package app\n",
		"internal/app/app_test.go": "package app\n",
		"package.json":             `{"name":"x","scripts":{"build":"bun run scripts/build.ts"}}`,
		"scripts/build.ts":         "export const x = 1;\n",
	})
	if len(ids) != 0 {
		t.Fatalf("tooling-only JS should be N/A; got %v", ids)
	}
}

func TestJSAppWithoutTestsFlagsBoth(t *testing.T) {
	// Real JS app source, no tests, no test script -> TEST + AUTO both fire.
	ids := ruleIDs(t, map[string]string{
		"package.json": `{"name":"app","scripts":{"build":"tsc"}}`,
		"src/index.ts": "export const app = () => 1;\n",
	})
	if !has(ids, "AE-TEST-001") || !has(ids, "AE-AUTO-001") {
		t.Fatalf("JS app with no tests and no test script should flag both, got %v", ids)
	}
}

func TestJSAppWithTestScriptAndTestsClean(t *testing.T) {
	ids := ruleIDs(t, map[string]string{
		"package.json":      `{"name":"app","scripts":{"test":"vitest run"}}`,
		"src/index.ts":      "export const app = () => 1;\n",
		"src/index.test.ts": "import {app} from './index';\n",
	})
	if len(ids) != 0 {
		t.Fatalf("JS app with tests + test script should be clean, got %v", ids)
	}
}

func TestPythonPytestConfiguredIsDiscoverable(t *testing.T) {
	ids := ruleIDs(t, map[string]string{
		"pyproject.toml":    "[tool.pytest.ini_options]\n",
		"src/app.py":        "def app():\n    return 1\n",
		"tests/test_app.py": "def test_app():\n    assert True\n",
	})
	if len(ids) != 0 {
		t.Fatalf("Python with tests + pytest config should be clean, got %v", ids)
	}
}

func TestRustInlineTestsCountAsTests(t *testing.T) {
	ids := ruleIDs(t, map[string]string{
		"Cargo.toml": "[package]\nname = \"x\"\n",
		"src/lib.rs": "pub fn add(a:i32,b:i32)->i32{a+b}\n#[cfg(test)]\nmod tests {\n  #[test]\n  fn t(){assert_eq!(2,1+1);}\n}\n",
	})
	if len(ids) != 0 {
		t.Fatalf("Rust with inline #[cfg(test)] should be clean (Cargo conventional + inline tests), got %v", ids)
	}
}

func TestMakefileTestTargetSatisfiesAutonomy(t *testing.T) {
	// A JS app with tests but the test command exposed only via a Makefile.
	ids := ruleIDs(t, map[string]string{
		"package.json":      `{"name":"app"}`,
		"src/index.ts":      "export const app = () => 1;\n",
		"src/index.test.ts": "test('x',()=>{});\n",
		"Makefile":          "test:\n\tvitest run\n",
	})
	if has(ids, "AE-AUTO-001") {
		t.Fatalf("Makefile test target should satisfy AE-AUTO-001, got %v", ids)
	}
}

func TestStrayLanguageFileWithoutManifestIsNotActive(t *testing.T) {
	// A Rust repo with a lone Homebrew .rb formula (no Gemfile) must not activate
	// Ruby — the ripgrep false positive from FP-validation.
	ids := ruleIDs(t, map[string]string{
		"Cargo.toml":           "[package]\nname=\"x\"\n",
		"src/main.rs":          "fn main(){}\n#[cfg(test)]\nmod t{#[test]fn a(){}}\n",
		"pkg/brew/tool-bin.rb": "class ToolBin < Formula\nend\n",
	})
	if len(ids) != 0 {
		t.Fatalf("a stray .rb (no Gemfile) must not activate Ruby; got %v", ids)
	}
}

func TestJSTestsInTestDirCount(t *testing.T) {
	// Tests in test/*.ts (AVA layout, no *.test.ts naming) must count — the ky
	// false positive from FP-validation.
	ids := ruleIDs(t, map[string]string{
		"package.json":    `{"name":"app","scripts":{"test":"ava"}}`,
		"source/index.ts": "export const app = () => 1;\n",
		"test/main.ts":    "import test from 'ava';\n",
	})
	if len(ids) != 0 {
		t.Fatalf("tests in test/*.ts should count; got %v", ids)
	}
}

func TestTestFindingDeductsAndIsHigh(t *testing.T) {
	root, inv := repo(t, map[string]string{
		"go.mod":              "module x\n\ngo 1.26\n",
		"internal/app/app.go": "package app\n",
	})
	fs := Run(root, inv)
	if got := scoring.Calculate(fs).Final; got != 90 {
		t.Fatalf("a single High AE-TEST-001 should score 90, got %d", got)
	}
}
