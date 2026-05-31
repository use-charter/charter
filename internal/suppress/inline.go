package suppress

import (
	"regexp"
	"strings"
)

var (
	directiveRE = regexp.MustCompile(`charter:ignore\s+(AE-[A-Z]+-\d+)\b(.*)`)
	reasonRE    = regexp.MustCompile(`reason="([^"]*)"`)
	approverRE  = regexp.MustCompile(`approver="([^"]*)"`)
	expiresRE   = regexp.MustCompile(`expires=(permanent|\d{4}-\d{2}-\d{2})`)
)

// parseInlineDirective extracts a charter:ignore directive from a single source
// line when it appears inside a line comment (#, //, or <!-- ... -->). It returns
// (entry, true) with Source=SourceInSource, the rule ID, and any reason/expires/
// approver fields. Requiring a comment leader is what prevents prose that merely
// mentions the syntax from being treated as a directive.
func parseInlineDirective(line string) (Entry, bool) {
	idx := commentIndex(line)
	if idx < 0 {
		return Entry{}, false
	}
	m := directiveRE.FindStringSubmatch(line[idx:])
	if m == nil {
		return Entry{}, false
	}
	e := Entry{Rule: m[1], Source: SourceInSource}
	tail := m[2]
	if r := reasonRE.FindStringSubmatch(tail); r != nil {
		e.Reason = strings.TrimSpace(r[1])
	}
	if a := approverRE.FindStringSubmatch(tail); a != nil {
		e.Approver = strings.TrimSpace(a[1])
	}
	if x := expiresRE.FindStringSubmatch(tail); x != nil {
		e.Expires = x[1]
	}
	return e, true
}

// commentIndex returns the byte index of the earliest line-comment leader
// (#, //, or <!--), or -1 when the line has none.
func commentIndex(line string) int {
	earliest := -1
	for _, leader := range []string{"#", "//", "<!--"} {
		if i := strings.Index(line, leader); i >= 0 && (earliest == -1 || i < earliest) {
			earliest = i
		}
	}
	return earliest
}
