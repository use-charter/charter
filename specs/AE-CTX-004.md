# AE-CTX-004

- Severity: Medium
- Category: Context
- Description: `.gitignore` should exclude common agent and local artifact noise.
- Detection logic: inspect `.gitignore` for local agent/session/cache patterns and verify that shared team config stays committed while local state does not.
- Pass example: `.charter/`, `*.charter-session`, `.claude/local/`, `.cursor/cache/`, `.hk/`, and `.env*` are handled correctly without ignoring committed team contracts.
- Fail example: local agent output, cache directories, or secret-bearing env files remain trackable by default.
- Evidence expectations: relevant `.gitignore` lines and any tracked local-agent artifacts that should not be in git.
- Edge cases: `.cursor/rules`, `.claude/settings.json`, and similar team-owned config should stay committed even while local cache/state stays ignored.
- Remediation: add precise ignore patterns and remove any accidentally tracked local artifacts from git.
- Related ADRs: ADR-0006
- Related evals: None yet
