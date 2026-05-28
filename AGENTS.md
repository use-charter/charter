# AGENTS.md

Last reviewed: 2026-05-28

## Current State

- Phase: Phase 1 Slice 1 implemented; first real `charter doctor` path exists
- Charter is an offline-first Go CLI that scans repositories for AI-agent readiness and scores deterministic repo safety signals.
- Tech stack: Go 1.26.3, Moonrepo, mise, hk, GitHub Actions.
- Product truth: `docs/internal/architecture/charter-architecture-2026.md`
- Module path: `go.charter.dev/charter`
- Current CLI: `charter doctor` text output with `--path`, `--threshold`, `--quiet`
- Implemented rules: `AE-CTX-001`, `AE-CTX-002`, `AE-CTX-004`, `AE-ENV-001`, `AE-CI-002`

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Verify: `moon run :check`
- Slice 1 smoke: `go run ./cmd/charter doctor --path . --threshold 80`

## Hard Constraints

- Before changing tools, SDKs, CI actions, APIs, MCP, schemas, or frameworks: inspect local manifests and lockfiles, fetch latest official docs, then inspect relevant installed skills or tool docs.
- Bootstrap keeps tracked MCP config absent until a pinned, reviewed integration exists.
- No LLM calls in Charter core.
- No silent mutation. Diff-first fixes only.
- Fail fast. No speculative refactors.

## Edit Scope

- Default edit zones: tracked docs, Go source, tests, specs, ADRs, RFCs, workflows, Moon config, mise config.
- Off-limits by default: `.env*`, `secrets/`, signing keys, credentials, production infra, generated local state
- Treat `docs/internal/architecture/charter-architecture-2026.md` as canonical for product behavior

## Repo Flow

- Hooks managed by `hk` via `hk.pkl`
- Pre-commit runs `moon run :lint` and `moon run :docs`
- Pre-push runs `moon run :test` and `moon run :security`

## Context Loading

- `ARCHITECTURE.md`: module layout, Slice 1 seams, error contracts
- `SECURITY.md`: secrets, MCP, supply-chain posture
- `CONTRIBUTING.md`: workflow, commits, PRs, ADR/RFC expectations
- `TESTING.md`: fixtures, evals, verification commands
- `PERMISSIONS.md`: off-limits paths, escalation, destructive-action policy
