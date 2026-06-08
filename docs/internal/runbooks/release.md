# Release Runbook

## Trigger

- preparing a release candidate or public tag
- validating the release pipeline after supply-chain or distribution changes

## Triage

1. Confirm versioning and changelog policy.
2. Verify the repo gate: `moon run :check`.
3. Verify the perf budget: `moon run :perf`.
4. Verify the release config: `moon run :release-check`.
5. Verify the offline release path: `moon run :release-snapshot`.
6. Verify docs/spec alignment, including HTML mirrors.
7. Confirm no secret-bearing artifacts were generated or staged.

## Artifact Checks

- release workflow present and actionlint/zizmor-clean
- `.goreleaser.yaml` current and valid
- signed-release pipeline still targets checksums, SBOM, and SLSA provenance
- `action/` contract matches the current binary/download flow
- install/distribution docs do not overclaim pre-launch publication state

## Escalation

- if `:check` or `:perf` fails: stop release work and fix the regression first
- if release signing/provenance behavior drifts from ADR-0016 or ADR-0018: update code and docs together before proceeding
- if launch-only dependencies are missing (public tag, tap publication, seeded action repo, docs deployment): hand off to the go-public checklist

## Follow-up

- record any release-surface drift in the relevant ADR/spec/runbook
- update the launch checklist inputs when a release concern is discovered here
