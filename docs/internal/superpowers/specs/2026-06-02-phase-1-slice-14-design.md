# Phase 1 Slice 14 Design — Agent Operability Expansion

## Goal

Take Charter from an agent-*setup* linter to an agent-*operability* scorecard by closing its biggest blind spot — **can the agent verify its work and run the project?** Adds three rules (`AE-TEST-001`, `AE-AUTO-001`, `AE-CTX-006`) in two new categories plus a Context quality nudge, and a per-category readiness breakdown in the report. Offline, deterministic, no-LLM, no-network (Commitments #4/#7). Implements ADR-0023; score formula (ADR-0008) unchanged.

## Audience

- coding agents implementing the slice
- maintainers reviewing detection accuracy + FP discipline (Commitment #9)

## Scope

### In scope

- `internal/rules/testing` → `AE-TEST-001` (tests present, stack-aware) — **High**
- `internal/rules/autonomy` → `AE-AUTO-001` (verification command discoverable/runnable) — **Medium**
- `internal/rules/context` → `AE-CTX-006` (instruction emphatic-density nudge) — **informational**
- category readiness breakdown in `internal/render/{text,markdown,json}` (+ `categories` in the JSON DTO); score formula untouched
- rule catalog + specs (`AE-TEST-001`/`AE-AUTO-001`/`AE-CTX-006`) + audit checklist + architecture §0.2 table (15→18) + roadmap renumber (14→15…19→20)
- testdata fixtures + unit/integration tests; an **FP-validation pass** against ≥10 real public repos (recorded), tuning thresholds before sign-off

### Out of scope

- coverage thresholds, pyramid ratio/tier-separation (maturity signals, deferred)
- review-readiness (CODEOWNERS/PR template/DangerJS — team/enterprise), dependency-freshness/`npm audit` (network), devcontainer, release automation
- per-dimension re-scoring (would break the stable ADR-0008 score contract — the scorecard is reporting only)

## Grounding (verified)

- **`doctor.Run`** aggregates rule findings into `Result{Findings, Score{Base,Final}, ...}`; `findings.Finding` has `Severity`, `Category`, `Informational`. Scoring (`internal/scoring`) = `100 − B×20 − H×10 − M×4 − L×1`, informational excluded (ADR-0008). New rules slot in like the existing rule packages.
- **Language detection**: reuse the manifest-reading pattern from `internal/scaffold.Detect` (go.mod→Go, package.json→JS/TS, pyproject.toml/setup.py/requirements.txt→Python, Cargo.toml→Rust, +pom.xml/build.gradle→Java/Kotlin, Gemfile→Ruby, *.csproj→C#, composer.json→PHP), evaluated over `repository.Inventory` (tracked files only).
- **External grounding** (native, no borrowed sources): the AGENTS.md open standard (commands = highest-value section; agents execute checks to self-verify); instruction-following research (arXiv 2603.25015; complexity-cliff) for AE-CTX-006.

## Rules

### AE-TEST-001 — Testing — High — tests present
For each detected language, require ≥1 recognized test artifact in the tracked inventory:

| Language | Manifest | Test signal |
|---|---|---|
| Go | `go.mod` | `**/*_test.go` |
| JS/TS | `package.json` | `**/*.{test,spec}.{js,jsx,ts,tsx,mjs,cjs}`, `**/__tests__/**` |
| Python | `pyproject.toml`/`setup.py`/`requirements.txt` | `**/test_*.py`, `**/*_test.py`, `**/tests/**`, `**/conftest.py` |
| Rust | `Cargo.toml` | `tests/**/*.rs`, or `#[test]`/`#[cfg(test)]` in a tracked `*.rs` |
| Java/Kotlin | `pom.xml`/`build.gradle*` | `**/src/test/**`, `**/*Test.{java,kt}`, `**/*Spec.kt` |
| Ruby | `Gemfile` | `**/*_spec.rb`, `**/*_test.rb`, `spec/**`, `test/**` |
| C# | `*.csproj` | `**/*Tests.cs`, `**/*Test.cs` |
| PHP | `composer.json` | `tests/**`, `**/*Test.php` |

- One finding per detected-language-without-tests (evidence names the language). **N/A** (no finding) when no recognized code language. Severity **High**. v1 = presence only.

### AE-AUTO-001 — Autonomy — Medium — verification command discoverable
Pass when the test/verification command is discoverable via **either**:
- a **task runner** declaring a test-ish entrypoint: `Makefile` (`test:`/`check:`), `justfile` (`test`/`check`), `Taskfile.yml` (`test`), `package.json` `scripts.test`, `mise.toml` `[tasks.test]`/`[tasks.check]`, `moon.yml` task, OR
- a **conventional zero-config toolchain**: Go (`go.mod` → `go test`), Rust (`Cargo.toml` → `cargo test`).

Fires only when tests are expected (a recognized language present) **and** neither path makes the test command discoverable (e.g. a JS/Python repo with tests but no `test` script/runner). **N/A** when no recognized language. Severity **Medium**. Complements AE-CTX-001 ("a verification command is *mentioned*") by checking it's *real and runnable*.

### AE-CTX-006 — Context — informational — instructions not over-emphasized
Count emphatic-directive tokens (`IMPORTANT|NEVER|MUST|CRITICAL|ALWAYS|EXTREMELY|ABSOLUTELY|FORBIDDEN|PROHIBITED`, word-boundary) per 1,000 words in the agent context file; flag when density ≥ 15/1K words. `Informational: true` (re-surfaces, never deducts — mirrors AE-SUPPRESS-003). Depends on a context file existing. Threshold tuned in the FP pass.

## Category scorecard (reporting only)
Group active findings by `Category`; for each, show finding count, total deduction (by severity), and worst severity. Text/markdown gain a "Readiness by category" block; JSON gains a `categories: [{category, findings, deduction, worst_severity}]` array (additive — SARIF + score formula untouched). Informational findings listed but contribute 0 deduction.

## Architecture / ownership
- New: `internal/rules/testing/`, `internal/rules/autonomy/` (pure detectors over the inventory). `internal/rules/context/` gains `AE-CTX-006`. `internal/render/*` gains the breakdown. `internal/render/json` DTO gains `categories`.
- Avoid: network/LLM; per-dimension re-scoring; FP on idiomatic Go/Rust (toolchain counts); penalizing non-code repos (N/A).

## Testing & verification strategy
- **Unit:** per-language test detection (present/absent/N/A); AUTO discoverability matrix (runner test target; conventional toolchain; JS-with-tests-no-script fires; non-code N/A); CTX-006 density (over/under threshold; informational, non-deducting via a scoring assertion); scorecard grouping.
- **Fixtures:** `pass-test-*`, `fail-test-missing`, `fail-auto-no-runner`, etc.
- **FP-validation:** scan ≥10 real public repos (mixed Go/JS/Python/Rust/polyglot); classify every AE-TEST-001/AE-AUTO-001 finding TP/FP; target ≤10% FP; tune; record in `docs/internal/catalog/`-style notes.
- **Dogfood:** Charter stays **100** (Go tests ✓, moon task contract ✓, declarative AGENTS.md ✓). `moon run :check` green.

## Success criteria
- 3 rules + scorecard shipped; rule set 15→18; Charter dogfoods 100; FP ≤10% on real repos (recorded).
- ADR-0023 + this spec + plan committed; architecture §0.2 table + roadmap renumber + specs + audit updated; HTML mirror regenerated (CF-2 gate); `moon run :check` green.

## References
- `docs/internal/decisions/0023-agent-operability-rules.md` (ADR-0023); `0008` (score, unchanged); `0013` (informational precedent)
- `internal/scaffold/detect.go` (lang detection pattern); `internal/scoring/score.go`; `docs/internal/superpowers/plans/2026-06-02-phase-1-slice-14.md`
