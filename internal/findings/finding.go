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
}
