// Package text renders a doctor.Result for terminal output.
//
// It has two paths, chosen solely by caps.ColorEnabled():
//
//   - The plain path (color disabled: non-TTY, NO_COLOR, TERM=dumb, or the Mono
//     tier) is byte-for-byte identical to Charter's historical `charter doctor`
//     text output. It NEVER touches lipgloss or emits an ANSI escape byte. This
//     is the load-bearing containment contract of Slice 15.
//   - The styled path (color enabled, i.e. a real TTY) uses
//     charm.land/lipgloss/v2 together with the internal/terminal palette to draw
//     a brand header, per-finding cards, a findings summary, a readiness
//     scorecard, and a score hero (with a progress bar and a cap notice).
//
// The package re-uses internal/terminal for capability and palette values; it
// performs no capability detection of its own and no I/O (the caller owns the
// TTY boundary and writes the returned bytes).
package text

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/rules/catalog"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/terminal"
	"go.use-charter.dev/charter/internal/version"
)

// Render projects a doctor.Result into terminal bytes. When caps reports color
// disabled it returns the exact plain format (no ANSI); otherwise it returns a
// styled render built from the palette.
func Render(result doctor.Result, caps terminal.Capabilities, pal terminal.Palette) []byte {
	if !caps.ColorEnabled() {
		return renderPlain(result)
	}
	return renderStyled(result, caps, pal)
}

// RenderFocused renders a filtered projection of result limited to the findings
// whose RuleID is in ruleIDs (ruleIDs also drives the filtered header, in the
// caller's order). It is the focused counterpart of Render for `doctor --rule`
// and DELIBERATELY omits the score and the readiness scorecard: a partial view
// must never imply a 0–100 verdict. Like Render it branches solely on
// caps.ColorEnabled() — plain (ANSI-free) when disabled, styled otherwise.
func RenderFocused(result doctor.Result, ruleIDs []string, caps terminal.Capabilities, pal terminal.Palette) []byte {
	if !caps.ColorEnabled() {
		return renderFocusedPlain(result, ruleIDs)
	}
	return renderFocusedStyled(result, ruleIDs, caps, pal)
}

// filterFindings keeps the findings whose RuleID is in ruleIDs, preserving the
// scan order so the focused view stays deterministic.
func filterFindings(all []findings.Finding, ruleIDs []string) []findings.Finding {
	want := make(map[string]bool, len(ruleIDs))
	for _, id := range ruleIDs {
		want[id] = true
	}
	matched := make([]findings.Finding, 0, len(all))
	for _, f := range all {
		if want[f.RuleID] {
			matched = append(matched, f)
		}
	}
	return matched
}

// focusedNote is shown when none of the selected rules produced a finding.
const focusedNote = "no findings for the selected rule(s)"

func renderFocusedPlain(result doctor.Result, ruleIDs []string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "charter doctor — filtered: %s\n", strings.Join(ruleIDs, ", "))
	matched := filterFindings(result.Findings, ruleIDs)
	if len(matched) == 0 {
		fmt.Fprintln(&buf, focusedNote)
		return buf.Bytes()
	}
	for _, f := range matched {
		writePlainFinding(&buf, f)
	}
	return buf.Bytes()
}

func renderFocusedStyled(result doctor.Result, ruleIDs []string, caps terminal.Capabilities, pal terminal.Palette) []byte {
	r := newStyler(caps, pal)

	var b bytes.Buffer
	heading := r.style(terminal.TextInfo).Bold(true).Render(r.g.brand+" charter") +
		r.style(terminal.TextTertiary).Render(" — filtered: ") +
		r.style(terminal.TextInfo).Bold(true).Render(strings.Join(ruleIDs, ", "))
	fmt.Fprintln(&b, heading)
	fmt.Fprintln(&b, r.divider())

	matched := filterFindings(result.Findings, ruleIDs)
	if len(matched) == 0 {
		fmt.Fprintln(&b, r.style(terminal.TextSuccess).Render(focusedNote))
		return b.Bytes()
	}
	for _, f := range matched {
		r.writeFinding(&b, result.Root, f)
	}
	return b.Bytes()
}

// renderPlain reproduces Charter's historical text output verbatim. Any change
// here breaks the byte-identical non-TTY contract; the containment test guards
// it. Keep it a pure projection with the standard library only.
func renderPlain(result doctor.Result) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "charter doctor: %s\n", result.Root)
	for _, finding := range result.Findings {
		writePlainFinding(&buf, finding)
	}
	for _, s := range result.Suppressed {
		fmt.Fprintf(&buf, "suppressed: %s (%s)", s.Finding.RuleID, s.Source)
		if s.Reason != "" {
			fmt.Fprintf(&buf, " — %s", s.Reason)
		}
		fmt.Fprintln(&buf)
	}
	if breakdown := scoring.ByCategory(result.Findings); len(breakdown) > 0 {
		fmt.Fprintln(&buf, "readiness by category:")
		for _, c := range breakdown {
			fmt.Fprintf(&buf, "  %-12s −%-3d (%d finding(s), worst %s)\n", c.Category, c.Deduction, c.Findings, c.WorstSeverity)
		}
	}
	fmt.Fprintf(&buf, "score: %d (threshold %d)\n", result.Score.Final, result.Threshold)
	return buf.Bytes()
}

