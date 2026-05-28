package scoring

import (
	"testing"

	"go.charter.dev/charter/internal/findings"
)

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
