package json

import (
	encodingjson "encoding/json"
	"testing"

	"go.charter.dev/charter/internal/doctor"
	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/scoring"
)

func TestRenderProducesStableJSONShape(t *testing.T) {
	result := doctor.Result{
		Root:      "D:/Projects/charter",
		Threshold: 80,
		Passed:    true,
		Findings: []findings.Finding{{
			RuleID:      "AE-CTX-001",
			Severity:    findings.SeverityBlocker,
			Category:    "Context",
			Summary:     "Agent context file is missing",
			Remediation: "Create AGENTS.md with the required sections",
			Evidence:    []string{"no supported root context file found"},
		}},
		Score: scoring.Result{Blocker: 1, Base: 80, Final: 59},
	}

	data, err := Render(result)
	if err != nil {
		t.Fatalf("expected render to succeed: %v", err)
	}

	var payload map[string]any
	if err := encodingjson.Unmarshal(data, &payload); err != nil {
		t.Fatalf("expected valid json: %v", err)
	}

	if payload["repo_path"] != "D:/Projects/charter" {
		t.Fatalf("unexpected repo_path: %#v", payload["repo_path"])
	}

	if payload["threshold"] != float64(80) {
		t.Fatalf("unexpected threshold: %#v", payload["threshold"])
	}

	if payload["passed"] != true {
		t.Fatalf("unexpected passed flag: %#v", payload["passed"])
	}

	findingsList, ok := payload["findings"].([]any)
	if !ok || len(findingsList) != 1 {
		t.Fatalf("unexpected findings payload: %#v", payload["findings"])
	}

	finding, ok := findingsList[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected finding payload: %#v", findingsList[0])
	}

	if finding["rule_id"] != "AE-CTX-001" {
		t.Fatalf("unexpected rule_id: %#v", finding["rule_id"])
	}

	if finding["severity"] != "BLOCKER" {
		t.Fatalf("unexpected severity: %#v", finding["severity"])
	}

	if _, exists := finding["RuleID"]; exists {
		t.Fatalf("unexpected Go field name in payload: %#v", finding)
	}
}
