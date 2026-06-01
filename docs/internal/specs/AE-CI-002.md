# AE-CI-002

- Severity: Low
- Category: CI
- Description: Repo should run Charter-related checks in CI.
- Detection logic: inspect `.github/workflows/` for Charter-related verification, workflow linting, supply-chain checks, and pinned third-party actions. Trusted SLSA reusable workflows (`slsa-framework/slsa-github-generator/.github/workflows/*.yml@vX.Y.Z`) are exempt from the SHA-pin requirement because `slsa-verifier` resolves the trusted builder identity from the semantic version tag and SHA-pinning is unsupported.
- Recognized forms (in addition to Charter's own `moon run :*` tasks, so non-`moon` repos are not false-flagged):
  - Repo quality: direct test/build commands — `go test`/`go build`, `npm test`/`npm run test`, `pnpm test`/`pnpm run test`, `yarn test`, `cargo test`/`cargo build`, `pytest`/`python -m pytest`, `bun test`, `make test`/`make check`.
  - Workflow lint: `actionlint` AND `zizmor`, each satisfied as a direct command or via the `rhysd/actionlint` / `zizmorcore/zizmor` actions.
  - Security: direct scanners `govulncheck`/`osv-scanner`/`gitleaks`/`trivy`/`grype`, or the `github/codeql-action`.
- First-party exemption: `use-charter/charter-action@<tag>` is exempt from the SHA-pin requirement — tag-pinning is the conventional consumer form for Charter's own action.
- Pass example: CI runs the repo quality gates, workflow security tools, and keeps mutable action tags out of the baseline.
- Fail example: no workflow coverage, no Charter entrypoint in CI, or unpinned third-party actions.
- Evidence expectations: workflow file path, job name, threshold or gate command, SARIF upload presence when applicable, and whether workflow hygiene tools are present.
- Edge cases: during pre-implementation bootstrap, CI may legitimately omit `charter doctor` itself if the scanner is not built yet. Once `charter doctor` exists, the repo should run a Charter-related CI gate, not just local quality gates.
- Remediation: add or harden workflows so the repo's intended quality path is enforced in CI, not just locally.
- Related ADRs: ADR-0006
- Related evals: None yet
