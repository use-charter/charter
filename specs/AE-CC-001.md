# AE-CC-001

- Severity: Blocker
- Category: Agent Config
- Description: Agent config must not contain dangerous shell patterns.
- Detection logic: flag shell injection, destructive commands, and unsafe escalation.
- Pass example: explicit scoped commands only.
- Fail example: `rm -rf` or untrusted interpolation in a hook.
- Remediation: replace with safe, bounded commands.
- Related ADRs: ADR-0006
- Related evals: None yet
