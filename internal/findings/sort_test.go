package findings

import "testing"

func TestSortByPriorityOrdersSeverityThenRuleID(t *testing.T) {
	in := []Finding{
		{RuleID: "AE-B", Severity: SeverityLow},
		{RuleID: "AE-A", Severity: SeverityHigh},
		{RuleID: "AE-C", Severity: SeverityHigh},
	}
	SortByPriority(in)
	if in[0].RuleID != "AE-A" || in[1].RuleID != "AE-C" || in[2].RuleID != "AE-B" {
		t.Fatalf("unexpected order: %#v", in)
	}
}
