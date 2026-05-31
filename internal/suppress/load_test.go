package suppress

import (
	"os"
	"path/filepath"
	"testing"

	"go.charter.dev/charter/internal/repository"
)

func writeFile(t *testing.T, path, content string) error {
	t.Helper()
	return os.WriteFile(path, []byte(content), 0o644)
}

func TestParseSuppressFile(t *testing.T) {
	data := []byte(`suppressions:
  - rule: AE-MCP-001
    reason: "vendored fixture"
    expires: 2099-01-01
  - rule: AE-CC-002
    approver: security-team
    path: AGENTS.md
`)
	entries, err := parseSuppressFile(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Rule != "AE-MCP-001" || entries[0].Reason != "vendored fixture" || entries[0].Expires != "2099-01-01" {
		t.Fatalf("entry 0 wrong: %+v", entries[0])
	}
	if entries[0].Source != SourceExternal {
		t.Fatalf("expected external source, got %q", entries[0].Source)
	}
	if entries[1].Path != "AGENTS.md" || entries[1].Approver != "security-team" {
		t.Fatalf("entry 1 wrong: %+v", entries[1])
	}
}

func TestParseSuppressFileMalformed(t *testing.T) {
	if _, err := parseSuppressFile([]byte("suppressions: [oops")); err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestParseSuppressFileBlankRuleFailsFast(t *testing.T) {
	if _, err := parseSuppressFile([]byte("suppressions:\n  - reason: x\n")); err == nil {
		t.Fatal("expected error for an entry missing the rule field")
	}
}

func TestLoadFileMissing(t *testing.T) {
	inv := repository.New(nil)
	entries, err := LoadFile(t.TempDir(), inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Fatalf("expected nil entries for missing file, got %+v", entries)
	}
}

func TestLoadFilePresent(t *testing.T) {
	root := t.TempDir()
	if err := writeFile(t, filepath.Join(root, File), "suppressions:\n  - rule: AE-MCP-001\n    reason: x\n"); err != nil {
		t.Fatal(err)
	}
	inv := repository.New([]string{File})
	entries, err := LoadFile(root, inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Rule != "AE-MCP-001" {
		t.Fatalf("got %+v", entries)
	}
}
