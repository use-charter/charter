# AGENTS.md

Last reviewed: 2026-06-01

## Current State

- Offline-first Go CLI scoring repos for AI-agent readiness (deterministic).
- Phase: Phase 1 Slice 11 implemented on top of the real `charter doctor` path.
- Stack: Go 1.26.3, Moonrepo, mise, hk, GitHub Actions, Bun TS scripts.
- License: Apache-2.0 OSS core; DCO-first contribution model
- Product truth: `docs/internal/architecture/charter-architecture-2026.md`; module: `go.use-charter.dev/charter`
- CLI: `charter init`, `charter doctor` (`--path --threshold --quiet --format text|json|markdown|sarif --out`), `charter suppress`, `charter version`
- Gate: `charter.yaml` `policy.profile`/`policy.threshold`; `--threshold` overrides
- Implemented rules (full v1 set): AE-CTX-001/002/004, AE-ENV-001, AE-CI-002, AE-SEC-001/002, AE-MCP-001/002/003, AE-CC-001/002, AE-SUPPRESS-001/002/003
- GitHub Action in `action/` (`use-charter/charter-action`); `moon run :perf` validates the 50k-file perf budget

## Documentation

- Topology: contract docs at root; engineering in `docs/internal/`; customer in `docs/product/`.
- Authority: architecture md owns behavior; ADRs hold irreversible constraints; HTML mirrors only.

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Verify: `moon run :check`
- Smoke: `go run ./cmd/charter doctor --path . --threshold 80`

## Hard Constraints

- Before changing tools/SDKs/CI/APIs/MCP/schemas/frameworks: inspect local manifests/lockfiles, then fetch latest docs.
- Bootstrap keeps tracked MCP config absent until a pinned, reviewed integration exists.
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
