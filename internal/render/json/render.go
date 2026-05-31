package json

import (
	encodingjson "encoding/json"
	"sort"

	"go.charter.dev/charter/internal/doctor"
	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/suppress"
)

type payload struct {
	RepoPath   string          `json:"repo_path"`
	Threshold  int             `json:"threshold"`
	Passed     bool            `json:"passed"`
	Findings   []findingDTO    `json:"findings"`
	Suppressed []suppressedDTO `json:"suppressed"`
	Summary    severitySummary `json:"summary"`
	Score      scoreSummary    `json:"score"`
}

type findingDTO struct {
	RuleID        string            `json:"rule_id"`
	Severity      findings.Severity `json:"severity"`
	Category      string            `json:"category"`
	Summary       string            `json:"summary"`
	Remediation   string            `json:"remediation"`
	Locations     []locationDTO     `json:"locations"`
	Evidence      []string          `json:"evidence"`
	Informational bool              `json:"informational,omitempty"`
}

type suppressedDTO struct {
	RuleID    string            `json:"rule_id"`
	Severity  findings.Severity `json:"severity"`
	Source    string            `json:"source"`
	Reason    string            `json:"reason,omitempty"`
	Approver  string            `json:"approver,omitempty"`
	Expires   string            `json:"expires,omitempty"`
	Locations []locationDTO     `json:"locations"`
}

type locationDTO struct {
	Path string `json:"path"`
	Line int    `json:"line"`
}

type severitySummary struct {
	Blocker int `json:"blocker"`
	High    int `json:"high"`
	Medium  int `json:"medium"`
	Low     int `json:"low"`
}

type scoreSummary struct {
	Base  int `json:"base"`
	Final int `json:"final"`
}

func Render(result doctor.Result) ([]byte, error) {
	ordered := append([]findings.Finding(nil), result.Findings...)
	sort.SliceStable(ordered, func(i, j int) bool {
		wi, wj := ordered[i].Severity.Weight(), ordered[j].Severity.Weight()
		if wi != wj {
			return wi > wj
		}

		return ordered[i].RuleID < ordered[j].RuleID
	})

	return encodingjson.Marshal(payload{
		RepoPath:   result.Root,
		Threshold:  result.Threshold,
		Passed:     result.Passed,
		Findings:   toFindingDTOs(ordered),
		Suppressed: toSuppressedDTOs(result.Suppressed),
		Summary: severitySummary{
			Blocker: result.Score.Blocker,
			High:    result.Score.High,
			Medium:  result.Score.Medium,
			Low:     result.Score.Low,
		},
		Score: scoreSummary{
			Base:  result.Score.Base,
			Final: result.Score.Final,
		},
	})
}

func toSuppressedDTOs(list []suppress.Suppressed) []suppressedDTO {
	ordered := append([]suppress.Suppressed(nil), list...)
	sort.SliceStable(ordered, func(i, j int) bool {
		wi, wj := ordered[i].Finding.Severity.Weight(), ordered[j].Finding.Severity.Weight()
		if wi != wj {
			return wi > wj
		}
		return ordered[i].Finding.RuleID < ordered[j].Finding.RuleID
	})

	dtos := make([]suppressedDTO, 0, len(ordered))
	for _, s := range ordered {
		locations := make([]locationDTO, 0, len(s.Finding.Locations))
		for _, loc := range s.Finding.Locations {
			locations = append(locations, locationDTO{Path: loc.Path, Line: loc.Line})
		}
		dtos = append(dtos, suppressedDTO{
			RuleID:    s.Finding.RuleID,
			Severity:  s.Finding.Severity,
			Source:    s.Source,
			Reason:    s.Reason,
			Approver:  s.Approver,
			Expires:   s.Expires,
			Locations: locations,
		})
	}
	return dtos
}

func toFindingDTOs(findingsList []findings.Finding) []findingDTO {
	dtos := make([]findingDTO, 0, len(findingsList))
	for _, finding := range findingsList {
		locations := make([]locationDTO, 0, len(finding.Locations))
		for _, loc := range finding.Locations {
			locations = append(locations, locationDTO{Path: loc.Path, Line: loc.Line})
		}

		dtos = append(dtos, findingDTO{
			RuleID:        finding.RuleID,
			Severity:      finding.Severity,
			Category:      finding.Category,
			Summary:       finding.Summary,
			Remediation:   finding.Remediation,
			Locations:     locations,
			Evidence:      finding.Evidence,
			Informational: finding.Informational,
		})
	}

	return dtos
}
