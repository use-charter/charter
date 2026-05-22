# AE-CC-002

- Severity: High
- Category: Agent Config
- Description: Agent edit scope must be explicitly constrained.
- Detection logic: look for protected path guidance and lack of broad unrestricted scope.
- Pass example: off-limits paths documented in `AGENTS.md` and `PERMISSIONS.md`.
- Fail example: blanket full-repo write guidance.
- Remediation: add path and action boundaries.
- Related ADRs: ADR-0006
- Related evals: None yet
