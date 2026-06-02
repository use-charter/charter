package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/rules/catalog"
)

// runDoctor executes `charter doctor <args...>` against a buffered root and
// returns stdout plus the command error.
func runDoctor(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(append([]string{"doctor"}, args...))
	cmd.SetContext(context.Background())
	err := cmd.Execute()
	return out.String(), err
}

// firedAndUnfired discovers, for a repo, one rule ID that produced a finding and
// one catalog rule ID that did not — so the --rule tests are robust to rule-set
// changes.
func firedAndUnfired(t *testing.T, repo string) (fired, unfired string) {
	t.Helper()
	result, err := doctor.Run(repo, 80, false)
	if err != nil {
		t.Fatalf("doctor.Run: %v", err)
	}
	if len(result.Findings) == 0 {
		t.Fatalf("fixture repo produced no findings; cannot exercise --rule exit 1")
	}
	firing := make(map[string]bool)
	for _, f := range result.Findings {
		firing[f.RuleID] = true
	}
	fired = result.Findings[0].RuleID
	for _, id := range catalog.IDs() {
		if !firing[id] {
			unfired = id
			break
		}
	}
	if unfired == "" {
		t.Fatalf("every catalog rule fired; cannot exercise --rule exit 0")
	}
	return fired, unfired
}

// TestDoctorRuleFiredExitsOne covers the focused view for a rule that produced a
// finding: filtered header, the finding, exit 1, and crucially no score line and
// no scorecard.
func TestDoctorRuleFiredExitsOne(t *testing.T) {
	repo := initTempRepo(t)
	fired, _ := firedAndUnfired(t, repo)

	out, err := runDoctor(t, "--path", repo, "--rule", fired)
	if code := exitCodeOf(t, err); code != 1 {
		t.Fatalf("--rule %s exit code = %d, want 1 (the rule fired)", fired, code)
	}
	if !strings.Contains(out, "charter doctor — filtered: "+fired) {
		t.Fatalf("expected a filtered header for %s, got:\n%s", fired, out)
	}
	if !strings.Contains(out, fired) {
		t.Fatalf("expected the focused view to include %s, got:\n%s", fired, out)
	}
	for _, unwanted := range []string{"score:", "readiness by category"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("filtered run must omit %q, got:\n%s", unwanted, out)
		}
	}
	if strings.IndexByte(out, 0x1b) != -1 {
		t.Fatalf("piped filtered run must be ANSI-free, got: %q", out)
	}
}

// TestDoctorRuleNotFiredExitsZero covers the focused view when the selected rule
// did not fire: the empty-match note and exit 0.
func TestDoctorRuleNotFiredExitsZero(t *testing.T) {
	repo := initTempRepo(t)
	_, unfired := firedAndUnfired(t, repo)

	out, err := runDoctor(t, "--path", repo, "--rule", unfired)
	if err != nil {
		t.Fatalf("--rule %s should exit 0 (no finding), got: %v", unfired, err)
	}
	if !strings.Contains(out, "charter doctor — filtered: "+unfired) {
		t.Fatalf("expected a filtered header for %s, got:\n%s", unfired, out)
	}
	if !strings.Contains(out, "no findings for the selected rule(s)") {
		t.Fatalf("expected the empty-match note for %s, got:\n%s", unfired, out)
	}
	if strings.Contains(out, "score:") {
		t.Fatalf("filtered run must omit the score, got:\n%s", out)
	}
}

// TestDoctorRuleMultiID covers a comma-separated --rule: the header lists both
// IDs, the fired rule is shown, and the run exits 1 because a named rule fired.
func TestDoctorRuleMultiID(t *testing.T) {
	repo := initTempRepo(t)
	fired, unfired := firedAndUnfired(t, repo)

	out, err := runDoctor(t, "--path", repo, "--rule", fired+","+unfired)
	if code := exitCodeOf(t, err); code != 1 {
		t.Fatalf("multi --rule exit code = %d, want 1", code)
	}
	if !strings.Contains(out, "filtered: "+fired+", "+unfired) {
		t.Fatalf("expected both IDs in the filtered header, got:\n%s", out)
	}
	if !strings.Contains(out, fired) {
		t.Fatalf("expected the fired finding in the focused view, got:\n%s", out)
	}
}

// TestDoctorRuleUnknownExitsTwo covers an unknown --rule ID (exit 2, lists valid
// IDs).
func TestDoctorRuleUnknownExitsTwo(t *testing.T) {
	repo := initTempRepo(t)
	_, err := runDoctor(t, "--path", repo, "--rule", "AE-NOPE-999")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("unknown --rule exit code = %d, want 2", code)
	}
	if !strings.Contains(err.Error(), "unknown rule") || !strings.Contains(err.Error(), "AE-SEC-001") {
		t.Fatalf("expected unknown-rule guidance listing valid IDs, got: %v", err)
	}
}

