package html

import (
	"flag"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/suppress"
)

var update = flag.Bool("update", false, "update golden files")

// fixedMeta is the injected, normalized provenance that keeps golden output
// byte-stable across machines and clocks.
func fixedMeta() meta {
	return meta{
		generatedAt: time.Date(2026, 6, 2, 18, 24, 37, 0, time.UTC),
		version:     "v1.0.0",
		commit:      "abc1234fdeadbeef0000",
	}
}

// multiFindingResult exercises every branch the report renders: a capped blocker,
// high/medium/low deductions, an informational governance finding, evidence that
// must be HTML-escaped, file-level and line-level locations, and two suppressions.
func multiFindingResult() doctor.Result {
	fs := []findings.Finding{
		{
			RuleID:      "AE-SEC-001",
			Severity:    findings.SeverityBlocker,
			Category:    "Secrets",
			Summary:     "Secret detected in agent-visible context file",
			Remediation: "Remove the key and reference it via an env var. Revoke the exposed key immediately.",
			Evidence:    []string{"OPENAI_API_KEY=sk-proj-•••• <redacted> & rotated"},
			Locations:   []findings.Location{{Path: "AGENTS.md", Line: 14}},
			Cap:         49,
		},
		{
			RuleID:      "AE-MCP-001",
			Severity:    findings.SeverityHigh,
			Category:    "MCP Safety",
			Summary:     "MCP server not pinned to a specific version",
			Remediation: "Pin to an exact CalVer version from the MCP catalog.",
			Evidence:    []string{`"mcp-server-git@latest"`},
			Locations:   []findings.Location{{Path: ".mcp.json", Line: 7}, {Path: ".mcp.json", Line: 12}},
		},
		{
			RuleID:      "AE-ENV-001",
			Severity:    findings.SeverityMedium,
			Category:    "Environment",
			Summary:     "Runtimes active but unpinned",
			Remediation: "Pin runtimes in mise.toml.",
		},
		{
			RuleID:    "AE-CTX-004",
			Severity:  findings.SeverityLow,
			Category:  "Context",
			Summary:   "Agent session artifacts are not git-ignored",
			Locations: []findings.Location{{Path: ".gitignore", Line: 0}},
		},
		{
			RuleID:        "AE-SUPPRESS-003",
			Severity:      findings.SeverityLow,
			Category:      "Governance",
			Summary:       "Suppression rate is elevated",
			Informational: true,
		},
	}

	supp := []suppress.Suppressed{
		{
			Finding: findings.Finding{RuleID: "AE-CI-002", Severity: findings.SeverityHigh, Category: "CI"},
			Source:  suppress.SourceExternal,
			Reason:  "Charter CI runs in the platform repo, not per-service",
			Expires: "2026-08-16",
		},
		{
			Finding:  findings.Finding{RuleID: "AE-CTX-006", Severity: findings.SeverityLow, Category: "Context"},
			Source:   suppress.SourceExternal,
			Reason:   "Accepted low-emphasis policy",
			Expires:  "permanent",
			Approver: "eng-lead@example.com",
		},
	}

	score := scoring.Calculate(fs)
	return doctor.Result{
		Root:       "/work/my-repo",
		Threshold:  80,
		Passed:     score.Final >= 80,
		Findings:   fs,
		Suppressed: supp,
		Score:      score,
	}
}

func cleanResult() doctor.Result {
	score := scoring.Calculate(nil)
	return doctor.Result{
		Root:      "/work/clean-repo",
		Threshold: 80,
		Passed:    score.Final >= 80,
		Score:     score,
	}
}

func render(t *testing.T, result doctor.Result) string {
	t.Helper()
	out, err := renderWith(result, fixedMeta())
	if err != nil {
		t.Fatalf("renderWith: %v", err)
	}
	return string(out)
}

// TestSelfContained is the load-bearing invariant (ADR-0025): the rendered HTML
// must reference zero external resources. Anchor links to the public rule docs are
// navigational and allowed; everything else that loads an asset is forbidden.
func TestSelfContained(t *testing.T) {
	out := render(t, multiFindingResult())

	// Match either quote style so a single-quoted external asset cannot bypass
	// the load-bearing self-containment check.
	attrRe := regexp.MustCompile(`(?i)\b(src|srcset|href)\s*=\s*["']([^"']*)["']`)
	for _, m := range attrRe.FindAllStringSubmatch(out, -1) {
		attr, val := strings.ToLower(m[1]), m[2]
		if strings.HasPrefix(val, "//") {
			t.Errorf("protocol-relative %s reference is not self-contained: %q", attr, val)
			continue
		}
		if !strings.HasPrefix(val, "http://") && !strings.HasPrefix(val, "https://") {
			continue
		}
		if attr != "href" {
			t.Errorf("external %s loads a remote asset: %q", attr, val)
			continue
		}
		u, err := url.Parse(val)
		if err != nil || u.Host != "use-charter.dev" {
			t.Errorf("external href is not an allowed doc link: %q", val)
		}
	}

	forbidden := []string{
		"<link",        // no external stylesheet/preload/icon
		"@import",      // no CSS @import
		"<script src",  // no external script
		"<script  src", // (defensive against whitespace)
		"url(http",     // no remote url() in CSS
		"url('http",    //
		"url(\"http",   //
		"<img",         // no remote images
		"<iframe",      // no embedded frames
	}
	low := strings.ToLower(out)
	for _, f := range forbidden {
		if strings.Contains(low, f) {
			t.Errorf("output contains forbidden external reference token %q", f)
		}
	}

	// Positive: assets are actually inlined, not just absent.
	if !strings.Contains(out, "<style>") || !strings.Contains(out, "</style>") {
		t.Error("expected inlined <style> block")
	}
	if !strings.Contains(out, "<script>") {
		t.Error("expected inlined <script> block")
	}
	if !strings.Contains(out, "viewBox=\"0 0 64 64\"") {
		t.Error("expected the inline brand mark SVG")
	}
}

