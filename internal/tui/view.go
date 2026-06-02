package tui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/rules/catalog"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/terminal"
	"go.use-charter.dev/charter/internal/version"
)

// Layout defaults and bounds. The browser starts at these dimensions so the
// view and the unit tests render before the first WindowSizeMsg.
const (
	defaultWidth  = 96
	defaultHeight = 30

	minBodyHeight = 4
	minListWidth  = 28
	scoreBarWidth = 18

	// brandMark is Charter's committed ASCII wordmark glyph; it renders on every
	// terminal tier, matching internal/render/text.
	brandMark = "[C]"
)

// glyphs is the icon set (box-drawing + status marks) with an ASCII fallback
// for the poorest terminal tiers. The glyph shapes mirror internal/render/text
// so the TUI and the static renderer stay visually consistent.
type glyphs struct {
	divider  string
	bar      string
	pass     string
	fail     string
	warn     string
	bullet   string
	barFull  string
	barEmpty string
}

var (
	unicodeGlyphs = glyphs{divider: "─", bar: "│", pass: "✓", fail: "✗", warn: "⚠", bullet: "•", barFull: "█", barEmpty: "░"}
	asciiGlyphs   = glyphs{divider: "-", bar: "|", pass: "+", fail: "x", warn: "!", bullet: "-", barFull: "#", barEmpty: "-"}
)

// theme carries the resolved presentation context for the browser: the
// capabilities, the palette, and the active glyph set.
type theme struct {
	caps    terminal.Capabilities
	pal     terminal.Palette
	unicode bool
	g       glyphs
}

func newTheme(caps terminal.Capabilities, pal terminal.Palette) theme {
	// Box-drawing and status glyphs need UTF-8, not 24-bit color, so any
	// 256-color-or-richer terminal gets the Unicode set; only ANSI16/Mono fall
	// back to ASCII. (This is deliberately looser than internal/render/text's
	// TrueColor-only gate — the interactive view leans on box-drawing glyphs.)
	t := theme{caps: caps, pal: pal, unicode: caps.Tier >= terminal.ANSI256}
	if t.unicode {
		t.g = unicodeGlyphs
	} else {
		t.g = asciiGlyphs
	}
	return t
}

// style turns a palette token into a lipgloss style, applying color only when
// the token carries one and always carrying the attribute fallbacks so
// hierarchy survives on poorer color tiers.
func (t theme) style(tok terminal.Token) lipgloss.Style {
	resolved := t.pal.Resolve(tok)
	st := lipgloss.NewStyle()
	if resolved.HasColor() {
		st = st.Foreground(resolved.Color)
	}
	if resolved.Bold {
		st = st.Bold(true)
	}
	if resolved.Faint {
		st = st.Faint(true)
	}
	if resolved.Reverse {
		st = st.Reverse(true)
	}
	return st
}

// severityTextToken maps a severity to its text palette token, matching the
// design + internal/render/text: BLOCKER → danger, HIGH/MEDIUM → warning,
// LOW (and anything else) → info.
func severityTextToken(sev findings.Severity) terminal.Token {
	switch sev {
	case findings.SeverityBlocker:
		return terminal.TextDanger
	case findings.SeverityHigh, findings.SeverityMedium:
		return terminal.TextWarning
	default:
		return terminal.TextInfo
	}
}

// tableStyles styles the findings table from the palette: a bold accent for the
// selected row, faint dividers for headers, no extra cell color (severity color
// is carried by the row content itself).
func (m Model) tableStyles() table.Styles {
	header := m.th.style(terminal.TextTertiary).Bold(true).Padding(0, 1)
	cell := lipgloss.NewStyle().Padding(0, 1)
	selected := m.th.style(terminal.TextInfo).Bold(true).Reverse(true)
	return table.Styles{Header: header, Cell: cell, Selected: selected}
}

