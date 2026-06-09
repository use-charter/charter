# AE-AUTO-001

- Severity: Medium
- Category: Autonomy
- Description: The command an agent runs to verify the project (its tests) must be **discoverable and runnable**. The AGENTS.md standard calls the exact CLI commands (install/test/lint/build) "the single highest-value section" for agents. AE-CTX-001 checks that a verification command is *mentioned* in the context file; AE-AUTO-001 checks the repo actually *exposes* a runnable one.
- Detection logic: applies only when a code language is active (see AE-TEST-001). Passes when the test command is discoverable via **either** (a) a task runner declaring a test/check entrypoint — `Makefile` (`test:`/`check:`), `justfile`/`.justfile` (`test`/`check` recipe), `Taskfile.yml`/`Taskfile.yaml` or `moon.yml` (`test:`/`check:` task), `mise.toml`/`.mise.toml` (`[tasks.test]`/`[tasks.check]`), or `package.json` `scripts.test`/`test:*` — **or** (b) the active language's conventional zero-config toolchain: Go (`go.mod` → `go test`), Rust (`Cargo.toml` → `cargo test`), Python when pytest is configured (`pytest.ini`/`tox.ini`/`[tool.pytest…]`/`[tool:pytest]` → `pytest`).
- Pass example: a Go module (`go.mod`) — `go test` is conventional and discoverable, passes with no runner required. A JS app with `package.json` `scripts.test` — passes.
- Fail example: a JS app with `src/index.ts` + tests but no `test` script and no task runner — flagged Medium; an agent has no discoverable way to run the checks.
- Evidence expectations: one finding stating no runner test target and no conventional toolchain applies.
- Edge cases: **N/A** when no language is active. **FP guard:** a single-language Go/Rust/Cargo repo is never penalized for lacking a `Makefile` — its toolchain is the contract. Conventional Python detection requires a pytest config (not merely `.py` files), since `python` alone is not a zero-config test command.
- Remediation: expose a test command via a task runner (Makefile/justfile/Taskfile/package.json scripts/mise/moon) so an agent can discover and run it.
- Scoring impact: `Medium` (−4); no hard cap.
- Why: Even with tests present, an agent needs to know how to run them. A discoverable command closes the agent's work loop — run, observe, fix, repeat — without guessing.
- Auto-fixable: No
- Related rules: AE-TEST-001, AE-ENV-001
- Related ADRs: ADR-0023, ADR-0008
- Related evals: None yet
