# AE-SEC-001

- Severity: Blocker
- Category: Secrets
- Description: No secret-like values in agent-visible context files.
- Detection logic: scan agent docs, prompts, and visible config for secret patterns.
- Pass example: only clearly fake placeholders.
- Fail example: live token in `AGENTS.md`.
- Remediation: remove and rotate the secret.
- Related ADRs: ADR-0006, ADR-0007
- Related evals: None yet
