package governance

import (
	"fmt"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/suppress"
)

func ids(fs []findings.Finding) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.RuleID)
	}
	return out
}

func TestMissingReasonFires(t *testing.T) {
	used := []suppress.Entry{{Rule: "AE-MCP-001", Expires: "2099-01-01", Source: suppress.SourceExternal}}
	// activeRuleCount high enough that the 1 suppression stays under the 30% rate.
	fs := Run(used, 10, 1)
	if got := ids(fs); len(got) != 1 || got[0] != "AE-SUPPRESS-001" {
		t.Fatalf("expected [AE-SUPPRESS-001], got %v", got)
	}
	if fs[0].Severity != findings.SeverityMedium {
		t.Fatalf("expected Medium, got %s", fs[0].Severity)
	}
}

func TestReasonPresentDoesNotFire001(t *testing.T) {
	used := []suppress.Entry{{Rule: "AE-MCP-001", Reason: "vendored", Expires: "2099-01-01", Source: suppress.SourceExternal}}
	for _, f := range Run(used, 1, 1) {
		if f.RuleID == "AE-SUPPRESS-001" {
			t.Fatalf("reason present must not fire AE-SUPPRESS-001")
		}
	}
}

func TestPermanentNoApproverFires(t *testing.T) {
	used := []suppress.Entry{{Rule: "AE-CC-002", Reason: "x", Expires: "permanent", Source: suppress.SourceExternal}}
	fs := Run(used, 1, 0)
	if got := ids(fs); len(got) != 1 || got[0] != "AE-SUPPRESS-002" {
		t.Fatalf("expected [AE-SUPPRESS-002], got %v", got)
	}
	if fs[0].Severity != findings.SeverityHigh {
		t.Fatalf("expected High, got %s", fs[0].Severity)
	}
}

func TestTimeBoundedNoApproverDoesNotFire002(t *testing.T) {
	used := []suppress.Entry{{Rule: "AE-CC-002", Reason: "x", Expires: "2099-01-01", Source: suppress.SourceExternal}}
	for _, f := range Run(used, 1, 1) {
		if f.RuleID == "AE-SUPPRESS-002" {
			t.Fatalf("time-bounded suppression must not fire AE-SUPPRESS-002: %v", ids(Run(used, 1, 1)))
		}
	}
}

func TestBareEntryDoesNotFire002(t *testing.T) {
	// A bare entry (no expires) is a default-TTL suppression, not a permanent
	// waiver, so AE-SUPPRESS-002 must not fire.
	used := []suppress.Entry{{Rule: "AE-CC-002", Reason: "x", Source: suppress.SourceExternal}}
	for _, f := range Run(used, 1, 1) {
		if f.RuleID == "AE-SUPPRESS-002" {
			t.Fatalf("bare entry must not fire AE-SUPPRESS-002")
		}
	}
}

func TestInlineSuppressionLocationAndEvidence(t *testing.T) {
	used := []suppress.Entry{
		{Rule: "AE-MCP-001", Source: suppress.SourceInSource, Path: "mcp.yml", Line: 3},                                     // missing reason -> 001
		{Rule: "AE-CC-002", Reason: "x", Expires: "permanent", Source: suppress.SourceInSource, Path: "AGENTS.md", Line: 5}, // permanent no approver -> 002
	}
	fs := Run(used, 10, 2)
	var got001, got002 *findings.Finding
	for i := range fs {
		switch fs[i].RuleID {
		case "AE-SUPPRESS-001":
			got001 = &fs[i]
		case "AE-SUPPRESS-002":
			got002 = &fs[i]
		}
	}
	if got001 == nil || got002 == nil {
		t.Fatalf("expected both 001 and 002, got %v", ids(fs))
	}
	if got001.Locations[0].Path != "mcp.yml" || got001.Locations[0].Line != 3 {
		t.Fatalf("001 inline location wrong: %+v", got001.Locations)
	}
	if !strings.Contains(got001.Evidence[0], "inline suppression of AE-MCP-001") {
		t.Fatalf("001 evidence wrong: %q", got001.Evidence[0])
	}
	if got002.Locations[0].Path != "AGENTS.md" || got002.Locations[0].Line != 5 {
		t.Fatalf("002 inline location wrong: %+v", got002.Locations)
	}
}

func TestMissingReasonSortedByLocation(t *testing.T) {
	used := []suppress.Entry{
		{Rule: "AE-CC-001", Source: suppress.SourceInSource, Path: "b.md", Line: 2},
		{Rule: "AE-MCP-001", Source: suppress.SourceInSource, Path: "a.md", Line: 9},
		{Rule: "AE-MCP-002", Source: suppress.SourceInSource, Path: "a.md", Line: 1},
	}
	fs := Run(used, 10, 3) // 3/13 ~ 23% keeps AE-SUPPRESS-003 from firing
	var locs []string
	for _, f := range fs {
		if f.RuleID == "AE-SUPPRESS-001" {
			locs = append(locs, fmt.Sprintf("%s:%d", f.Locations[0].Path, f.Locations[0].Line))
		}
	}
	want := []string{"a.md:1", "a.md:9", "b.md:2"}
	if len(locs) != 3 || locs[0] != want[0] || locs[1] != want[1] || locs[2] != want[2] {
		t.Fatalf("expected %v, got %v", want, locs)
	}
}

func TestHighRateFiresInformational(t *testing.T) {
	// 6 suppressed of 9 total = 67% > 30%.
	fs := Run(nil, 3, 6)
	if got := ids(fs); len(got) != 1 || got[0] != "AE-SUPPRESS-003" {
		t.Fatalf("expected [AE-SUPPRESS-003], got %v", got)
	}
	if !fs[0].Informational {
		t.Fatalf("AE-SUPPRESS-003 must be informational")
	}
	if !strings.Contains(fs[0].Evidence[0], "67%") {
		t.Fatalf("expected rate in evidence, got %q", fs[0].Evidence[0])
	}
}

func TestLowRateDoesNotFire003(t *testing.T) {
	for _, f := range Run(nil, 10, 1) { // ~9%
		if f.RuleID == "AE-SUPPRESS-003" {
			t.Fatalf("low rate must not fire AE-SUPPRESS-003")
		}
	}
}

func TestZeroDenominatorNoFire(t *testing.T) {
	if fs := Run(nil, 0, 0); len(fs) != 0 {
		t.Fatalf("no suppressions/findings should yield no governance findings, got %v", ids(fs))
	}
}
