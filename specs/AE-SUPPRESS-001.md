# AE-SUPPRESS-001

- Severity: Medium
- Category: Governance
- Description: Suppressions require a human-readable reason.
- Detection logic: inspect suppression comments or files for missing reason fields.
- Pass example: suppression includes explicit reason.
- Fail example: bare ignore directive.
- Remediation: add meaningful reason text.
- Related ADRs: ADR-0006
- Related evals: None yet
