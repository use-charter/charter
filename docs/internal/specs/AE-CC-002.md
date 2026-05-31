# AE-CC-002

- Severity: High
- Category: Agent Config
- Description: The agent context must explicitly constrain the agent's edit scope by declaring concrete off-limits paths, so an agent is not granted implicit full-repo write access (OWASP MCP Top 10 beta, MCP02 Privilege Escalation via Scope Creep).
- Detection logic: reads the agent context files (the shared `agentcontext` set — AGENTS.md, CLAUDE.md, .cursor/rules, .windsurfrules, .github/copilot-instructions.md, opencode.md, codex.md, DESIGN.md, SKILL.md) plus `PERMISSIONS.md` when present. Passes when the context declares a concrete off-limits / protected-path boundary — recognized sensitive-path tokens presented as off-limits (`.env`, `secrets`, `.github/workflows`, `terraform`, `infra`, `db/migrations`, `credentials`) or an explicit reference to `PERMISSIONS.md`. Flags High when no concrete off-limits-path declaration is found in any context file. This is intentionally stricter than `AE-CTX-001`, which only requires a generic edit-boundary mention.
- Pass example: `AGENTS.md` with an "Edit Scope / Off-limits" section listing `.env*`, `secrets/`, and signing keys (Charter's own AGENTS.md), or a context file that points to `PERMISSIONS.md` — passes.
- Fail example: an `AGENTS.md` that describes the project and stack but never declares any off-limits paths or edit boundaries — flagged High.
- Evidence expectations: a file-level location (the context file evaluated) and an evidence string stating that no concrete off-limits-path declaration was found (and which context files were checked).
- Edge cases: a single-purpose repo with no sensitive paths may legitimately have broad scope (documented false-positive risk); the check is presence-based on concrete sensitive-path tokens, not a semantic policy evaluation; when no agent context file exists at all, `AE-CTX-001` already fires (Blocker) and `AE-CC-002` does not duplicate the absence finding.
- Remediation: add an explicit "Off-limits for agents" section to the agent context (or `PERMISSIONS.md`) listing at minimum `.github/workflows/`, `terraform/` or `infra/`, `db/migrations/`, `.env*`, and `secrets/`, then commit the change.
- Scoring impact: each finding is `High` (−10); no hard cap.
- Related ADRs: ADR-0012, ADR-0006, ADR-0009
- Related evals: None yet
