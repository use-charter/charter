# AE-SUPPRESS-002

- Severity: High
- Category: Governance
- Description: Permanent suppressions require an approver.
- Detection logic: inspect permanent suppression entries for approver metadata.
- Pass example: permanent suppression includes approver.
- Fail example: permanent suppression without approver.
- Remediation: add approver or convert to time-bounded suppression.
- Related ADRs: ADR-0006
- Related evals: None yet
