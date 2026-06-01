package catalog

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLookup(t *testing.T) {
	e, ok := Lookup("AE-SEC-001")
	if !ok {
		t.Fatal("expected AE-SEC-001 in the catalog")
	}
	if e.Category != "Secrets" || e.HelpURI != "https://use-charter.dev/rules/AE-SEC-001" || e.ShortDescription == "" {
		t.Fatalf("unexpected entry: %+v", e)
	}
	if _, ok := Lookup("AE-NOPE-999"); ok {
		t.Fatal("did not expect an unknown rule")
	}
}

// TestCatalogMatchesSpecs is the drift guard: the catalog's rule IDs must exactly
// match the behavioral spec files under docs/internal/specs/.
func TestCatalogMatchesSpecs(t *testing.T) {
	specDir := filepath.Join("..", "..", "..", "docs", "internal", "specs")
	matches, err := filepath.Glob(filepath.Join(specDir, "AE-*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatalf("no spec files found under %s", specDir)
	}

	specIDs := map[string]bool{}
	for _, m := range matches {
		specIDs[strings.TrimSuffix(filepath.Base(m), ".md")] = true
	}
	catIDs := map[string]bool{}
	for _, id := range IDs() {
		catIDs[id] = true
	}

	for id := range specIDs {
		if !catIDs[id] {
			t.Errorf("spec %s has no catalog entry", id)
		}
	}
	for id := range catIDs {
		if !specIDs[id] {
			t.Errorf("catalog entry %s has no spec file", id)
		}
	}
}
