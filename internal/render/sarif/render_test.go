package sarif

import (
	"encoding/json"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/suppress"
)

func sampleResult() doctor.Result {
	return doctor.Result{
		Root: "/repo", Threshold: 80, Passed: false,
		Findings: []findings.Finding{
			{RuleID: "AE-SEC-001", Severity: findings.SeverityBlocker, Category: "Secrets", Summary: "secret in context", Locations: []findings.Location{{Path: "AGENTS.md", Line: 14}}},
			{RuleID: "AE-ENV-001", Severity: findings.SeverityMedium, Category: "Environment", Summary: "no toolchain"},
			{RuleID: "AE-SUPPRESS-003", Severity: findings.SeverityMedium, Category: "Governance", Summary: "high suppression rate", Informational: true, Locations: []findings.Location{{Path: ".charter-suppress.yml"}}},
		},
		Suppressed: []suppress.Suppressed{
			{Finding: findings.Finding{RuleID: "AE-MCP-001", Severity: findings.SeverityHigh, Summary: "unpinned", Locations: []findings.Location{{Path: ".mcp.json", Line: 1}}}, Source: suppress.SourceExternal, Reason: "vendored"},
		},
		Score: scoring.Result{Final: 49},
	}
}

type sarifDoc struct {
	Version string `json:"version"`
	Runs    []struct {
		AutomationDetails struct {
			ID string `json:"id"`
		} `json:"automationDetails"`
		Tool struct {
			Driver struct {
				Name    string `json:"name"`
				Version string `json:"version"`
				Rules   []struct {
					ID                   string `json:"id"`
					DefaultConfiguration struct {
						Level string `json:"level"`
					} `json:"defaultConfiguration"`
					Properties struct {
						Category         string   `json:"category"`
						Severity         string   `json:"severity"`
						Tags             []string `json:"tags"`
						SecuritySeverity string   `json:"security-severity"`
					} `json:"properties"`
				} `json:"rules"`
			} `json:"driver"`
		} `json:"tool"`
		Results []struct {
			RuleID              string            `json:"ruleId"`
			RuleIndex           int               `json:"ruleIndex"`
			Level               string            `json:"level"`
			Kind                string            `json:"kind"`
			PartialFingerprints map[string]string `json:"partialFingerprints"`
			Suppressions        []struct {
				Kind string `json:"kind"`
			} `json:"suppressions"`
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct {
						URI string `json:"uri"`
					} `json:"artifactLocation"`
					Region *struct {
						StartLine int `json:"startLine"`
					} `json:"region"`
				} `json:"physicalLocation"`
			} `json:"locations"`
		} `json:"results"`
	} `json:"runs"`
}

func TestRenderShapeAndMappings(t *testing.T) {
	data, err := Render(sampleResult())
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	var log sarifDoc
	if err := json.Unmarshal(data, &log); err != nil {
		t.Fatalf("invalid SARIF JSON: %v", err)
	}
	if log.Version != "2.1.0" {
		t.Fatalf("version = %q, want 2.1.0", log.Version)
	}
	run := log.Runs[0]
	if run.Tool.Driver.Name != "Charter" || run.Tool.Driver.Version == "" {
		t.Fatalf("driver wrong: %+v", run.Tool.Driver)
	}
	if run.AutomationDetails.ID != "charter" {
		t.Fatalf("automationDetails.id = %q, want charter", run.AutomationDetails.ID)
	}

	// 4 distinct rules, sorted by ID (AE-ENV-001 first).
	if len(run.Tool.Driver.Rules) != 4 || run.Tool.Driver.Rules[0].ID != "AE-ENV-001" {
		t.Fatalf("rules wrong: %+v", run.Tool.Driver.Rules)
	}
	for _, r := range run.Tool.Driver.Rules {
		switch r.ID {
		case "AE-SEC-001":
			if r.Properties.SecuritySeverity != "9.5" || !contains(r.Properties.Tags, "security") {
				t.Fatalf("AE-SEC-001 security classification wrong: %+v", r.Properties)
			}
		case "AE-SUPPRESS-003":
			if r.Properties.SecuritySeverity != "" || len(r.Properties.Tags) != 0 {
				t.Fatalf("informational rule must omit security-severity/tags: %+v", r.Properties)
			}
		}
	}

	if len(run.Results) != 4 {
		t.Fatalf("want 4 results, got %d", len(run.Results))
	}
	for _, r := range run.Results {
		if r.PartialFingerprints["primaryLocationLineHash"] == "" {
			t.Fatalf("%s missing fingerprint", r.RuleID)
		}
		if r.RuleIndex < 0 || r.RuleIndex >= len(run.Tool.Driver.Rules) {
			t.Fatalf("%s ruleIndex out of range", r.RuleID)
		}
		switch r.RuleID {
		case "AE-SEC-001":
			if r.Level != "error" || len(r.Locations) != 1 || r.Locations[0].PhysicalLocation.Region == nil || r.Locations[0].PhysicalLocation.Region.StartLine != 14 {
				t.Fatalf("AE-SEC-001 wrong: %+v", r)
			}
		case "AE-ENV-001":
			if r.Level != "warning" || len(r.Locations) != 0 {
				t.Fatalf("AE-ENV-001 should be warning with no locations: %+v", r)
			}
		case "AE-SUPPRESS-003":
			if r.Kind != "informational" || r.Level != "note" {
				t.Fatalf("informational mapping wrong: %+v", r)
			}
		case "AE-MCP-001":
			if len(r.Suppressions) != 1 || r.Suppressions[0].Kind != "external" {
				t.Fatalf("suppressed result should carry suppressions[external]: %+v", r)
			}
		}
	}

	// Consistency: active results carry an (empty) suppressions array, never null.
	if !strings.Contains(string(data), `"suppressions":[]`) {
		t.Fatalf("expected empty suppressions arrays on active results: %s", data)
	}
}

func TestFingerprintDeterministicAndDistinct(t *testing.T) {
	a := fingerprint(findings.Finding{RuleID: "AE-MCP-001", Locations: []findings.Location{{Path: "a.json", Line: 1}}})
	b := fingerprint(findings.Finding{RuleID: "AE-MCP-001", Locations: []findings.Location{{Path: "a.json", Line: 1}}})
	c := fingerprint(findings.Finding{RuleID: "AE-MCP-001", Locations: []findings.Location{{Path: "a.json", Line: 2}}})
	d := fingerprint(findings.Finding{RuleID: "AE-CTX-001"})
	if a != b {
		t.Fatal("same finding must produce the same fingerprint")
	}
	if a == c || a == d || d == "" {
		t.Fatal("different findings must produce different, non-empty fingerprints")
	}
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
