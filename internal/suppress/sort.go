package suppress

import "sort"

// SortByPriority orders suppressed findings in place: severity weight descending,
// then RuleID ascending (stable).
func SortByPriority(in []Suppressed) {
	sort.SliceStable(in, func(i, j int) bool {
		wi, wj := in[i].Finding.Severity.Weight(), in[j].Finding.Severity.Weight()
		if wi != wj {
			return wi > wj
		}
		return in[i].Finding.RuleID < in[j].Finding.RuleID
	})
}

// SortedByPriority returns a copy sorted by SortByPriority.
func SortedByPriority(in []Suppressed) []Suppressed {
	out := append([]Suppressed(nil), in...)
	SortByPriority(out)
	return out
}
