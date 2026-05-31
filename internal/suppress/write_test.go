package suppress

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpsertFileAppendsAndReplaces(t *testing.T) {
	root := t.TempDir()
	out, err := UpsertFile(root, FileEntry{Rule: "AE-MCP-001", Reason: "first", Expires: "2099-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, File), out, 0o644); err != nil {
		t.Fatal(err)
	}

	// Replace the same rule rather than duplicate it.
	out, err = UpsertFile(root, FileEntry{Rule: "AE-MCP-001", Reason: "second", Expires: "2099-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := parseSuppressFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Reason != "second" {
		t.Fatalf("expected one replaced entry, got %+v", entries)
	}
	if err := os.WriteFile(filepath.Join(root, File), out, 0o644); err != nil {
		t.Fatal(err)
	}

	// A different rule appends.
	out, err = UpsertFile(root, FileEntry{Rule: "AE-CC-002", Reason: "x", Expires: "2099-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "AE-CC-002") {
		t.Fatalf("expected appended entry, got:\n%s", out)
	}
	entries, err = parseSuppressFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected two entries after append, got %d", len(entries))
	}
}

func TestUpsertFileMissingIsEmpty(t *testing.T) {
	out, err := UpsertFile(t.TempDir(), FileEntry{Rule: "AE-MCP-001", Reason: "x", Expires: "2099-01-01"})
	if err != nil {
		t.Fatalf("missing file should be treated as empty: %v", err)
	}
	if !strings.Contains(string(out), "AE-MCP-001") {
		t.Fatalf("expected the new entry, got:\n%s", out)
	}
}
