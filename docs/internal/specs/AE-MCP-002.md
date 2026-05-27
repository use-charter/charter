# AE-MCP-002

- Severity: High
- Category: MCP Safety
- Description: Remote MCP origins must be known or allowlisted.
- Detection logic: compare remote endpoints to catalog or repo allowlist.
- Pass example: trusted local or allowlisted remote.
- Fail example: unknown remote server origin.
- Remediation: allowlist or replace the server.
- Related ADRs: ADR-0006, ADR-0007
- Related evals: None yet
