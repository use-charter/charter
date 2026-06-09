# Context Map

## Start Here

- `AGENTS.md`: repo contract and working loop
- `docs/internal/architecture/charter-architecture-2026.md`: product truth

## Documentation Topology

- Root contract docs stay at repo root.
- `docs/internal/` holds repo engineering knowledge.
- `docs/product/` holds the customer-facing Mintlify documentation site.

## Load by Task

- If editing Go code: read `ARCHITECTURE.md`, then the relevant spec or ADR, then `TESTING.md`
- If editing workflows or toolchain config: read `SECURITY.md`, `PERMISSIONS.md`, and `CONTRIBUTING.md`
- If changing contracts, rules, or machine-readable outputs: read `ARCHITECTURE.md`, `docs/internal/specs/`, `schemas/`, and related ADRs or RFCs
- If changing docs or process: read `CONTRIBUTING.md`, then linked ADRs, RFCs, and playbooks
- If planning a new subsystem: read `docs/internal/architecture/charter-architecture-2026.md`, `ARCHITECTURE.md`, `docs/internal/decisions/`, and `docs/internal/rfcs/`

## Decision Graph

- `docs/internal/decisions/`: accepted architecture decisions
- `docs/internal/rfcs/`: proposed changes before broad implementation
- `docs/internal/specs/`: rule-level behavior contracts
- `evals/`: verification and acceptance artifacts
- `docs/internal/architecture/c4/`: architecture maps
- `docs/internal/runbooks/`: operational responses
- `docs/internal/playbooks/`: repeatable task flows

## Linking Rules

- ADRs should link related RFCs and specs
- Specs should link related ADRs and evals
- RFCs should link acceptance checks and affected playbooks
- Runbooks should link relevant playbooks, ADRs, and workflows
