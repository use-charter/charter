package repository

import (
	"path/filepath"
	"testing"
)

func TestResolveRootFromNestedPath(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "pass-basic")
	nested := filepath.Join(root, "nested", "deeper")

	resolved, err := ResolveRoot(nested)
	if err != nil {
		t.Fatalf("expected root resolution to succeed: %v", err)
	}

	expected, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("expected fixture path to resolve: %v", err)
	}

	if resolved != expected {
		t.Fatalf("expected %q, got %q", expected, resolved)
	}
}
