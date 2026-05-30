package findings

type Severity string

const (
	SeverityBlocker Severity = "BLOCKER"
	SeverityHigh    Severity = "HIGH"
	SeverityMedium  Severity = "MEDIUM"
	SeverityLow     Severity = "LOW"
)

// Location identifies a physical site implicated by a finding. Line is 1-based;
// a Line of 0 means the finding is file-level (no specific line). Locations map
// directly onto SARIF result.locations[].physicalLocation.
type Location struct {
	Path string
	Line int
}

type Finding struct {
	RuleID      string
	Severity    Severity
	Category    string
	Summary     string
	Remediation string
	Evidence    []string
	// Locations are the physical sites implicated by the finding, in order.
	// Empty for absence findings (e.g. a missing context or toolchain file).
	Locations []Location
	// Cap, when greater than zero, caps the final score at this value whenever
	// this finding is present. Rules that own a hard score ceiling (e.g. raw
	// secret detection) set it so scoring stays a pure formula engine.
	Cap int
}
