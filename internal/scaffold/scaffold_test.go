package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// estTokens mirrors the estimate the AE-CTX-001 rule uses so the template
// budget assertion tracks the gate exactly.
func estTokens(s string) int {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0
	}
	return (len(trimmed) + 3) / 4
}

func countNonEmptyLines(s string) int {
	n := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestDetectGoProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/x\n\ngo 1.26\n")
	writeFile(t, filepath.Join(root, ".github", "workflows", "ci.yml"), "name: ci\n")
	writeFile(t, filepath.Join(root, ".claude", "settings.json"), "{}\n")

	p, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	if p.Name != filepath.Base(root) {
		t.Errorf("Name = %q, want %q", p.Name, filepath.Base(root))
	}
	if len(p.Langs) != 1 {
		t.Fatalf("Langs = %+v, want exactly one (Go)", p.Langs)
	}
	if p.Langs[0].Name != "Go" || p.Langs[0].Version != "1.26" {
		t.Errorf("Langs[0] = %+v, want {Go 1.26}", p.Langs[0])
	}
	if p.CI != "GitHub Actions" {
		t.Errorf("CI = %q, want %q", p.CI, "GitHub Actions")
	}
	if !slices.Contains(p.Agents, "claude") {
		t.Errorf("Agents = %v, want it to contain claude", p.Agents)
	}
}

func TestDetectGoToolchainFallback(t *testing.T) {
	root := t.TempDir()
	// No `go` directive: version must fall back to the toolchain line.
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/x\n\ntoolchain go1.26.3\n")

	p, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(p.Langs) != 1 || p.Langs[0].Name != "Go" || p.Langs[0].Version != "1.26.3" {
		t.Fatalf("Langs = %+v, want {Go 1.26.3}", p.Langs)
	}
}

func TestDetectMultiLanguage(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/x\ngo 1.26\ntoolchain go1.26.3\n")
	writeFile(t, filepath.Join(root, "package.json"), `{"name":"x","engines":{"node":">=20"}}`)
	writeFile(t, filepath.Join(root, "pyproject.toml"), "[project]\nrequires-python = \">=3.11\"\n")
	writeFile(t, filepath.Join(root, "Cargo.toml"), "[package]\nname = \"x\"\n")
	writeFile(t, filepath.Join(root, ".cursor", "rules", "a.mdc"), "x\n")
	writeFile(t, filepath.Join(root, ".github", "copilot-instructions.md"), "x\n")

	p, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	want := []Lang{
		{"Go", "1.26"},
		{"JavaScript/TypeScript", ">=20"},
		{"Python", ">=3.11"},
		{"Rust", ""},
	}
	if !slices.Equal(p.Langs, want) {
		t.Fatalf("Langs = %+v, want %+v (stable order)", p.Langs, want)
	}
	if !slices.Contains(p.Agents, "cursor") {
		t.Errorf("Agents = %v, want cursor", p.Agents)
	}
	if !slices.Contains(p.Agents, "copilot") {
		t.Errorf("Agents = %v, want copilot", p.Agents)
	}
}

func TestDetectEmptyDir(t *testing.T) {
	root := t.TempDir()

	p, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(p.Langs) != 0 {
		t.Errorf("Langs = %+v, want none", p.Langs)
	}
	if p.CI != "" {
		t.Errorf("CI = %q, want empty", p.CI)
	}
	if len(p.Agents) != 0 {
		t.Errorf("Agents = %v, want none", p.Agents)
	}
}

