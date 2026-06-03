// Package operability implements the agent-operability rules: AE-TEST-001
// (automated tests are present) and AE-AUTO-001 (the verification command is
// discoverable and runnable). Together they answer "can an agent verify its own
// work and run this project?" — the operability axis that complements Charter's
// context and safety rules. Detection is pure, offline, and deterministic over
// the tracked inventory; no network, no LLM (Commitments #4/#7).
//
// A language is "active" only when it has BOTH a recognizing manifest (go.mod,
// package.json, Cargo.toml, …) AND non-test SOURCE outside tooling directories
// (scripts/, tools/, testdata/, …). The manifest gate rejects a stray secondary
// file (e.g. a single Homebrew `.rb` formula in a Rust repo); the source-outside-
// tooling gate rejects a tooling-only manifest (e.g. a package.json that only
// drives build scripts in a Go repo); and the embedded-asset gate drops files
// referenced by a //go:embed directive — a bundled resource of the host program,
// not an independent language surface (e.g. a web report.js embedded into a Go
// binary). All three are core false-positive guards.
package operability

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

const (
	langGo     = "Go"
	langJS     = "JavaScript/TypeScript"
	langPython = "Python"
	langRust   = "Rust"
	langJVM    = "Java/Kotlin"
	langRuby   = "Ruby"
	langCSharp = "C#"
	langPHP    = "PHP"
)

// stackInfo counts a language's non-test source files and test files (outside
// tooling dirs).
type stackInfo struct{ source, tests int }

// Run evaluates AE-TEST-001 and AE-AUTO-001 over the repository.
func Run(root string, inv repository.Inventory) []findings.Finding {
	stacks := scanStacks(root, inv)
	active := activeLanguages(stacks, inv)
	if len(active) == 0 {
		// No real code surface: both rules are N/A (a docs/config/tooling-only
		// repo is not penalized for having no tests or test command).
		return nil
	}

	var result []findings.Finding
	if f, ok := checkTestsPresent(root, inv, stacks, active); ok {
		result = append(result, f)
	}
	if f, ok := checkVerificationDiscoverable(root, inv, active); ok {
		result = append(result, f)
	}
	return result
}

// toolingSegments are path segments whose contents are build tooling, fixtures,
// or vendored/generated code — not the tested application surface.
var toolingSegments = map[string]struct{}{
	"scripts": {}, "tools": {}, "testdata": {}, "examples": {}, "example": {},
	"third_party": {}, "vendor": {}, "node_modules": {}, "dist": {}, "build": {},
	".github": {}, "gen": {}, "generated": {},
}

func inToolingDir(p string) bool {
	for _, seg := range strings.Split(p, "/") {
		if _, ok := toolingSegments[seg]; ok {
			return true
		}
	}
	return false
}

// scanStacks classifies every tracked source file by language and whether it is
// a test, skipping tooling directories and //go:embed'd assets.
func scanStacks(root string, inv repository.Inventory) map[string]*stackInfo {
	embedded := embeddedAssets(root, inv)
	m := map[string]*stackInfo{}
	for _, p := range inv.Paths {
		if inToolingDir(p) {
			continue
		}
		if _, ok := embedded[p]; ok {
			// A //go:embed'd file is a bundled resource of the host program,
			// not an independent language source surface — skip it.
			continue
		}
		lang, isTest, ok := classifySource(p)
		if !ok {
			continue
		}
		if m[lang] == nil {
			m[lang] = &stackInfo{}
		}
		if isTest {
			m[lang].tests++
		} else {
			m[lang].source++
		}
	}
	return m
}

