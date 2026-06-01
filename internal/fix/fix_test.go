package fix

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

func writeFileT(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFileT(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func planRuleIDs(plans []FilePlan) []string {
	ids := make([]string, 0, len(plans))
	for _, p := range plans {
		ids = append(ids, p.RuleID)
	}
	sort.Strings(ids)
	return ids
}

// --- registry / Fixable ----------------------------------------------------

func TestRegistryHoldsExactlyTheV1Fixers(t *testing.T) {
	if len(registry) != 3 {
		t.Fatalf("single-file registry has %d fixers, want 3: %v", len(registry), registry)
	}
	if len(multiRegistry) != 1 {
		t.Fatalf("multi-file registry has %d fixers, want 1: %v", len(multiRegistry), multiRegistry)
	}
	for _, id := range []string{"AE-CTX-001", "AE-CTX-004", "AE-CI-002", "AE-MCP-001"} {
		if !Fixable(id) {
			t.Errorf("Fixable(%q) = false, want true", id)
		}
	}
}

func TestFixableRejectsSecretAndDangerousRules(t *testing.T) {
	for _, id := range []string{"AE-SEC-001", "AE-SEC-002", "AE-CC-001", "AE-CTX-002", "unknown"} {
		if Fixable(id) {
			t.Errorf("Fixable(%q) = true, want false (must never be fixable)", id)
		}
	}
}

// --- Plan -------------------------------------------------------------------

func TestPlanBuildsOnlyForFixableActiveFindings(t *testing.T) {
	root := t.TempDir()
	inv := repository.New(nil)
	result := doctor.Result{Findings: []findings.Finding{
		{RuleID: "AE-CTX-001"},
		{RuleID: "AE-SEC-001"}, // not fixable: must be ignored
		{RuleID: "AE-CI-002"},
		{RuleID: "AE-CC-001"}, // not fixable: must be ignored
	}}

	plans, err := Plan(result, root, inv, Options{})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	got := planRuleIDs(plans)
	want := []string{"AE-CI-002", "AE-CTX-001"}
	if !slicesEqual(got, want) {
		t.Fatalf("Plan rule IDs = %v, want %v", got, want)
	}
}

func TestPlanFiltersByRuleOption(t *testing.T) {
	root := t.TempDir()
	inv := repository.New(nil)
	result := doctor.Result{Findings: []findings.Finding{
		{RuleID: "AE-CTX-001"},
		{RuleID: "AE-CI-002"},
	}}

	plans, err := Plan(result, root, inv, Options{Rule: "AE-CTX-001"})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(plans) != 1 || plans[0].RuleID != "AE-CTX-001" {
		t.Fatalf("filtered Plan = %v, want exactly AE-CTX-001", planRuleIDs(plans))
	}
}

func TestPlanProducesOnePlanPerRule(t *testing.T) {
	root := t.TempDir()
	inv := repository.New(nil)
	result := doctor.Result{Findings: []findings.Finding{
		{RuleID: "AE-CTX-001"},
		{RuleID: "AE-CTX-001"}, // duplicate finding for same rule
	}}

	plans, err := Plan(result, root, inv, Options{})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("got %d plans, want 1 (one plan per rule)", len(plans))
	}
}

func TestPlanSkipsRuleWhenFixerReportsNothingToDo(t *testing.T) {
	root := t.TempDir()
	// AGENTS.md already present -> AE-CTX-001 fixer reports nothing to do.
	inv := repository.New([]string{"AGENTS.md"})
	result := doctor.Result{Findings: []findings.Finding{{RuleID: "AE-CTX-001"}}}

	plans, err := Plan(result, root, inv, Options{})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(plans) != 0 {
		t.Fatalf("got %d plans, want 0 when fixer is satisfied", len(plans))
	}
}

// --- Apply: backup proof ----------------------------------------------------

func TestApplyBacksUpExistingTargetBeforeAppend(t *testing.T) {
	root := t.TempDir()

	original := []byte("# existing ignore\n.charter/\n*.charter-session\n")
	gitignore := filepath.Join(root, ".gitignore")
	writeFileT(t, gitignore, string(original))

	// An UNRELATED sibling file: it must be byte-identical afterwards.
	sibling := filepath.Join(root, "keep.txt")
	siblingBefore := []byte("do not touch me\n")
	writeFileT(t, sibling, string(siblingBefore))

	added := []byte("\n# Charter / agent session artifacts\n.hk/\n.env*\n")
	plan := FilePlan{RuleID: "AE-CTX-004", Path: ".gitignore", Action: Append, Contents: added}

	written, backupDir, err := Apply(root, []FilePlan{plan})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if backupDir == "" {
		t.Fatal("backupDir is empty, want a .charter/backups/<ts> directory")
	}
	if !strings.HasPrefix(filepath.ToSlash(backupDir), ".charter/backups/") {
		t.Errorf("backupDir = %q, want it under .charter/backups/", backupDir)
	}
	if !slicesEqual(written, []string{".gitignore"}) {
		t.Errorf("written = %v, want [.gitignore]", written)
	}

	// (a) Backup holds the ORIGINAL sentinel bytes, untouched.
	backupPath := filepath.Join(root, filepath.FromSlash(backupDir), ".gitignore")
	if got := readFileT(t, backupPath); !bytes.Equal(got, original) {
		t.Errorf("backup contents = %q, want original %q", got, original)
	}

	// (b) Live .gitignore == original + appended.
	wantLive := append(append([]byte{}, original...), added...)
	if got := readFileT(t, gitignore); !bytes.Equal(got, wantLive) {
		t.Errorf("live .gitignore = %q, want %q", got, wantLive)
	}

	// (c) Unrelated sibling is byte-identical (never deleted/truncated/touched).
	if got := readFileT(t, sibling); !bytes.Equal(got, siblingBefore) {
		t.Errorf("sibling keep.txt = %q, want unchanged %q", got, siblingBefore)
	}
}

// --- Apply: Create semantics ------------------------------------------------

func TestApplyCreateWritesAbsentPathAndReportsNoBackup(t *testing.T) {
	root := t.TempDir()
	plan := FilePlan{RuleID: "AE-CI-002", Path: ".github/workflows/charter.yaml", Action: Create, Contents: []byte("name: Charter\n")}

	written, backupDir, err := Apply(root, []FilePlan{plan})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if backupDir != "" {
		t.Errorf("backupDir = %q, want empty (nothing pre-existed)", backupDir)
	}
	if !slicesEqual(written, []string{".github/workflows/charter.yaml"}) {
		t.Errorf("written = %v", written)
	}
	got := readFileT(t, filepath.Join(root, ".github", "workflows", "charter.yaml"))
	if string(got) != "name: Charter\n" {
		t.Errorf("created file = %q", got)
	}
}

func TestApplyCreateNeverOverwritesExistingPath(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "AGENTS.md")
	writeFileT(t, existing, "KEEP ME\n")

	plan := FilePlan{RuleID: "AE-CTX-001", Path: "AGENTS.md", Action: Create, Contents: []byte("OVERWRITE\n")}
	written, _, err := Apply(root, []FilePlan{plan})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if got := readFileT(t, existing); string(got) != "KEEP ME\n" {
		t.Errorf("AGENTS.md = %q, want it left untouched (KEEP ME)", got)
	}
	if len(written) != 0 {
		t.Errorf("written = %v, want empty (create skipped on existing path)", written)
	}
}