// writePlainFinding emits one finding in the historical plain format (rule line,
// locations, evidence, remediation). It is the shared projection used by both
// renderPlain and the focused plain view; changing it changes the byte-identical
// non-TTY contract, which the containment test guards.
func writePlainFinding(buf *bytes.Buffer, finding findings.Finding) {
	fmt.Fprintf(buf, "%s %s %s\n", finding.RuleID, finding.Severity, finding.Summary)
	for _, loc := range finding.Locations {
		if loc.Line > 0 {
			fmt.Fprintf(buf, "  location: %s:%d\n", loc.Path, loc.Line)
		} else {
			fmt.Fprintf(buf, "  location: %s\n", loc.Path)
		}
	}
	for _, evidence := range finding.Evidence {
		fmt.Fprintf(buf, "  - %s\n", evidence)
	}
	fmt.Fprintf(buf, "  remediation: %s\n", finding.Remediation)
}

// dividerWidth is the fixed rule length under the header and above the score.
const dividerWidth = 48

// styler carries the resolved presentation context for one styled render.
type styler struct {
	caps    terminal.Capabilities
	pal     terminal.Palette
	unicode bool
	g       glyphs
}

// glyphs is a small icon set with an ASCII fallback for terminals that are not
// known to be unicode-safe.
type glyphs struct {
	brand    string
	divider  string
	bar      string
	pass     string
	fail     string
	warn     string
	bullet   string
	barFull  string
	barEmpty string
}

// brandMark is Charter's committed wordmark glyph (docs/internal/designs/brand/
// mark.svg). It is plain ASCII so it renders on every terminal, hence it is the
// brand on every tier rather than degrading like the decorative glyphs.
const brandMark = "[C]"

var (
	unicodeGlyphs = glyphs{brand: brandMark, divider: "─", bar: "│", pass: "✓", fail: "✗", warn: "⚠", bullet: "•", barFull: "█", barEmpty: "░"}
	asciiGlyphs   = glyphs{brand: brandMark, divider: "-", bar: "|", pass: "+", fail: "x", warn: "!", bullet: "-", barFull: "#", barEmpty: "-"}
)

// newStyler resolves the presentation context for a styled render: the unicode
// glyph set on TrueColor, the ASCII fallback on every poorer tier so the layout
// stays intact on conservative emulators.
func newStyler(caps terminal.Capabilities, pal terminal.Palette) styler {
	r := styler{caps: caps, pal: pal, unicode: caps.Tier == terminal.TrueColor}
	if r.unicode {
		r.g = unicodeGlyphs
	} else {
		r.g = asciiGlyphs
	}
	return r
}

func renderStyled(result doctor.Result, caps terminal.Capabilities, pal terminal.Palette) []byte {
	r := newStyler(caps, pal)

	var b bytes.Buffer

	r.writeHeader(&b, result)
	r.writeFindings(&b, result)
	r.writeSuppressed(&b, result)
	r.writeSummary(&b, result)
	r.writeScorecard(&b, result)
	r.writeScore(&b, result)

	return b.Bytes()
}

