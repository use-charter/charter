# Testdata

Use this directory for deterministic fixtures used by rules, renderers, fixes, and evals.

First-slice expectation:

- the first implemented rule should add at least one pass fixture and one fail fixture here, or document why a direct code-level failing test is sufficient
- fixture names should describe the rule and scenario, not implementation detail
- fixtures must remain secret-safe and fully reviewable in git
- each fixture should map back to a spec or rule contract

Proof model reference:

- `docs/internal/superpowers/checklists/2026-05-28-first-slice-proof-model.md`

## Secret-Rule Fixtures (AE-SEC-001 and AE-SEC-002)

Two pass fixtures prove that the secret detector correctly neutralizes environment-variable references and placeholders:

- **`repos/pass-secrets-agent/`**: Demonstrates that AE-SEC-001 **passes** when agent-visible files (AGENTS.md, CLAUDE.md, etc.) contain only environment-variable references (`${OPENAI_API_KEY}`, `$OPENAI_API_KEY`) and placeholder strings (`your-api-key-here`). These patterns are never flagged as secrets.

- **`repos/pass-secrets-config/`**: Demonstrates that AE-SEC-002 **passes** when MCP config files (.mcp.json, etc.) use environment-variable references in `env` objects (`"OPENAI_API_KEY": "${OPENAI_API_KEY}"`) rather than literal secret values. Placeholders are also safe.

Both fixtures are git-safe (no real or fake high-confidence secret patterns) and fully reviewable. Fail-case scenarios are tested inline in `internal/rules/secrets/sec001_test.go` and `sec002_test.go` using table-driven tests with generated fake patterns.