// TestDoctorRuleEmptyListExitsTwo covers a --rule value that contains no real
// IDs (only separators).
func TestDoctorRuleEmptyListExitsTwo(t *testing.T) {
	repo := initTempRepo(t)
	_, err := runDoctor(t, "--path", repo, "--rule", ",,")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("empty --rule list exit code = %d, want 2", code)
	}
}

// TestDoctorRuleRejectsNonTextFormat covers the contract that --rule is
// text-only (json/sarif/markdown carry the score by contract).
func TestDoctorRuleRejectsNonTextFormat(t *testing.T) {
	repo := initTempRepo(t)
	_, err := runDoctor(t, "--path", repo, "--rule", "AE-SEC-001", "--format", "json")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("--rule with --format json exit code = %d, want 2", code)
	}
}

// TestDoctorQuietRuleRejected covers the rejected --quiet + --rule combination:
// --quiet is a whole-repo CI gate and --rule is a focused, score-free view, so
// they don't compose and the command must exit 2 with a clear message.
func TestDoctorQuietRuleRejected(t *testing.T) {
	repo := initTempRepo(t)
	_, err := runDoctor(t, "--path", repo, "--rule", "AE-SEC-001", "--quiet")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("--quiet + --rule exit code = %d, want 2", code)
	}
	if !strings.Contains(err.Error(), "--quiet cannot be combined with --rule") {
		t.Fatalf("expected a clear --quiet/--rule conflict message, got: %v", err)
	}
}

// TestDoctorColorPrecedence covers --color always (styled even when piped),
// --no-color (plain), and --no-color winning over --color=always — using a
// passing fixture so the command itself exits 0.
func TestDoctorColorPrecedence(t *testing.T) {
	repo, ferr := makeTempDoctorRepo(t)
	if ferr != nil {
		t.Fatalf("fixture: %v", ferr)
	}

	always, err := runDoctor(t, "--path", repo, "--color", "always")
	if err != nil {
		t.Fatalf("doctor --color always: %v", err)
	}
	if strings.IndexByte(always, 0x1b) == -1 {
		t.Fatalf("--color always must force styled output (ANSI) even when piped, got: %q", always)
	}

	off, err := runDoctor(t, "--path", repo, "--no-color")
	if err != nil {
		t.Fatalf("doctor --no-color: %v", err)
	}
	if strings.IndexByte(off, 0x1b) != -1 {
		t.Fatalf("--no-color must produce zero ANSI, got: %q", off)
	}

	conflict, err := runDoctor(t, "--path", repo, "--color", "always", "--no-color")
	if err != nil {
		t.Fatalf("doctor --color always --no-color: %v", err)
	}
	if strings.IndexByte(conflict, 0x1b) != -1 {
		t.Fatalf("--no-color must win over --color=always (zero ANSI), got: %q", conflict)
	}
}

// TestDoctorInvalidColorExitsTwo covers --color validation.
func TestDoctorInvalidColorExitsTwo(t *testing.T) {
	repo, ferr := makeTempDoctorRepo(t)
	if ferr != nil {
		t.Fatalf("fixture: %v", ferr)
	}
	_, err := runDoctor(t, "--path", repo, "--color", "chartreuse")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("invalid --color exit code = %d, want 2", code)
	}
}

// TestDoctorRuleOutWritesFile covers the focused view written to --out: the
// file holds the filtered (score-free) report and the run still exits 1.
func TestDoctorRuleOutWritesFile(t *testing.T) {
	repo := initTempRepo(t)
	fired, _ := firedAndUnfired(t, repo)
	outPath := filepath.Join(t.TempDir(), "focused.txt")

	stdout, err := runDoctor(t, "--path", repo, "--rule", fired, "--out", outPath)
	if code := exitCodeOf(t, err); code != 1 {
		t.Fatalf("filtered --out exit code = %d, want 1", code)
	}
	if stdout != "" {
		t.Fatalf("expected nothing on stdout when --out is set, got %q", stdout)
	}
	data, rerr := os.ReadFile(outPath)
	if rerr != nil {
		t.Fatalf("expected %s written: %v", outPath, rerr)
	}
	if !strings.Contains(string(data), "filtered: "+fired) {
		t.Fatalf("expected the filtered header in the file, got:\n%s", data)
	}
	if strings.Contains(string(data), "score:") {
		t.Fatalf("filtered --out report must omit the score, got:\n%s", data)
	}
}

// TestDoctorColorAlwaysWithFilteredView confirms the focused view also honors
// --color always (styled focused output).
func TestDoctorColorAlwaysFilteredStyled(t *testing.T) {
	repo := initTempRepo(t)
	fired, _ := firedAndUnfired(t, repo)
	out, err := runDoctor(t, "--path", repo, "--rule", fired, "--color", "always")
	if code := exitCodeOf(t, err); code != 1 {
		t.Fatalf("filtered --color always exit code = %d, want 1", code)
	}
	if strings.IndexByte(out, 0x1b) == -1 {
		t.Fatalf("filtered --color always must emit ANSI, got: %q", out)
	}
}
