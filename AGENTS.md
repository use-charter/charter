# AGENTS.md

Charter — an offline-first Go CLI that scores a repo 0–100 on how safely an AI
coding agent can work in it, then hands back the fix. Deterministic. No LLM.

Last reviewed: 2026-06-15

## How to work here

Bias to caution over speed on real work; trivial edits can skip the ceremony.

- **Think before coding.** Don't assume. Don't hide confusion. Surface tradeoffs.
- **Simplicity first.** Minimum code that solves the problem. Nothing speculative.
- **Surgical changes.** Touch only what you must. Match the surrounding style. No drive-by refactors.
- **Goal-driven.** Define success criteria, then loop until verified — close a gate (`moon run :check`), don't narrate steps.

## What Charter is

- Mintlify docs in `docs/product/`; site, dashboard, blog in `web/` (Astro).
- Stack: Go 1.26.3, Moonrepo, mise, hk, GHA, Bun. License: Apache-2.0; DCO-first.
- Product truth: `docs/internal/architecture/charter-architecture-2026.md`.
- CLI: init, doctor, explain, report, fix, suppress, version.
- Gate: `charter.yaml` `policy.profile`/`policy.threshold`; `--threshold` overrides.
- 18 rules / 9 categories (`AE-*`); full list in `docs/internal/specs/` or `charter explain`.
- GitHub Action in `action/`; `moon run :perf` checks the 50k-file budget.

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Verify: `moon run :check` · Smoke: `go run ./cmd/charter doctor --path . --threshold 80`

## Hard constraints

- Before changing tools/SDKs/CI/APIs/MCP/schemas/frameworks: inspect local manifests/lockfiles, then fetch latest docs. Never a version from memory.
- Tracked MCP config stays absent until a pinned, reviewed integration exists.
- No LLM in core. No silent mutation — diff-first fixes only. Fail fast.

## Edit scope

- Default: tracked docs, Go source, tests, specs, ADRs, RFCs, workflows, Moon/mise config.
- Off-limits: `.env*`, `secrets/`, signing keys, credentials, production infra, generated state.

## Repo flow

- Hooks via `hk` (`hk.pkl`). Pre-commit: `:lint` + `:docs`. Pre-push: `:test` + `:security`.

## Read before you touch it

- `ARCHITECTURE.md` — module layout, seams, error contracts.
- `SECURITY.md` / `PERMISSIONS.md` — secrets, MCP, supply-chain; off-limits, escalation.
- `CONTRIBUTING.md` / `TESTING.md` — workflow, commits, ADR/RFC; fixtures, evals, verification.
- `CONTEXT_MAP.md` — what to read per task.
