package findings

import "testing"

func TestFindingCarriesStableFields(t *testing.T) {
	finding := Finding{
		RuleID:      "AE-CTX-001",
		Severity:    SeverityBlocker,
		Category:    "Context",
		Summary:     "missing AGENTS.md",
		Remediation: "create AGENTS.md",
		Evidence:    []string{"no root context file found"},
	}

	if finding.RuleID == "" || len(finding.Evidence) == 0 {
		t.Fatalf("expected stable fields to be populated")
	}
}
