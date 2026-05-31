package suppress

import "testing"

func TestParseInlineDirective(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		wantOK   bool
		wantRule string
		wantRsn  string
		wantExp  string
		wantApp  string
	}{
		{"hash with reason", `db_url: "x" # charter:ignore AE-MCP-001 reason="vendored"`, true, "AE-MCP-001", "vendored", "", ""},
		{"slash bare", `// charter:ignore AE-CC-001`, true, "AE-CC-001", "", "", ""},
		{"html with all", `<!-- charter:ignore AE-CC-002 reason="legacy" expires=2026-09-01 approver="sec" -->`, true, "AE-CC-002", "legacy", "2026-09-01", "sec"},
		{"permanent", `# charter:ignore AE-MCP-002 expires=permanent approver="alice"`, true, "AE-MCP-002", "", "permanent", "alice"},
		{"no comment leader", `charter:ignore AE-MCP-001 reason="x"`, false, "", "", "", ""},
		{"comment but no directive", `# just a normal comment`, false, "", "", "", ""},
		{"prose mentions syntax without a comment", "text about charter:ignore AE-MCP-001 without a comment", false, "", "", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e, ok := parseInlineDirective(c.line)
			if ok != c.wantOK {
				t.Fatalf("ok = %v, want %v (line %q)", ok, c.wantOK, c.line)
			}
			if !ok {
				return
			}
			if e.Rule != c.wantRule || e.Reason != c.wantRsn || e.Expires != c.wantExp || e.Approver != c.wantApp {
				t.Fatalf("got %+v", e)
			}
			if e.Source != SourceInSource {
				t.Fatalf("expected inSource, got %q", e.Source)
			}
		})
	}
}
