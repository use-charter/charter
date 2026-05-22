# AE-CTX-002

- Severity: Medium
- Category: Context
- Description: Agent context must remain current with repo reality.
- Detection logic: compare declared stack, commands, and boundaries to repo state.
- Pass example: docs match current toolchain and workflow.
- Fail example: AGENTS says one stack, repo uses another.
- Remediation: update context files and decision links.
- Related ADRs: ADR-0006
- Related evals: None yet
