package findings

type Severity string

const (
	SeverityBlocker Severity = "BLOCKER"
	SeverityHigh    Severity = "HIGH"
	SeverityMedium  Severity = "MEDIUM"
	SeverityLow     Severity = "LOW"
)

type Finding struct {
	RuleID      string
	Severity    Severity
	Category    string
	Summary     string
	Remediation string
	Evidence    []string
	// Cap, when greater than zero, caps the final score at this value whenever
	// this finding is present. Rules that own a hard score ceiling (e.g. raw
	// secret detection) set it so scoring stays a pure formula engine.
	Cap int
}
