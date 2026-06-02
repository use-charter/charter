package tui

import (
	"testing"

	"go.use-charter.dev/charter/internal/findings"
)

func TestBuildItemsClassifiesKinds(t *testing.T) {
	items := buildItems(sampleResult())
	if len(items) != 5 {
		t.Fatalf("buildItems = %d, want 5", len(items))
	}
	kinds := map[string]itemKind{}
	for _, it := range items {
		kinds[it.finding.RuleID] = it.kind
	}
	if kinds["AE-SEC-001"] != kindActive {
		t.Fatalf("AE-SEC-001 should be active")
	}
	if kinds["AE-CTX-006"] != kindInformational {
		t.Fatalf("AE-CTX-006 should be informational")
	}
	if kinds["AE-MCP-001"] != kindSuppressed {
		t.Fatalf("AE-MCP-001 should be suppressed")
	}
}

func TestUniqueCategoriesSorted(t *testing.T) {
	got := uniqueCategories(buildItems(sampleResult()))
	want := []string{"CI", "Context", "MCP Safety", "Secrets"}
	if len(got) != len(want) {
		t.Fatalf("categories = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("categories[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFilterItemsMutedExcludedByDefault(t *testing.T) {
	items := buildItems(sampleResult())
	visible := filterItems(items, "", "", false, "")
	if len(visible) != 3 {
		t.Fatalf("un-muted filter = %d, want 3", len(visible))
	}
	withMuted := filterItems(items, "", "", true, "")
	if len(withMuted) != 5 {
		t.Fatalf("muted-included filter = %d, want 5", len(withMuted))
	}
}

func TestFilterItemsSeverityAndQuery(t *testing.T) {
	items := buildItems(sampleResult())
	if got := filterItems(items, findings.SeverityBlocker, "", false, ""); len(got) != 1 {
		t.Fatalf("blocker filter = %d, want 1", len(got))
	}
	if got := filterItems(items, "", "", false, "weak"); len(got) != 1 {
		t.Fatalf("query 'weak' (un-muted) = %d, want 1", len(got))
	}
	// "secret" appears in the AE-SEC-001 summary + evidence.
	if got := filterItems(items, "", "", false, "secret"); len(got) != 1 {
		t.Fatalf("query 'secret' = %d, want 1", len(got))
	}
}

func TestSortItemsModes(t *testing.T) {
	items := filterItems(buildItems(sampleResult()), "", "", false, "")

	sortItems(items, sortBySeverity)
	if items[0].finding.RuleID != "AE-SEC-001" {
		t.Fatalf("severity sort first = %q, want AE-SEC-001", items[0].finding.RuleID)
	}

	sortItems(items, sortByCategory)
	if items[0].finding.RuleID != "AE-CI-002" {
		t.Fatalf("category sort first = %q, want AE-CI-002", items[0].finding.RuleID)
	}
}

func TestFirstLocation(t *testing.T) {
	cases := []struct {
		name string
		f    findings.Finding
		want string
	}{
		{"path and line", findings.Finding{Locations: []findings.Location{{Path: "a.env", Line: 3}}}, "a.env:3"},
		{"path only", findings.Finding{Locations: []findings.Location{{Path: "AGENTS.md"}}}, "AGENTS.md"},
		{"no location", findings.Finding{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstLocation(tc.f); got != tc.want {
				t.Fatalf("firstLocation = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSeverityForDigit(t *testing.T) {
	cases := map[string]findings.Severity{
		"1": findings.SeverityBlocker,
		"2": findings.SeverityHigh,
		"3": findings.SeverityMedium,
		"4": findings.SeverityLow,
		"9": "",
	}
	for in, want := range cases {
		if got := severityForDigit(in); got != want {
			t.Fatalf("severityForDigit(%q) = %q, want %q", in, got, want)
		}
	}
}
