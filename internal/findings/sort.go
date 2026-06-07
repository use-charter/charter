package findings

import "sort"

// SortByPriority orders findings in place: severity weight descending, then
// RuleID ascending. Stable so equal-priority items keep their relative order.
func SortByPriority(in []Finding) {
	sort.SliceStable(in, func(i, j int) bool {
		wi, wj := in[i].Severity.Weight(), in[j].Severity.Weight()
		if wi != wj {
			return wi > wj
		}
		return in[i].RuleID < in[j].RuleID
	})
}

// SortedByPriority returns a copy of in sorted by SortByPriority.
func SortedByPriority(in []Finding) []Finding {
	out := append([]Finding(nil), in...)
	SortByPriority(out)
	return out
}

// LessByPriority reports whether a sorts before b under SortByPriority.
func LessByPriority(a, b Finding) bool {
	wa, wb := a.Severity.Weight(), b.Severity.Weight()
	if wa != wb {
		return wa > wb
	}
	return a.RuleID < b.RuleID
}

// FirstLocation returns the first physical site of a finding, or a zero Location
// when the finding has no locations (an absence finding).
func FirstLocation(f Finding) Location {
	if len(f.Locations) == 0 {
		return Location{}
	}
	return f.Locations[0]
}
