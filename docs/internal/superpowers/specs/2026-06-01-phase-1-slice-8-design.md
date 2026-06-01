# Phase 1 Slice 8 Design — SARIF Output & Policy Profiles

## Goal

Ship two M1.1/M1.5-prep capabilities: (1) a **SARIF 2.1.0** renderer (`charter doctor --format sarif [--out file]`) backed by a new **rule catalog**, so Charter findings can be uploaded to GitHub Code Scanning; and (2) **policy profiles** in `charter.yaml` (`policy.profile`/`policy.threshold`) so teams gate pipelines without per-invocation flags. Implements ADR-0014 (SARIF + catalog) and ADR-0015 (policy profiles).

This slice does not build `charter init`/`charter fix` (M1.4), the GitHub Action (M1.5), `charter explain`, or rule-level enable/disable.

## Audience

- coding agents implementing this vertical slice
- maintainers reviewing the SARIF contract, the rule catalog, and the policy-resolution precedence
- future contributors building the GitHub Action (consumes SARIF), `charter explain` (reuses the catalog), and a richer policy engine

## Scope

### In scope

- `internal/rules/catalog`: a static `{ID, Name, Category, ShortDescription, HelpURI}` table for all 15 rules + a spec-sync test (catalog IDs == `docs/internal/specs/AE-*.md`)
- `internal/render/sarif`: `Render(doctor.Result) ([]byte, error)` → SARIF 2.1.0 (tool.driver + rules[] + results[] for active & suppressed + partialFingerprints)
- `internal/version`: an ldflags-injectable `Version` var (default `0.0.0-dev`) for `tool.driver.version`
- `internal/config` + a small policy resolver: load `policy.{profile,threshold}`; resolve the effective threshold by precedence; built-in `strict=90/standard=80/relaxed=60`
- `internal/doctor`: `Run(path, threshold, thresholdSet)` resolves and reports the effective threshold
- `cmd/charter/doctor.go`: `--format sarif`, `--out <file>`, threshold precedence via `Changed`
- `schemas/charter-config.schema.json`; fixtures; docs sync (AGENTS/README/ARCHITECTURE/architecture-doc)

### Out of scope

- `charter init` / `charter fix` (M1.4); the GitHub Action + GoReleaser (M1.5)
- `charter explain`, `--rule` filtering, SARIF 2.2, `artifacts[]`/`invocation` enrichment
- rule-level enable/disable, per-rule severity overrides, `rules.ignore` (larger policy engine)
- `--no-color` / plain-CI text variant (separate small feature)

## Why this slice

SARIF is the spine of the Phase 1 distribution surface: the M1.5 Action uploads it to Code Scanning, which is the hardest-to-fake validation signal (organic CI adoption). Slices 4 (Location) and 7 (suppression source + Informational) already made the model SARIF-shaped, so this is mostly a renderer + a metadata catalog. Policy profiles are small and unblock `charter init` (M1.4) and CI gating. Both share the `charter doctor` CLI surface, so they ship together cleanly.

## Grounding (verified live 2026-06-01, not from memory)

- **GitHub Code Scanning** ingests **SARIF 2.1.0** and uses `tool.driver` (incl. `rules[]` of `reportingDescriptor`), `results[]`, and `partialFingerprints` — it reads only `primaryLocationLineHash` and backfills from source if absent. `shortDescription`/`fullDescription` show at the top of an alert; `helpUri` renders as a link; `result.locations[].physicalLocation.{artifactLocation.uri, region.startLine}` drives code annotations.
- **SARIF suppression** (from Slice 7 research): `suppression.kind ∈ {inSource, external}`; a run's results must have `suppressions` all-null or all-non-null. `result.kind: "informational"` → importance note.
- **Library vs hand-roll:** `owenrumney/go-sarif` v3.3.0 (Go 1.24, Unlicense) is maintained, but adds a dependency; the needed subset is small → hand-roll (ADR-0014).
- **Existing seams:** `cmd/charter/doctor.go` validates `--format text|json` (then markdown in Slice 6); `internal/config` loads only `mcp.trustedRemotes` via yaml.v3; `doctor.Run(path, threshold)` resolves root, builds inventory, runs rules, partitions suppressions, scores; `internal/version` is a doc-only stub.

## Rule catalog (`internal/rules/catalog`)

- `Entry` = `{ID, Name, Category, ShortDescription, HelpURI}` (no severity — see below). `HelpURI` = `https://use-charter.dev/rules/<ID>`.
- `All() []Entry` (or a map keyed by ID) covering the 15 rules: AE-CTX-001/002/004, AE-SEC-001/002, AE-MCP-001/002/003, AE-CC-001/002, AE-ENV-001, AE-CI-002, AE-SUPPRESS-001/002/003. `Lookup(id) (Entry, bool)`.
- **No severity in the catalog**: a result's `level` and a rule's `defaultConfiguration.level` are derived from the actual finding `Severity` (single source of truth). The catalog supplies stable rule-level `Name`/`ShortDescription`/`HelpURI`/`Category`.
- **Spec-sync test**: enumerate `docs/internal/specs/AE-*.md` (excluding README), assert the filename stems exactly equal the catalog ID set — a drift guard binding the catalog to the behavioral contracts.

