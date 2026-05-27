# AGENTS.md

Last reviewed: 2026-05-28

## Current State

- Phase: Phase 0 repo-executable closure complete; Phase 1 implementation not started
- Product truth: `docs/internal/architecture/charter-architecture-2026.md`
- Module path: `go.charter.dev/charter`
- Current CLI: bootstrap placeholder only
- First Phase 1 slice start point is defined in `docs/internal/superpowers/checklists/2026-05-28-phase-1-admission.md`

This file is the root operating contract for agents. Keep it universal, compact, and routing-first.

## Core Principles

- Think before coding. State assumptions, surface ambiguity, and ask when the repo or request is unclear.
- Simplicity first. Prefer the smallest correct change; do not add speculative abstraction or configurability.
- Surgical changes. Touch only what the task requires, and clean up only what your changes make obsolete.
- Verify before claiming success. Run the narrowest meaningful checks, then the repo-standard gate when the change is broad enough.

## Documentation Topology

- Root contract docs stay at repo root.
- Internal engineering docs live under `docs/internal/`.
- Future customer-facing docs live under `docs/product/`.

## Documentation Authority

1. `docs/internal/architecture/charter-architecture-2026.md` for product behavior
2. `docs/internal/audit/charter-v1-audit-checklist.md` for manual audit companion detail
3. ADRs in `docs/internal/decisions/` for irreversible constraints
4. root companion docs for execution guidance only
5. HTML artifacts as presentation mirrors only

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Golden path: `moon run :check`
- Focused: `moon run :test`, `moon run :vet`, `moon run :lint`, `moon run :build`, `moon run :docs`, `moon run :security`, `moon run :eval`

## Hard Constraints

- Model knowledge stale by default.
- Before changing tools, SDKs, CI actions, APIs, MCP, schemas, or frameworks: inspect local manifests and lockfiles, fetch latest official docs, then inspect relevant installed skills or tool docs.
- If latest-docs lookup is unavailable, stop and report reduced confidence.
- Prefer repo evidence over memory.
- Prefer pointers to deeper docs over duplicating process detail here.
- Bootstrap keeps tracked MCP config absent until a pinned, reviewed integration exists.
- No LLM calls in Charter core.
- No silent mutation. Diff-first fixes only.
- No secrets in docs, prompts, configs, tests, or logs.
- Fail fast rather than hide broken behavior with generic fallback code.
- No speculative refactors outside task scope.

## Edit Scope

- Default edit zones: repo docs, Go source, tests, specs, ADRs, RFCs, workflows, Moon config, mise config
- Off-limits by default: `.env*`, `secrets/`, signing keys, credentials, production infra, generated local state
- Treat `docs/internal/architecture/charter-architecture-2026.md` as canonical for product behavior

## Architecture

- Single root Go module. No `go.work`. No extra modules.
- Command entrypoint in `cmd/charter/`.
- Non-public code in `internal/`.
- Public Go API deferred until a stable external integration surface exists.
- Contract-first for APIs and schemas.
- ADR before irreversible architecture changes. RFC before cross-cutting changes.
- Read the relevant ADR or RFC before changing package boundaries, trust model, config loading, or release posture.
- `charter fix` must always diff before apply; never silent mutation.

## Repo Flow

- Hooks managed by `hk` via `hk.pkl`
- Install hooks with `./scripts/install-hooks.sh`
- Pre-commit runs `moon run :lint` and `moon run :docs`
- Commit-msg enforces Conventional Commits
- Pre-push runs `moon run :test` and `moon run :security`

## Context Loading

Use progressive disclosure. Load only the deeper docs relevant to the current task.

- `CONTEXT_MAP.md`: load map
- `ARCHITECTURE.md`: module layout, slices, error contracts
- `SECURITY.md`: secrets, MCP, supply-chain posture
- `CONTRIBUTING.md`: workflow, commits, PRs, ADR/RFC expectations
- `TESTING.md`: fixtures, evals, verification commands
- `PERMISSIONS.md`: off-limits paths, escalation, destructive-action policy
