package markdown

import (
	"fmt"
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/suppress"
)

// Render projects a doctor.Result into GitHub-PR-comment-friendly Markdown.
func Render(result doctor.Result) ([]byte, error) {
	ordered := append([]findings.Finding(nil), result.Findings...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if wi, wj := ordered[i].Severity.Weight(), ordered[j].Severity.Weight(); wi != wj {
			return wi > wj
		}
		return ordered[i].RuleID < ordered[j].RuleID
	})

	var b strings.Builder
	status := "FAIL"
	if result.Passed {
		status = "PASS"
	}
	b.WriteString("# Charter\n\n")
	fmt.Fprintf(&b, "**Score: %d / 100** (threshold %d) — **%s**\n\n", result.Score.Final, result.Threshold, status)

	if len(ordered) == 0 {
		b.WriteString("No findings.\n")
	} else {
		b.WriteString("| Rule | Severity | Location | Summary |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, f := range ordered {
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", f.RuleID, f.Severity, location(f), escapePipes(f.Summary))
		}
	}

	writeCategoryBreakdown(&b, result.Findings)
	writeSuppressed(&b, result.Suppressed)
	return []byte(b.String()), nil
}

func writeCategoryBreakdown(b *strings.Builder, all []findings.Finding) {
	breakdown := scoring.ByCategory(all)
	if len(breakdown) == 0 {
		return
	}
	b.WriteString("\n**Readiness by category**\n\n")
	b.WriteString("| Category | Findings | Deduction | Worst |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, c := range breakdown {
		fmt.Fprintf(b, "| %s | %d | −%d | %s |\n", c.Category, c.Findings, c.Deduction, c.WorstSeverity)
	}
}

func writeSuppressed(b *strings.Builder, list []suppress.Suppressed) {
	if len(list) == 0 {
		return
	}
	ordered := append([]suppress.Suppressed(nil), list...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if wi, wj := ordered[i].Finding.Severity.Weight(), ordered[j].Finding.Severity.Weight(); wi != wj {
			return wi > wj
		}
		return ordered[i].Finding.RuleID < ordered[j].Finding.RuleID
	})

	fmt.Fprintf(b, "\n**Suppressed (%d)**\n\n", len(ordered))
	b.WriteString("| Rule | Source | Reason | Expires |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, s := range ordered {
		reason := s.Reason
		if reason == "" {
			reason = "—"
		}
		expires := s.Expires
		if expires == "" {
			expires = "default"
		}
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n", s.Finding.RuleID, s.Source, escapePipes(reason), expires)
	}
}

func location(f findings.Finding) string {
	if len(f.Locations) == 0 {
		return "—"
	}
	loc := f.Locations[0]
	if loc.Line > 0 {
		return fmt.Sprintf("`%s:%d`", loc.Path, loc.Line)
	}
	if loc.Path == "" {
		return "—"
	}
	return "`" + loc.Path + "`"
}

func escapePipes(s string) string { return strings.ReplaceAll(s, "|", `\|`) }
