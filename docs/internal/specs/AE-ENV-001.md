# AE-ENV-001

- Severity: Medium
- Category: Environment
- Description: Toolchain, lockfiles, and hooks should make the repo reproducible.
- Detection logic: inspect runtime pin declarations, committed lockfiles, and committed hook manager config; fail when active languages or repo tooling cannot be reproduced from tracked files.
- Pass example: `mise.toml`, `mise.lock`, `go.mod`, and `hk.pkl` align with repo automation and verification commands.
- Fail example: floating runtimes, missing lockfiles, or task/hook behavior that depends on untracked local files.
- Evidence expectations: which toolchain file covers which language, which lockfiles exist, and which hook manager is committed.
- Edge cases: partial language coverage is acceptable only if the uncovered language is not active in the repo; local-only helper scripts do not satisfy reproducibility for tracked task config.
- Remediation: pin the runtime in tracked config, commit the lockfile, and ensure local/hook/CI behavior all derive from the same tracked baseline.
- Why: An agent that cannot reproduce the project's toolchain installs the wrong version, runs tests against the wrong runtime, and produces fixes that pass locally but break in CI or on a colleague's machine.
- Auto-fixable: No
- Related rules: AE-CI-002, AE-AUTO-001
- Related ADRs: ADR-0006
- Related evals: None yet
