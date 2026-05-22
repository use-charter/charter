# AE-CTX-004

- Severity: Medium
- Category: Context
- Description: `.gitignore` should exclude common agent and local artifact noise.
- Detection logic: inspect ignore patterns for env files and local agent artifacts.
- Pass example: local env and cache files ignored, committed contracts preserved.
- Fail example: local agent outputs and env files tracked by default.
- Remediation: add safe ignore patterns.
- Related ADRs: ADR-0006
- Related evals: None yet
