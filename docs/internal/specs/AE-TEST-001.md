# AE-TEST-001

- Severity: High
- Category: Testing
- Description: A repository's active code language(s) must have automated tests. Agent-readiness depends on the agent being able to *verify its own work*; the AGENTS.md standard expects agents to run programmatic checks and fix failures before finishing a task. A language with source but no tests gives the agent nothing to verify against.
- Detection logic: scans the tracked inventory. A language is **active** only when **both** (a) its project manifest is present (`go.mod`, `package.json`, `Cargo.toml`, `Gemfile`, `pyproject.toml`/`setup.py`/`setup.cfg`/`requirements.txt`, `pom.xml`/`build.gradle*`, `*.csproj`, `composer.json`) **and** (b) it has ≥1 non-test source file *outside tooling directories* (`scripts/`, `tools/`, `testdata/`, `examples/`, `third_party/`, `vendor/`, `node_modules/`, `dist/`, `build/`, `.github/`, `gen/`, `generated/`). The manifest gate rejects a stray secondary-language file (e.g. a Homebrew `.rb` formula in a Rust repo); the source-outside-tooling gate rejects a tooling-only manifest (e.g. a `package.json` driving build scripts in a Go repo). For each active language, require ≥1 recognized test artifact: a file under a `test/`/`tests/`/`spec/`/`__tests__/` segment (any language), or a per-file name convention — Go `*_test.go`; JS/TS `*.{test,spec}.*`; Python `test_*.py`/`*_test.py`/`conftest.py`; Rust `tests/**.rs` or an inline `#[test]`/`#[cfg(test)]` signal; Java/Kotlin `*Test.{java,kt}`/`*Spec.kt`; Ruby `*_spec.rb`/`*_test.rb`; C# `*Tests.cs`/`*Test.cs`; PHP `*Test.php`.
- Pass example: a Go module (`go.mod` + `internal/**/*.go`) with `*_test.go` files — tests present, passes.
- Fail example: a Go module with `internal/app/app.go` but no `*_test.go` anywhere — flagged High; evidence: `no test files detected for active language: Go`.
- Evidence expectations: one finding listing each active language lacking tests. No location (repo-level).
- Edge cases: **N/A** (no finding) when no recognized code language is active (docs/config/tooling-only repos). Tooling-only manifests (a `package.json` with only `scripts/*.ts`) and stray secondary-language files (a lone `*.rb` Homebrew formula with no `Gemfile`) are not active. Rust inline unit tests count via the `#[cfg(test)]`/`#[test]` content signal. v1 judges *presence*, not coverage, quantity, or pyramid shape. Validated at 0% FP across 10 real public repos (`docs/internal/operability-fp-validation.md`).
- Remediation: add tests for the active language(s) so an agent can run them and self-verify before finishing a task.
- Scoring impact: `High` (−10); no hard cap.
- Related ADRs: ADR-0023, ADR-0008
- Related evals: None yet
