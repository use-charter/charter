# AE-TEST-001

- Severity: High
- Category: Testing
- Description: A repository's active code language(s) must have automated tests. Agent-readiness depends on the agent being able to *verify its own work*; the AGENTS.md standard expects agents to run programmatic checks and fix failures before finishing a task. A language with source but no tests gives the agent nothing to verify against.
- Detection logic: scans the tracked inventory. A language is **active** when it has ≥1 non-test source file *outside tooling directories* (`scripts/`, `tools/`, `testdata/`, `examples/`, `third_party/`, `vendor/`, `node_modules/`, `dist/`, `build/`, `.github/`, `gen/`, `generated/`) — a manifest alone (e.g. a `package.json` driving build scripts in a Go repo) does **not** make a language a tested surface. For each active language, require ≥1 recognized test artifact: Go `*_test.go`; JS/TS `*.{test,spec}.{js,jsx,ts,tsx,mjs,cjs}` or `__tests__/`; Python `test_*.py`/`*_test.py`/`tests/`/`conftest.py`; Rust `tests/**.rs` or an inline `#[test]`/`#[cfg(test)]` signal; Java/Kotlin `src/test/**`/`*Test.{java,kt}`/`*Spec.kt`; Ruby `*_spec.rb`/`*_test.rb`/`spec/`/`test/`; C# `*Tests.cs`/`*Test.cs`; PHP `tests/`/`*Test.php`.
- Pass example: a Go module (`go.mod` + `internal/**/*.go`) with `*_test.go` files — tests present, passes.
- Fail example: a Go module with `internal/app/app.go` but no `*_test.go` anywhere — flagged High; evidence: `no test files detected for active language: Go`.
- Evidence expectations: one finding listing each active language lacking tests. No location (repo-level).
- Edge cases: **N/A** (no finding) when no recognized code language is active (docs/config/tooling-only repos). Tooling-only language footprints (only `scripts/*.ts`) are not active. Rust inline unit tests count via the `#[cfg(test)]`/`#[test]` content signal. v1 judges *presence*, not coverage, quantity, or pyramid shape.
- Remediation: add tests for the active language(s) so an agent can run them and self-verify before finishing a task.
- Scoring impact: `High` (−10); no hard cap.
- Related ADRs: ADR-0023, ADR-0008
- Related evals: None yet
