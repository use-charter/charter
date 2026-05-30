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
- agent context drift guard: `agentcontext` package is the single canonical source; a new context file type cannot exist without also being scanned by AE-SEC-001 and AE-SEC-002
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

## Secret-Rule Testing Convention

AE-SEC-001 and AE-SEC-002 tests use a **generate-in-test** pattern: fake secret patterns (matching high-confidence prefixes: `sk-`, `ghp_`, `AKIA`, `xoxb-`, PEM headers) are constructed inside test code, never committed as literals in fixtures. Pass/fail fixtures like `testdata/repos/pass-secrets-agent/` and `testdata/repos/pass-secrets-config/` prove that environment-variable references (`${VAR}`, `$VAR`) and the placeholder `your-api-key-here` are correctly neutralized and **do not trigger findings**. These fixtures are git-safe and fully reviewable.

To add a test:
1. Use table-driven tests (see `internal/rules/secrets/sec001_test.go`)
2. Generate fake patterns matching the detector's high-confidence set in test cases, not as committed files
3. For pass-case fixtures, commit files with env-refs and placeholders only (no fake secrets)
4. For fail-case scenarios, generate the fake pattern at test runtime

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
