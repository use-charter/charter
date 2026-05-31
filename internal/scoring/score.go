package scoring

import "go.charter.dev/charter/internal/findings"

type Result struct {
	Blocker int
	High    int
	Medium  int
	Low     int
	Base    int
	Final   int
}

func Calculate(all []findings.Finding) Result {
	result := Result{}

	for _, finding := range all {
		if finding.Informational {
			continue
		}
		switch finding.Severity {
		case findings.SeverityBlocker:
			result.Blocker++
		case findings.SeverityHigh:
			result.High++
		case findings.SeverityMedium:
			result.Medium++
		case findings.SeverityLow:
			result.Low++
		}
	}

	result.Base = 100 - (result.Blocker * 20) - (result.High * 10) - (result.Medium * 4) - result.Low
	if result.Base < 0 {
		result.Base = 0
	}

	result.Final = result.Base
	if result.Blocker > 0 && result.Final > 59 {
		result.Final = 59
	}

	for _, finding := range all {
		if finding.Informational {
			continue
		}
		if finding.Cap > 0 && result.Final > finding.Cap {
			result.Final = finding.Cap
		}
	}

	return result
}
