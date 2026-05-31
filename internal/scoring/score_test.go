package scoring

import (
	"testing"

	"go.charter.dev/charter/internal/findings"
)

func TestCalculateSkipsInformational(t *testing.T) {
	all := []findings.Finding{
		{Severity: findings.SeverityMedium, Informational: true},
		{Severity: findings.SeverityLow},
	}
	got := Calculate(all)
	if got.Final != 99 {
		t.Fatalf("informational finding must not deduct: got %d, want 99", got.Final)
	}
}

func TestCalculateInformationalIgnoresCap(t *testing.T) {
	all := []findings.Finding{
		{Severity: findings.SeverityMedium, Informational: true, Cap: 10},
	}
	got := Calculate(all)
	if got.Final != 100 {
		t.Fatalf("informational finding must not cap: got %d, want 100", got.Final)
	}
}

func TestScoreAppliesFormulaAndBlockerCap(t *testing.T) {
	result := Calculate([]findings.Finding{{Severity: findings.SeverityBlocker}})

	if result.Base != 80 {
		t.Fatalf("expected base 80, got %d", result.Base)
	}

	if result.Final != 59 {
		t.Fatalf("expected blocker cap final 59, got %d", result.Final)
	}
}

func TestScoreWithoutFindingsStaysAtOneHundred(t *testing.T) {
	result := Calculate(nil)
	if result.Base != 100 || result.Final != 100 {
		t.Fatalf("expected 100/100, got %d/%d", result.Base, result.Final)
	}
}

func TestScoreAppliesSecretCapAtFortyNine(t *testing.T) {
	result := Calculate([]findings.Finding{{RuleID: "AE-SEC-001", Severity: findings.SeverityBlocker, Cap: 49}})

	if result.Final != 49 {
		t.Fatalf("expected final score 49 when secret rule fires, got %d", result.Final)
	}
}

func TestScoreAppliesLowestFindingCap(t *testing.T) {
	result := Calculate([]findings.Finding{
		{RuleID: "AE-SEC-001", Severity: findings.SeverityBlocker, Cap: 49},
		{RuleID: "AE-CUSTOM", Severity: findings.SeverityLow, Cap: 30},
	})

	if result.Final != 30 {
		t.Fatalf("expected lowest cap 30 to win, got %d", result.Final)
	}
}