// embeddedAssets returns the set of inventory paths referenced by any //go:embed
// directive in a tracked .go file. Such a file is a bundled resource compiled
// into the host program (e.g. a web report.js embedded into a Go binary), not an
// independent language surface, so it must not make a language "active" — the
// embedded-asset false-positive guard. Resolution is offline and deterministic
// over the inventory; real JS/Python apps do not //go:embed their own source, so
// the manifest and source-outside-tooling gates remain the primary activators.
func embeddedAssets(root string, inv repository.Inventory) map[string]struct{} {
	// //go:embed is Go-only, so the host program is always Go and only a
	// *secondary* (non-Go) language's manifest can be falsely activated by an
	// embedded asset. A language with no manifest is never active, so without a
	// non-Go manifest the gate cannot change any outcome — skip the .go scan
	// entirely (also keeps a pure-Go monorepo off the per-file read path).
	if !hasNonGoManifest(inv) {
		return nil
	}
	embedded := map[string]struct{}{}
	for _, p := range inv.Paths {
		if !strings.HasSuffix(p, ".go") {
			continue
		}
		content, ok := readRepoFile(root, inv, p)
		if !ok || !strings.Contains(content, "//go:embed") {
			continue // skip quickly: no directive in this file
		}
		dir := pathDir(p)
		for _, line := range strings.Split(content, "\n") {
			// A real //go:embed directive must sit at column 0 (no leading
			// whitespace) — the Go compiler ignores indented //go:embed comments
			// inside function bodies, so we must too, or we would over-exclude a
			// genuine source file. Strip only a trailing \r for Windows endings.
			rest, ok := strings.CutPrefix(strings.TrimRight(line, "\r"), "//go:embed")
			if !ok || rest == "" || (rest[0] != ' ' && rest[0] != '\t') {
				continue
			}
			for _, pat := range splitEmbedPatterns(rest) {
				pat = strings.TrimPrefix(pat, "all:")
				if pat == "" {
					continue
				}
				resolved := path.Join(dir, pat)
				for _, candidate := range inv.Paths {
					if matchesEmbedPattern(resolved, candidate) {
						embedded[candidate] = struct{}{}
					}
				}
			}
		}
	}
	return embedded
}

// matchesEmbedPattern reports whether an inventory path is covered by a resolved
// //go:embed pattern: an exact file, any file under an embedded directory subtree
// (go:embed of a directory embeds it recursively), or a path.Match glob (slash-
// based, like the stdlib embed matcher; '*' does not cross '/').
func matchesEmbedPattern(pattern, candidate string) bool {
	if candidate == pattern {
		return true
	}
	if strings.HasPrefix(candidate, pattern+"/") {
		return true
	}
	if ok, err := path.Match(pattern, candidate); err == nil && ok {
		return true
	}
	return false
}

// splitEmbedPatterns splits a //go:embed argument list into individual patterns,
// honoring Go's "..." and `...` quoting so a pattern containing spaces stays a
// single token. Surrounding quotes are stripped from the returned patterns.
func splitEmbedPatterns(s string) []string {
	var out []string
	for i, n := 0, len(s); i < n; {
		for i < n && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= n {
			break
		}
		switch s[i] {
		case '"':
			j := i + 1
			for j < n && s[j] != '"' {
				j++
			}
			out = append(out, s[i+1:j])
			i = j + 1
		case '`':
			j := i + 1
			for j < n && s[j] != '`' {
				j++
			}
			out = append(out, s[i+1:j])
			i = j + 1
		default:
			j := i
			for j < n && s[j] != ' ' && s[j] != '\t' {
				j++
			}
			out = append(out, s[i:j])
			i = j
		}
	}
	return out
}

func activeLanguages(stacks map[string]*stackInfo, inv repository.Inventory) []string {
	var out []string
	for lang, info := range stacks {
		if info.source > 0 && hasManifest(inv, lang) {
			out = append(out, lang)
		}
	}
	sort.Strings(out)
	return out
}

// nonGoLangs are the languages whose source could be //go:embed'd into a Go
// binary and thus need the embedded-asset gate (Go itself is never embedded as a
// language surface — its package files build the host).
var nonGoLangs = []string{langJS, langPython, langRust, langJVM, langRuby, langCSharp, langPHP}

// hasNonGoManifest reports whether any non-Go language manifest is present. It is
// a cheap, file-read-free gate (path lookups only) on the embedded-asset scan.
func hasNonGoManifest(inv repository.Inventory) bool {
	for _, lang := range nonGoLangs {
		if hasManifest(inv, lang) {
			return true
		}
	}
	return false
}

// hasManifest reports whether a language's project manifest is present — the
// gate that distinguishes a real language surface from a stray source file.
func hasManifest(inv repository.Inventory, lang string) bool {
	switch lang {
	case langGo:
		return inv.Has("go.mod")
	case langJS:
		return inv.Has("package.json")
	case langPython:
		return inv.Has("pyproject.toml") || inv.Has("setup.py") || inv.Has("setup.cfg") || inv.Has("requirements.txt")
	case langRust:
		return inv.Has("Cargo.toml")
	case langJVM:
		return inv.Has("pom.xml") || inv.Has("build.gradle") || inv.Has("build.gradle.kts")
	case langRuby:
		return inv.Has("Gemfile")
	case langCSharp:
		for _, p := range inv.Paths {
			if strings.HasSuffix(p, ".csproj") {
				return true
			}
		}
		return false
	case langPHP:
		return inv.Has("composer.json")
	default:
		return false
	}
}

