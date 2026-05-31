// Package governance implements the AE-SUPPRESS-001/002/003 rules that audit the
// suppressions applied during a scan.
package governance

import (
	"fmt"
	"sort"
	"strings"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/suppress"
)

// rateThreshold is the AE-SUPPRESS-003 suppression-rate trigger (30%, per the
// audit checklist).
const rateThreshold = 0.30

// Run audits the suppressions applied this scan and emits AE-SUPPRESS-001/002/003.
// activeRuleCount is the number of active findings from the scanning rules
// (excluding governance findings); suppressedCount is the number suppressed.
// Expiry is pre-filtered by suppress.Apply (used carries only auditable entries),
// so no clock is needed here.
func Run(used []suppress.Entry, activeRuleCount, suppressedCount int) []findings.Finding {
	var result []findings.Finding
	result = append(result, checkMissingReason(used)...)
	result = append(result, checkPermanentNoApprover(used)...)
	if f, ok := checkHighRate(activeRuleCount, suppressedCount); ok {
		result = append(result, f)
	}
	return result
}

func checkMissingReason(used []suppress.Entry) []findings.Finding {
	var out []findings.Finding
	for _, e := range used {
		if strings.TrimSpace(e.Reason) != "" {
			continue
		}
		out = append(out, findings.Finding{
			RuleID:      "AE-SUPPRESS-001",
			Severity:    findings.SeverityMedium,
			Category:    "Governance",
			Summary:     "Suppression is missing a required reason",
			Remediation: "Add a reason to the .charter-suppress.yml entry or the inline charter:ignore directive explaining why the finding is accepted.",
			Evidence:    []string{evidence(e) + " has no reason"},
			Locations:   []findings.Location{location(e)},
		})
	}
	sortFindings(out)
	return out
}

func checkPermanentNoApprover(used []suppress.Entry) []findings.Finding {
	var out []findings.Finding
	for _, e := range used {
		if !suppress.IsPermanent(e) || strings.TrimSpace(e.Approver) != "" {
			continue
		}
		out = append(out, findings.Finding{
			RuleID:      "AE-SUPPRESS-002",
			Severity:    findings.SeverityHigh,
			Category:    "Governance",
			Summary:     "Permanent suppression has no approver",
			Remediation: "Add approver=\"<name>\" to the permanent suppression, or set an expires date so it re-surfaces for review; until then it is not honored and the finding stays active.",
			Evidence:    []string{evidence(e) + " is permanent without an approver (not honored — re-fires)"},
			Locations:   []findings.Location{location(e)},
		})
	}
	sortFindings(out)
	return out
}

func checkHighRate(activeRuleCount, suppressedCount int) (findings.Finding, bool) {
	total := activeRuleCount + suppressedCount
	if total == 0 {
		return findings.Finding{}, false
	}
	rate := float64(suppressedCount) / float64(total)
	if rate <= rateThreshold {
		return findings.Finding{}, false
	}
	pct := int(rate*100 + 0.5)
	return findings.Finding{
		RuleID:        "AE-SUPPRESS-003",
		Severity:      findings.SeverityMedium,
		Category:      "Governance",
		Summary:       "High suppression rate",
		Remediation:   "Review suppressed findings: recalibrate noisy rules, fix the underlying issues, or document accepted risks with reasons and approvers.",
		Evidence:      []string{fmt.Sprintf("%d of %d findings suppressed (%d%%) — high suppression rate", suppressedCount, total, pct)},
		Locations:     []findings.Location{{Path: suppress.File}},
		Informational: true,
	}, true
}

func evidence(e suppress.Entry) string {
	kind := "external suppression"
	if e.Source == suppress.SourceInSource {
		kind = "inline suppression"
	}
	return fmt.Sprintf("%s of %s", kind, e.Rule)
}

func location(e suppress.Entry) findings.Location {
	if e.Source == suppress.SourceInSource {
		return findings.Location{Path: e.Path, Line: e.Line}
	}
	return findings.Location{Path: suppress.File}
}

func sortFindings(fs []findings.Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		li, lj := first(fs[i]), first(fs[j])
		if li.Path != lj.Path {
			return li.Path < lj.Path
		}
		return li.Line < lj.Line
	})
}

func first(f findings.Finding) findings.Location {
	if len(f.Locations) == 0 {
		return findings.Location{}
	}
	return f.Locations[0]
}
