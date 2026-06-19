package text

import (
	"bytes"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/suppress"
	"go.use-charter.dev/charter/internal/terminal"
)

// sampleResult is a representative failing result that exercises every plain
// branch: locations with and without a line, evidence, a remediation-only
// finding, suppressions with and without a reason, and a multi-category
// scorecard.
func sampleResult() doctor.Result {
	return doctor.Result{
		Root:      "/repo",
		Threshold: 80,
		Passed:    false,
		Findings: []findings.Finding{
			{
				RuleID:      "AE-SEC-001",
				Severity:    findings.SeverityBlocker,
				Category:    "Secrets",
				Summary:     "Secret detected",
				Remediation: "Remove key",
				Evidence:    []string{"OPENAI_API_KEY=..."},
				Locations:   []findings.Location{{Path: "AGENTS.md", Line: 14}},
			},
			{
				RuleID:      "AE-MCP-001",
				Severity:    findings.SeverityHigh,
				Category:    "MCP Safety",
				Summary:     "MCP not pinned",
				Remediation: "charter fix",
				Locations:   []findings.Location{{Path: ".mcp.json", Line: 7}},
			},
			{
				RuleID:      "AE-ENV-001",
				Severity:    findings.SeverityMedium,
				Category:    "Environment",
				Summary:     "Toolchain incomplete",
				Remediation: "Pin runtimes",
			},
			{
				RuleID:      "AE-CI-002",
				Severity:    findings.SeverityLow,
				Category:    "CI",
				Summary:     "No workflows",
				Remediation: "Add workflow",
				Locations:   []findings.Location{{Path: ".github/workflows/"}},
			},
		},
		Suppressed: []suppress.Suppressed{
			{Finding: findings.Finding{RuleID: "AE-CC-001"}, Source: suppress.SourceExternal, Reason: "legacy"},
			{Finding: findings.Finding{RuleID: "AE-CC-002"}, Source: suppress.SourceInSource},
		},
		Score:        scoring.Result{Final: 49},
		PathsScanned: 8,
	}
}

func disabledCaps() terminal.Capabilities { return terminal.Capabilities{Tier: terminal.Mono} }
func ansi16Caps() terminal.Capabilities {
	return terminal.Capabilities{Tier: terminal.ANSI16, IsTTY: true}
}

func paletteFor(c terminal.Capabilities) terminal.Palette {
	return terminal.NewPalette(c, true)
}

func trueColorCaps(hyperlinks bool) terminal.Capabilities {
	return terminal.Capabilities{Tier: terminal.TrueColor, IsTTY: true, Hyperlinks: hyperlinks}
}

