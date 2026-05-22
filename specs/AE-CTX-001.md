# AE-CTX-001

- Severity: Blocker
- Category: Context
- Description: Agent context file must exist, be meaningful, and fit budget.
- Detection logic: check canonical agent-visible files for presence, content, and size budget.
- Pass example: `AGENTS.md` exists with commands, constraints, and context loading.
- Fail example: no agent context file or oversized empty boilerplate.
- Remediation: create or tighten `AGENTS.md`.
- Related ADRs: ADR-0006
- Related evals: None yet
