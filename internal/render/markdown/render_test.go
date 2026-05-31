package markdown

import (
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/suppress"
)

func TestRenderSuppressedSection(t *testing.T) {
	result := doctor.Result{
		Root: "/repo", Threshold: 80, Passed: true,
		Suppressed: []suppress.Suppressed{
			{Finding: findings.Finding{RuleID: "AE-MCP-001", Severity: findings.SeverityHigh}, Source: suppress.SourceExternal, Reason: "vendored", Expires: "2099-01-01"},
		},
		Score: scoreResult(100, 100),
	}
	out := string(mustRender(t, result))
	if !strings.Contains(out, "Suppressed (1)") {
		t.Fatalf("expected suppressed section:\n%s", out)
	}
	if !strings.Contains(out, "AE-MCP-001") || !strings.Contains(out, "vendored") || !strings.Contains(out, "external") {
		t.Fatalf("expected suppressed row:\n%s", out)
	}
}

func scoreResult(base, final int) scoring.Result {
	return scoring.Result{Base: base, Final: final}
}

func mustRender(t *testing.T, r doctor.Result) []byte {
	t.Helper()
	out, err := Render(r)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return out
}

func TestRenderWithFindings(t *testing.T) {
	result := doctor.Result{
		Root: "/repo", Threshold: 80, Passed: false,
		Findings: []findings.Finding{
			{RuleID: "AE-CC-001", Severity: findings.SeverityBlocker, Category: "Agent Config", Summary: "dangerous command", Locations: []findings.Location{{Path: ".claude/settings.json", Line: 7}}},
			{RuleID: "AE-CTX-002", Severity: findings.SeverityMedium, Category: "Context", Summary: "drift"},
			{RuleID: "AE-ENV-001", Severity: findings.SeverityMedium, Category: "Environment", Summary: "missing a | pipe", Locations: []findings.Location{{Path: "mise.toml"}}},
		},
		Score: scoreResult(59, 47),
	}
	out := string(mustRender(t, result))
	if !strings.Contains(out, "# Charter") || !strings.Contains(out, "47") {
		t.Fatalf("missing header/score:\n%s", out)
	}
	iCC := strings.Index(out, "AE-CC-001")
	iCTX := strings.Index(out, "AE-CTX-002")
	if iCC < 0 || iCTX < 0 {
		t.Fatalf("both findings should appear:\n%s", out)
	}
	if iCC > iCTX {
		t.Fatalf("findings not severity-ordered:\n%s", out)
	}
	if !strings.Contains(out, ".claude/settings.json:7") {
		t.Fatalf("missing location:\n%s", out)
	}
	if !strings.Contains(out, `missing a \| pipe`) {
		t.Fatalf("pipe not escaped in summary:\n%s", out)
	}
	if !strings.Contains(out, "`mise.toml`") {
		t.Fatalf("file-level location should render backticked path without line:\n%s", out)
	}
}

func TestRenderNoFindings(t *testing.T) {
	out := string(mustRender(t, doctor.Result{Root: "/repo", Threshold: 80, Passed: true, Score: scoreResult(100, 100)}))
	if !strings.Contains(out, "100") || !strings.Contains(strings.ToLower(out), "no findings") {
		t.Fatalf("expected clean report:\n%s", out)
	}
}
