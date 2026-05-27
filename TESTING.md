# Testing

## Golden Commands

- `moon run :check`
- `moon run :test`
- `moon run :vet`
- `moon run :lint`
- `moon run :build`
- `moon run :docs`
- `moon run :eval`
- `moon run :security`

## Test Matrix

Every meaningful change should consider:

- unit behavior: table-driven Go tests close to the slice being changed
- fixture coverage: deterministic fixtures under `testdata/`
- contract validation: specs, schemas, and machine-readable behavior
- docs/spec drift: architecture, ADR, RFC, and rule contract updates
- workflow verification: Moon task mapping and workflow command parity
- security posture: secret-safety, supply-chain, and least-privilege checks
- eval impact: non-trivial prompt, workflow, or agent-facing behavior
- future scale checks: large-repo fixtures, benchmarks, and performance gates when real scanner code lands

Treat the test matrix as part of the product contract, not a final cleanup step.

## Expectations

- Prefer table-driven Go tests
- Keep tests deterministic and secret-safe
- Add fixtures under `testdata/`
- Add eval artifacts for non-trivial prompt or workflow behavior
- Run `go test -race ./...` through `moon run :test`
- Run `go vet ./...` through `moon run :vet` when touching Go packages or command wiring

## First Slice Proof Model

The first Phase 1 slice must use the proof model in `docs/internal/superpowers/checklists/2026-05-28-first-slice-proof-model.md`.

Minimum proof for the first slice:

- a failing test or failing fixture-driven assertion first
- the smallest implementation required to pass it
- one exact verification command recorded in the same slice
- rule spec or companion doc updated if behavior changed
- no hidden expansion from one rule into many unrelated concerns

## CLI Quality

Before output-heavy work begins, follow `docs/internal/superpowers/checklists/2026-05-28-cli-quality-principles.md`.

Treat quiet mode, redaction, rule labeling, and machine-readable output as product surfaces, not incidental logging details.
