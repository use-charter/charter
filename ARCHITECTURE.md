# Architecture

## Bootstrap Contract

- Single root Go module: `go.charter.dev/charter`
- No `go.work`
- No extra modules during bootstrap
- Charter core performs no LLM calls
- Scanner remains deterministic and offline-first
- Fixes are diff-first and never silently mutate files
- Tracked MCP server config stays absent until a pinned, reviewed, least-privilege integration exists (distinct from the AE-MCP-* rules, which scan MCP config when present)

## Root Layout

- Root contract docs stay at repo root for contributor and agent entry.
- Repo-internal engineering docs live under `docs/internal/`.
- Future customer-facing docs live under `docs/product/`.

- `cmd/`: binary entrypoints and command wiring
- `internal/`: non-public implementation details
  - `agentcontext`: canonical agent-visible context file registry (drift guard for context and secret rules)
  - `config`: `charter.yaml` loader (MCP trusted-remote allowlist)
  - `doctor`: scan orchestration pipeline
  - `findings`: finding model with Location support (path:line)
  - `repository`: repo resolution and file inventory
  - `rules/`: rule implementations (context, environment, ci, secrets, mcp, agentconfig)
  - `scoring`: score calculation and caps
  - `render/`: output formatters (text, JSON, Markdown)
  - `secrets`: secret pattern detection and redaction
- `api/openapi/`: future API contracts before implementation
- `schemas/`: machine-readable config and report contracts (includes `doctor-result.schema.json`)
- `docs/internal/specs/`: rule-level behavior contracts
- `docs/internal/decisions/`: accepted ADRs
- `docs/internal/rfcs/`: proposals for cross-cutting or risky changes
- `evals/`: acceptance fixtures for prompt, workflow, and future agent behavior

## Command Tree Intent

The expected CLI tree follows the product authority in `docs/internal/architecture/charter-architecture-2026.md`:

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
- location data: structured finding locations (path:line) per ADR-0009, with 1-based line numbers (0 = file-level absence findings)
- score cap: hard ceiling applied when finding is present
- no secret-bearing payloads

Example Finding shape:

```text
RuleID: AE-CTX-001
Severity: BLOCKER
Summary: Missing agent context file
Remediation: Create AGENTS.md with project state, tech stack, edit boundaries, and verification command
Evidence: ["no agent-visible context file found in repo root"]
Locations: [] (empty for absence findings)
Cap: 0
```

Another example with locations:

```text
RuleID: AE-SEC-001
Severity: BLOCKER
Summary: Secret detected in agent context file
Remediation: Remove the literal secret, rotate it externally, use an environment variable reference instead
Evidence: ["sk-p… (redacted token)"]
Locations: [{Path: "AGENTS.md", Line: 14}]
Cap: 49
```

Every user-facing rule or command error should be backed by pass/fail examples in `docs/internal/specs/`.
