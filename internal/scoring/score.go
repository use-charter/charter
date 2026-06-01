package scoring

import (
	"sort"

	"go.use-charter.dev/charter/internal/findings"
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
