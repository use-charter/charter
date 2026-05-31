# AGENTS.md

Last reviewed: 2026-06-01

## Current State

- Offline-first Go CLI that scans repos for AI-agent readiness with deterministic scoring.
- Phase: Phase 1 Slice 7 implemented on top of the real `charter doctor` path.
- Stack: Go 1.26.3, Moonrepo, mise, hk, GitHub Actions, Bun TS scripts.
- License: Apache-2.0 OSS core; DCO-first contribution model
- Product truth: `docs/internal/architecture/charter-architecture-2026.md`; module: `go.charter.dev/charter`
- CLI: `charter doctor` (`--path --threshold --quiet --format text|json|markdown`), `charter suppress`
- Implemented rules (full v1 set): AE-CTX-001/002/004, AE-ENV-001, AE-CI-002, AE-SEC-001/002, AE-MCP-001/002/003, AE-CC-001/002, AE-SUPPRESS-001/002/003

## Documentation

- Topology: root contract docs at repo root; engineering docs under `docs/internal/`; customer docs under `docs/product/`.
- Authority: architecture md owns behavior; audit md is companion; ADRs hold irreversible constraints; HTML mirrors only.

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Verify: `moon run :check`
- Smoke: `go run ./cmd/charter doctor --path . --threshold 80`

## Hard Constraints

- Before changing tools, SDKs, CI, APIs, MCP, schemas, or frameworks: inspect local manifests/lockfiles, fetch latest official docs, then check installed skills/tool docs.
- Bootstrap keeps tracked MCP config absent until a pinned, reviewed integration exists.
- No LLM calls in Charter core.
- No silent mutation. Diff-first fixes only.
- Fail fast. No speculative refactors.

## Edit Scope

- Default edit zones: tracked docs, Go source, tests, specs, ADRs, RFCs, workflows, Moon/mise config.
- Off-limits: `.env*`, `secrets/`, signing keys, credentials, production infra, generated local state.
- Canonical behavior owner: `docs/internal/architecture/charter-architecture-2026.md`

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
