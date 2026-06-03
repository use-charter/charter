package operability

import (
	"os"
	"path/filepath"
	"slices"
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

// embedGo is a minimal Go source file declaring a //go:embed directive over the
// given patterns, mirroring Charter's own internal/render/html/render.go.
func embedGo(patterns string) string {
	return "package render\n\nimport \"embed\"\n\n//go:embed " + patterns + "\nvar assets embed.FS\n"
}

func TestEmbeddedJSAssetIsNotActive(t *testing.T) {
	// The Charter dogfood case: a Go repo whose only JS is a single web asset
	// embedded into the binary via //go:embed must NOT activate JavaScript, so
	// AE-TEST-001 does not fire for an "untested" JS surface that is really a
	// bundled resource. A package.json (build tooling) is present.
	ids := ruleIDs(t, map[string]string{
		"go.mod":                           "module x\n\ngo 1.26\n",
		"internal/app/app.go":              "package app\n",
		"internal/app/app_test.go":         "package app\n",
		"package.json":                     `{"name":"x","scripts":{"build":"bun build"}}`,
		"internal/render/render.go":        embedGo("assets/report.js"),
		"internal/render/assets/report.js": "export const x = 1;\n",
	})
	if len(ids) != 0 {
		t.Fatalf("a //go:embed'd .js asset must not activate JS (no AE-TEST-001); got %v", ids)
	}
}

func TestNonEmbeddedJSStillActivates(t *testing.T) {
	// No regression: real, non-embedded JS source with a package.json and no JS
	// tests must still flag AE-TEST-001 (the embedded-asset gate is precise).
	ids := ruleIDs(t, map[string]string{
		"go.mod":                   "module x\n\ngo 1.26\n",
		"internal/app/app.go":      "package app\n",
		"internal/app/app_test.go": "package app\n",
		"package.json":             `{"name":"x","scripts":{"build":"tsc"}}`,
		"web/index.js":             "export const app = () => 1;\n",
	})
	if !has(ids, "AE-TEST-001") {
		t.Fatalf("real non-embedded JS with no tests should still flag AE-TEST-001; got %v", ids)
	}
}

func TestMixedEmbeddedAndRealJSStillActivates(t *testing.T) {
	// One embedded .js (excluded) alongside one real non-embedded .js (counts):
	// JS stays active and AE-TEST-001 fires for the genuine untested surface.
	ids := ruleIDs(t, map[string]string{
		"go.mod":                           "module x\n\ngo 1.26\n",
		"internal/app/app.go":              "package app\n",
		"internal/app/app_test.go":         "package app\n",
		"package.json":                     `{"name":"x"}`,
		"internal/render/render.go":        embedGo("assets/report.js"),
		"internal/render/assets/report.js": "export const x = 1;\n",
		"web/app.js":                       "export const app = () => 1;\n",
	})
	if !has(ids, "AE-TEST-001") {
		t.Fatalf("a real non-embedded .js alongside an embedded one keeps JS active; got %v", ids)
	}
}

func TestDirectoryEmbedExcludesSubtree(t *testing.T) {
	// A directory embed (`//go:embed all:assets`) drops every web file under the
	// embedded subtree recursively, so JS is not activated.
	ids := ruleIDs(t, map[string]string{
		"go.mod":                              "module x\n\ngo 1.26\n",
		"internal/app/app.go":                 "package app\n",
		"internal/app/app_test.go":            "package app\n",
		"package.json":                        `{"name":"x"}`,
		"internal/render/render.go":           embedGo("all:assets"),
		"internal/render/assets/report.js":    "export const x = 1;\n",
		"internal/render/assets/js/vendor.js": "export const y = 2;\n",
	})
	if len(ids) != 0 {
		t.Fatalf("a directory //go:embed must exclude the whole subtree; got %v", ids)
	}
}

func TestGlobEmbedExcludesMatches(t *testing.T) {
	// A glob embed (`//go:embed assets/*.js`) drops the matched files.
	ids := ruleIDs(t, map[string]string{
		"go.mod":                           "module x\n\ngo 1.26\n",
		"internal/app/app.go":              "package app\n",
		"internal/app/app_test.go":         "package app\n",
		"package.json":                     `{"name":"x"}`,
		"internal/render/render.go":        embedGo("assets/*.js"),
		"internal/render/assets/report.js": "export const x = 1;\n",
	})
	if len(ids) != 0 {
		t.Fatalf("a glob //go:embed must exclude matched files; got %v", ids)
	}
}

func TestIndentedEmbedDirectiveIsIgnored(t *testing.T) {
	// A //go:embed comment indented inside a function body is ignored by the Go
	// compiler (real directives sit at column 0), so it must NOT exclude the
	// referenced asset. The .js here resolves to the indented pattern's target,
	// so only the column-0 fix keeps it counted: JS activates, AE-TEST-001 fires.
	ids := ruleIDs(t, map[string]string{
		"go.mod":                    "module x\n\ngo 1.26\n",
		"internal/app/app.go":       "package app\n",
		"internal/app/app_test.go":  "package app\n",
		"package.json":              `{"name":"x"}`,
		"internal/render/render.go": "package render\n\nimport \"embed\"\n\nfunc load() embed.FS {\n\t//go:embed app.js\n\tvar fsys embed.FS\n\treturn fsys\n}\n",
		"internal/render/app.js":    "export const app = () => 1;\n",
	})
	if !has(ids, "AE-TEST-001") {
		t.Fatalf("an indented //go:embed must not exclude real source; JS should activate; got %v", ids)
	}
}

func TestRootLevelEmbedResolves(t *testing.T) {
	// A root-level .go file (pathDir == "") embedding a root-level asset must
	// resolve via path.Join("", asset) and be excluded — so JS does not activate.
	ids := ruleIDs(t, map[string]string{
		"go.mod":       "module x\n\ngo 1.26\n",
		"main.go":      "package main\n\nimport \"embed\"\n\n//go:embed report.js\nvar assets embed.FS\n\nfunc main() {}\n",
		"main_test.go": "package main\n",
		"package.json": `{"name":"x"}`,
		"report.js":    "export const x = 1;\n",
	})
	if len(ids) != 0 {
		t.Fatalf("a root-level //go:embed asset must be excluded (pathDir==\"\"); got %v", ids)
	}
}

func TestSplitEmbedPatterns(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"bare", "assets/report.js", []string{"assets/report.js"}},
		{"multiple bare", "a.js  b.css\tc.html", []string{"a.js", "b.css", "c.html"}},
		{"double quoted", `"assets/report.js"`, []string{"assets/report.js"}},
		{"backtick quoted", "`assets/report.js`", []string{"assets/report.js"}},
		{"double quoted with space", `"my assets/report.js"`, []string{"my assets/report.js"}},
		{"backtick with space", "`my dir/x.js`", []string{"my dir/x.js"}},
		{"mixed quoting", "a.js \"b c.js\" `d e.js`", []string{"a.js", "b c.js", "d e.js"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := splitEmbedPatterns(tc.in); !slices.Equal(got, tc.want) {
				t.Fatalf("splitEmbedPatterns(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
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
