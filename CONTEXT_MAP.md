# Context Map

## Start Here

- `AGENTS.md`: repo contract and working loop
- `docs/architecture/charter-architecture-2026.md`: product truth

## Load by Task

- If editing Go code: read `ARCHITECTURE.md`, then the relevant spec or ADR, then `TESTING.md`
- If editing workflows or toolchain config: read `SECURITY.md`, `PERMISSIONS.md`, and `CONTRIBUTING.md`
- If changing contracts, rules, or machine-readable outputs: read `ARCHITECTURE.md`, `specs/`, `schemas/`, and related ADRs or RFCs
- If changing docs or process: read `CONTRIBUTING.md`, then linked ADRs, RFCs, and playbooks
- If planning a new subsystem: read `docs/architecture/charter-architecture-2026.md`, `ARCHITECTURE.md`, `decisions/`, and `rfcs/`

## Decision Graph

- `decisions/`: accepted architecture decisions
- `rfcs/`: proposed changes before broad implementation
- `specs/`: rule-level behavior contracts
- `evals/`: verification and acceptance artifacts
- `docs/architecture/c4/`: architecture maps
- `runbooks/`: operational responses
- `playbooks/`: repeatable task flows

## Linking Rules

- ADRs should link related RFCs and specs
- Specs should link related ADRs and evals
- RFCs should link acceptance checks and affected playbooks
- Runbooks should link relevant playbooks, ADRs, and workflows
