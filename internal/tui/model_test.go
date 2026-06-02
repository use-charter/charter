package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/suppress"
	"go.use-charter.dev/charter/internal/terminal"
)

// sampleResult builds a deterministic doctor.Result with a finding of each
// severity plus an informational finding and a suppressed finding, so the
// filter/sort/selection transitions can be asserted without a real scan.
func sampleResult() doctor.Result {
	active := []findings.Finding{
		{RuleID: "AE-SEC-001", Severity: findings.SeverityBlocker, Category: "Secrets", Summary: "raw secret in context", Remediation: "remove the secret", Evidence: []string{"key=****"}, Locations: []findings.Location{{Path: "app.env", Line: 3}}},
		{RuleID: "AE-CTX-001", Severity: findings.SeverityHigh, Category: "Context", Summary: "weak agent context", Remediation: "expand AGENTS.md", Locations: []findings.Location{{Path: "AGENTS.md"}}},
		{RuleID: "AE-CI-002", Severity: findings.SeverityMedium, Category: "CI", Summary: "charter not in CI", Remediation: "add CI step"},
		{RuleID: "AE-CTX-006", Severity: findings.SeverityLow, Category: "Context", Summary: "over-emphatic instructions", Remediation: "tone it down", Informational: true},
	}
	score := scoring.Calculate(active)
	return doctor.Result{
		Root:      "/tmp/repo",
		Threshold: 80,
		Passed:    score.Final >= 80,
		Findings:  active,
		Suppressed: []suppress.Suppressed{
			{Finding: findings.Finding{RuleID: "AE-MCP-001", Severity: findings.SeverityHigh, Category: "MCP Safety", Summary: "unpinned MCP server"}, Source: suppress.SourceExternal, Reason: "tracked in JIRA-1"},
		},
		Score: score,
	}
}

func testCaps() (terminal.Capabilities, terminal.Palette) {
	caps := terminal.Detect(terminal.Env{ColorTerm: "truecolor", Term: "xterm-256color"}, true, terminal.ColorAuto)
	return caps, terminal.NewPalette(caps, true)
}

func newTestModel(t *testing.T) Model {
	t.Helper()
	caps, pal := testCaps()
	return New(sampleResult(), nil, caps, pal)
}

// step drives Update once and casts the result back to the concrete Model.
func step(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	got, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want tui.Model", next)
	}
	return got, cmd
}

func runeKey(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }
func codeKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func TestNewSeedsDefaultView(t *testing.T) {
	m := newTestModel(t)
	// Default view hides muted (informational + suppressed): 3 scored findings.
	if len(m.filtered) != 3 {
		t.Fatalf("default filtered = %d, want 3", len(m.filtered))
	}
	// Default sort is severity-desc, so the blocker leads.
	if got := m.filtered[0].finding.RuleID; got != "AE-SEC-001" {
		t.Fatalf("default first row = %q, want AE-SEC-001", got)
	}
	if m.selected != 0 {
		t.Fatalf("default selection = %d, want 0", m.selected)
	}
}

func TestSeverityFilterNarrowsList(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('1')) // 1 → BLOCKER
	if m.sevFilter != findings.SeverityBlocker {
		t.Fatalf("sevFilter = %q, want BLOCKER", m.sevFilter)
	}
	if len(m.filtered) != 1 || m.filtered[0].finding.RuleID != "AE-SEC-001" {
		t.Fatalf("blocker filter = %+v, want [AE-SEC-001]", ruleIDs(m.filtered))
	}
	// Pressing the active severity again clears the filter.
	m, _ = step(t, m, runeKey('1'))
	if m.sevFilter != "" {
		t.Fatalf("re-pressing severity should clear it, got %q", m.sevFilter)
	}
	if len(m.filtered) != 3 {
		t.Fatalf("cleared filter = %d rows, want 3", len(m.filtered))
	}
}

func TestSearchNarrowsList(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('/'))
	if !m.searching {
		t.Fatalf("'/' should enter search mode")
	}
	for _, r := range "sec" {
		m, _ = step(t, m, runeKey(r))
	}
	if m.query != "sec" {
		t.Fatalf("query = %q, want sec", m.query)
	}
	if len(m.filtered) != 1 || m.filtered[0].finding.RuleID != "AE-SEC-001" {
		t.Fatalf("search 'sec' = %v, want [AE-SEC-001]", ruleIDs(m.filtered))
	}
	// Enter commits the query and exits search mode.
	m, _ = step(t, m, codeKey(tea.KeyEnter))
	if m.searching {
		t.Fatalf("enter should exit search mode")
	}
	if m.query != "sec" {
		t.Fatalf("query after enter = %q, want sec", m.query)
	}
	// Re-enter and esc clears the query.
	m, _ = step(t, m, runeKey('/'))
	m, _ = step(t, m, codeKey(tea.KeyEsc))
	if m.searching || m.query != "" {
		t.Fatalf("esc should cancel+clear search, got searching=%v query=%q", m.searching, m.query)
	}
	if len(m.filtered) != 3 {
		t.Fatalf("cleared search = %d rows, want 3", len(m.filtered))
	}
}

