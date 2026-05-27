# Architecture

## Bootstrap Contract

- Single root Go module: `go.charter.dev/charter`
- No `go.work`
- No extra modules during bootstrap
- Charter core performs no LLM calls
- Scanner remains deterministic and offline-first
- Fixes are diff-first and never silently mutate files
- MCP support is deferred until there is a pinned, local-first, least-privilege integration to expose

## Root Layout

- `cmd/`: binary entrypoints and command wiring
- `internal/`: non-public implementation details, detectors, fix engines, and orchestration helpers
- `api/openapi/`: future API contracts before implementation
- `schemas/`: machine-readable config and report contracts
- `specs/`: rule-level behavior contracts
- `decisions/`: accepted ADRs
- `rfcs/`: proposals for cross-cutting or risky changes
- `evals/`: acceptance fixtures for prompt, workflow, and future agent behavior

## Command Tree Intent

The expected CLI tree follows the product authority in `docs/architecture/charter-architecture-2026.md`:

- `charter init`
- `charter doctor`
- `charter report`
- `charter fix`
- `charter suppress`
- `charter version`

Phase 0 should set package boundaries so each command can own a vertical slice without later module churn. Cobra is the intended command framework. Koanf is the intended config-loading stack once real runtime config lands.

## Slice Rule

Vertical slice means a user-facing capability should keep its parser, domain model, verification logic, and remediation logic close together instead of spreading shared behavior into generic utility packages. Shared code is allowed only when at least two slices need the same stable abstraction.

Boundary rules:

- keep external packages private by default under `internal/`
- introduce a public Go package only when a stable external integration surface is proven
- avoid catch-all `util`, `helpers`, or `common` packages
- cross-cutting behavior that changes package boundaries requires an ADR or RFC first

## Error Contract

Agent-readable errors must expose:

- stable error code
- short machine-readable summary
- human remediation guidance
- evidence fields that explain the failure
- no secret-bearing payloads

Example shape:

```text
code: AE-REPRO-001
summary: Missing pinned toolchain version
remediation: Add explicit versions to the bootstrap toolchain config before enabling CI gates.
evidence: ["mise.toml: govulncheck uses an unpinned selector"]
```

Every user-facing rule or command error should be backed by pass/fail examples in `specs/`.