// --- Apply: safety guards ---------------------------------------------------

func TestApplyRefusesToWriteOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "escape.md")
	_ = os.Remove(outside) // ensure clean slate

	plan := FilePlan{RuleID: "AE-CTX-001", Path: "../escape.md", Action: Create, Contents: []byte("x")}
	_, _, err := Apply(root, []FilePlan{plan})
	if err == nil {
		t.Fatal("Apply accepted a path outside root, want error")
	}
	if _, statErr := os.Stat(outside); statErr == nil {
		t.Errorf("escape file was created at %s, want it absent", outside)
		_ = os.Remove(outside)
	}
}

func TestApplyRefusesUnregisteredRule(t *testing.T) {
	root := t.TempDir()
	plan := FilePlan{RuleID: "AE-SEC-001", Path: "leak.txt", Action: Create, Contents: []byte("x")}
	_, _, err := Apply(root, []FilePlan{plan})
	if err == nil {
		t.Fatal("Apply accepted an unregistered (secret) rule, want error")
	}
	if _, statErr := os.Stat(filepath.Join(root, "leak.txt")); statErr == nil {
		t.Error("unregistered-rule plan wrote a file, want none")
	}
}

func TestApplyAppendToAbsentFileDegradesToCreateWithoutBackup(t *testing.T) {
	root := t.TempDir()
	plan := FilePlan{RuleID: "AE-CTX-004", Path: ".gitignore", Action: Append, Contents: []byte("body\n")}

	written, backupDir, err := Apply(root, []FilePlan{plan})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if backupDir != "" {
		t.Errorf("backupDir = %q, want empty (nothing to back up)", backupDir)
	}
	if got := readFileT(t, filepath.Join(root, ".gitignore")); string(got) != "body\n" {
		t.Errorf(".gitignore = %q, want body", got)
	}
	if !slicesEqual(written, []string{".gitignore"}) {
		t.Errorf("written = %v", written)
	}
}