// View implements tea.Model. It composes header + body (list | detail) + footer
// into the alternate screen with synchronized output (the v2 default).
func (m Model) View() tea.View {
	var content string
	if !m.ready || m.width <= 0 || m.height <= 0 {
		content = m.th.style(terminal.TextSecondary).Render(brandMark + " charter — loading…")
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			m.headerView(),
			m.bodyView(),
			m.footerView(),
		)
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// layout sizes the sub-components from the current terminal dimensions. The
// findings list takes ~40% of the width; the detail pane takes the rest; both
// share the body height between the header and footer. Heights come from state
// (headerHeight/footerHeight) rather than rendering, so layout stays in the
// model layer and never calls back into the view.
func (m Model) layout() Model {
	if !m.ready {
		return m
	}
	bodyH := m.height - m.headerHeight() - m.footerHeight()
	if bodyH < minBodyHeight {
		bodyH = minBodyHeight
	}
	innerH := bodyH - 2 // each pane box has a top + bottom border row
	if innerH < 1 {
		innerH = 1
	}

	const paneBorders = 4 // two pane boxes, one column of border on each side
	listW := (m.width - paneBorders) * 2 / 5
	if listW < minListWidth {
		listW = minListWidth
	}
	detailW := m.width - paneBorders - listW
	if detailW < minListWidth {
		detailW = minListWidth
	}

	m.table.SetColumns(m.columns(listW))
	m.table.SetWidth(listW)
	m.table.SetHeight(innerH)

	m.detail.SetWidth(detailW)
	m.detail.SetHeight(innerH)

	m.help.SetWidth(m.width)
	m.search.SetWidth(max(8, m.width-4))
	return m
}

// headerHeight is the line count of headerView, derived from state to avoid a
// render pass in layout. headerView is the brand line, the score line, and the
// divider (3 rows), plus the per-category scorecard row whenever the result
// carries any findings (scorecardLine is empty only for a zero-finding result).
func (m Model) headerHeight() int {
	h := 3
	if len(m.result.Findings) > 0 {
		h++
	}
	return h
}

// footerHeight is the line count of footerView, derived from state. The footer
// is the status/search line (always one row) above the help keybar: one row for
// the compact short help, or the tallest FullHelp column when full help is
// shown. It reads the binding groups directly (help's own layout source) rather
// than rendering; width-based truncation can only shorten the real footer, so
// the computed height never under-sizes the body into an overlap.
func (m Model) footerHeight() int {
	helpRows := 1
	if m.showHelp {
		for _, col := range m.keys.FullHelp() {
			if len(col) > helpRows {
				helpRows = len(col)
			}
		}
	}
	return 1 + helpRows
}

// columns derives the findings table columns to fit the list width. Each cell
// carries one column of padding on each side (Padding(0,1)), so the four
// columns plus their padding sum to the list width; the summary column absorbs
// the remainder.
func (m Model) columns(listW int) []table.Column {
	const (
		flagW = 1
		sevW  = 7
		ruleW = 12
		pad   = 2 * 4 // four columns, two padding cells each
	)
	summaryW := listW - flagW - sevW - ruleW - pad
	if summaryW < 6 {
		summaryW = 6
	}
	return []table.Column{
		{Title: "", Width: flagW},
		{Title: "Sev", Width: sevW},
		{Title: "Rule", Width: ruleW},
		{Title: "Summary", Width: summaryW},
	}
}

// rowFor renders one finding as a table row. The first column flags muted rows
// (~ suppressed, i informational) so they are distinguishable without relying
// on color; the severity column always carries its text label (WCAG 1.4.1).
func (m Model) rowFor(it item) table.Row {
	flag := " "
	switch it.kind {
	case kindSuppressed:
		flag = "~"
	case kindInformational:
		flag = "i"
	case kindActive:
		flag = " "
	}
	return table.Row{flag, string(it.finding.Severity), it.finding.RuleID, it.finding.Summary}
}

// headerView renders the brand line, the score hero, and the per-category
// scorecard. It uses the active result so the header reflects rescans.
func (m Model) headerView() string {
	brand := m.th.style(terminal.TextInfo).Bold(true).Render(brandMark + " charter")
	meta := m.th.style(terminal.TextSecondary).Render(fmt.Sprintf("  %s · %s", version.Version(), m.result.Root))

	scoreTok := terminal.TextDanger
	verdict := "FAIL"
	if m.result.Passed {
		scoreTok = terminal.TextSuccess
		verdict = "PASS"
	}
	if m.unicodeVerdict() {
		if m.result.Passed {
			verdict += " " + m.th.g.pass
		} else {
			verdict += " " + m.th.g.fail
		}
	}
	scoreLine := m.th.style(terminal.TextSecondary).Render("Score ") +
		m.th.style(scoreTok).Bold(true).Render(strconv.Itoa(m.result.Score.Final)) +
		m.th.style(terminal.TextTertiary).Render("/100  ") +
		m.scoreBar(m.result.Score.Final, scoreTok) + "  " +
		m.th.style(scoreTok).Bold(true).Render(verdict) +
		m.th.style(terminal.TextTertiary).Render(fmt.Sprintf("   threshold %d", m.result.Threshold))

	lines := []string{brand + meta, scoreLine}
	if card := m.scorecardLine(); card != "" {
		lines = append(lines, card)
	}
	lines = append(lines, m.th.style(terminal.TextTertiary).Render(strings.Repeat(m.th.g.divider, max(8, m.width))))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// scorecardLine renders the per-category readiness breakdown as a single
// compact line of "Category −D" cells colored by the category's worst severity.
func (m Model) scorecardLine() string {
	breakdown := scoring.ByCategory(m.result.Findings)
	if len(breakdown) == 0 {
		return ""
	}
	cells := make([]string, 0, len(breakdown))
	for _, c := range breakdown {
		tok := severityTextToken(c.WorstSeverity)
		cells = append(cells, m.th.style(terminal.TextSecondary).Render(c.Category+" ")+
			m.th.style(tok).Render(fmt.Sprintf("−%d", c.Deduction)))
	}
	sep := m.th.style(terminal.TextTertiary).Render(" · ")
	return strings.Join(cells, sep)
}

// scoreBar renders a fixed-width progress bar proportional to score/100.
func (m Model) scoreBar(score int, tok terminal.Token) string {
	filled := max(0, min(scoreBarWidth, score*scoreBarWidth/100))
	var b strings.Builder
	if filled > 0 {
		b.WriteString(m.th.style(tok).Render(strings.Repeat(m.th.g.barFull, filled)))
	}
	if filled < scoreBarWidth {
		b.WriteString(m.th.style(terminal.TextTertiary).Render(strings.Repeat(m.th.g.barEmpty, scoreBarWidth-filled)))
	}
	return b.String()
}

func (m Model) unicodeVerdict() bool { return m.th.unicode }

// bodyView composes the findings list and detail pane side by side, each in its
// own box whose border brightens when that pane holds focus.
func (m Model) bodyView() string {
	list := m.paneBox(focusList).Render(m.table.View())
	detail := m.paneBox(focusDetail).Render(m.detail.View())
	return lipgloss.JoinHorizontal(lipgloss.Top, list, detail)
}

// paneBox returns the bordered box style for a pane, accenting the border of
// the focused pane with the info token and leaving the other subdued.
func (m Model) paneBox(area focusArea) lipgloss.Style {
	st := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	tok := terminal.BorderTertiary
	if m.focus == area {
		tok = terminal.BorderInfo
	}
	if resolved := m.pal.Resolve(tok); resolved.HasColor() {
		st = st.BorderForeground(resolved.Color)
	}
	return st
}

// footerView renders the status/search line above the help keybar.
func (m Model) footerView() string {
	var top string
	switch {
	case m.searching:
		top = m.search.View()
	case m.status != "":
		top = m.th.style(terminal.TextInfo).Render(m.status)
	default:
		top = m.th.style(terminal.TextTertiary).Render(m.filterSummary())
	}
	return lipgloss.JoinVertical(lipgloss.Left, top, m.help.View(m.keys))
}

// filterSummary describes the active filters and the visible/total counts so
// the footer always communicates what subset of findings is shown.
func (m Model) filterSummary() string {
	parts := []string{fmt.Sprintf("%d/%d findings", len(m.filtered), len(m.items))}
	if m.sevFilter != "" {
		parts = append(parts, "severity "+string(m.sevFilter))
	}
	if m.catFilter != "" {
		parts = append(parts, "category "+m.catFilter)
	}
	if m.showMuted {
		parts = append(parts, "+suppressed")
	}
	if m.query != "" {
		parts = append(parts, "search "+strconv.Quote(m.query))
	}
	if m.sort == sortByCategory {
		parts = append(parts, "by category")
	}
	return strings.Join(parts, " · ")
}

// renderEmptyDetail is shown when no finding is selected (the filters matched
// nothing).
func (m Model) renderEmptyDetail() string {
	if len(m.items) == 0 {
		return m.th.style(terminal.TextSuccess).Render(m.th.g.pass + " No findings — clean scan.")
	}
	return m.th.style(terminal.TextTertiary).Render("No findings match the active filters. Press esc to clear.")
}

// renderDetail renders the detail pane for one finding: rule ID, a text
// severity badge, category, summary, evidence (already redacted), remediation,
// each path:line, and the catalog short description ("why this matters").
func (m Model) renderDetail(it item) string {
	f := it.finding
	textTok := severityTextToken(f.Severity)

	var b strings.Builder
	b.WriteString(m.th.style(terminal.TextInfo).Bold(true).Render(f.RuleID))
	if f.Category != "" {
		b.WriteString(m.th.style(terminal.TextTertiary).Render("  " + f.Category))
	}
	b.WriteString("\n")

	badge := m.th.style(textTok).Bold(true).Render(m.severityIcon(f.Severity) + " " + string(f.Severity))
	b.WriteString(badge)
	if it.kind == kindInformational {
		b.WriteString(m.th.style(terminal.TextTertiary).Render("  (informational — not scored)"))
	}
	if it.kind == kindSuppressed {
		b.WriteString(m.th.style(terminal.TextTertiary).Render("  (suppressed)"))
	}
	b.WriteString("\n\n")

	b.WriteString(m.th.style(terminal.TextPrimary).Render(f.Summary))
	b.WriteString("\n")

	if len(f.Locations) > 0 {
		b.WriteString("\n" + m.detailLabel("locations"))
		for _, loc := range f.Locations {
			b.WriteString("\n  " + m.th.style(terminal.TextSecondary).Render(locationText(loc)))
		}
		b.WriteString("\n")
	}

	if len(f.Evidence) > 0 {
		b.WriteString("\n" + m.detailLabel("evidence"))
		for _, ev := range f.Evidence {
			b.WriteString("\n  " + m.th.style(terminal.TextTertiary).Render(m.th.g.bullet+" ") +
				m.th.style(terminal.TextSecondary).Render(ev))
		}
		b.WriteString("\n")
	}

	if f.Remediation != "" {
		b.WriteString("\n" + m.detailLabel("remediation"))
		b.WriteString("\n  " + m.th.style(terminal.TextSecondary).Render(f.Remediation))
		b.WriteString("\n")
	}

	if it.kind == kindSuppressed && it.reason != "" {
		b.WriteString("\n" + m.detailLabel("suppression"))
		b.WriteString("\n  " + m.th.style(terminal.TextTertiary).Render(it.suppressionDetail()))
		b.WriteString("\n")
	}

	if entry, ok := catalog.Lookup(f.RuleID); ok {
		b.WriteString("\n" + m.detailLabel("why this matters"))
		b.WriteString("\n  " + m.th.style(terminal.TextSecondary).Render(entry.ShortDescription))
		b.WriteString("\n  " + m.th.style(terminal.TextTertiary).Render(entry.HelpURI))
	}

	return b.String()
}

// suppressionDetail composes the human-readable suppression line for the detail
// pane (source, reason, approver, expiry) from the item's metadata.
func (it item) suppressionDetail() string {
	parts := []string{it.source}
	if it.reason != "" {
		parts = append(parts, "reason: "+it.reason)
	}
	if it.approver != "" {
		parts = append(parts, "approver: "+it.approver)
	}
	if it.expires != "" {
		parts = append(parts, "expires: "+it.expires)
	}
	return strings.Join(parts, " · ")
}

// detailLabel renders a faint section label in the detail pane.
func (m Model) detailLabel(name string) string {
	return m.th.style(terminal.TextTertiary).Bold(true).Render(name)
}

// severityIcon mirrors internal/render/text: BLOCKER is the danger cross,
// HIGH/MEDIUM the warning triangle, everything else the neutral bullet.
func (m Model) severityIcon(sev findings.Severity) string {
	switch sev {
	case findings.SeverityBlocker:
		return m.th.g.fail
	case findings.SeverityHigh, findings.SeverityMedium:
		return m.th.g.warn
	default:
		return m.th.g.bullet
	}
}

func locationText(loc findings.Location) string {
	if loc.Line > 0 {
		return fmt.Sprintf("%s:%d", loc.Path, loc.Line)
	}
	return loc.Path
}
