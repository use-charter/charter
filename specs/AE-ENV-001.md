# AE-ENV-001

- Severity: Medium
- Category: Environment
- Description: Toolchain, lockfiles, and hooks should make the repo reproducible.
- Detection logic: inspect runtime pins, lockfiles, and committed hook config.
- Pass example: `.mise.toml` plus hook config present.
- Fail example: floating or undeclared runtime environment.
- Remediation: pin runtimes and commit the relevant config.
- Related ADRs: ADR-0006
- Related evals: None yet