// classifySource maps a path to its language and whether it is a test file. A
// file under a test/tests/spec/__tests__ directory counts as a test for any
// language (catches AVA/tap/node:test/RSpec-style layouts that don't use a
// per-file *.test.* naming convention), in addition to language-specific names.
func classifySource(p string) (lang string, isTest, ok bool) {
	base := pathBase(p)
	inTestDir := hasSegment(p, "test") || hasSegment(p, "tests") || hasSegment(p, "spec") || hasSegment(p, "__tests__")
	switch {
	case strings.HasSuffix(p, ".go"):
		return langGo, strings.HasSuffix(p, "_test.go") || inTestDir, true
	case strings.HasSuffix(base, ".d.ts"):
		return "", false, false // type declarations, not a source surface
	case hasAnySuffix(base, ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"):
		return langJS, jsTestFile(base) || inTestDir, true
	case strings.HasSuffix(p, ".py"):
		t := (strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py")) ||
			strings.HasSuffix(base, "_test.py") || base == "conftest.py" || inTestDir
		return langPython, t, true
	case strings.HasSuffix(p, ".rs"):
		return langRust, inTestDir, true // inline #[test] handled separately
	case hasAnySuffix(base, ".java", ".kt"):
		t := inTestDir || strings.HasSuffix(base, "Test.java") ||
			strings.HasSuffix(base, "Test.kt") || strings.HasSuffix(base, "Spec.kt")
		return langJVM, t, true
	case strings.HasSuffix(p, ".rb"):
		t := strings.HasSuffix(base, "_spec.rb") || strings.HasSuffix(base, "_test.rb") || inTestDir
		return langRuby, t, true
	case strings.HasSuffix(p, ".cs"):
		return langCSharp, strings.HasSuffix(base, "Tests.cs") || strings.HasSuffix(base, "Test.cs") || inTestDir, true
	case strings.HasSuffix(p, ".php"):
		return langPHP, strings.HasSuffix(base, "Test.php") || inTestDir, true
	}
	return "", false, false
}

func jsTestFile(base string) bool {
	for _, ext := range []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"} {
		if strings.HasSuffix(base, ".test"+ext) || strings.HasSuffix(base, ".spec"+ext) {
			return true
		}
	}
	return false
}

// ---- AE-TEST-001 : tests present -------------------------------------------

func checkTestsPresent(root string, inv repository.Inventory, stacks map[string]*stackInfo, active []string) (findings.Finding, bool) {
	var without []string
	for _, lang := range active {
		if stacks[lang].tests > 0 {
			continue
		}
		// Rust unit tests are usually inline (#[cfg(test)]/#[test]), so a zero
		// test-file count does not mean "no tests" — scan for the inline signal.
		if lang == langRust && rustHasInlineTests(root, inv) {
			continue
		}
		without = append(without, lang)
	}
	if len(without) == 0 {
		return findings.Finding{}, false
	}
	evidence := make([]string, 0, len(without))
	for _, lang := range without {
		evidence = append(evidence, "no test files detected for active language: "+lang)
	}
	return findings.Finding{
		RuleID:      "AE-TEST-001",
		Severity:    findings.SeverityHigh,
		Category:    "Testing",
		Summary:     "Repository has no automated tests for an active language — an agent cannot verify its changes",
		Remediation: "Add tests for the active language(s) so an agent can run them and self-verify before finishing a task.",
		Evidence:    evidence,
	}, true
}

func rustHasInlineTests(root string, inv repository.Inventory) bool {
	for _, p := range inv.Paths {
		if !strings.HasSuffix(p, ".rs") || inToolingDir(p) {
			continue
		}
		if c, ok := readRepoFile(root, inv, p); ok && (strings.Contains(c, "#[test]") || strings.Contains(c, "#[cfg(test)]")) {
			return true
		}
	}
	return false
}

// ---- AE-AUTO-001 : verification command discoverable -----------------------

func checkVerificationDiscoverable(root string, inv repository.Inventory, active []string) (findings.Finding, bool) {
	if verificationDiscoverable(root, inv, active) {
		return findings.Finding{}, false
	}
	return findings.Finding{
		RuleID:      "AE-AUTO-001",
		Severity:    findings.SeverityMedium,
		Category:    "Autonomy",
		Summary:     "No discoverable command to run the project's tests — an agent cannot find how to verify the repo",
		Remediation: "Expose a test command via a task runner (Makefile/justfile/Taskfile/package.json scripts/mise/moon) so an agent can discover and run it.",
		Evidence:    []string{"no test target in a recognized task runner and no conventional test toolchain for the active language(s)"},
	}, true
}

func verificationDiscoverable(root string, inv repository.Inventory, active []string) bool {
	for _, lang := range active {
		if conventionalTestToolchain(root, inv, lang) {
			return true
		}
	}
	return runnerHasTestTarget(root, inv)
}

// conventionalTestToolchain reports whether the language has a discoverable
// zero-config test command, so a task runner is not required.
func conventionalTestToolchain(root string, inv repository.Inventory, lang string) bool {
	switch lang {
	case langGo:
		return inv.Has("go.mod") // `go test ./...`
	case langRust:
		return inv.Has("Cargo.toml") // `cargo test`
	case langPython:
		return pytestConfigured(root, inv) // `pytest` discoverable when configured
	default:
		return false
	}
}

func pytestConfigured(root string, inv repository.Inventory) bool {
	if inv.Has("pytest.ini") || inv.Has("tox.ini") {
		return true
	}
	if c, ok := readRepoFile(root, inv, "pyproject.toml"); ok && strings.Contains(c, "[tool.pytest") {
		return true
	}
	if c, ok := readRepoFile(root, inv, "setup.cfg"); ok && strings.Contains(c, "[tool:pytest]") {
		return true
	}
	return false
}

var (
	makeTestTarget = regexp.MustCompile(`(?m)^(test|check)[A-Za-z0-9_-]*[ \t]*:`)
	justRecipe     = regexp.MustCompile(`(?m)^(test|check)[A-Za-z0-9_-]*\s*[: ]`)
	yamlTestKey    = regexp.MustCompile(`(?m)^\s{2,}(test|check)[A-Za-z0-9_-]*\s*:`)
)

func runnerHasTestTarget(root string, inv repository.Inventory) bool {
	if c, ok := readRepoFile(root, inv, "Makefile"); ok && makeTestTarget.MatchString(c) {
		return true
	}
	for _, p := range []string{"justfile", ".justfile"} {
		if c, ok := readRepoFile(root, inv, p); ok && justRecipe.MatchString(c) {
			return true
		}
	}
	for _, p := range []string{"Taskfile.yml", "Taskfile.yaml", "moon.yml"} {
		if c, ok := readRepoFile(root, inv, p); ok && yamlTestKey.MatchString(c) {
			return true
		}
	}
	for _, p := range []string{"mise.toml", ".mise.toml"} {
		if c, ok := readRepoFile(root, inv, p); ok && (strings.Contains(c, "[tasks.test") || strings.Contains(c, "[tasks.check")) {
			return true
		}
	}
	return packageJSONHasTestScript(root, inv)
}

func packageJSONHasTestScript(root string, inv repository.Inventory) bool {
	content, ok := readRepoFile(root, inv, "package.json")
	if !ok {
		return false
	}
	var manifest struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return false
	}
	for name, cmd := range manifest.Scripts {
		if strings.TrimSpace(cmd) == "" {
			continue
		}
		if name == "test" || strings.HasPrefix(name, "test:") {
			return true
		}
	}
	return false
}

// ---- helpers ---------------------------------------------------------------

func pathBase(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// pathDir returns the slash-based directory of p ("" for a root-level file), so
// a //go:embed pattern can be resolved relative to its .go file's directory.
func pathDir(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[:i]
	}
	return ""
}

func hasAnySuffix(s string, suffixes ...string) bool {
	for _, suf := range suffixes {
		if strings.HasSuffix(s, suf) {
			return true
		}
	}
	return false
}

// hasSegment reports whether name is a full path segment of p.
func hasSegment(p, name string) bool {
	for _, seg := range strings.Split(p, "/") {
		if seg == name {
			return true
		}
	}
	return false
}

func readRepoFile(root string, inv repository.Inventory, path string) (string, bool) {
	if !inv.Has(path) {
		return "", false
	}
	// #nosec G304 -- path is a tracked inventory path joined to the resolved root.
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return "", false
	}
	return string(data), true
}