func TestNavigationMovesSelectionAndDetail(t *testing.T) {
	m := newTestModel(t)
	if got := m.detail.GetContent(); !strings.Contains(got, "AE-SEC-001") {
		t.Fatalf("initial detail should show AE-SEC-001, got:\n%s", got)
	}
	m, _ = step(t, m, runeKey('j')) // down
	if m.selected != 1 {
		t.Fatalf("after j, selected = %d, want 1", m.selected)
	}
	if got := m.detail.GetContent(); !strings.Contains(got, "AE-CTX-001") {
		t.Fatalf("after j, detail should show AE-CTX-001, got:\n%s", got)
	}
	m, _ = step(t, m, runeKey('k')) // back up
	if m.selected != 0 {
		t.Fatalf("after k, selected = %d, want 0", m.selected)
	}
}

func TestMutedToggleRevealsSuppressedAndInformational(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('s'))
	if !m.showMuted {
		t.Fatalf("'s' should enable showMuted")
	}
	if len(m.filtered) != 5 {
		t.Fatalf("with muted shown = %d rows, want 5 (%v)", len(m.filtered), ruleIDs(m.filtered))
	}
}

func TestCategoryCycleFiltersByCategory(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('c')) // first category alphabetically = "CI"
	if m.catFilter != "CI" {
		t.Fatalf("category filter = %q, want CI", m.catFilter)
	}
	if len(m.filtered) != 1 || m.filtered[0].finding.RuleID != "AE-CI-002" {
		t.Fatalf("CI filter = %v, want [AE-CI-002]", ruleIDs(m.filtered))
	}
}

func TestSortToggleGroupsByCategory(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('o'))
	if m.sort != sortByCategory {
		t.Fatalf("'o' should switch to sortByCategory")
	}
	// By category ascending: CI < Context < Secrets.
	if got := m.filtered[0].finding.RuleID; got != "AE-CI-002" {
		t.Fatalf("category sort first = %q, want AE-CI-002", got)
	}
}

func TestClearFiltersResets(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('1')) // narrow to blocker
	m, _ = step(t, m, codeKey(tea.KeyEsc))
	if m.sevFilter != "" || m.catFilter != "" || m.showMuted || m.query != "" {
		t.Fatalf("esc should clear all filters, got %+v", m)
	}
	if len(m.filtered) != 3 {
		t.Fatalf("after clear = %d rows, want 3", len(m.filtered))
	}
}

func TestRescanTriggersCommandAndRebuilds(t *testing.T) {
	caps, pal := testCaps()
	// Inject a deterministic scan returning a clean result (no findings).
	clean := doctor.Result{Root: "/tmp/repo", Threshold: 80, Passed: true, Score: scoring.Calculate(nil)}
	m := New(sampleResult(), func() (doctor.Result, error) { return clean, nil }, caps, pal)

	m, cmd := step(t, m, runeKey('r'))
	if cmd == nil {
		t.Fatalf("'r' should return a rescan command")
	}
	if m.status != "rescanning…" {
		t.Fatalf("status = %q, want rescanning…", m.status)
	}

	msg := cmd()
	done, ok := msg.(rescanDoneMsg)
	if !ok {
		t.Fatalf("rescan command produced %T, want rescanDoneMsg", msg)
	}
	m, _ = step(t, m, done)
	if len(m.items) != 0 {
		t.Fatalf("after rescan items = %d, want 0", len(m.items))
	}
	if m.status != "rescanned" {
		t.Fatalf("status = %q, want rescanned", m.status)
	}
	if !m.result.Passed {
		t.Fatalf("rescan result should be the injected clean (passing) result")
	}
}

func TestRescanWithoutScanFuncIsNoop(t *testing.T) {
	m := newTestModel(t) // scan == nil
	m, cmd := step(t, m, runeKey('r'))
	if cmd != nil {
		t.Fatalf("rescan with no scan func should not return a command")
	}
	if m.status != "rescan unavailable" {
		t.Fatalf("status = %q, want rescan unavailable", m.status)
	}
}

func TestQuitKeyQuits(t *testing.T) {
	m := newTestModel(t)
	m, cmd := step(t, m, runeKey('q'))
	if !m.quitting {
		t.Fatalf("'q' should set quitting")
	}
	if cmd == nil {
		t.Fatalf("'q' should return the quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("quit command should produce tea.QuitMsg")
	}
}

func TestCopySelectionSetsClipboard(t *testing.T) {
	m := newTestModel(t)
	// Default selection is the blocker, which has app.env:3.
	m, cmd := step(t, m, runeKey('y'))
	if cmd == nil {
		t.Fatalf("'y' on a located finding should return a clipboard command")
	}
	if !strings.Contains(m.status, "app.env:3") {
		t.Fatalf("status = %q, want it to mention app.env:3", m.status)
	}
}

func TestCopySelectionWithoutLocation(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('3')) // MEDIUM → AE-CI-002, which has no location
	m, cmd := step(t, m, runeKey('y'))
	if cmd != nil {
		t.Fatalf("'y' with no location should not return a command")
	}
	if !strings.Contains(m.status, "no path:line") {
		t.Fatalf("status = %q, want no path:line", m.status)
	}
}