func TestGolden(t *testing.T) {
	out, err := renderWith(multiFindingResult(), fixedMeta())
	if err != nil {
		t.Fatalf("renderWith: %v", err)
	}
	golden := filepath.Join("testdata", "golden.html")
	if *update {
		if err := os.MkdirAll("testdata", 0o750); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(golden, out, 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(golden) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read golden (run with -update to generate): %v", err)
	}
	if string(out) != string(want) {
		t.Errorf("golden mismatch: rendered output differs from %s (run with -update to refresh)", golden)
	}
}

func TestDeterministic(t *testing.T) {
	a := render(t, multiFindingResult())
	b := render(t, multiFindingResult())
	if a != b {
		t.Error("render is not deterministic for identical input + meta")
	}
}

func TestContent(t *testing.T) {
	out := render(t, multiFindingResult())

	for _, id := range []string{"AE-SEC-001", "AE-MCP-001", "AE-ENV-001", "AE-CTX-004", "AE-SUPPRESS-003"} {
		if !strings.Contains(out, id) {
			t.Errorf("missing finding rule id %s", id)
		}
	}
	for _, id := range []string{"AE-CI-002", "AE-CTX-006"} {
		if !strings.Contains(out, id) {
			t.Errorf("missing suppression rule id %s", id)
		}
	}
	for _, cat := range []string{"Secrets", "MCP Safety", "Environment", "Context", "Governance", "Testing"} {
		if !strings.Contains(out, cat) {
			t.Errorf("missing category %s", cat)
		}
	}

	// Evidence must be HTML-escaped, never injected raw.
	if !strings.Contains(out, "&lt;redacted&gt;") || !strings.Contains(out, "&amp; rotated") {
		t.Error("evidence was not HTML-escaped")
	}
	if strings.Contains(out, "<redacted>") {
		t.Error("raw unescaped evidence leaked into output")
	}

	// Locations rendered as path:line and file-level.
	if !strings.Contains(out, "AGENTS.md:14") || !strings.Contains(out, ".mcp.json:7") {
		t.Error("missing path:line location")
	}
	if !strings.Contains(out, "Why this matters") {
		t.Error("missing catalog-sourced 'Why this matters'")
	}
	if !strings.Contains(out, "https://use-charter.dev/rules/AE-SEC-001") {
		t.Error("missing rule doc link")
	}
}

func TestCapAlertPresentWhenCapped(t *testing.T) {
	out := render(t, multiFindingResult())
	if !strings.Contains(out, `role="alert"`) {
		t.Error("expected a cap alert (role=alert) when score is capped")
	}
	if !strings.Contains(out, "AE-SEC-001 caps the score at 49") {
		t.Errorf("cap alert should name the binding rule and value")
	}
	// Formula shows the cap transition.
	if !strings.Contains(out, "cap") || !strings.Contains(out, "<strong>65</strong>") {
		t.Error("formula should show base 65 and a cap transition")
	}
}

func TestBlockerOnlyCap(t *testing.T) {
	fs := []findings.Finding{{
		RuleID: "AE-CC-001", Severity: findings.SeverityBlocker, Category: "Agent Config",
		Summary: "Dangerous hook command", Locations: []findings.Location{{Path: "hooks.json", Line: 3}},
	}}
	score := scoring.Calculate(fs) // base 80, blocker cap -> 59
	res := doctor.Result{Root: "/r/x", Threshold: 80, Passed: false, Findings: fs, Score: score}

	cap := buildCap(res)
	if !cap.Active || !strings.Contains(cap.Title, "capped at 59") {
		t.Errorf("expected blocker cap at 59, got %+v", cap)
	}
	out := render(t, res)
	if !strings.Contains(out, "Blocker present") {
		t.Error("expected blocker-cap callout in output")
	}
}

func TestAllClean(t *testing.T) {
	out := render(t, cleanResult())

	if !strings.Contains(out, "No active findings") {
		t.Error("clean report should show the empty-findings state")
	}
	if strings.Contains(out, `role="alert"`) {
		t.Error("clean report must not render a cap alert")
	}
	if !strings.Contains(out, "100/100 PASS") {
		t.Error("clean report title should report a perfect passing score")
	}
	if !strings.Contains(out, "all passed") {
		t.Error("clean report should mark categories as all passed")
	}
	// No findings list cards.
	if strings.Contains(out, `class="fc"`) {
		t.Error("clean report should not render finding cards")
	}
	// No suppressions section when none exist.
	if strings.Contains(out, `id="suppressions"`) {
		t.Error("clean report should omit the suppressions section")
	}
}

// TestUncatalogedCategory covers a finding whose Category is absent from the rule
// catalog: the category must still render in the scorecard, with its rule total
// inferred from the distinct rule IDs seen in the findings.
func TestUncatalogedCategory(t *testing.T) {
	fs := []findings.Finding{
		{RuleID: "AE-EXP-001", Severity: findings.SeverityHigh, Category: "Experimental", Summary: "Experimental rule A"},
		{RuleID: "AE-EXP-002", Severity: findings.SeverityMedium, Category: "Experimental", Summary: "Experimental rule B"},
	}

	score := scoring.Calculate(fs)
	res := doctor.Result{Root: "/work/exp-repo", Threshold: 80, Passed: false, Findings: fs, Score: score}
	out := render(t, res)

	// HIGH (−10) + MEDIUM (−4) = −14; 2 distinct rules, none clean → 0 of 2.
	if !strings.Contains(out, "Experimental: 0 of 2 rules clean, 2 findings, worst HIGH, minus 14 points") {
		t.Error("uncataloged category did not render with the inferred rule total")
	}
	if !strings.Contains(out, "Experimental") || !strings.Contains(out, "2 rule") {
		t.Error("uncataloged category should appear in the scorecard with its distinct rule count")
	}
}

func TestA11yStructure(t *testing.T) {
	out := render(t, multiFindingResult())
	required := []string{
		`role="banner"`,
		`role="progressbar"`,
		`aria-valuenow=`,
		`role="button"`,
		`aria-expanded="true"`,  // first card open by default
		`aria-expanded="false"`, // other cards collapsed
		`aria-pressed="true"`,   // "All" filter pressed
		`aria-pressed="false"`,
		`aria-label="Search findings`,
		`aria-labelledby="findings-h"`,
		`role="list"`,
		`<a class="skip-link"`,
		`<main `,   // main landmark
		`<footer>`, // contentinfo landmark (top-level <footer>)
	}
	for _, r := range required {
		if !strings.Contains(out, r) {
			t.Errorf("missing a11y marker: %s", r)
		}
	}
}

func TestSuppressionStatuses(t *testing.T) {
	supp := []suppress.Suppressed{
		{Finding: findings.Finding{RuleID: "AE-A", Severity: findings.SeverityLow}, Expires: "permanent"}, // no approver
		{Finding: findings.Finding{RuleID: "AE-B", Severity: findings.SeverityLow}},                       // no reason, default expiry
	}
	vms := buildSuppressions(supp)
	got := map[string]suppressionVM{}
	for _, v := range vms {
		got[v.RuleID] = v
	}
	if got["AE-A"].Status != "no approver" {
		t.Errorf("AE-A status = %q, want 'no approver'", got["AE-A"].Status)
	}
	if got["AE-B"].Status != "no reason" || got["AE-B"].Expires != "default" {
		t.Errorf("AE-B = %+v, want status 'no reason' expires 'default'", got["AE-B"])
	}
}

func TestPublicRenderSmoke(t *testing.T) {
	out, err := Render(multiFindingResult())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	s := string(out)
	if !strings.HasPrefix(s, "<!doctype html>") {
		t.Error("expected an HTML5 document")
	}
	if !strings.Contains(s, "self-contained · offline") {
		t.Error("expected footer self-contained marker")
	}
}

func TestHelpers(t *testing.T) {
	if got := repoName("/work/my-repo/"); got != "my-repo" {
		t.Errorf("repoName trailing slash = %q", got)
	}
	if got := repoName(""); got != "repository" {
		t.Errorf("repoName empty = %q", got)
	}
	if got := shortCommit(""); got != "unknown" {
		t.Errorf("shortCommit empty = %q", got)
	}
	if got := shortCommit("abc1234fdeadbeef"); got != "abc1234fdead" {
		t.Errorf("shortCommit long = %q", got)
	}
	if got := shortCommit("abc1234"); got != "abc1234" {
		t.Errorf("shortCommit short = %q", got)
	}
	for score, wantZone := range map[int]string{95: "success", 70: "warning", 30: "danger"} {
		if z, _ := zoneOf(score); z != wantZone {
			t.Errorf("zoneOf(%d) = %q, want %q", score, z, wantZone)
		}
	}
	if categoryIcon("Nonexistent Category") != "dot" {
		t.Error("unknown category should fall back to the dot icon")
	}
	if severityIcon(findings.SeverityBlocker) != "x" || severityClass(findings.SeverityLow) != "b" {
		t.Error("severity icon/class mapping incorrect")
	}
	if string(icon("definitely-missing")) != string(icon("dot")) {
		t.Error("unknown icon should fall back to dot")
	}
}
