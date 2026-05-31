package suppress

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.use-charter.dev/charter/internal/findings"
)

func now(t *testing.T) time.Time {
	t.Helper()
	return time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
}

func mcpFinding() findings.Finding {
	return findings.Finding{RuleID: "AE-MCP-001", Severity: findings.SeverityHigh, Locations: []findings.Location{{Path: "mcp.yml", Line: 3}}}
}

// tempRepo gives a root with a benign mcp.yml (no directive) so that the inline
// detector can read the finding's source line, mirroring a real scan where every
// finding location points at a tracked, readable file.
func tempRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "mcp.yml"), []byte("servers:\n  db:\n    command: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestApplyFileHonored(t *testing.T) {
	all := []findings.Finding{mcpFinding()}
	entries := []Entry{{Rule: "AE-MCP-001", Reason: "ok", Expires: "2099-01-01", Source: SourceExternal}}
	active, suppressed, used, err := Apply(t.TempDir(), all, entries, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 || len(suppressed) != 1 {
		t.Fatalf("active=%d suppressed=%d", len(active), len(suppressed))
	}
	if len(used) != 1 {
		t.Fatalf("expected 1 used entry, got %d", len(used))
	}
}

func TestApplyFileExpiredStaysActive(t *testing.T) {
	all := []findings.Finding{mcpFinding()}
	entries := []Entry{{Rule: "AE-MCP-001", Reason: "ok", Expires: "2025-01-01", Source: SourceExternal}}
	active, suppressed, used, err := Apply(tempRepo(t), all, entries, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || len(suppressed) != 0 {
		t.Fatalf("expired entry should not suppress: active=%d suppressed=%d", len(active), len(suppressed))
	}
	if len(used) != 0 {
		t.Fatalf("expired entry must not be audited, got %d used", len(used))
	}
}

func TestApplyPermanentNoApproverNotHonoredButAudited(t *testing.T) {
	all := []findings.Finding{mcpFinding()}
	entries := []Entry{{Rule: "AE-MCP-001", Reason: "ok", Expires: "permanent", Source: SourceExternal}}
	active, suppressed, used, err := Apply(tempRepo(t), all, entries, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || len(suppressed) != 0 {
		t.Fatalf("permanent-no-approver must not suppress: active=%d suppressed=%d", len(active), len(suppressed))
	}
	if len(used) != 1 {
		t.Fatalf("permanent-no-approver must be audited, got %d used", len(used))
	}
}

func TestApplyPermanentWithApproverHonored(t *testing.T) {
	all := []findings.Finding{mcpFinding()}
	entries := []Entry{{Rule: "AE-MCP-001", Reason: "ok", Expires: "permanent", Approver: "sec", Source: SourceExternal}}
	active, suppressed, _, err := Apply(t.TempDir(), all, entries, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 || len(suppressed) != 1 {
		t.Fatalf("permanent+approver should suppress: active=%d suppressed=%d", len(active), len(suppressed))
	}
}

func TestApplyPathScope(t *testing.T) {
	all := []findings.Finding{mcpFinding()}
	entries := []Entry{{Rule: "AE-MCP-001", Reason: "ok", Expires: "2099-01-01", Path: "other.yml", Source: SourceExternal}}
	active, suppressed, _, err := Apply(tempRepo(t), all, entries, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || len(suppressed) != 0 {
		t.Fatalf("path-scope mismatch should not suppress: active=%d suppressed=%d", len(active), len(suppressed))
	}
}

func TestApplyInlineHonored(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "mcp.yml"), []byte("a\nb\nurl: x # charter:ignore AE-MCP-001 reason=\"vendored\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	all := []findings.Finding{mcpFinding()} // location mcp.yml:3
	active, suppressed, used, err := Apply(root, all, nil, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 || len(suppressed) != 1 || suppressed[0].Source != SourceInSource {
		t.Fatalf("inline directive should suppress inSource: active=%d suppressed=%+v", len(active), suppressed)
	}
	if len(used) != 1 {
		t.Fatalf("matched inline directive must be audited, got %d used", len(used))
	}
}

func TestApplyNoMatchActive(t *testing.T) {
	all := []findings.Finding{mcpFinding()}
	active, suppressed, used, err := Apply(tempRepo(t), all, nil, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || len(suppressed) != 0 || len(used) != 0 {
		t.Fatalf("no suppression should leave finding active: active=%d suppressed=%d used=%d", len(active), len(suppressed), len(used))
	}
}

func TestApplyBareEntryHonored(t *testing.T) {
	// A bare file entry (no expires) is a default-TTL suppression: honored, and
	// not flagged as permanent (only explicit expires: permanent needs an approver).
	all := []findings.Finding{mcpFinding()}
	entries := []Entry{{Rule: "AE-MCP-001", Reason: "ok", Source: SourceExternal}}
	active, suppressed, used, err := Apply(t.TempDir(), all, entries, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 || len(suppressed) != 1 {
		t.Fatalf("bare entry should be honored: active=%d suppressed=%d", len(active), len(suppressed))
	}
	if len(used) != 1 {
		t.Fatalf("bare entry must be audited, got %d used", len(used))
	}
}

func TestApplySecretSuppressibleWithApprover(t *testing.T) {
	all := []findings.Finding{{RuleID: "AE-SEC-001", Severity: findings.SeverityBlocker, Cap: 49, Locations: []findings.Location{{Path: "AGENTS.md", Line: 2}}}}
	entries := []Entry{{Rule: "AE-SEC-001", Reason: "test fixture key", Expires: "permanent", Approver: "sec", Source: SourceExternal}}
	active, suppressed, _, err := Apply(t.TempDir(), all, entries, now(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 || len(suppressed) != 1 {
		t.Fatalf("approved secret suppression should suppress: active=%d suppressed=%d", len(active), len(suppressed))
	}
}