func TestFocusToggleAndDetailScroll(t *testing.T) {
	m := newTestModel(t)
	if m.focus != focusList {
		t.Fatalf("initial focus should be the list")
	}
	m, _ = step(t, m, codeKey(tea.KeyTab))
	if m.focus != focusDetail {
		t.Fatalf("tab should switch focus to the detail pane")
	}
	// A nav key in detail focus must not move the list selection.
	m, _ = step(t, m, runeKey('j'))
	if m.selected != 0 {
		t.Fatalf("detail-focused nav should not change list selection, got %d", m.selected)
	}
	m, _ = step(t, m, codeKey(tea.KeyTab))
	if m.focus != focusList {
		t.Fatalf("tab should switch focus back to the list")
	}
}

func TestHelpToggle(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('?'))
	if !m.showHelp || !m.help.ShowAll {
		t.Fatalf("'?' should enable full help")
	}
	m, _ = step(t, m, runeKey('?'))
	if m.showHelp || m.help.ShowAll {
		t.Fatalf("'?' should toggle full help off")
	}
}

func TestWindowResizeRelayouts(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.width != 120 || m.height != 40 {
		t.Fatalf("resize not applied: %dx%d", m.width, m.height)
	}
	if v := m.View(); v.Content == "" {
		t.Fatalf("View should render content after resize")
	}
}

func TestViewRendersHeaderAndDetail(t *testing.T) {
	m := newTestModel(t)
	content := m.View().Content
	for _, want := range []string{"charter", "Score", "AE-SEC-001"} {
		if !strings.Contains(content, want) {
			t.Fatalf("view missing %q in:\n%s", want, content)
		}
	}
}

func TestEmptyFilterShowsEmptyDetail(t *testing.T) {
	m := newTestModel(t)
	m, _ = step(t, m, runeKey('/'))
	for _, r := range "zzzznomatch" {
		m, _ = step(t, m, runeKey(r))
	}
	if len(m.filtered) != 0 {
		t.Fatalf("no-match search should empty the list, got %d", len(m.filtered))
	}
	if _, ok := m.selectedItem(); ok {
		t.Fatalf("no selection should exist for an empty list")
	}
	if got := m.detail.GetContent(); !strings.Contains(got, "No findings match") {
		t.Fatalf("empty list detail = %q", got)
	}
}

// TestNewWithCleanResultShowsCleanDetail covers the zero-findings path: the
// list is empty, there is no selection, the detail pane shows the clean-scan
// copy, and the view still renders (headerHeight's no-scorecard branch).
func TestNewWithCleanResultShowsCleanDetail(t *testing.T) {
	caps, pal := testCaps()
	clean := doctor.Result{Root: "/tmp/repo", Threshold: 80, Passed: true, Score: scoring.Calculate(nil)}
	m := New(clean, nil, caps, pal)

	if len(m.items) != 0 {
		t.Fatalf("clean result items = %d, want 0", len(m.items))
	}
	if _, ok := m.selectedItem(); ok {
		t.Fatalf("clean result should have no selection")
	}
	if got := m.detail.GetContent(); !strings.Contains(got, "No findings — clean scan.") {
		t.Fatalf("clean-scan detail = %q, want the clean-scan copy", got)
	}
	if v := m.View(); v.Content == "" {
		t.Fatalf("clean result View should still render content")
	}
}

// TestRenderDetailCatalogPresence covers both arms of the catalog.Lookup branch
// in the detail pane: an unknown rule omits the "why this matters" block, while
// a catalogued rule renders it.
func TestRenderDetailCatalogPresence(t *testing.T) {
	caps, pal := testCaps()

	unknown := findings.Finding{RuleID: "AE-NOPE-000", Severity: findings.SeverityHigh, Category: "Context", Summary: "synthetic finding"}
	mUnknown := New(doctor.Result{Root: "/tmp/repo", Threshold: 80, Findings: []findings.Finding{unknown}, Score: scoring.Calculate([]findings.Finding{unknown})}, nil, caps, pal)
	got := mUnknown.detail.GetContent()
	if !strings.Contains(got, "AE-NOPE-000") {
		t.Fatalf("detail should show the selected finding, got:\n%s", got)
	}
	if strings.Contains(got, "why this matters") {
		t.Fatalf("unknown rule must omit the catalog block, got:\n%s", got)
	}

	known := findings.Finding{RuleID: "AE-SEC-001", Severity: findings.SeverityBlocker, Category: "Secrets", Summary: "raw secret"}
	mKnown := New(doctor.Result{Root: "/tmp/repo", Threshold: 80, Findings: []findings.Finding{known}, Score: scoring.Calculate([]findings.Finding{known})}, nil, caps, pal)
	if got := mKnown.detail.GetContent(); !strings.Contains(got, "why this matters") {
		t.Fatalf("catalogued rule must render the catalog block, got:\n%s", got)
	}
}

// ruleIDs is a tiny assertion helper.
func ruleIDs(items []item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.finding.RuleID
	}
	return out
}
