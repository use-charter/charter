# Phase 1 Slice 10 Design — GitHub Action & Performance Validation

## Goal

Deliver Charter's primary product surface (architecture M1.5 T1.5.1) and its performance guarantee (T1.5.3): (1) a composite **GitHub Action** — developed in `action/`, consumed as `use-charter/charter-action@v1` — that downloads the signed `charter` binary, verifies it (cosign + sha256), runs `charter doctor --format sarif`, and uploads to GitHub Code Scanning; and (2) a **perf validation** proving `charter doctor` meets the ≤ 2 s / ≤ 256 MB RSS budget on a 50k-file repo. Implements ADR-0018. Consumes Slice 9's release artifacts.

## Audience

- coding agents implementing this slice
- maintainers reviewing the Action contract (inputs, verification, upload) and the perf harness
- the founder seeding `use-charter/charter-action` and switching CI to dogfood it at launch

## Scope

### In scope

- `action/action.yml`: composite action (branding, inputs, download → verify → run → upload steps)
- `action/README.md`: copy-paste usage + required permissions
- action verification logic (portable bash): version resolve, OS/arch asset map, download, cosign `verify-blob`, sha256 check, extract, run, upload
- a Bun TS validator (`scripts/…`) asserting `action.yml` structure + full-SHA-pinned `uses:`; wired into `action/moon.yml` and `:check`
- perf validation: a synthetic 50k-file fixture generator + a build-tagged perf test (`charter doctor` ≤ 2 s; Linux peak RSS ≤ 256 MB) + a Moon `:perf` task
- docs sync (AGENTS/README/ARCHITECTURE), architecture T1.5.1/T1.5.3 reconciliation, Slice 17 roadmap checklist item

### Out of scope

- creating/seeding the `use-charter/charter-action` repo + first end-to-end run (launch / Slice 19, armed-not-fired)
- automated monorepo→action-repo sync; Marketplace listing
- switching Charter's own CI to `use-charter/charter-action@` (launch; CI keeps `go run ./cmd/charter doctor`)
- cutting any release tag; `charter init`/`fix` (Slices 11/12)

## Why this slice

The Action is the surface Phase-1 adoption is measured on (`uses: use-charter/charter-action`). Slice 9 built the signed binary it consumes; Slice 8 built the SARIF it uploads. Perf validation is the T1.5.3 guarantee that Charter won't be a CI bottleneck. Both are M1.5 and ride the same release/CI story, so they ship together.

## Grounding (verified live 2026-06-01, not from memory)

- **upload-sarif is v4** (`github/codeql-action/upload-sarif@v4`, v4.35.1, Node 24; v3 deprecated Dec 2026). Inputs: `sarif_file`, optional `category`. Consumer needs `security-events: write` (+ `actions: read`/`contents: read` for private repos).
- **cosign v3 verify-blob:** `cosign verify-blob --bundle <f>.sigstore.json --certificate-identity <ref> --certificate-oidc-issuer https://token.actions.githubusercontent.com <f>`. No `id-token` needed to verify. Our Slice 9 signs `checksums.txt` → one bundle covers all artifacts by hash.
- **Composite vs dedicated:** Marketplace requires a dedicated repo with `action.yml` at root; same-repo actions are `./path`, cross-repo are `owner/repo/path@ref`. Branded `use-charter/charter-action@v1` ⇒ dedicated repo (ADR-0018) seeded from `action/`.
- **Existing seams:** `cmd/charter/doctor.go` supports `--format sarif --out --threshold` with exit 0/1/2; `action/` is an empty placeholder (`moon.yml: id: action`); `AE-CI-002` credits `use-charter/charter-action@` and flags non-SHA-pinned actions; release identity is `github.com/use-charter/charter/.github/workflows/release.yml` (ADR-0016).

## The Action (`action/action.yml`)

