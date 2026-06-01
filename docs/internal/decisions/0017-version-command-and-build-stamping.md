# ADR-0017 `charter version` Command & Build Stamping

- Status: Accepted
- Context: `charter version` is one of the six v1 commands (architecture §1.8) and the only one still unwired — `cmd/charter/root.go` registers `doctor` and `suppress`. ADR-0014/Slice 8 already added `internal/version.Version()` (ldflags-injected `injected` → `runtime/debug.ReadBuildInfo().Main.Version` → `0.0.0-dev`) for SARIF `tool.driver.version`. The architecture documents the exact output the command should print:
  ```
  charter   1.0.0
  commit    abc1234f
  built     2026-05-18T09:42:00Z
  go        1.26.3
  platform  darwin/arm64
  ```
  So three fields beyond `version` are needed — `commit`, `built`, `go`, `platform` — and a build-stamp mechanism that is accurate both for release binaries (GoReleaser, ADR-0016) and for plain `go build`/`go run` during development. Verified (2026-06-01): `runtime/debug.ReadBuildInfo()` exposes VCS data in `Settings` (`vcs.revision`, `vcs.time`, `vcs.modified`) for binaries built from a checkout, and `runtime.Version()` / `runtime.GOOS` / `runtime.GOARCH` give the toolchain and platform with no inputs.
- Decision:
  1. **Extend `internal/version`, don't fork it.** Add package-level `commit` and `date` vars (ldflags targets, mirroring `injected`) and `Commit()` / `Date()` accessors. Precedence per field: **ldflags-injected value → `runtime/debug` VCS setting (`vcs.revision`/`vcs.time`) → `"unknown"`.** `Version()` is unchanged. This is the 2026 idiom: releases are authoritative via ldflags (ADR-0016 injects `{{.Version}}`/`{{.FullCommit}}`/`{{.CommitDate}}`), while dev builds still self-describe from embedded VCS info, and the binary never carries secrets or network calls.
  2. **`go` and `platform` are computed at runtime**, never injected: `go` = `strings.TrimPrefix(runtime.Version(), "go")`, `platform` = `runtime.GOOS + "/" + runtime.GOARCH`.
  3. **`cmd/charter/version.go` adds a pure `version` subcommand** registered in `root.go`, printing the five aligned `label  value` lines above to stdout (deterministic, no flags in v1). It depends only on `internal/version` + `runtime` — no scan, no I/O beyond stdout — so it always exits 0.
- Consequences: `internal/version` gains `commit`/`date` vars + `Commit()`/`Date()` (unit-tested for the `"unknown"` dev fallback, matching the existing `Version()` test); `cmd/charter/version.go` + a CLI test assert the labels and that `version` reports `version.Version()`. ADR-0016's GoReleaser ldflags line is the injection point. Output is intentionally minimal — no `--format json`/`--short` in v1 (deferred). Related: ADR-0014 (introduced `Version()`), ADR-0016 (release-time injection), architecture §1.8.