// style turns a resolved palette token into a lipgloss style. Color is applied
// only when the token actually carries one; the attribute fallbacks still apply
// so hierarchy survives on poorer tiers.
func (r styler) style(tok terminal.Token) lipgloss.Style {
	resolved := r.pal.Resolve(tok)
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

// link adds an OSC 8 hyperlink to a style only when the terminal supports them
// and a target is available.
func (r styler) link(st lipgloss.Style, url string) lipgloss.Style {
	if r.caps.Hyperlinks && url != "" {
		return st.Hyperlink(url)
	}
	return st
}

func (r styler) divider() string {
	return r.style(terminal.TextTertiary).Render(strings.Repeat(r.g.divider, dividerWidth))
}

func (r styler) writeHeader(b *bytes.Buffer, result doctor.Result) {
	brand := r.style(terminal.TextInfo).Bold(true).Render(r.g.brand + " charter")
	meta := r.style(terminal.TextSecondary).Render(fmt.Sprintf("  %s · %s", version.Version(), result.Root))
	fmt.Fprintln(b, brand+meta)
	fmt.Fprintln(b, r.divider())
}

func (r styler) writeFindings(b *bytes.Buffer, result doctor.Result) {
	if len(result.Findings) == 0 {
		return
	}
	fmt.Fprintln(b, r.style(terminal.TextSecondary).Render("Findings"))
	for _, f := range result.Findings {
		r.writeFinding(b, result.Root, f)
	}
}

func (r styler) writeFinding(b *bytes.Buffer, root string, f findings.Finding) {
	textTok, borderTok := severityTokens(f.Severity)
	bar := r.style(borderTok).Render(r.g.bar) + " "

	icon := r.style(textTok).Render(r.severityIcon(f.Severity))
	badge := r.style(textTok).Bold(true).Render(string(f.Severity))
	rule := r.link(r.style(terminal.TextInfo).Bold(true), ruleURL(f.RuleID)).Render(f.RuleID)

	header := bar + icon + " " + badge + "  " + rule
	if f.Category != "" {
		header += "  " + r.style(terminal.TextTertiary).Render(f.Category)
	}
	fmt.Fprintln(b, header)

	fmt.Fprintln(b, bar+r.style(terminal.TextPrimary).Render(f.Summary))

	for _, loc := range f.Locations {
		locStyle := r.link(r.style(terminal.TextTertiary), fileURL(root, loc))
		fmt.Fprintln(b, bar+r.label("loc")+locStyle.Render(locationText(loc)))
	}
	for _, ev := range f.Evidence {
		fmt.Fprintln(b, bar+r.style(terminal.TextTertiary).Render(r.g.bullet+" ")+r.style(terminal.TextSecondary).Render(ev))
	}
	fmt.Fprintln(b, bar+r.label("fix")+r.style(terminal.TextSecondary).Render(f.Remediation))
	fmt.Fprintln(b)
}

// label renders a short, faint detail label padded to a stable width.
func (r styler) label(name string) string {
	return r.style(terminal.TextTertiary).Render(fmt.Sprintf("%-4s", name))
}

func (r styler) writeSuppressed(b *bytes.Buffer, result doctor.Result) {
	for _, s := range result.Suppressed {
		line := r.style(terminal.TextTertiary).Render("suppressed ") +
			r.style(terminal.TextSecondary).Render(s.Finding.RuleID) + " " +
			r.style(terminal.TextTertiary).Render("("+s.Source+")")
		if s.Reason != "" {
			line += " " + r.style(terminal.TextTertiary).Render("— "+s.Reason)
		}
		fmt.Fprintln(b, line)
	}
}

// writeSummary renders the findings rollup line, e.g.
// "Checked 18 rules · 3 findings · 1 BLOCKER · 1 HIGH · 1 MEDIUM",
// omitting zero severity buckets. A clean run reads "… · 0 findings ✓".
func (r styler) writeSummary(b *bytes.Buffer, result doctor.Result) {
	dot := r.style(terminal.TextTertiary).Render(" · ")
	total := len(result.Findings)

	checked := r.style(terminal.TextSecondary).Render(fmt.Sprintf("Checked %d rules", len(catalog.IDs())))

	countTok := terminal.TextSuccess
	countText := fmt.Sprintf("%d findings", total)
	if total > 0 {
		countTok, _ = severityTokens(worstSeverity(result.Findings))
	} else if r.unicode {
		countText += " " + r.g.pass
	}
	line := checked + dot + r.style(countTok).Render(countText)

	counts := severityCounts(result.Findings)
	for _, sev := range []findings.Severity{
		findings.SeverityBlocker, findings.SeverityHigh, findings.SeverityMedium, findings.SeverityLow,
	} {
		if n := counts[sev]; n > 0 {
			tok, _ := severityTokens(sev)
			line += dot + r.style(tok).Render(fmt.Sprintf("%d %s", n, sev))
		}
	}
	fmt.Fprintln(b, line)
}

func (r styler) writeScorecard(b *bytes.Buffer, result doctor.Result) {
	breakdown := scoring.ByCategory(result.Findings)
	if len(breakdown) == 0 {
		return
	}
	fmt.Fprintln(b)
	fmt.Fprintln(b, r.style(terminal.TextSecondary).Render("readiness by category"))
	for _, c := range breakdown {
		textTok, _ := severityTokens(c.WorstSeverity)
		name := r.style(terminal.TextSecondary).Render(fmt.Sprintf("  %-12s", c.Category))
		deduction := r.style(textTok).Render(fmt.Sprintf("−%-3d", c.Deduction))
		detail := r.style(terminal.TextTertiary).Render(fmt.Sprintf("%d finding(s), worst %s", c.Findings, c.WorstSeverity))
		fmt.Fprintln(b, name+" "+deduction+" "+detail)
	}
}

// scoreBarWidth is the cell width of the score progress bar in the hero.
const scoreBarWidth = 24

func (r styler) writeScore(b *bytes.Buffer, result doctor.Result) {
	fmt.Fprintln(b, r.divider())

	scoreTok := terminal.TextDanger
	verdict := "FAIL"
	if result.Passed {
		scoreTok = terminal.TextSuccess
		verdict = "PASS"
	}
	if r.unicode {
		if result.Passed {
			verdict += " " + r.g.pass
		} else {
			verdict += " " + r.g.fail
		}
	}

	label := r.style(terminal.TextSecondary).Render("Score ")
	score := r.style(scoreTok).Bold(true).Render(strconv.Itoa(result.Score.Final))
	maxScore := r.style(terminal.TextTertiary).Render("/100")
	bar := r.scoreBar(result.Score.Final, scoreTok)
	badge := r.style(scoreTok).Bold(true).Render(verdict)
	fmt.Fprintln(b, label+score+maxScore+"  "+bar+"  "+badge)

	// A cap is active when the final score sits below the formula base (blocker
	// ceiling or a rule cap pulled it down).
	if result.Score.Final < result.Score.Base {
		fmt.Fprintln(b, "      "+r.style(terminal.TextDanger).Render(fmt.Sprintf("cap   score capped at %d", result.Score.Final)))
	}
	fmt.Fprintln(b, "      "+r.style(terminal.TextTertiary).Render(fmt.Sprintf("threshold %d", result.Threshold)))
}

// scoreBar renders a fixed-width progress bar whose filled portion is
// proportional to score/100 and carries the score's severity color.
func (r styler) scoreBar(score int, tok terminal.Token) string {
	filled := max(0, min(scoreBarWidth, score*scoreBarWidth/100))
	bar := ""
	if filled > 0 {
		bar += r.style(tok).Render(strings.Repeat(r.g.barFull, filled))
	}
	if filled < scoreBarWidth {
		bar += r.style(terminal.TextTertiary).Render(strings.Repeat(r.g.barEmpty, scoreBarWidth-filled))
	}
	return bar
}

// severityIcon mirrors the design (charter-doctor-init-fix.html): only BLOCKER
// is the danger cross; HIGH and MEDIUM share the warning triangle; everything
// else (LOW, informational, unknown) uses the neutral bullet.
func (r styler) severityIcon(sev findings.Severity) string {
	switch sev {
	case findings.SeverityBlocker:
		return r.g.fail
	case findings.SeverityHigh, findings.SeverityMedium:
		return r.g.warn
	default:
		return r.g.bullet
	}
}

// severityTokens maps a severity to its text and border palette tokens, per the
// design: BLOCKER → danger, HIGH/MEDIUM → warning, LOW → info. Unknown
// severities fall through to the informational (info) tokens.
func severityTokens(sev findings.Severity) (text, border terminal.Token) {
	switch sev {
	case findings.SeverityBlocker:
		return terminal.TextDanger, terminal.BorderDanger
	case findings.SeverityHigh, findings.SeverityMedium:
		return terminal.TextWarning, terminal.BorderWarning
	case findings.SeverityLow:
		return terminal.TextInfo, terminal.BorderInfo
	default:
		return terminal.TextInfo, terminal.BorderInfo
	}
}

// severityCounts tallies findings by severity for the summary rollup.
func severityCounts(fs []findings.Finding) map[findings.Severity]int {
	counts := make(map[findings.Severity]int, 4)
	for _, f := range fs {
		counts[f.Severity]++
	}
	return counts
}

// worstSeverity returns the highest-weight severity among findings (the zero
// Severity when there are none), used to color the summary findings count.
func worstSeverity(fs []findings.Finding) findings.Severity {
	worst := findings.Severity("")
	for _, f := range fs {
		if f.Severity.Weight() > worst.Weight() {
			worst = f.Severity
		}
	}
	return worst
}

func locationText(loc findings.Location) string {
	if loc.Line > 0 {
		return fmt.Sprintf("%s:%d", loc.Path, loc.Line)
	}
	return loc.Path
}

// fileURL builds an OSC 8 file:// target for a location, resolving repo-relative
// paths against the scanned root. It returns "" when there is no path to link.
func fileURL(root string, loc findings.Location) string {
	if loc.Path == "" {
		return ""
	}
	path := loc.Path
	if !filepath.IsAbs(path) && !strings.HasPrefix(path, "/") {
		path = filepath.Join(root, path)
	}
	path = filepath.ToSlash(path)
	if len(path) >= 2 && path[1] == ':' {
		path = "/" + path
	}
	return "file://" + path
}

// ruleURL returns the catalog help URI for a rule ID, or "" when the rule is not
// in the catalog (so no hyperlink is emitted).
func ruleURL(id string) string {
	if e, ok := catalog.Lookup(id); ok {
		return e.HelpURI
	}
	return ""
}
