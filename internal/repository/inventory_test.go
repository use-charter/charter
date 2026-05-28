package repository

import (
	"path/filepath"
	"testing"
)

func TestBuildInventoryRespectsGitignore(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "pass-basic")

	inv, err := BuildInventory(root)
	if err != nil {
		t.Fatalf("expected inventory build to succeed: %v", err)
	}

	if inv.Has(".charter/local.txt") {
		t.Fatalf("expected ignored local artifact to be excluded")
	}

	if !inv.Has("AGENTS.md") {
		t.Fatalf("expected AGENTS.md to be present in inventory")
	}
}
