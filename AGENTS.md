# AGENTS.md

Last reviewed: 2026-06-15

## Current State

- Offline-first Go CLI scoring repos for AI-agent readiness (deterministic).
- Public Mintlify docs in `docs/product/`; marketing site, founder dashboard, and blog in `web/` (Astro).
- Stack: Go 1.26.3, Moonrepo, mise, hk, GHA, Bun.
- License: Apache-2.0 OSS core; DCO-first
- Product truth: `docs/internal/architecture/charter-architecture-2026.md`
- CLI: init, doctor, explain, report, fix, suppress, version (flags: architecture §1.8)
- Gate: `charter.yaml` `policy.profile`/`policy.threshold`; `--threshold` overrides
- Rules (v1, 18): AE-CTX-001/002/004/006, AE-ENV-001, AE-CI-002, AE-SEC-001/002, AE-MCP-001/002/003, AE-CC-001/002, AE-TEST-001, AE-AUTO-001, AE-SUPPRESS-001/002/003
- GitHub Action in `action/` (`use-charter/charter-action`); `moon run :perf` checks the 50k-file budget

## Documentation

- Topology: contract docs at root; engineering in `docs/internal/`; customer in `docs/product/`.
- Authority: architecture md owns behavior; ADRs hold irreversible constraints; HTML mirrors only.

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Verify: `moon run :check`
- Smoke: `go run ./cmd/charter doctor --path . --threshold 80`

## Hard Constraints

- Before changing tools/SDKs/CI/APIs/MCP/schemas/frameworks: inspect local manifests/lockfiles, then fetch latest docs.
- Tracked MCP config stays absent until a pinned, reviewed integration exists.
- No LLM calls in Charter core.
- No silent mutation. Diff-first fixes only.
- Fail fast. No speculative refactors.

## Edit Scope

- Default edit zones: tracked docs, Go source, tests, specs, ADRs, RFCs, workflows, Moon/mise config.
- Off-limits: `.env*`, `secrets/`, signing keys, credentials, production infra, generated local state.

## Repo Flow

- Hooks managed by `hk` via `hk.pkl`
- Pre-commit: `moon run :lint` and `moon run :docs`
- Pre-push: `moon run :test` and `moon run :security`

## Context Loading

- `ARCHITECTURE.md`: module layout, seams, error contracts
- `SECURITY.md`: secrets, MCP, supply-chain posture
- `CONTRIBUTING.md`: workflow, commits, PRs, ADR/RFC rules
- `TESTING.md`: fixtures, evals, verification commands
- `PERMISSIONS.md`: off-limits paths, escalation, destructive-action policy