// TestRenderPlainIsByteIdenticalAndAnsiFree is the load-bearing containment
// test: with color disabled the renderer must produce the exact historical
// plain format and never emit an ANSI escape byte.
func TestRenderPlainIsByteIdenticalAndAnsiFree(t *testing.T) {
	t.Parallel()

	const want = "charter doctor: /repo\n" +
		"AE-SEC-001 BLOCKER Secret detected\n" +
		"  location: AGENTS.md:14\n" +
		"  - OPENAI_API_KEY=...\n" +
		"  remediation: Remove key\n" +
		"AE-MCP-001 HIGH MCP not pinned\n" +
		"  location: .mcp.json:7\n" +
		"  remediation: charter fix\n" +
		"AE-ENV-001 MEDIUM Toolchain incomplete\n" +
		"  remediation: Pin runtimes\n" +
		"AE-CI-002 LOW No workflows\n" +
		"  location: .github/workflows/\n" +
		"  remediation: Add workflow\n" +
		"suppressed: AE-CC-001 (external) — legacy\n" +
		"suppressed: AE-CC-002 (inSource)\n" +
		"readiness by category:\n" +
		"  Secrets      −20  (1 finding(s), worst BLOCKER)\n" +
		"  MCP Safety   −10  (1 finding(s), worst HIGH)\n" +
		"  Environment  −4   (1 finding(s), worst MEDIUM)\n" +
		"  CI           −1   (1 finding(s), worst LOW)\n" +
		"score: 49 (threshold 80)\n"

	got := Render(sampleResult(), disabledCaps(), paletteFor(disabledCaps()))
	if string(got) != want {
		t.Fatalf("plain render drifted from the documented format\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
	if bytes.IndexByte(got, 0x1b) != -1 {
		t.Fatalf("plain render must contain zero ANSI escape bytes, got: %q", got)
	}
}

// TestRenderPlainPassNoFindings covers the empty-findings/empty-scorecard plain
// branches and the passing score line.
func TestRenderPlainPassNoFindings(t *testing.T) {
	t.Parallel()

	result := doctor.Result{Root: "/repo", Threshold: 80, Passed: true, Score: scoring.Result{Final: 100}}
	const want = "charter doctor: /repo\nscore: 100 (threshold 80)\n"

	got := Render(result, disabledCaps(), paletteFor(disabledCaps()))
	if string(got) != want {
		t.Fatalf("plain pass render = %q, want %q", got, want)
	}
}

// TestRenderStyledTrueColor asserts the styled path keeps the load-bearing
// content (rule IDs, score, scorecard categories) and emits ANSI, without
// pinning a brittle full golden.
func TestRenderStyledTrueColor(t *testing.T) {
	t.Parallel()

	caps := trueColorCaps(false)
	got := Render(sampleResult(), caps, paletteFor(caps))

	if bytes.IndexByte(got, 0x1b) == -1 {
		t.Fatalf("styled render must contain ANSI escape bytes, got: %q", got)
	}

	out := string(got)
	for _, want := range []string{
		"AE-SEC-001", "AE-MCP-001", "AE-ENV-001", "AE-CI-002", // rule IDs
		"BLOCKER", "HIGH", "MEDIUM", "LOW", // severity text labels (never color-only)
		"Secret detected", "MCP not pinned", // summaries
		"AGENTS.md:14", ".mcp.json:7", // locations carry line numbers
		"OPENAI_API_KEY=...",          // evidence
		"Remove key",                  // remediation
		"Secrets", "MCP Safety", "CI", // scorecard categories
		"readiness by category", // scorecard heading
		"1/2",                   // per-category rules-clean fraction (Secrets: 1 of 2 rules clean)
		"49", "/100", "FAIL",    // score hero
		"min 80", "0 network calls", // status bar (threshold marker label + offline guarantee)
		"suppressed", "AE-CC-001", "legacy", // suppressions
		"charter", // brand header
	} {
		if !strings.Contains(out, want) {
			t.Errorf("styled render missing %q\nfull output:\n%s", want, out)
		}
	}

	// The brand wordmark is the committed [C] mark on every tier.
	if !strings.Contains(out, "[C] charter") {
		t.Errorf("expected the [C] brand wordmark, got:\n%s", out)
	}
	// TrueColor still uses the unicode severity glyph set (e.g. the ✗ cross for
	// the BLOCKER finding).
	if !strings.Contains(out, "✗") {
		t.Errorf("expected unicode glyphs on the TrueColor path, got:\n%s", out)
	}
	// Hyperlinks were disabled for this case: no OSC 8 file targets.
	if strings.Contains(out, "file://") {
		t.Errorf("did not expect hyperlinks when caps.Hyperlinks is false, got:\n%s", out)
	}
}

// TestRenderStyledHyperlinks covers the OSC 8 link branches for both the rule
// help URI and file locations (relative and absolute).
func TestRenderStyledHyperlinks(t *testing.T) {
	t.Parallel()

	result := sampleResult()
	result.Findings = append(result.Findings, findings.Finding{
		RuleID:      "AE-CTX-001",
		Severity:    findings.SeverityHigh,
		Category:    "Context",
		Summary:     "absolute location",
		Remediation: "fix it",
		Locations:   []findings.Location{{Path: "/abs/path.go", Line: 3}},
	})

	caps := trueColorCaps(true)
	out := string(Render(result, caps, paletteFor(caps)))

	for _, want := range []string{
		"https://use-charter.dev/rules/AE-SEC-001", // rule help URI hyperlink
		"file:///repo/AGENTS.md",                   // relative path resolved against root
		"file:///abs/path.go",                      // already-absolute path preserved
	} {
		if !strings.Contains(out, want) {
			t.Errorf("styled hyperlink render missing %q\nfull output:\n%s", want, out)
		}
	}
}

// TestRenderStyledANSI16 covers the ASCII glyph fallback and the no-mark verdict
// used on non-TrueColor tiers.
func TestRenderStyledANSI16(t *testing.T) {
	t.Parallel()

	caps := ansi16Caps()
	out := string(Render(sampleResult(), caps, paletteFor(caps)))

	if strings.Contains(out, "✦") || strings.Contains(out, "✗") {
		t.Errorf("expected ASCII glyph fallback on ANSI16, got:\n%s", out)
	}
	if !strings.Contains(out, "[C] charter") {
		t.Errorf("expected the [C] brand wordmark on ANSI16, got:\n%s", out)
	}
	if !strings.Contains(out, "|") {
		t.Errorf("expected ASCII card bar on ANSI16, got:\n%s", out)
	}
	// Verdict has no trailing glyph on non-unicode tiers.
	if !strings.Contains(out, "FAIL") || strings.Contains(out, "FAIL ✗") {
		t.Errorf("expected bare FAIL verdict on ANSI16, got:\n%s", out)
	}
}

// TestRenderStyledPass covers the passing verdict and the clean-pass scorecard:
// the Findings section is omitted (nothing to list), but the readiness scorecard
// still renders every category as all-clean (reassuring, and matching the
// styled status bar), and the status bar reports the offline guarantee.
func TestRenderStyledPass(t *testing.T) {
	t.Parallel()

	result := doctor.Result{Root: "/repo", Threshold: 80, Passed: true, Score: scoring.Result{Final: 100}}
	caps := trueColorCaps(false)
	out := string(Render(result, caps, paletteFor(caps)))

	if !strings.Contains(out, "PASS") || !strings.Contains(out, "100") {
		t.Errorf("expected passing score hero, got:\n%s", out)
	}
	// No finding cards on a clean pass.
	if strings.Contains(out, "Findings") {
		t.Errorf("did not expect a Findings section for a clean pass, got:\n%s", out)
	}
	// The full readiness scorecard renders, every category clean (e.g. Context 4/4).
	for _, want := range []string{"readiness by category", "Context", "4/4", "0 network calls"} {
		if !strings.Contains(out, want) {
			t.Errorf("clean pass: expected %q in scorecard/status bar, got:\n%s", want, out)
		}
	}
}

// truecolor dark SGR foreground prefixes for the semantic tokens this package
// uses, derived from the palette hex values (DESIGN-TOKENS.md).
const (
	dangerRGB  = "38;2;248;113;113" // #f87171 text-danger (dark)
	warningRGB = "38;2;251;191;36"  // #fbbf24 text-warning (dark)
)

// TestRenderStyledHighUsesWarningNotDanger pins the design's severity mapping:
// a HIGH finding renders with the warning token and warn glyph, never the
// danger token or the ✗ cross.
func TestRenderStyledHighUsesWarningNotDanger(t *testing.T) {
	t.Parallel()

	result := doctor.Result{
		Root:      "/repo",
		Threshold: 80,
		Passed:    true, // a single HIGH (−10) still passes, so the hero is green
		Findings: []findings.Finding{
			{RuleID: "AE-CC-002", Severity: findings.SeverityHigh, Category: "Agent Config", Summary: "high finding", Remediation: "fix"},
		},
		Score: scoring.Result{Final: 90, Base: 90},
	}
	caps := trueColorCaps(false)
	out := string(Render(result, caps, paletteFor(caps)))

	if !strings.Contains(out, "⚠") {
		t.Errorf("expected the warning glyph ⚠ for a HIGH finding, got:\n%s", out)
	}
	if strings.Contains(out, "✗") {
		t.Errorf("HIGH must not emit the danger cross ✗ on a passing run, got:\n%s", out)
	}
	if !strings.Contains(out, warningRGB) {
		t.Errorf("expected warning color %q for HIGH, got:\n%s", warningRGB, out)
	}
	if strings.Contains(out, dangerRGB) {
		t.Errorf("did not expect danger color %q anywhere for a passing HIGH-only run, got:\n%s", dangerRGB, out)
	}
}

// TestRenderStyledSummaryLine covers the findings-count rollup, including the
// per-severity buckets and zero-bucket omission.
func TestRenderStyledSummaryLine(t *testing.T) {
	t.Parallel()

	caps := trueColorCaps(false)
	out := string(Render(sampleResult(), caps, paletteFor(caps)))

	for _, want := range []string{
		"Checked 8 paths", " rules", " categories", // scan breadth (paths from result, rule/category counts catalog-driven)
		"4 findings",                               // total findings
		"1 BLOCKER", "1 HIGH", "1 MEDIUM", "1 LOW", // per-severity buckets
	} {
		if !strings.Contains(out, want) {
			t.Errorf("summary line missing %q\nfull output:\n%s", want, out)
		}
	}
}

// TestRenderStyledSummaryCleanPass covers the zero-findings summary variant.
func TestRenderStyledSummaryCleanPass(t *testing.T) {
	t.Parallel()

	result := doctor.Result{Root: "/repo", Threshold: 80, Passed: true, Score: scoring.Result{Final: 100, Base: 100}}
	caps := trueColorCaps(false)
	out := string(Render(result, caps, paletteFor(caps)))

	if !strings.Contains(out, "0 findings") {
		t.Errorf("expected a zero-findings summary, got:\n%s", out)
	}
	for _, unwanted := range []string{"BLOCKER", "HIGH", "MEDIUM", "LOW"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("did not expect severity bucket %q for a clean pass, got:\n%s", unwanted, out)
		}
	}
}

// TestRenderStyledScoreCapAndBar covers the cap notice (Final < Base) and the
// presence of a proportional progress bar in the score hero.
func TestRenderStyledScoreCapAndBar(t *testing.T) {
	t.Parallel()

	result := doctor.Result{
		Root:      "/repo",
		Threshold: 80,
		Passed:    false,
		Findings: []findings.Finding{
			{RuleID: "AE-SEC-001", Severity: findings.SeverityBlocker, Category: "Secrets", Summary: "secret", Remediation: "remove"},
		},
		Score: scoring.Result{Base: 80, Final: 49},
	}
	caps := trueColorCaps(false)
	out := string(Render(result, caps, paletteFor(caps)))

	if !strings.Contains(out, "capped at 49") {
		t.Errorf("expected a cap notice when Final < Base, got:\n%s", out)
	}
	if !strings.Contains(out, "█") {
		t.Errorf("expected a unicode progress bar in the score hero, got:\n%s", out)
	}
	if !strings.Contains(out, "░") {
		t.Errorf("expected an unfilled bar segment for a sub-100 score, got:\n%s", out)
	}
}

// TestRenderStyledNoCapWhenFinalEqualsBase ensures the cap notice is omitted
// when no cap is active.
func TestRenderStyledNoCapWhenFinalEqualsBase(t *testing.T) {
	t.Parallel()

	result := doctor.Result{Root: "/repo", Threshold: 80, Passed: true, Score: scoring.Result{Base: 100, Final: 100}}
	caps := trueColorCaps(false)
	out := string(Render(result, caps, paletteFor(caps)))

	if strings.Contains(out, "capped at") {
		t.Errorf("did not expect a cap notice when Final == Base, got:\n%s", out)
	}
}

// TestRenderStyledUnknownSeverity covers the default token/icon arms.
func TestRenderStyledUnknownSeverity(t *testing.T) {
	t.Parallel()

	result := doctor.Result{
		Root:      "/repo",
		Threshold: 80,
		Findings: []findings.Finding{
			{RuleID: "AE-X-000", Severity: findings.Severity("WEIRD"), Category: "Other", Summary: "odd", Remediation: "n/a"},
		},
		Score: scoring.Result{Final: 100},
	}
	caps := trueColorCaps(false)
	out := string(Render(result, caps, paletteFor(caps)))

	if !strings.Contains(out, "AE-X-000") || !strings.Contains(out, "WEIRD") {
		t.Errorf("expected unknown-severity finding rendered with its label, got:\n%s", out)
	}
}

// TestRenderFocusedPlainSingleRule pins the focused plain projection for one
// rule: a filtered header plus that rule's plain finding block, and crucially
// no score line and no scorecard (a partial view must not imply a verdict).
func TestRenderFocusedPlainSingleRule(t *testing.T) {
	t.Parallel()

	const want = "charter doctor — filtered: AE-SEC-001\n" +
		"AE-SEC-001 BLOCKER Secret detected\n" +
		"  location: AGENTS.md:14\n" +
		"  - OPENAI_API_KEY=...\n" +
		"  remediation: Remove key\n"

	got := RenderFocused(sampleResult(), []string{"AE-SEC-001"}, disabledCaps(), paletteFor(disabledCaps()))
	if string(got) != want {
		t.Fatalf("focused plain render drifted\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
	if bytes.IndexByte(got, 0x1b) != -1 {
		t.Fatalf("focused plain render must contain zero ANSI escape bytes, got: %q", got)
	}
	out := string(got)
	for _, unwanted := range []string{"score:", "readiness by category", "AE-MCP-001"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("focused plain render must omit %q, got:\n%s", unwanted, out)
		}
	}
}

// TestRenderFocusedPlainMultiRule covers comma-joined headers and multiple
// matched findings in scan order.
func TestRenderFocusedPlainMultiRule(t *testing.T) {
	t.Parallel()

	got := string(RenderFocused(sampleResult(), []string{"AE-SEC-001", "AE-CI-002"}, disabledCaps(), paletteFor(disabledCaps())))

	if !strings.HasPrefix(got, "charter doctor — filtered: AE-SEC-001, AE-CI-002\n") {
		t.Fatalf("expected a comma-joined filtered header, got:\n%s", got)
	}
	for _, want := range []string{"AE-SEC-001 BLOCKER", "AE-CI-002 LOW"} {
		if !strings.Contains(got, want) {
			t.Fatalf("focused render missing %q, got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "AE-MCP-001") || strings.Contains(got, "AE-ENV-001") {
		t.Fatalf("focused render must include only the selected rules, got:\n%s", got)
	}
	// AE-SEC-001 precedes AE-CI-002 in the scan order; the focused view keeps it.
	if strings.Index(got, "AE-SEC-001") > strings.Index(got, "AE-CI-002") {
		t.Fatalf("focused render must preserve scan order, got:\n%s", got)
	}
}

// TestRenderFocusedNoMatch covers the empty-match note shown when a selected
// rule produced no finding.
func TestRenderFocusedNoMatch(t *testing.T) {
	t.Parallel()

	const want = "charter doctor — filtered: AE-CTX-001\n" + focusedNote + "\n"
	got := string(RenderFocused(sampleResult(), []string{"AE-CTX-001"}, disabledCaps(), paletteFor(disabledCaps())))
	if got != want {
		t.Fatalf("focused no-match render = %q, want %q", got, want)
	}
}

// TestRenderFocusedStyled covers the styled focused path: ANSI present, the
// brand + filtered header, the selected finding, and no score/scorecard.
func TestRenderFocusedStyled(t *testing.T) {
	t.Parallel()

	caps := trueColorCaps(false)
	got := RenderFocused(sampleResult(), []string{"AE-SEC-001"}, caps, paletteFor(caps))

	if bytes.IndexByte(got, 0x1b) == -1 {
		t.Fatalf("focused styled render must contain ANSI escape bytes, got: %q", got)
	}
	out := string(got)
	for _, want := range []string{"[C] charter", "filtered:", "AE-SEC-001", "Secret detected", "Remove key"} {
		if !strings.Contains(out, want) {
			t.Errorf("focused styled render missing %q\nfull output:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"readiness by category", "/100", "AE-MCP-001"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("focused styled render must omit %q, got:\n%s", unwanted, out)
		}
	}
}

// TestRenderFocusedStyledNoMatch covers the styled empty-match note.
func TestRenderFocusedStyledNoMatch(t *testing.T) {
	t.Parallel()

	caps := trueColorCaps(false)
	out := string(RenderFocused(sampleResult(), []string{"AE-CTX-001"}, caps, paletteFor(caps)))
	if !strings.Contains(out, focusedNote) {
		t.Errorf("expected the empty-match note in the styled focused view, got:\n%s", out)
	}
}
