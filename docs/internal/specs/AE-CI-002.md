# AE-CI-002

- Severity: Low
- Category: CI
- Description: Repo should run Charter-related checks in CI.
- Detection logic: inspect `.github/workflows/` for Charter-related verification, workflow linting, supply-chain checks, and pinned third-party actions.
- Pass example: CI runs the repo quality gates, workflow security tools, and keeps mutable action tags out of the baseline.
- Fail example: no workflow coverage, no Charter entrypoint in CI, or unpinned third-party actions.
- Evidence expectations: workflow file path, job name, threshold or gate command, SARIF upload presence when applicable, and whether workflow hygiene tools are present.
- Edge cases: during pre-implementation bootstrap, CI may legitimately omit `charter doctor` itself if the scanner is not built yet, but the repo should still keep workflow lint/security gates active and document the deferred product gate clearly.
- Remediation: add or harden workflows so the repo's intended quality path is enforced in CI, not just locally.
- Related ADRs: ADR-0006
- Related evals: None yet
