package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRootFromNestedPath(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, ".git")
	nested := filepath.Join(root, "nested", "deeper")

	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("expected .git marker to be created: %v", err)
	}

	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("expected nested path to be created: %v", err)
	}

	resolved, err := ResolveRoot(nested)
	if err != nil {
		t.Fatalf("expected root resolution to succeed: %v", err)
	}

	if resolved != root {
		t.Fatalf("expected %q, got %q", root, resolved)
	}
}
