package json

import (
	encodingjson "encoding/json"
	"sort"

	"go.charter.dev/charter/internal/doctor"
	"go.charter.dev/charter/internal/findings"
)

type payload struct {
	RepoPath  string          `json:"repo_path"`
	Threshold int             `json:"threshold"`
	Passed    bool            `json:"passed"`
	Findings  []findingDTO    `json:"findings"`
	Summary   severitySummary `json:"summary"`
	Score     scoreSummary    `json:"score"`
}

type findingDTO struct {
	RuleID      string            `json:"rule_id"`
	Severity    findings.Severity `json:"severity"`
	Category    string            `json:"category"`
	Summary     string            `json:"summary"`
	Remediation string            `json:"remediation"`
	Evidence    []string          `json:"evidence"`
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
		wi := weight(ordered[i].Severity)
		wj := weight(ordered[j].Severity)
		if wi != wj {
			return wi > wj
		}

		return ordered[i].RuleID < ordered[j].RuleID
	})

	return encodingjson.Marshal(payload{
		RepoPath:  result.Root,
		Threshold: result.Threshold,
		Passed:    result.Passed,
		Findings:  toFindingDTOs(ordered),
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

func toFindingDTOs(findingsList []findings.Finding) []findingDTO {
	dtos := make([]findingDTO, 0, len(findingsList))
	for _, finding := range findingsList {
		dtos = append(dtos, findingDTO{
			RuleID:      finding.RuleID,
			Severity:    finding.Severity,
			Category:    finding.Category,
			Summary:     finding.Summary,
			Remediation: finding.Remediation,
			Evidence:    finding.Evidence,
		})
	}

	return dtos
}

func weight(severity findings.Severity) int {
	switch severity {
	case findings.SeverityBlocker:
		return 4
	case findings.SeverityHigh:
		return 3
	case findings.SeverityMedium:
		return 2
	case findings.SeverityLow:
		return 1
	default:
		return 0
	}
}