// --- diff golden ------------------------------------------------------------

func TestBuildCreateDiffGolden(t *testing.T) {
	got := buildCreateDiff("AGENTS.md", []byte("# Title\nbody line\n"))
	want := "--- /dev/null\n" +
		"+++ b/AGENTS.md\n" +
		"@@ -0,0 +1,2 @@\n" +
		"+# Title\n" +
		"+body line\n"
	if got != want {
		t.Errorf("buildCreateDiff mismatch:\n got=%q\nwant=%q", got, want)
	}
}

func TestBuildAppendDiffGolden(t *testing.T) {
	existing := []byte("a\nb\nc\nd\n")
	added := []byte("\n# block\nx\n")
	got := buildAppendDiff(".gitignore", existing, added)
	want := "--- a/.gitignore\n" +
		"+++ b/.gitignore\n" +
		"@@ -2,3 +2,6 @@\n" +
		" b\n" +
		" c\n" +
		" d\n" +
		"+\n" +
		"+# block\n" +
		"+x\n"
	if got != want {
		t.Errorf("buildAppendDiff mismatch:\n got=%q\nwant=%q", got, want)
	}
}

func TestSplitLinesHandlesTrailingNewlineAndEmpty(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a\n", []string{"a"}},
		{"a\nb", []string{"a", "b"}},
		{"a\n\n", []string{"a", ""}},
	}
	for _, c := range cases {
		got := splitLines([]byte(c.in))
		if !slicesEqual(got, c.want) {
			t.Errorf("splitLines(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// --- fixCTX001 --------------------------------------------------------------

func TestFixCTX001CreatesAGENTSWhenNoContextExists(t *testing.T) {
	root := t.TempDir()
	writeFileT(t, filepath.Join(root, "go.mod"), "module example.com/x\n\ngo 1.26\n")
	inv := repository.New([]string{"go.mod"})

	plan, ok, err := fixCTX001(root, inv)
	if err != nil {
		t.Fatalf("fixCTX001: %v", err)
	}
	if !ok {
		t.Fatal("fixCTX001 ok = false, want a plan when no context exists")
	}
	if plan.Path != "AGENTS.md" || plan.Action != Create {
		t.Errorf("plan = {Path:%q Action:%v}, want {AGENTS.md Create}", plan.Path, plan.Action)
	}
	for _, want := range []string{"# AGENTS.md", "charter doctor"} {
		if !strings.Contains(string(plan.Contents), want) {
			t.Errorf("AGENTS.md contents missing %q", want)
		}
	}
	if !strings.HasPrefix(plan.Diff, "--- /dev/null\n+++ b/AGENTS.md\n") {
		t.Errorf("diff preview = %q, want a create diff header", plan.Diff)
	}
}

func TestFixCTX001SkipsWhenSingleFileContextExists(t *testing.T) {
	root := t.TempDir()
	inv := repository.New([]string{"CLAUDE.md"})

	if _, ok, err := fixCTX001(root, inv); err != nil || ok {
		t.Fatalf("fixCTX001 with CLAUDE.md present = (ok=%v, err=%v), want (false, nil)", ok, err)
	}
}

func TestFixCTX001SkipsWhenCursorRulesExist(t *testing.T) {
	root := t.TempDir()
	inv := repository.New([]string{".cursor/rules/base.mdc"})

	if _, ok, err := fixCTX001(root, inv); err != nil || ok {
		t.Fatalf("fixCTX001 with .cursor/rules present = (ok=%v, err=%v), want (false, nil)", ok, err)
	}
}

// --- fixCTX004 --------------------------------------------------------------

func TestFixCTX004AppendsMissingPatterns(t *testing.T) {
	root := t.TempDir()
	// Present: all but .hk/ and .env* (each on its own line).
	gi := "# project ignores\n.charter/\n*.charter-session\n.claude/local/\n.cursor/cache/\n"
	writeFileT(t, filepath.Join(root, ".gitignore"), gi)
	inv := repository.New([]string{".gitignore"})

	plan, ok, err := fixCTX004(root, inv)
	if err != nil {
		t.Fatalf("fixCTX004: %v", err)
	}
	if !ok {
		t.Fatal("fixCTX004 ok = false, want an append plan")
	}
	if plan.Action != Append {
		t.Errorf("Action = %v, want Append", plan.Action)
	}
	wantContents := "\n# Charter / agent session artifacts\n.hk/\n.env*\n"
	if string(plan.Contents) != wantContents {
		t.Errorf("append Contents = %q, want %q", plan.Contents, wantContents)
	}
	// Exactly the missing patterns, and nothing already-present.
	for _, missing := range []string{".hk/", ".env*"} {
		if !strings.Contains(string(plan.Contents), missing) {
			t.Errorf("append Contents missing %q", missing)
		}
	}
	if strings.Contains(string(plan.Contents), ".charter/") {
		t.Errorf("append Contents re-added an already-present pattern: %q", plan.Contents)
	}
}

func TestFixCTX004NoOpWhenAllPatternsPresent(t *testing.T) {
	root := t.TempDir()
	var gi strings.Builder
	gi.WriteString("# complete\n")
	for _, pat := range ctx004Required {
		gi.WriteString(pat)
		gi.WriteByte('\n')
	}
	writeFileT(t, filepath.Join(root, ".gitignore"), gi.String())
	inv := repository.New([]string{".gitignore"})

	if _, ok, err := fixCTX004(root, inv); err != nil || ok {
		t.Fatalf("fixCTX004 on complete .gitignore = (ok=%v, err=%v), want (false, nil)", ok, err)
	}
}

func TestFixCTX004CreatesWhenGitignoreAbsent(t *testing.T) {
	root := t.TempDir()
	inv := repository.New(nil)

	plan, ok, err := fixCTX004(root, inv)
	if err != nil {
		t.Fatalf("fixCTX004: %v", err)
	}
	if !ok {
		t.Fatal("fixCTX004 ok = false, want a create plan when .gitignore is absent")
	}
	if plan.Path != ".gitignore" || plan.Action != Create {
		t.Errorf("plan = {Path:%q Action:%v}, want {.gitignore Create}", plan.Path, plan.Action)
	}
	for _, pat := range ctx004Required {
		if !strings.Contains(string(plan.Contents), pat) {
			t.Errorf("created .gitignore missing required pattern %q", pat)
		}
	}
}

// --- fixCI002 ---------------------------------------------------------------

func TestFixCI002CreatesWorkflowWhenGateAbsent(t *testing.T) {
	root := t.TempDir()
	inv := repository.New(nil)

	plan, ok, err := fixCI002(root, inv)
	if err != nil {
		t.Fatalf("fixCI002: %v", err)
	}
	if !ok {
		t.Fatal("fixCI002 ok = false, want a create plan when no gate exists")
	}
	if plan.Path != ".github/workflows/charter.yaml" || plan.Action != Create {
		t.Errorf("plan = {Path:%q Action:%v}", plan.Path, plan.Action)
	}
	for _, want := range []string{
		"name: Charter",
		"use-charter/charter-action@v1",
		"actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2",
		`threshold: "80"`,
		"security-events: write",
	} {
		if !strings.Contains(string(plan.Contents), want) {
			t.Errorf("charter.yaml missing %q", want)
		}
	}
}

func TestFixCI002SkipsWhenCharterDoctorGatePresent(t *testing.T) {
	root := t.TempDir()
	writeFileT(t, filepath.Join(root, ".github", "workflows", "ci.yml"), "jobs:\n  x:\n    steps:\n      - run: charter doctor --path .\n")
	inv := repository.New([]string{".github/workflows/ci.yml"})

	if _, ok, err := fixCI002(root, inv); err != nil || ok {
		t.Fatalf("fixCI002 with charter doctor present = (ok=%v, err=%v), want (false, nil)", ok, err)
	}
}

func TestFixCI002SkipsWhenCharterActionPresent(t *testing.T) {
	root := t.TempDir()
	writeFileT(t, filepath.Join(root, ".github", "workflows", "gate.yaml"), "jobs:\n  x:\n    steps:\n      - uses: use-charter/charter-action@v1\n")
	inv := repository.New([]string{".github/workflows/gate.yaml"})

	if _, ok, err := fixCI002(root, inv); err != nil || ok {
		t.Fatalf("fixCI002 with charter-action present = (ok=%v, err=%v), want (false, nil)", ok, err)
	}
}

// --- end-to-end Plan -> Apply ----------------------------------------------

func TestPlanThenApplyCreatesFixableFiles(t *testing.T) {
	root := t.TempDir()
	writeFileT(t, filepath.Join(root, "go.mod"), "module example.com/x\n\ngo 1.26\n")
	inv := repository.New([]string{"go.mod"})
	result := doctor.Result{Findings: []findings.Finding{
		{RuleID: "AE-CTX-001"},
		{RuleID: "AE-CI-002"},
	}}

	plans, err := Plan(result, root, inv, Options{})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	written, backupDir, err := Apply(root, plans)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if backupDir != "" {
		t.Errorf("backupDir = %q, want empty (all creates on absent paths)", backupDir)
	}

	sort.Strings(written)
	want := []string{".github/workflows/charter.yaml", "AGENTS.md"}
	sort.Strings(want)
	if !slicesEqual(written, want) {
		t.Fatalf("written = %v, want %v", written, want)
	}
	if _, statErr := os.Stat(filepath.Join(root, "AGENTS.md")); statErr != nil {
		t.Errorf("AGENTS.md not written: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".github", "workflows", "charter.yaml")); statErr != nil {
		t.Errorf("charter.yaml not written: %v", statErr)
	}
}

// --- fixMCP001 (in-place version bumps) -------------------------------------

func TestFixMCP001BumpsAdvisoryPinToFixedVersion(t *testing.T) {
	root := t.TempDir()
	cfg := `{ "mcpServers": { "git": { "command": "uvx", "args": ["mcp-server-git@2025.8.0"] } } }` + "\n"
	writeFileT(t, filepath.Join(root, ".mcp.json"), cfg)
	inv := repository.New([]string{".mcp.json"})

	plans, err := fixMCP001(root, inv)
	if err != nil {
		t.Fatalf("fixMCP001: %v", err)
	}
	if len(plans) != 1 || plans[0].Action != Replace || plans[0].Path != ".mcp.json" {
		t.Fatalf("plans = %+v, want one Replace of .mcp.json", plans)
	}
	if !strings.Contains(string(plans[0].Contents), "mcp-server-git@2026.1.14") {
		t.Errorf("bumped contents = %q, want mcp-server-git@2026.1.14", plans[0].Contents)
	}
	if strings.Contains(string(plans[0].Contents), "2025.8.0") {
		t.Errorf("vulnerable version still present: %q", plans[0].Contents)
	}
	if !strings.Contains(plans[0].Diff, "-") || !strings.Contains(plans[0].Diff, "+") {
		t.Errorf("diff should show the change: %q", plans[0].Diff)
	}
}

func TestFixMCP001PinsUnpinnedCatalogedPackage(t *testing.T) {
	root := t.TempDir()
	cfg := `{ "mcpServers": { "fs": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem@latest"] } } }` + "\n"
	writeFileT(t, filepath.Join(root, ".mcp.json"), cfg)
	inv := repository.New([]string{".mcp.json"})

	plans, err := fixMCP001(root, inv)
	if err != nil {
		t.Fatalf("fixMCP001: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("plans = %+v, want one", plans)
	}
	if !strings.Contains(string(plans[0].Contents), "@modelcontextprotocol/server-filesystem@2026.1.14") {
		t.Errorf("expected pin to catalog stable, got %q", plans[0].Contents)
	}
}

func TestFixMCP001NeverRewritesDeprecatedPackage(t *testing.T) {
	root := t.TempDir()
	cfg := `{ "mcpServers": { "gh": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-github@1.2.3"] } } }` + "\n"
	writeFileT(t, filepath.Join(root, ".mcp.json"), cfg)
	inv := repository.New([]string{".mcp.json"})

	plans, err := fixMCP001(root, inv)
	if err != nil {
		t.Fatalf("fixMCP001: %v", err)
	}
	if len(plans) != 0 {
		t.Fatalf("deprecated package must not be auto-rewritten; got %+v", plans)
	}
}

func TestFixMCP001ApplyBacksUpAndRewritesInPlace(t *testing.T) {
	root := t.TempDir()
	cfg := `{ "mcpServers": { "git": { "command": "uvx", "args": ["mcp-server-git@2025.8.0"] } } }` + "\n"
	target := filepath.Join(root, ".mcp.json")
	writeFileT(t, target, cfg)
	inv := repository.New([]string{".mcp.json"})

	plans, err := fixMCP001(root, inv)
	if err != nil {
		t.Fatalf("fixMCP001: %v", err)
	}
	written, backupDir, err := Apply(root, plans)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !slicesEqual(written, []string{".mcp.json"}) || backupDir == "" {
		t.Fatalf("written=%v backupDir=%q, want [.mcp.json] + a backup", written, backupDir)
	}
	if got := string(readFileT(t, target)); !strings.Contains(got, "2026.1.14") || strings.Contains(got, "2025.8.0") {
		t.Errorf("live .mcp.json not bumped: %q", got)
	}
	// Backup preserves the original vulnerable pin.
	backup := filepath.Join(root, filepath.FromSlash(backupDir), ".mcp.json")
	if got := string(readFileT(t, backup)); !strings.Contains(got, "2025.8.0") {
		t.Errorf("backup should hold the original: %q", got)
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