func TestDetectPackageJSONWithoutEngines(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"name":"x"}`)
	// Malformed pyproject should not crash detection; version best-effort empty.
	writeFile(t, filepath.Join(root, "pyproject.toml"), "name = no-version-here\n")

	p, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	want := []Lang{
		{"JavaScript/TypeScript", ""},
		{"Python", ""},
	}
	if !slices.Equal(p.Langs, want) {
		t.Fatalf("Langs = %+v, want %+v", p.Langs, want)
	}
}

func TestDetectCIIgnoresNonYAML(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "workflows", "README.txt"), "not a workflow\n")

	p, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if p.CI != "" {
		t.Errorf("CI = %q, want empty (no yaml workflow)", p.CI)
	}
}

func TestAGENTSMarkdownSatisfiesGates(t *testing.T) {
	p := Project{
		Name:  "acme",
		Langs: []Lang{{"Go", "1.26"}},
		CI:    "GitHub Actions",
	}
	out := string(AGENTSMarkdown(p))

	for _, want := range []string{
		"charter doctor",
		".env*",
		"secrets/",
		".github/workflows/",
		"Off-limits",
		"Safe for agents",
		"Go", // primary language / tech-stack mention
	} {
		if !strings.Contains(out, want) {
			t.Errorf("AGENTS.md missing required token %q\n---\n%s", want, out)
		}
	}

	if got := countNonEmptyLines(out); got < 5 {
		t.Errorf("AGENTS.md has %d non-empty lines, want >= 5", got)
	}

	tokens := estTokens(out)
	if tokens >= 600 {
		t.Errorf("AGENTS.md est tokens = %d, want < 600", tokens)
	}
	t.Logf("AGENTS.md est tokens: %d (non-empty lines: %d)", tokens, countNonEmptyLines(out))
}

func TestAGENTSMarkdownPerPrimaryLanguage(t *testing.T) {
	cases := []Project{
		{Name: "go-repo", Langs: []Lang{{"Go", "1.26"}}},
		{Name: "js-repo", Langs: []Lang{{"JavaScript/TypeScript", ">=20"}}},
		{Name: "py-repo", Langs: []Lang{{"Python", ">=3.11"}}},
		{Name: "rust-repo", Langs: []Lang{{"Rust", ""}}},
		{Name: "blank-repo"}, // no languages -> generic
	}
	for _, p := range cases {
		out := string(AGENTSMarkdown(p))
		for _, want := range []string{"charter doctor", ".env*", "secrets/", ".github/workflows/", "Off-limits", "Safe for agents", "## Tech Stack"} {
			if !strings.Contains(out, want) {
				t.Errorf("[%s] AGENTS.md missing %q", p.Name, want)
			}
		}
		if got := countNonEmptyLines(out); got < 5 {
			t.Errorf("[%s] non-empty lines = %d, want >= 5", p.Name, got)
		}
		if tokens := estTokens(out); tokens >= 600 {
			t.Errorf("[%s] est tokens = %d, want < 600", p.Name, tokens)
		}
	}

	// The generic (no-language) overview must still read as prose.
	blank := string(AGENTSMarkdown(Project{Name: "blank-repo"}))
	if !strings.Contains(blank, "Primary language: (edit)") {
		t.Errorf("blank project AGENTS.md missing the generic tech-stack line\n%s", blank)
	}
}

func TestClaudeSettingsIsValidMCPFreeJSON(t *testing.T) {
	raw := ClaudeSettings()

	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("ClaudeSettings is not valid JSON: %v", err)
	}

	s := string(raw)
	for _, deny := range []string{"Read(./.env)", "Read(./.env.*)", "Read(./secrets/**)"} {
		if !strings.Contains(s, deny) {
			t.Errorf("ClaudeSettings missing deny entry %q", deny)
		}
	}
	if strings.Contains(strings.ToLower(s), "mcp") {
		t.Errorf("ClaudeSettings must be MCP-free, got:\n%s", s)
	}
}

func TestCharterYAML(t *testing.T) {
	out := string(CharterYAML("strict"))
	want := "# yaml-language-server: $schema=" + SchemaURL + "\npolicy:\n  profile: strict\n"
	if out != want {
		t.Fatalf("CharterYAML(strict) = %q", out)
	}
	if !strings.Contains(out, "profile: strict") {
		t.Errorf("CharterYAML missing profile line")
	}
}

func TestGitignoreAndEnvExampleAndArchitecture(t *testing.T) {
	gi := string(Gitignore())
	for _, want := range []string{".charter/", "*.charter-session", ".claude/local/", ".cursor/cache/"} {
		if !strings.Contains(gi, want) {
			t.Errorf("Gitignore missing %q", want)
		}
	}

	env := string(EnvExample())
	if !strings.Contains(env, "#") {
		t.Errorf("EnvExample missing a header comment")
	}
	if !strings.Contains(env, "EXAMPLE_VAR") {
		t.Errorf("EnvExample missing the example var line")
	}

	arch := string(ArchitectureMarkdown(Project{Name: "acme"}))
	for _, want := range []string{"# acme", "## Overview", "## Layout"} {
		if !strings.Contains(arch, want) {
			t.Errorf("ArchitectureMarkdown missing %q", want)
		}
	}
}

func TestPlanCreatesAllInStableOrderWhenAbsent(t *testing.T) {
	p := Project{Name: "acme", Langs: []Lang{{"Go", "1.26"}}}
	actions := Plan(p, Options{Profile: "standard", Agents: []string{"claude"}}, func(string) bool { return false })

	wantOrder := []string{
		"AGENTS.md",
		"charter.yaml",
		".gitignore",
		"ARCHITECTURE.md",
		".env.example",
		".claude/settings.json",
	}
	if len(actions) != len(wantOrder) {
		t.Fatalf("got %d actions, want %d: %+v", len(actions), len(wantOrder), actions)
	}
	for i, fa := range actions {
		if fa.Path != wantOrder[i] {
			t.Errorf("action[%d].Path = %q, want %q", i, fa.Path, wantOrder[i])
		}
		if fa.Action != Create {
			t.Errorf("action[%d] (%s) = %v, want Create", i, fa.Path, fa.Action)
		}
		if len(fa.Contents) == 0 {
			t.Errorf("action[%d] (%s) has empty Contents on Create", i, fa.Path)
		}
	}
}

func TestPlanSkipsExisting(t *testing.T) {
	p := Project{Name: "acme"}
	exists := func(rel string) bool { return rel == "AGENTS.md" }
	actions := Plan(p, Options{Profile: "standard", Agents: []string{"claude"}}, exists)

	var agents FileAction
	found := false
	for _, fa := range actions {
		if fa.Path == "AGENTS.md" {
			agents = fa
			found = true
		}
	}
	if !found {
		t.Fatal("AGENTS.md not present in plan")
	}
	if agents.Action != Skip {
		t.Errorf("AGENTS.md Action = %v, want Skip", agents.Action)
	}
	if len(agents.Contents) != 0 {
		t.Errorf("Skip action should carry no Contents, got %d bytes", len(agents.Contents))
	}
}

func TestPlanClaudeSettingsGatedOnAgent(t *testing.T) {
	p := Project{Name: "acme"}

	with := Plan(p, Options{Profile: "standard", Agents: []string{"claude", "cursor"}}, func(string) bool { return false })
	if !hasPath(with, ".claude/settings.json") {
		t.Errorf("expected .claude/settings.json when claude requested")
	}

	without := Plan(p, Options{Profile: "standard", Agents: []string{"cursor"}}, func(string) bool { return false })
	if hasPath(without, ".claude/settings.json") {
		t.Errorf("did not expect .claude/settings.json without claude")
	}
	if len(without) != 5 {
		t.Errorf("non-claude plan = %d actions, want 5", len(without))
	}
}

func hasPath(actions []FileAction, path string) bool {
	for _, fa := range actions {
		if fa.Path == path {
			return true
		}
	}
	return false
}
