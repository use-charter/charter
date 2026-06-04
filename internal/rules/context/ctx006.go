package context

import (
	"fmt"
	"regexp"
	"strings"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

// emphaticDirective matches stacked emphatic directives. Instruction-following
// research shows their overuse creates a fragile, competitive instruction
// topology that degrades adherence; concise declarative guidance transfers more
// reliably. Match the all-caps emphatic form actually used for shouting.
var emphaticDirective = regexp.MustCompile(`\b(IMPORTANT|NEVER|MUST|CRITICAL|ALWAYS|EXTREMELY|ABSOLUTELY|FORBIDDEN|PROHIBITED)\b`)

// emphaticDensityThreshold is the AE-CTX-006 trigger: emphatic directives per
// 1,000 words. Conservative; the finding is informational (never deducts).
const emphaticDensityThreshold = 15.0

func checkCTX006(root string, inv repository.Inventory) (findings.Finding, bool) {
	locations := supportedContextLocations(inv)
	if len(locations) == 0 {
		return findings.Finding{}, false // absence is AE-CTX-001's concern
	}
	candidate := locations[0]
	content, ok := readContextCandidate(root, inv, candidate)
	if !ok {
		return findings.Finding{}, false
	}
	words := len(strings.Fields(content))
	if words == 0 {
		return findings.Finding{}, false
	}
	hits := len(emphaticDirective.FindAllString(content, -1))
	density := float64(hits) / float64(words) * 1000
	if density < emphaticDensityThreshold {
		return findings.Finding{}, false
	}
	return findings.Finding{
		RuleID:        "AE-CTX-006",
		Severity:      findings.SeverityLow,
		Category:      "Context",
		Summary:       "Agent context file over-uses emphatic directives, which can degrade instruction adherence",
		Remediation:   "Prefer concise, declarative guidance over stacked emphatic directives (IMPORTANT/NEVER/MUST/…); state constraints plainly.",
		Evidence:      []string{fmt.Sprintf("%d emphatic directives in %d words (~%.0f per 1K; threshold %.0f)", hits, words, density, emphaticDensityThreshold)},
		Locations:     []findings.Location{{Path: candidate}},
		Informational: true,
	}, true
}
