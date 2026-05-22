# AE-SUPPRESS-003

- Severity: Medium
- Category: Governance
- Description: High suppression rate is an informational governance signal.
- Detection logic: compare suppressed finding count to total findings.
- Pass example: suppression rate stays below threshold or is explicitly justified.
- Fail example: suppressions dominate findings over time.
- Remediation: review rules, calibrate false positives, or revisit accepted risk.
- Related ADRs: ADR-0006
- Related evals: None yet
