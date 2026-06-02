package tui

import (
	"fmt"
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
)

// itemKind classifies a row in the findings list. A finding is either an active
// scored finding, an informational finding (listed but excluded from the score),
// or a suppressed finding (muted by a suppression). Informational and suppressed
// items are hidden behind their own filter (the `s` toggle).
type itemKind int

const (
	kindActive itemKind = iota
	kindInformational
	kindSuppressed
)

// item is one row of the browser: a finding plus its kind and, for suppressed
// rows, the suppression metadata needed by the detail pane.
type item struct {
	finding findings.Finding
	kind    itemKind

	// Suppression metadata, populated only for kindSuppressed.
	source   string
	reason   string
	approver string
	expires  string
}

// muted reports whether the item is hidden by default (informational or
// suppressed). The list shows only un-muted, score-bearing findings until the
// `s` filter is toggled on.
func (it item) muted() bool { return it.kind != kindActive }

// buildItems flattens a doctor.Result into the browser's item list: every active
// finding (informational ones flagged as such) followed by every suppressed
// finding. Scan order is preserved; sorting is applied later by the model.
func buildItems(result doctor.Result) []item {
	items := make([]item, 0, len(result.Findings)+len(result.Suppressed))
	for _, f := range result.Findings {
		kind := kindActive
		if f.Informational {
			kind = kindInformational
		}
		items = append(items, item{finding: f, kind: kind})
	}
	for _, s := range result.Suppressed {
		items = append(items, item{
			finding:  s.Finding,
			kind:     kindSuppressed,
			source:   s.Source,
			reason:   s.Reason,
			approver: s.Approver,
			expires:  s.Expires,
		})
	}
	return items
}

// uniqueCategories returns the sorted, de-duplicated category names across all
// items (including muted ones) so category cycling is stable regardless of the
// active filters.
func uniqueCategories(items []item) []string {
	seen := map[string]bool{}
	cats := make([]string, 0)
	for _, it := range items {
		c := it.finding.Category
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return cats
}

// sortMode controls the list ordering. Both modes are deterministic.
type sortMode int

const (
	// sortBySeverity orders worst-severity first, then category, then rule ID.
	sortBySeverity sortMode = iota
	// sortByCategory groups by category, then worst-severity first within each.
	sortByCategory
)

// filterItems projects items down to those matching the active filters: the
// muted toggle, an optional severity, an optional category, and a free-text
// query. It allocates a new slice and never mutates the input.
func filterItems(items []item, sev findings.Severity, cat string, showMuted bool, query string) []item {
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]item, 0, len(items))
	for _, it := range items {
		if it.muted() && !showMuted {
			continue
		}
		if sev != "" && it.finding.Severity != sev {
			continue
		}
		if cat != "" && it.finding.Category != cat {
			continue
		}
		if q != "" && !itemMatches(it, q) {
			continue
		}
		out = append(out, it)
	}
	return out
}

// itemMatches reports whether the item matches the lowercased query across its
// rule ID, summary, category, evidence, and location paths.
func itemMatches(it item, q string) bool {
	f := it.finding
	if strings.Contains(strings.ToLower(f.RuleID), q) ||
		strings.Contains(strings.ToLower(f.Summary), q) ||
		strings.Contains(strings.ToLower(f.Category), q) {
		return true
	}
	for _, ev := range f.Evidence {
		if strings.Contains(strings.ToLower(ev), q) {
			return true
		}
	}
	for _, loc := range f.Locations {
		if strings.Contains(strings.ToLower(loc.Path), q) {
			return true
		}
	}
	return false
}

// sortItems orders items in place per mode. The comparison is a total order on
// (category?, severity weight desc, category, rule ID) so the result is stable
// and reproducible across runs.
func sortItems(items []item, mode sortMode) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i].finding, items[j].finding
		if mode == sortByCategory && a.Category != b.Category {
			return a.Category < b.Category
		}
		if wa, wb := a.Severity.Weight(), b.Severity.Weight(); wa != wb {
			return wa > wb
		}
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		return a.RuleID < b.RuleID
	})
}

// firstLocation returns the "path:line" (or bare "path") of a finding's first
// location, or "" when the finding has no physical site (an absence finding).
func firstLocation(f findings.Finding) string {
	if len(f.Locations) == 0 {
		return ""
	}
	loc := f.Locations[0]
	if loc.Path == "" {
		return ""
	}
	if loc.Line > 0 {
		return fmt.Sprintf("%s:%d", loc.Path, loc.Line)
	}
	return loc.Path
}

// severityForDigit maps the 1-4 filter keys to their severity. It returns the
// empty severity for any other key.
func severityForDigit(s string) findings.Severity {
	switch s {
	case "1":
		return findings.SeverityBlocker
	case "2":
		return findings.SeverityHigh
	case "3":
		return findings.SeverityMedium
	case "4":
		return findings.SeverityLow
	}
	return ""
}
