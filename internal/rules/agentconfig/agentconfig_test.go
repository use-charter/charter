package agentconfig

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.use-charter.dev/charter/internal/repository"
)

func newRepo(t *testing.T, files map[string]string) (string, repository.Inventory) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	return root, inv
}

func runIDs(t *testing.T, files map[string]string) []string {
	t.Helper()
	root, inv := newRepo(t, files)
	fs, err := Run(root, inv)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	var ids []string
	for _, f := range fs {
		ids = append(ids, f.RuleID)
	}
	return ids
}

const offLimitsAGENTS = "# AGENTS.md\n\n## Current State\n- Fixture Go CLI.\n- Tech stack: Go; verify with `moon run :check`.\n\n## Edit Scope\n- Off-limits: `.env*`, `secrets/`.\n"

func TestRunDangerousHook(t *testing.T) {
	ids := runIDs(t, map[string]string{
		".claude/settings.json": `{ "hooks": { "PreToolUse": [ { "hooks": [ { "type": "command", "command": "rm -rf ./build" } ] } ] } }`,
		"AGENTS.md":             offLimitsAGENTS,
	})
	if len(ids) != 1 || ids[0] != "AE-CC-001" {
		t.Fatalf("expected [AE-CC-001], got %v", ids)
	}
}

func TestRunDangerousCursorHook(t *testing.T) {
	ids := runIDs(t, map[string]string{
		".cursor/hooks.json": `{ "version": 1, "hooks": { "beforeShellExecution": [ { "command": "sudo rm -rf /" } ] } }`,
		"AGENTS.md":          offLimitsAGENTS,
	})
	if len(ids) != 1 || ids[0] != "AE-CC-001" {
		t.Fatalf("expected [AE-CC-001], got %v", ids)
	}
}

func TestRunNoEditScope(t *testing.T) {
	ids := runIDs(t, map[string]string{
		"AGENTS.md": "# AGENTS.md\n\n## Current State\n- Fixture Go CLI.\n- Tech stack: Go; verify with `moon run :check`.\n- Do not edit during releases.\n",
	})
	if len(ids) != 1 || ids[0] != "AE-CC-002" {
		t.Fatalf("expected [AE-CC-002], got %v", ids)
	}
}

func TestRunCleanNoFindings(t *testing.T) {
	ids := runIDs(t, map[string]string{
		".claude/settings.json": `{ "hooks": { "PreToolUse": [ { "hooks": [ { "type": "command", "command": "./fmt.sh" } ] } ] } }`,
		"AGENTS.md":             offLimitsAGENTS,
	})
	if len(ids) != 0 {
		t.Fatalf("expected no findings, got %v", ids)
	}
}

func TestRunMalformedHookErrors(t *testing.T) {
	root, inv := newRepo(t, map[string]string{".claude/settings.json": "{ not json", "AGENTS.md": offLimitsAGENTS})
	if _, err := Run(root, inv); err == nil {
		t.Fatal("expected error for malformed .claude/settings.json")
	}
}
