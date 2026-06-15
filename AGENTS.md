# AGENTS.md

Charter — an offline-first Go CLI that scores a repository 0–100 on how safely an
AI coding agent can work in it, then hands back the exact fix for every gap.
Deterministic. No LLM in the loop.

Last reviewed: 2026-06-15

## How to work here

Four principles. They bias toward caution over speed — worth it on real work, skip
the ceremony on trivial edits.

1. **Think before coding.** Don't assume. Don't hide confusion. Surface tradeoffs.
   State your interpretation and ask when the request is ambiguous — Charter is a
   deterministic contract that people gate CI on, so a wrong guess ships as a wrong number.
2. **Simplicity first.** Minimum code that solves the problem. Nothing speculative —
   no unasked abstractions, flags, or error handling. If a senior engineer would call
   it overcomplicated, cut it.
3. **Surgical changes.** Touch only what you must. Match the surrounding style. Remove
   only the imports and variables your change orphaned. No drive-by refactors of code
   that isn't broken.
4. **Goal-driven execution.** Define success criteria, then loop until verified — don't
   narrate steps, close a gate. The gates here are concrete: `moon run :check` green ·
   the rule's pass/fail examples in `docs/internal/specs/` · the score threshold.

## What Charter is

- Offline-first Go CLI scoring repos for AI-agent readiness (deterministic).
- Public Mintlify docs in `docs/product/`; marketing site, founder dashboard, and blog in `web/` (Astro).
- Stack: Go 1.26.3, Moonrepo, mise, hk, GHA, Bun.
- License: Apache-2.0 OSS core; DCO-first.
- Product truth: `docs/internal/architecture/charter-architecture-2026.md`.
- CLI: init, doctor, explain, report, fix, suppress, version (flags: architecture §1.8).
- Gate: `charter.yaml` `policy.profile`/`policy.threshold`; `--threshold` overrides.
- Rules (v1, 18): AE-CTX-001/002/004/006, AE-ENV-001, AE-CI-002, AE-SEC-001/002, AE-MCP-001/002/003, AE-CC-001/002, AE-TEST-001, AE-AUTO-001, AE-SUPPRESS-001/002/003.
- GitHub Action in `action/` (`use-charter/charter-action`); `moon run :perf` checks the 50k-file budget.

## Documentation

- Topology: contract docs at root; engineering in `docs/internal/`; customer in `docs/product/`.
- Authority: architecture md owns behavior; ADRs hold irreversible constraints; HTML mirrors only.

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Verify: `moon run :check`
- Smoke: `go run ./cmd/charter doctor --path . --threshold 80`

## Hard constraints

Non-negotiable. Breaking one is a bug, not a style choice.

- Before changing tools/SDKs/CI/APIs/MCP/schemas/frameworks: inspect local manifests/lockfiles, then fetch latest docs. Never a version from memory.
- Tracked MCP config stays absent until a pinned, reviewed integration exists.
- No LLM calls in Charter core.
- No silent mutation. Diff-first fixes only.
- Fail fast. No speculative refactors.

## Edit scope

- Default zones: tracked docs, Go source, tests, specs, ADRs, RFCs, workflows, Moon/mise config.
- Off-limits: `.env*`, `secrets/`, signing keys, credentials, production infra, generated local state.

## Repo flow

- Hooks managed by `hk` via `hk.pkl`.
- Pre-commit: `moon run :lint` and `moon run :docs`.
- Pre-push: `moon run :test` and `moon run :security`.

## Read before you touch it

- `ARCHITECTURE.md` — module layout, seams, error contracts.
- `SECURITY.md` — secrets, MCP, supply-chain posture.
- `CONTRIBUTING.md` — workflow, commits, PRs, ADR/RFC rules.
- `TESTING.md` — fixtures, evals, verification commands.
- `PERMISSIONS.md` — off-limits paths, escalation, destructive-action policy.
