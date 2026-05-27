# AE-MCP-001

- Severity: High
- Category: MCP Safety
- Description: MCP dependencies and servers must be pinned.
- Detection logic: flag `@latest`, ranges, and floating git refs in MCP config.
- Pass example: exact version or digest.
- Fail example: unpinned server package.
- Remediation: pin exact version or digest.
- Related ADRs: ADR-0006, ADR-0007
- Related evals: None yet
