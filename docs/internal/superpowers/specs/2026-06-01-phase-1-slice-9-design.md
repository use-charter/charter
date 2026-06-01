# Phase 1 Slice 9 Design — Release Pipeline & `charter version`

## Goal

Ship Charter's supply-chain release rails (architecture M1.5 T1.5.2) and the last unwired v1 command. Concretely: (1) a GoReleaser v2 pipeline producing multi-OS/arch binaries, `checksums.txt`, syft SPDX SBOMs, cosign keyless signatures, and SLSA L3 provenance, distributed as a signed GitHub Release + Homebrew cask; and (2) a `charter version` command printing version/commit/build-date/go/platform from build-stamped metadata. Implements ADR-0016 (release pipeline) and ADR-0017 (version command & build stamping). This unblocks Slice 10 (the Action downloads the released binary).

## Audience

- coding agents implementing this slice
- maintainers reviewing the supply-chain posture (signing identity, provenance, SBOM, distribution) and the version contract
- founder/admin executing the out-of-band prerequisites (org/tap/secrets) and the first real tag in go-public ops

## Scope

### In scope

- `internal/version`: `commit`/`date` ldflags vars + `Commit()`/`Date()`; `"unknown"` dev fallback via `runtime/debug`
- `cmd/charter/version.go` + registration in `root.go`: the 5-line `version` subcommand; CLI test
- root `.goreleaser.yaml` (`version: 2`): builds, archives, checksum, sboms (syft), signs (cosign bundle), homebrew_casks, release (prerelease auto), changelog, snapshot
- `.github/workflows/release.yml` rewrite: `goreleaser` job (build/sign/release + emit base64 hashes) + `provenance` job (SLSA generic generator @v2.1.0)
- `mise.toml` + `mise.lock`: pin `goreleaser`, `cosign`, `syft`
- root `moon.yml`: `release`, `release-snapshot`, `release-check` tasks; `release-check` added to `:check`
- zizmor: a single documented `unpinned-uses` ignore for the SLSA reusable workflow tag-pin
- docs: ADR-0016/0017 (done), this spec, the plan; `AGENTS.md` (+`charter version`), `README.md`, `ARCHITECTURE.md`, and architecture-doc reconciliation (SPDX version, SLSA L3)

### Out of scope

- cutting a real tag / `v0.9.0`; live cosign/Rekor/SLSA execution; the tap push (all go-public ops — "armed, not fired")
- Apple notarization; `go install` vanity-path redirect (needs `go.use-charter.dev` host)
- GHCR/container images; Scoop/Nix/apt/winget; announcements (Twitter/Slack)
- the GitHub Action itself + perf validation (Slice 10); `charter version --format json`/`--short`

## Why this slice