- **Type:** composite. **Branding:** icon + color (Marketplace). **name/description** reflect "scan for AI-agent readiness + upload SARIF."
- **Inputs** (all optional): `version` (`latest`), `path` (`.`), `threshold` (`""` → charter.yaml policy/default), `format` (`sarif`), `sarif-file` (`charter.sarif`), `upload` (`true`), `verify` (`true`), `fail-below` (`true`), `category` (`""`).
- **Steps** (bash `run:` + pinned `uses:`):
  1. **Resolve version** — if `latest`, query the latest `use-charter/charter` release tag (via `gh`/API); else use the input. Compute the asset base name (`charter_<version>_<os>_<arch>`).
  2. **Download** archive + `checksums.txt` (+ `.sigstore.json`) from the release.
  3. **Verify** (`if verify`): `sigstore/cosign-installer@<sha>` then `cosign verify-blob --bundle checksums.txt.sigstore.json --certificate-identity https://github.com/use-charter/charter/.github/workflows/release.yml@refs/tags/<version> --certificate-oidc-issuer https://token.actions.githubusercontent.com checksums.txt`; then `sha256sum -c` (or `shasum -a 256 -c`) the archive against `checksums.txt`.
  4. **Extract** the `charter` binary; add to PATH.
  5. **Run** `charter doctor --path <path> --format <format> --out <sarif-file>` (+ `--threshold <n>` only when provided); capture the exit code (don't fail the step yet).
  6. **Upload** (`if upload && format == sarif`): `github/codeql-action/upload-sarif@<sha v4>` with `sarif_file: <sarif-file>` and `category` if set.
  7. **Gate**: if the captured doctor exit is 1 and `fail-below`, exit 1 (so alerts upload first, then the gate fails); exit 2 on scan error.
- **Outputs:** `exit-code`, `sarif-file` (for downstream steps). (`score` deferred — not shipped.)
- **Pinning:** `cosign-installer`, `upload-sarif`, and any `checkout` pinned to full SHAs.

## Action validator (`action/moon.yml` + Bun TS)

- A Bun TS script parses `action/action.yml` and asserts: `runs.using == "composite"`; required inputs present with documented defaults; every `uses:` is a 40-hex SHA pin (no floating tags) — the SHA-pin discipline `AE-CI-002` enforces on workflows, applied to the action that zizmor/actionlint don't scan.
- `action/moon.yml` gains a `validate` (or `lint`) task running it; add it to the root `:check` dep set (or as an `action:`-scoped check rolled into `:check`).

## Perf validation (T1.5.3)

- **Fixture generator** (test helper): synthesize a repo tree with ~50,000 files (nested dirs, a realistic mix incl. a minimal `AGENTS.md`/`go.mod` so the scan is representative) in `t.TempDir()` — generated, never committed.
- **Perf test** (build tag e.g. `//go:build perf`, or a `-run TestPerf` gate): run the doctor pipeline (repository resolve → inventory → rules → score) and assert **wall-clock ≤ 2 s**; on Linux, read peak RSS from `/proc/self/status` `VmHWM` and assert **≤ 256 MB** (skip the RSS assertion on non-Linux); always `-race`-clean. Report `runtime.MemStats` for visibility.
- **Moon task** `:perf` runs the tagged test; it is NOT in `:check` (kept out of the hot path) but runs in CI as its own step (and is available locally).

## Architecture / ownership

- `action/` — `action.yml` (composite), `README.md`, bash step logic; source of truth, seeded to `use-charter/charter-action` at launch.
- `scripts/` — Bun TS `action.yml` validator (repo dev helper).
- perf: a `_test.go` (build-tagged) + fixture helper under the scanner packages (`internal/doctor` or a dedicated `internal/perf` test); `:perf` Moon task.
- Avoid: a JS/Bun runtime dependency in the action steps (portable bash only); committing 50k fixture files; cutting a release; referencing `use-charter/charter-action@v1` from Charter's own workflows before it's seeded.

## Go alignment (per golang-patterns / golang-testing)

- Perf test is deterministic and hermetic (`t.TempDir()`, no network); RSS read wrapped/Linux-gated with `runtime.GOOS`; `-race` clean; errors wrapped with `%w`; no `_` discards.
- No change to the Charter binary's deps or core behavior; the scanner is exercised as-is.

## Testing & verification strategy

- **Action validator:** `action.yml` parses, is composite, inputs/defaults present, all `uses:` SHA-pinned → `:check` green.
- **Perf:** `moon run :perf` passes (≤ 2 s; Linux RSS ≤ 256 MB), `-race` clean.
- **Static action review:** the documented usage block in `action/README.md` is internally consistent (permissions, inputs); bash steps reviewed for injection-safety (no unquoted interpolation of untrusted input; inputs are action-controlled).
- **Gate:** `moon run :check` green (lint, vet, test, build, docs, security, eval, actionlint, zizmor, release-check, action validator).
- **Dogfood:** `charter doctor` still 100 clean; `charter version` unaffected.
- **Deferred (launch / Slice 19):** real end-to-end — the action downloads a signed `v0.9.0` binary, verifies it, and SARIF appears in the Security tab — documented in the release runbook.

## Risks

- **No release to download yet** — control: armed-not-fired; validate structure + SHA-pins + perf now; first e2e at the signed release. Charter CI keeps `go run ./cmd/charter doctor`.
- **Action `uses:` not covered by zizmor/actionlint** (they scan only workflows) — control: a repo-owned validator enforces SHA-pinning + structure on `action.yml` in `:check`.
- **Bash injection in composite steps** — control: quote all expansions; only action-controlled inputs flow into `run:`; no `${{ github.event.* }}` in shell.
- **RSS not portable in Go** — control: Linux-gated `VmHWM` read (CI is Linux) + a `MemStats` proxy elsewhere; wall-clock asserted on all platforms.
- **Version-resolution / asset-name drift vs GoReleaser naming** — control: the asset name template matches `.goreleaser.yaml` exactly (`charter_<version>_<os>_<arch>`); validated against the snapshot artifact names from Slice 9.

## Success criteria

- `action/action.yml` is a complete, branded composite action (download → verify → run → upload), all `uses:` SHA-pinned, validated in `:check`; `action/README.md` shows a correct usage block + permissions.
- `moon run :perf` proves `charter doctor` ≤ 2 s and (Linux) ≤ 256 MB RSS on a synthesized 50k-file repo, race-clean.
- ADR-0018 + this spec + the plan committed; AGENTS/README/ARCHITECTURE reflect the Action + perf; architecture T1.5.1/T1.5.3 reconciled; Slice 17 checklist gains the seed/dogfood step.
- `moon run :check` green; dogfood `charter doctor` still 100; no release tag cut.

## References

- `docs/internal/decisions/0018-github-action-and-sarif-upload.md` (ADR-0018); `0016-release-pipeline-and-supply-chain.md` (release/identity); `0014-sarif-output-and-rule-catalog.md` (SARIF)
- `docs/internal/architecture/charter-architecture-2026.md` (M1.5 T1.5.1 Action, T1.5.3 perf; §1.8 SARIF/CI)
- `docs/internal/superpowers/plans/2026-06-01-v1-launch-roadmap.md` (Slice 10 row; Slice 17 checklist)
- `github/codeql-action/upload-sarif@v4`; cosign v3 `verify-blob --bundle`; GitHub composite-action + Marketplace docs; GitHub OIDC
- `docs/internal/superpowers/plans/2026-06-01-phase-1-slice-10.md` (derived implementation plan)
