package json

import (
	encodingjson "encoding/json"
	"strings"
	"testing"

	"go.charter.dev/charter/internal/doctor"
	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/scoring"
	"go.charter.dev/charter/internal/suppress"
)

func TestRenderEmitsSuppressedAndInformational(t *testing.T) {
	result := doctor.Result{
		Root: "/repo", Threshold: 80, Passed: true,
		Findings: []findings.Finding{
			{RuleID: "AE-SUPPRESS-003", Severity: findings.SeverityMedium, Category: "Governance", Summary: "High suppression rate", Informational: true, Locations: []findings.Location{{Path: ".charter-suppress.yml"}}},
		},
		Suppressed: []suppress.Suppressed{
			{Finding: findings.Finding{RuleID: "AE-MCP-001", Severity: findings.SeverityHigh, Locations: []findings.Location{{Path: ".mcp.json", Line: 1}}}, Source: suppress.SourceExternal, Reason: "vendored", Expires: "2099-01-01"},
		},
		Score: scoring.Result{Medium: 0, Base: 100, Final: 100},
	}

	data, err := Render(result)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, `"informational":true`) {
		t.Fatalf("expected informational flag, got %s", s)
	}

	var payload struct {
		Suppressed []struct {
			RuleID string `json:"rule_id"`
			Source string `json:"source"`
			Reason string `json:"reason"`
		} `json:"suppressed"`
	}
	if err := encodingjson.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload.Suppressed) != 1 || payload.Suppressed[0].RuleID != "AE-MCP-001" || payload.Suppressed[0].Source != "external" || payload.Suppressed[0].Reason != "vendored" {
		t.Fatalf("unexpected suppressed payload: %#v", payload.Suppressed)
	}
}

func TestRenderEmitsEmptySuppressedArray(t *testing.T) {
	data, err := Render(doctor.Result{Root: "/repo", Threshold: 80, Passed: true, Score: scoring.Result{Final: 100}})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(string(data), `"suppressed":[]`) {
		t.Fatalf("expected empty suppressed array (never null) for SARIF consistency, got %s", data)
	}
}

func TestRenderEmitsStructuredLocations(t *testing.T) {
	result := doctor.Result{
		Root:      "/repo",
		Threshold: 80,
		Passed:    false,
		Findings: []findings.Finding{
			{
				RuleID:    "AE-SEC-001",
				Severity:  findings.SeverityBlocker,
				Category:  "Secrets",
				Summary:   "Secret detected in agent-visible context file",
				Locations: []findings.Location{{Path: "AGENTS.md", Line: 14}},
				Evidence:  []string{"AGENTS.md: sk-a…"},
				Cap:       49,
			},
			{
				RuleID:   "AE-ENV-001",
				Severity: findings.SeverityMedium,
				Category: "Environment",
				Summary:  "No reproducible toolchain declaration found",
				Evidence: []string{"no toolchain file detected"},
			},
		},
		Score: scoring.Result{Blocker: 1, Medium: 1, Base: 76, Final: 49},
	}

	data, err := Render(result)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	// Absence finding must serialize an empty array, never null (SARIF needs an array).
	if !strings.Contains(string(data), `"locations":[]`) {
		t.Fatalf("expected an empty locations array for the absence finding, got %s", data)
	}

	var payload struct {
		Findings []struct {
			RuleID    string `json:"rule_id"`
			Locations []struct {
				Path string `json:"path"`
				Line int    `json:"line"`
			} `json:"locations"`
		} `json:"findings"`
	}
	if err := encodingjson.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Deterministic order: BLOCKER (AE-SEC-001) sorts before MEDIUM (AE-ENV-001).
	sec := payload.Findings[0]
	if sec.RuleID != "AE-SEC-001" {
		t.Fatalf("expected AE-SEC-001 first, got %q", sec.RuleID)
	}
	if len(sec.Locations) != 1 || sec.Locations[0].Path != "AGENTS.md" || sec.Locations[0].Line != 14 {
		t.Fatalf("expected location AGENTS.md:14, got %#v", sec.Locations)
	}

	if env := payload.Findings[1]; len(env.Locations) != 0 {
		t.Fatalf("expected no locations for absence finding, got %#v", env.Locations)
	}
}

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