The Action (Slice 10, the primary product surface and Phase 1 Signal 1) is meaningless without a real, verifiable binary to run, and `@v1` must resolve to a signed release. Charter also scores others on exactly this posture (AE-CI-002, AE-ENV-001, Commitment #6), so the release pipeline is dogfooding. `charter version` rides along because it shares the build-stamp mechanism the release injects.

## Grounding (verified live 2026-06-01, not from memory)

- **GoReleaser v2.16.0:** `brews` deprecated since v2.10 → use `homebrew_casks` (Casks dir is default; add `tap_migrations.json` only when migrating an existing formula — N/A for greenfield). v2 renamed `snapshot.name_template` → `version_template`. `sboms` shells out to syft; `signs` shells out to cosign. `release.prerelease: auto` marks `vX.Y.Z-rc/alpha/beta` as a prerelease.
- **cosign v3.0.6:** v3 standardized on `cosign sign-blob --bundle <file>.sigstore.json --yes <file>` and `verify-blob --bundle …`. `id-token: write` → cosign auto-detects the GitHub OIDC token (no `COSIGN_EXPERIMENTAL`, no keys). cosign-installer (v4.1.1) is unnecessary because mise provides the pinned cosign binary.
- **slsa-github-generator v2.1.0:** `…/generator_generic_slsa3.yml@v2.1.0` consumes `base64-subjects` (base64 of the sha256 checksums file) and uploads provenance to the release; needs `actions: read`, `id-token: write`, `contents: write`. Must be **tag-pinned** (slsa-verifier requires it) — the one SHA-pin exception. Yields SLSA Build **L3** (isolated reusable workflow), satisfying the doc's L2.
- **Repo seams:** `mise.toml` pins all tools via aqua + `[env] HK_MISE=1`; root `moon.yml` aggregates gates in `:check` and shells tasks to Bun scripts or direct commands; `release.yml` is a `moon run :build` stub (`contents: read`); `ci.yml` SHA-pins every action and runs `charter doctor` as a gate; `internal/version` has `Version()` (+ `injected` var) and a doc stub; `root.go` registers only `doctor`/`suppress`. Module path is the vanity `go.use-charter.dev/charter`; remote is `github.com/use-charter/charter`.

## `charter version` (`cmd/charter/version.go` + `internal/version`)

- `internal/version`: add `var commit, date string` (ldflags targets) + `Commit() string` / `Date() string`. Each: injected value → `runtime/debug.ReadBuildInfo()` setting (`vcs.revision` / `vcs.time`) → `"unknown"`. `Version()` unchanged.
- `cmd/charter/version.go`: `newVersionCommand()` → `cobra.Command{Use: "version", Short: …}` whose `RunE` writes five aligned lines to `cmd.OutOrStdout()`:
  ```
  charter   <Version()>
  commit    <Commit()>
  built     <Date()>
  go        <trimmed runtime.Version()>
  platform  <GOOS/GOARCH>
  ```
  No flags, no scan, always exit 0. Register via `cmd.AddCommand(newVersionCommand())` in `root.go`.
- `runtime`-derived: `go` = `strings.TrimPrefix(runtime.Version(), "go")`; `platform` = `runtime.GOOS + "/" + runtime.GOARCH`.

## `.goreleaser.yaml` (root)

- `version: 2`, `project_name: charter`; `before.hooks: [go mod tidy, go mod verify]`.
- **builds:** one build, `main: ./cmd/charter`, `binary: charter`, `env: [CGO_ENABLED=0]`, `flags: [-trimpath]`, `mod_timestamp: "{{ .CommitTimestamp }}"`, `goos: [linux, darwin, windows]`, `goarch: [amd64, arm64]`, `ldflags: -s -w -X go.use-charter.dev/charter/internal/version.injected={{.Version}} -X go.use-charter.dev/charter/internal/version.commit={{.FullCommit}} -X go.use-charter.dev/charter/internal/version.date={{.CommitDate}}`.
- **archives:** default (tar.gz; zip for windows via `formats`/`format_overrides`); include `LICENSE`, `README.md`.
- **checksum:** `name_template: checksums.txt` (sha256 — SLSA input).
- **sboms:** `- artifacts: archive` (syft, SPDX-JSON). Verify the emitted SPDX version at implementation; reconcile the architecture doc to it (don't over-claim 3.0.1).
- **signs:** cosign keyless over the checksum file:
  ```yaml
  signs:
    - cmd: cosign
      signature: "${artifact}.sigstore.json"
      args: ["sign-blob", "--bundle=${signature}", "${artifact}", "--yes"]
      artifacts: checksum
      output: true
  ```
- **homebrew_casks:** repository `{owner: "{{ .Env.TAP_OWNER }}", name: homebrew-tap, token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"}`, `homepage: https://use-charter.dev`, `description`, `license: Apache-2.0`; a post-install hook stripping `com.apple.quarantine` from the staged `charter` binary. (Verify the exact v2.16 cask hook field names at implementation.)
- **release:** `prerelease: auto`; **changelog:** grouped by conventional-commit type.
- **snapshot:** `version_template: "{{ incpatch .Version }}-snapshot"`.

## `.github/workflows/release.yml` (rewrite)

- `on: push: tags: ["v*"]`; top-level `permissions: contents: read`; concurrency as today.
- **job `goreleaser`** — `permissions: {contents: write, id-token: write}`:
  - checkout (`fetch-depth: 0`, `persist-credentials: false`); mise-action (pinned, provides go/goreleaser/cosign/syft); `bun install --frozen-lockfile` only if needed (not for release).
  - `mise x -- moon run :release` (or `goreleaser release --clean`). Env: `GITHUB_TOKEN`, `TAP_OWNER`, `HOMEBREW_TAP_TOKEN`.
  - a step computing `hashes` = base64 of `dist/checksums.txt` → `outputs.hashes` (the documented GoReleaser→SLSA bridge).
- **job `provenance`** — `needs: goreleaser`, `permissions: {actions: read, id-token: write, contents: write}`:
  - `uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.1.0` with `base64-subjects: ${{ needs.goreleaser.outputs.hashes }}`, `upload-assets: true`. Inline `# zizmor: ignore[unpinned-uses] tag-pin required by slsa-verifier (ADR-0016)`.
- All other actions full-SHA-pinned (reuse the SHAs already in `ci.yml`).

## mise + Moon wiring

- `mise.toml`: add `"aqua:goreleaser/goreleaser" = "2.16.0"`, `"aqua:sigstore/cosign" = "3.0.6"`, `"aqua:anchore/syft" = "<verify latest>"`; run `mise install` to refresh `mise.lock`.
- root `moon.yml` tasks:
  - `release`: `goreleaser release --clean` (`options.cache: false`), not in `:check`.
  - `release-snapshot`: `goreleaser release --snapshot --clean --skip=sign,publish` (offline local/CI dry-run), not in `:check`.
  - `release-check`: `goreleaser check` (validates `.goreleaser.yaml`); **add to `check.deps`** so config breakage fails the standard gate.

## Architecture / ownership

- `internal/version/` — add `commit`/`date` + `Commit()`/`Date()`; pure, no new deps.
- `cmd/charter/version.go` — the `version` subcommand; `root.go` registration.
- `.goreleaser.yaml`, `.github/workflows/release.yml`, `mise.toml`/`mise.lock`, `moon.yml` — release rails.
- Avoid: a `main`-package version var (keep it in `internal/version`); `goreleaser-action`/`cosign-installer` (mise owns tools); hardcoded `use-charter`/`tashfiqul-islam` owner; the deprecated `brews` block; cutting a real tag.

## Go alignment (per golang-patterns / golang-testing)

- `internal/version` stays pure (no I/O, no globals beyond the ldflags vars, which are write-once at link time and read-only at runtime); accessors return concrete strings with a deterministic fallback.
- `version` command is pure-stdout, always exit 0; errors (none expected) would wrap with `%w`.
- Tests: table-free direct assertions for `Commit()/Date()` `"unknown"` dev fallback (mirrors `TestVersionDefaultsInTestBuild`); a CLI test asserting the five labels and that the version line equals `version.Version()`. `-race`, `gofumpt`/`go vet` clean.

## Testing & verification strategy

- **Unit:** `internal/version` — `Commit()`/`Date()` non-empty and default to `"unknown"` in a test build; `cmd/charter` — `version` output contains all five labels and the `version.Version()` value; exit 0.
- **Config:** `goreleaser check` passes (`moon run :release-check`).
- **Offline build:** `moon run :release-snapshot` builds all 6 targets + `checksums.txt` + SBOMs into `dist/` with no network, no signing.
- **Workflow:** `actionlint` + `zizmor` clean on the rewritten `release.yml` (SLSA line ignored with justification).
- **Gate:** `moon run :check` green (now incl. `release-check`).
- **Dogfood:** `go run ./cmd/charter version` prints the table (dev: `0.0.0-dev` / VCS commit/date or `unknown`); `charter doctor --path . --threshold 80` still 100; `AE-CI-002` not regressed.
- **Deferred verification (go-public ops):** the first real tag `v0.9.0` exercises live cosign keyless (`cosign verify-blob --bundle …`), SLSA provenance (`slsa-verifier verify-artifact …`), and the Homebrew cask install — documented in the release runbook, not run here.

## Risks

- **Can't fully verify signing/provenance/tap offline** — control: split deliverables; verify everything non-Rekor (`goreleaser check`, `--snapshot`, actionlint/zizmor, `:check`) now; gate the live path behind the first tag in go-public ops with explicit verify commands in the runbook. Honest "armed, not fired".
- **SLSA tag-pin vs zizmor `--pedantic`** — control: a single inline `unpinned-uses` ignore citing ADR-0016; every other action SHA-pinned; the exception is the SLSA project's documented requirement.
- **SPDX version drift from the doc ("3.0.1")** — control: emit syft's actual SPDX-JSON version and reconcile the architecture doc to it; never claim a version we don't emit.
- **Owner/identity churn (vanity path vs remote vs future org)** — control: parametric owner (`github.repository*`, `TAP_OWNER`); a later org move changes only the cosign cert-identity string used at verify time, documented in the runbook.
- **macOS Gatekeeper on unsigned-by-Apple binaries** — control: cask post-install strips `com.apple.quarantine`; notarization deferred (needs Apple cert).
- **Homebrew tap / token prerequisites** — control: snapshot skips upload; the real publish documents the `homebrew-tap` repo + `HOMEBREW_TAP_TOKEN` as out-of-band founder steps.

## Success criteria

- `charter version` prints version/commit/built/go/platform; dev build shows `0.0.0-dev` + VCS-or-`unknown` commit/date; unit + CLI tests pass.
- `.goreleaser.yaml` passes `goreleaser check`; `moon run :release-snapshot` builds all six targets + checksums + SBOMs offline.
- `release.yml` is actionlint/zizmor-clean (SLSA exception documented); `moon run :check` green.
- ADR-0016/0017 + this spec + the plan committed; `AGENTS.md`/`README.md`/`ARCHITECTURE.md` reflect `charter version` + the release pipeline; the architecture doc's SBOM/SLSA wording reconciled to as-built.
- No tag pushed; the go-public release path is documented for `v0.9.0`.

## References

- `docs/internal/decisions/0016-release-pipeline-and-supply-chain.md`, `0017-version-command-and-build-stamping.md`
- `docs/internal/decisions/0014-sarif-output-and-rule-catalog.md` (introduced `Version()`), `0006-latest-docs-first.md`, `0002-single-root-go-module.md`
- `docs/internal/architecture/charter-architecture-2026.md` (§1.8 command gallery + `charter version`; M1.5 T1.5.2 release pipeline)
- `docs/internal/superpowers/plans/2026-06-01-v1-launch-roadmap.md` (Slice 9 row)
- GoReleaser v2 docs (homebrew_casks, sboms, signs, snapshot); cosign v3 sign-blob/verify-blob `--bundle`; slsa-github-generator v2.1.0 `generator_generic_slsa3.yml`; GitHub OIDC `id-token: write`
- `docs/internal/superpowers/plans/2026-06-01-phase-1-slice-9.md` (derived implementation plan)
