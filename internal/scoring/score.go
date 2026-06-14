package scoring

import (
	"sort"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/rules/catalog"
)

// CategoryScore is one row of the per-category readiness breakdown: how many
// findings a category carries, the raw deduction they sum to (informational
// findings deduct 0), and the worst severity present.
type CategoryScore struct {
	Category      string
	Findings      int
	Deduction     int
	WorstSeverity findings.Severity
}

// DeductionFor returns the point deduction of a severity (matching Calculate).
func DeductionFor(s findings.Severity) int {
	switch s {
	case findings.SeverityBlocker:
		return 20
	case findings.SeverityHigh:
		return 10
	case findings.SeverityMedium:
		return 4
	case findings.SeverityLow:
		return 1
	default:
		return 0
	}
}

// ByCategory groups findings into a per-category readiness breakdown, sorted by
// deduction (desc) then category (asc). It is a presentation aid over the same
// findings Calculate scores; it does not change the score formula. Informational
// findings count toward Findings but contribute 0 to Deduction.
func ByCategory(all []findings.Finding) []CategoryScore {
	type acc struct {
		count, deduction, worstWeight int
		worst                         findings.Severity
	}
	m := map[string]*acc{}
	order := []string{}
	for _, f := range all {
		a := m[f.Category]
		if a == nil {
			a = &acc{}
			m[f.Category] = a
			order = append(order, f.Category)
		}
		a.count++
		if !f.Informational {
			a.deduction += DeductionFor(f.Severity)
		}
		if w := f.Severity.Weight(); w > a.worstWeight {
			a.worstWeight = w
			a.worst = f.Severity
		}
	}
	out := make([]CategoryScore, 0, len(order))
	for _, cat := range order {
		a := m[cat]
		out = append(out, CategoryScore{Category: cat, Findings: a.count, Deduction: a.deduction, WorstSeverity: a.worst})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Deduction != out[j].Deduction {
			return out[i].Deduction > out[j].Deduction
		}
		return out[i].Category < out[j].Category
	})
	return out
}

// CategoryReadiness is one row of the full readiness scorecard: every catalog
// category (not only those with findings), with the count of rules that scanned
// clean out of the category total. Passed/Total drives the per-category mini bar
// and the "N/M" fraction; the deduction and worst severity colour the row.
// Informational is true when a category has findings but none of them deduct.
type CategoryReadiness struct {
	Category      string
	Total         int // catalog rules in this category
	Passed        int // rules with no finding (Total − Failed)
	Failed        int // distinct rules with at least one finding
	Findings      int
	Deduction     int
	WorstSeverity findings.Severity
	Informational bool
}

// Readiness assembles the full per-category scorecard in canonical category
// order. Unlike ByCategory (which lists only categories that carry a finding),
// Readiness covers every catalog category so a clean category still reports
// "N/N rules clean". Passed counts catalog rules with no finding; a rule that
// fired only an informational finding counts as not-clean but does not deduct.
func Readiness(all []findings.Finding) []CategoryReadiness {
	byCat := map[string]CategoryScore{}
	for _, c := range ByCategory(all) {
		byCat[c.Category] = c
	}
	failed := map[string]map[string]bool{}
	for _, f := range all {
		if failed[f.Category] == nil {
			failed[f.Category] = map[string]bool{}
		}
		failed[f.Category][f.RuleID] = true
	}
	counts := catalog.RuleCountByCategory()

	// Canonical catalog categories first, then any category seen only in the
	// findings (defensive: an uncataloged category must still appear, with its
	// rule total inferred from the distinct rule IDs observed).
	cats := catalog.Categories()
	inCatalog := make(map[string]bool, len(cats))
	for _, c := range cats {
		inCatalog[c] = true
	}
	extra := make([]string, 0)
	for cat := range failed {
		if !inCatalog[cat] {
			extra = append(extra, cat)
		}
	}
	sort.Strings(extra)
	cats = append(cats, extra...)

	out := make([]CategoryReadiness, 0, len(cats))
	for _, cat := range cats {
		failedN := len(failed[cat])
		total := counts[cat]
		if total == 0 {
			total = failedN // uncataloged category: infer total from observed rules
		}
		passed := max(0, total-failedN)
		cs := byCat[cat]
		out = append(out, CategoryReadiness{
			Category:      cat,
			Total:         total,
			Passed:        passed,
			Failed:        failedN,
			Findings:      cs.Findings,
			Deduction:     cs.Deduction,
			WorstSeverity: cs.WorstSeverity,
			Informational: cs.Findings > 0 && cs.Deduction == 0,
		})
	}
	return out
}

type Result struct {
	Blocker int
	High    int
	Medium  int
	Low     int
	Base    int
	Final   int
}

func Calculate(all []findings.Finding) Result {
	result := Result{}

	for _, finding := range all {
		if finding.Informational {
			continue
		}
		switch finding.Severity {
		case findings.SeverityBlocker:
			result.Blocker++
		case findings.SeverityHigh:
			result.High++
		case findings.SeverityMedium:
			result.Medium++
		case findings.SeverityLow:
			result.Low++
		}
	}

	result.Base = 100 - (result.Blocker * 20) - (result.High * 10) - (result.Medium * 4) - result.Low
	if result.Base < 0 {
		result.Base = 0
	}

	result.Final = result.Base
	if result.Blocker > 0 && result.Final > 59 {
		result.Final = 59
	}

	for _, finding := range all {
		if finding.Informational {
			continue
		}
		if finding.Cap > 0 && result.Final > finding.Cap {
			result.Final = finding.Cap
		}
	}

	return result
}