## SARIF renderer (`internal/render/sarif`)

- `Render(doctor.Result) ([]byte, error)` builds a SARIF 2.1.0 log: `{version: "2.1.0", $schema, runs: [run]}`.
- **run**: `automationDetails.id: "charter"` (logical analysis id for clean re-run replacement / multi-tool coexistence).
- **tool.driver**: `name: "Charter"`, `informationUri: "https://use-charter.dev"`, `version: version.Version()`, `rules: [...]`.
- **rules[]**: the distinct rule IDs across active + suppressed findings, deduped, sorted by ID; each = `{id, name, shortDescription.text, helpUri, defaultConfiguration.level, properties}` (metadata from the catalog; level/severity from the finding). `properties` = `{category, severity}` always; non-informational rules also add `tags: ["security"]` + `security-severity` (Blocker 9.5 / High 7.5 / Medium 5.0 / Low 2.0) for GitHub security classification/ranking. Results reference rules via `ruleIndex`.
- **results[]**: for each finding (active and suppressed), `{ruleId, ruleIndex, level, message.text: Summary, locations[], partialFingerprints}`; suppressed findings add `suppressions: [{kind}]`; informational findings add `kind: "informational"`. When any result is suppressed, active results carry `suppressions: []` (consistency rule).
- **level mapping:** Blocker/High → `error`, Medium → `warning`, Low → `note`.
- **locations[]**: only when the finding has a location with a non-empty path; `physicalLocation.artifactLocation.uri` = repo-relative forward-slash path; `region.startLine` only when line > 0.
- **partialFingerprints.primaryLocationLineHash**: hex `sha256(ruleId\x00path\x00line)` from the finding's primary location (line 0 for file-level), or `sha256(ruleId)` with no location. Position-based and computed purely from the finding — no source I/O — so the renderer stays pure and never fails on fingerprinting. (Content-based line-shift resilience is deferred.)
- Deterministic ordering mirrors the JSON renderer (severity weight desc, then rule_id asc) for results; rules[] sorted by ID.

## Policy profiles (`internal/config` + resolver)

- `charter.yaml`: `policy: {profile: strict|standard|relaxed, threshold: 0..100}` (both optional). `config.LoadPolicy(root, inv) (Policy, error)` parses it (missing file/section → zero Policy, no error; malformed → wrapped error).
- Built-in profiles: `strict=90, standard=80, relaxed=60`.
- `Resolve(p Policy, flagThreshold int, flagSet bool) (int, error)`: `flagSet` → flagThreshold; else `p.Threshold` if set (validate 0..100); else `p.Profile` mapped (unknown → error); else 80. Out-of-range/unknown → wrapped error.
- `doctor.Run(path string, threshold int, thresholdSet bool)`: after building the inventory, `LoadPolicy` + `Resolve`; the effective threshold sets `Result.Threshold` and `Passed`.

## CLI surface (`cmd/charter/doctor.go`)

- `--format` accepts `text|json|markdown|sarif` (reject others, exit 2).
- `--out <file>`: when set, write the rendered bytes to the file (0o644) and print a one-line summary to stderr; when absent, write to stdout (current behavior). Applies to any format.
- Pass `threshold` + `cmd.Flags().Changed("threshold")` into `doctor.Run`. Below the effective threshold → exit 1 (silent); scan/render/IO error → exit 2.
- Optional: SARIF `run.properties.charter.{profile,threshold}` for traceability.

## Architecture / ownership

- `internal/rules/catalog/` (new) — pure metadata table + lookup; imported by `render/sarif` (and later `charter explain`). No imports of rule packages.
- `internal/render/sarif/` (new) — `Render(doctor.Result) ([]byte, error)`; depends on `doctor`, `findings`, `suppress`, `catalog`, `version`. Mirrors `render/json`.
- `internal/version/` — add a `Version()` func: ldflags-injected value → else `runtime/debug.ReadBuildInfo().Main.Version` (accurate for `go install …@vX`) → else `0.0.0-dev`.
- `internal/config/` — `Policy` type + `LoadPolicy`; built-in profile table; `Resolve` (here or a thin `internal/policy`).
- `internal/doctor/run.go` — `Run(path, threshold, thresholdSet)`; resolve effective threshold.
- `cmd/charter/doctor.go` — `--format sarif`, `--out`, `Changed("threshold")`.
- `schemas/charter-config.schema.json` (new).

