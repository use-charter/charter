# AE-SEC-002

- Severity: Blocker
- Category: Secrets
- Description: No secret-like values in MCP configuration or adjacent configs.
- Detection logic: scan MCP and tool configs for raw credential values.
- Pass example: env var references only.
- Fail example: raw API key in config.
- Remediation: move to env var or secret store.
- Related ADRs: ADR-0006, ADR-0007
- Related evals: None yet
