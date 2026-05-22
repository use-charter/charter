# AE-MCP-003

- Severity: High
- Category: MCP Safety
- Description: Sensitive remote MCP servers must declare auth metadata.
- Detection logic: inspect remote MCP config for OAuth 2.1 / PKCE metadata where needed.
- Pass example: protected remote server declares auth.
- Fail example: sensitive remote endpoint with no auth declaration.
- Remediation: add explicit auth metadata or switch integration mode.
- Related ADRs: ADR-0006, ADR-0007
- Related evals: None yet