Avoid: a SARIF library dependency, a charter-authored SARIF schema, a generic policy/rules engine, or putting severity in the catalog.

## Go alignment (per golang-patterns / golang-testing)

- Pure functions take data and return concrete types; the SARIF renderer is pure (position-based fingerprints, no source I/O); the only disk touch is the config load.
- Errors are values, wrapped with `%w`, fail fast (unknown profile, out-of-range threshold, malformed config, unreadable `--out`); nothing discarded with `_`.
- No package-level mutable state (catalog table, profile map, level map are read-only).
- Deterministic output: stable result + rules ordering; fingerprints content-derived; JSON marshaled from typed structs.
- Testing: table-driven `t.Run` subtests, `t.Helper()`, `t.TempDir()`, golden SARIF, the catalog spec-sync test, `-race`, ≥85% line coverage for the new packages.

## Testing strategy

- catalog: spec-sync (IDs == specs/AE-*.md), `Lookup` hit/miss.
- sarif unit: valid 2.1.0 shape; severity→level; rules[] deduped/sorted from the catalog with `ruleIndex` wiring; `automationDetails.id`; non-informational rules carry `tags:["security"]` + `security-severity`, informational omit both; suppressed result carries `suppressions[{kind}]` and active results carry `suppressions: []`; informational → `kind: informational`; fingerprint determinism + position/no-location fallbacks; absence finding emits no `locations`.
- sarif conformance: every emitted document validates against the vendored official `sarif-2.1.0` JSON schema using the test-only `github.com/santhosh-tekuri/jsonschema/v6`; plus a golden-file test locking the exact bytes for a known fixture.
- config/policy: `LoadPolicy` (present/missing/malformed); `Resolve` precedence matrix (flag>threshold>profile>default), unknown profile → error, out-of-range → error.
- doctor: `Run` honors the resolved threshold (fixture `charter.yaml` with `policy.profile: strict` → threshold 90; flag overrides config).
- CLI: `--format sarif` to stdout is valid SARIF; `--out` writes the file + nothing to stdout; invalid format → exit 2.
- dogfood: Charter's own repo — `charter doctor --format sarif` is valid SARIF with 0 results (score 100); default profile (none) keeps the gate at 80; `--format sarif --out charter.sarif` writes a file.
- `moon run :check` stays green.

## Risks

- **SARIF correctness** — control: every emitted document is validated against the vendored official `sarif-2.1.0` schema (test-only `santhosh-tekuri/jsonschema/v6`), plus golden-file + structural tests; mapping kept to the documented Code-Scanning subset; real upload check at M1.5. Cost: one pinned test-only dependency (never shipped) — a deliberate, documented exception to dependency-minimalism (ADR-0014).
- **Catalog drift from specs** — control: the spec-sync test fails CI if the ID sets diverge.
- **partialFingerprints churn** — control: content-based hash (line-shift resilient) with deterministic fallbacks; documented.
- **Threshold-precedence ambiguity** — control: `Changed("threshold")` distinguishes a set flag from the default; explicit precedence matrix tested.
- **Backward compatibility of the gate** — control: no config + no flag still yields 80; existing tests updated to the new `Run` signature with `thresholdSet=true` (meaning 80) preserved.

## Success criteria

- `charter doctor --format sarif` emits valid SARIF 2.1.0 with `tool.driver.rules[]`, results for active + suppressed findings (suppressed carrying `suppressions[]`), informational `kind`, correct levels, and `partialFingerprints`; `--out` writes it to a file.
- A repo with `charter.yaml` `policy.profile: strict` gates at 90 with no flags; `--threshold` overrides it; an unknown profile fails fast (exit 2).
- the catalog spec-sync test passes; `schemas/charter-config.schema.json` documents the config; AGENTS/README/ARCHITECTURE reflect SARIF + profiles.
- `moon run :check` green; dogfood `charter doctor` still scores 100 and emits valid empty-result SARIF.

## References

- `docs/internal/decisions/0014-sarif-output-and-rule-catalog.md`, `0015-policy-profiles.md`
- `docs/internal/decisions/0009-finding-location.md`, `0013-suppression-model.md`, `0008-score-formula.md`, `0004-contract-first-interfaces.md`
- `docs/internal/architecture/charter-architecture-2026.md` (§0 brief, §1.8 SARIF + `charter init` profile, M1.1 T1.1.3/T1.1.5, M1.5 T1.5.1)
- GitHub code-scanning SARIF support (2.1.0 subset; `primaryLocationLineHash`); SARIF 2.1.0 OASIS spec
- `docs/internal/superpowers/specs/2026-06-01-phase-1-slice-7-design.md`
- `docs/internal/superpowers/plans/2026-06-01-phase-1-slice-8.md` (derived implementation plan)
