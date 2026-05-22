# Contributing

## Workflow

1. Read `AGENTS.md`.
2. Load only the context files needed for the task.
3. Check local manifests and latest official docs before code.
4. Make the smallest viable change.
5. Run explicit verification commands.
6. Update ADRs, RFCs, specs, docs, and evals when behavior changes.

## Change Contracts

- Commits: Conventional Commits
- Versioning: SemVer
- Cross-cutting or risky work: RFC first
- Irreversible architecture decisions: ADR first
- New API surfaces: contract-first
- Non-trivial agent workflows or prompts: eval-driven

Trigger rules:

- Add or update an ADR when changing package boundaries, command ownership, trust model, config loading, release posture, or other hard-to-reverse architecture decisions
- Add or update an RFC when changing cross-cutting behavior, introducing a new subsystem, or accepting migration/compatibility risk
- Add or update specs when command behavior, rule semantics, fix behavior, suppressions, or error contracts change
- Add or update evals when prompt, workflow, or agent-facing behavior changes in a non-trivial way
- Add or update API contracts before implementing new HTTP or machine-readable interfaces

## Pull Requests

- Link task, ADR, or RFC when applicable
- Include verification commands and results
- Call out risks, generated code, and docs/spec changes
- Keep diffs surgical unless the task explicitly needs breadth
- Route verification through the root Moon task family so local and CI checks stay consistent

## Review Culture

- Agents are collaborators, not authorities
- Repo evidence beats memory
- Fresh docs beat assumptions
- Verification beats confidence
