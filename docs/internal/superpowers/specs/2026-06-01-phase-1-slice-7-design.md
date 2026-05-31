# Phase 1 Slice 7 Design — Governance & Suppression

## Goal

Implement the three governance rules — `AE-SUPPRESS-001` (suppression missing `reason`; Medium), `AE-SUPPRESS-002` (permanent suppression missing `approver`; High), and `AE-SUPPRESS-003` (high suppression rate; informational, score-neutral) — on top of a real suppression engine that reads both `.charter-suppress.yml` and inline `charter:ignore` comment directives, excludes suppressed findings from the score, and lists them separately. Add the `charter suppress` write command. With AE-SUPPRESS, the **full 15-rule v1 catalog is complete**. Implements ADR-0013; fulfills the architecture's standing "suppressed findings excluded from the base score" contract.

## Audience

- coding agents implementing this vertical slice
- maintainers reviewing the suppression contract, the score-exclusion change, and the new write command
- future contributors extending inline forms (next-line/block), profile-tunable thresholds, and SARIF `suppressions[]` emission

## Scope

### In scope

- a new `internal/suppress` package: `.charter-suppress.yml` loader (YAML), inline `charter:ignore` directive parsing (per-finding detection), and an `Apply` partition (file + inline, expiry via injectable clock, non-suppressible secrets)
- a new `internal/rules/governance` package: `AE-SUPPRESS-001/002/003`
- `internal/findings`: add `Informational bool`; `internal/scoring`: skip informational findings; `doctor.Result`: add a `Suppressed` list; `internal/doctor/run.go`: partition before scoring, then run governance
- all three renderers (`text`, `json`, `markdown`): a "suppressed (N)" section + surface informational findings without scoring them
- a new `charter suppress <RULE-ID>` Cobra command writing `.charter-suppress.yml` (`--reason --expires --approver --path --dry-run`)
- fixtures under `testdata/repos/`, as-built `AE-SUPPRESS-001/002/003` specs, audit checklist rows (+ HTML mirror), AGENTS.md/README/ARCHITECTURE updates, architecture-doc rule-list verification

### Out of scope

- next-line / block inline forms (`charter:ignore-next-line`, begin/end) — same-line anchoring only in v1
- profile-tunable `AE-SUPPRESS-003` thresholds — fixed v1 constants; tuning ships with the policy-profiles slice
- `--format sarif` emission of `suppressions[]` (M1.5); `charter fix` (M1.4)
- suppressing governance findings (self-suppression) — not supported in v1
- the "unscanned/unknown repo state ≤ 79" cap — unrelated scoring concern

## Why this slice

AE-SUPPRESS-001/002/003 are the last three v1 rules; landing them completes the 15-rule catalog. They cannot exist without a suppression engine, and that engine pays off a contract the canonical architecture already publishes (suppressed findings excluded from score, listed separately) but that `scoring.Calculate` does not yet honor. Doing suppression now (before SARIF/M1.5) means the later SARIF renderer emits correct active-vs-suppressed results on first build (`suppression.kind` maps directly).

## Grounding (verified live 2026-06-01, not from memory)

- **SARIF 2.1.0 `result.suppressions[]`** — each element has `kind ∈ {inSource, external}` and optional `status ∈ {accepted, underReview, rejected}` (default `accepted`); a result is suppressed iff it has suppressions and none are `underReview`/`rejected`; within a run all results' `suppressions` are all-null or all-non-null. `result.kind: "informational"` ⇒ importance `note`. → inline ⇒ `inSource`, file ⇒ `external`; `AE-SUPPRESS-003` ⇒ `kind: informational` (forward-compat; not emitted in v1).
- **Inline directive prior art** — gitleaks `gitleaks:allow` (same-line), golangci `//nolint:rule // reason` (same-line + reason), eslint `// eslint-disable-line rule -- reason`, semgrep `// nosemgrep`. Consensus: same-line anchoring, explicit rule ID, an attached reason. → Charter `charter:ignore <RULE-ID> reason="…"`.
- **Architecture truth** — §0.2 wording (`reason="…"`, `approver="…"`) confirms inline `key="value"` directives; §1.8 command gallery confirms `charter suppress … --expires 90d` writing `.charter-suppress.yml` and "re-surfaces on expiry"; Score Formula + Hard Caps confirm "excluded from the base score; listed separately".
- **Existing seams** — `findings.Finding` (+ `Locations`, `Cap`), `scoring.Calculate` (blocker cap ≤ 59 + per-finding `Cap`), `repository.Inventory` (`Has`, `Paths`), `config` already depends on `gopkg.in/yaml.v3`, the JSON/markdown renderers sort by severity-weight desc then rule_id asc.

## Suppression model (`internal/suppress`)

- **`Entry`** — `Rule, Reason, Approver string; Expires string` (ISO `YYYY-MM-DD`, the literal `permanent`, or empty ⇒ permanent); `Path string` (file source: optional finding-path scope; inSource: the directive's file, used for the governance location); `Source` (`external` | `inSource`); `Line int` (inSource directive line, 0 for file).
- **File load** — if `.charter-suppress.yml` is tracked, parse `suppressions: []`; malformed YAML ⇒ wrapped error (fail fast). Missing file ⇒ no external entries, no error.
- **Inline detection (per-finding)** — for a finding with a concrete `path:line`, read that line; a match requires the line to contain, inside a `#` / `//` / `<!-- … -->` comment, `charter:ignore <finding.RuleID>` plus optional `reason="…"`, `expires=YYYY-MM-DD`, `approver="…"`. Cache file contents per path within a single `Apply`.
- **Honoring** — an entry suppresses (is "honored") when: it has an explicit, non-expired, well-formed `expires` date; **or** it is permanent (`expires: permanent` or no `expires`) **and** carries an `approver`. A permanent entry without an `approver` is **not honored** (the finding stays active) but is still audited so `AE-SUPPRESS-002` flags it. An expired or malformed-date entry is inert (fail closed): not honored, not audited. There is no non-suppressible rule class — secrets follow the same rules (a permanent secret suppression without an approver is not honored, so the ≤ 49 cap holds).
- **`Apply(root string, all []findings.Finding, fileEntries []Entry, now time.Time) (active []findings.Finding, suppressed []Suppressed, used []Entry, err error)`** — for each finding, in inventory/finding order: a matching **honored** file entry (`rule` equal; `path` empty or equal to a finding location path) ⇒ suppressed (`external`); else a matching **honored** inline directive at a finding location ⇒ suppressed (`inSource`); else active. `used` (governance input) = every non-expired file entry (honored or permanent-unapproved) + every inline directive discovered at a finding location (honored or permanent-unapproved). Read errors wrap with `%w`.
- **`Suppressed`** — `{Finding findings.Finding; Source string; Reason, Approver, Expires string}` (consumed by `doctor.Result` and the renderers).

## Governance rules (`internal/rules/governance`)

- **`Run(used []suppress.Entry, activeRuleCount, suppressedCount int) []findings.Finding`** (pure; no disk; expiry is already pre-filtered by `Apply`, so `used` carries only non-expired file entries + matched inline directives — governance needs no clock):
  - `AE-SUPPRESS-001` (Medium): one finding per `used` entry whose `Reason` is blank. Evidence names the suppressed rule + source; location is `.charter-suppress.yml` (file entry) or the directive `path:line` (inline).
  - `AE-SUPPRESS-002` (High): one finding per `used` entry that is permanent (`Expires` empty or `permanent`) **and** has a blank `Approver`. Evidence notes the suppression is not honored long-term and will re-fire.
  - `AE-SUPPRESS-003` (informational): one finding when `suppressedCount / (activeRuleCount + suppressedCount) > 0.30` (the audit checklist's 30% threshold; guarded against a zero denominator). `Informational: true`, Severity Medium, score-neutral; evidence states the rate.
- `Apply` builds `used` as: every non-expired `.charter-suppress.yml` entry (audited even if its rule did not fire) plus each inline directive that matched a finding. Governance findings are appended to the active set after the partition and are not themselves suppressible.

## Score & result changes

- `findings.Finding` gains `Informational bool`.
- `scoring.Calculate` skips findings with `Informational == true` (they neither deduct nor cap). Suppressed findings never reach `Calculate` (partitioned out upstream), so the blocker cap (≤ 59) and per-finding `Cap` (secret ≤ 49) reflect active findings only.
- `doctor.Result` gains `Suppressed []suppress.Suppressed`. `Findings` stays active-only.

## doctor pipeline order (`internal/doctor/run.go`)

1. resolve root, build inventory (unchanged).
2. run all existing rules → `all` (unchanged).
3. `fileEntries, err := suppress.LoadFile(root, inv)` — fail fast.
4. `active, suppressed, used, err := suppress.Apply(root, all, fileEntries, time.Now())` — fail fast.
5. `gov := governance.Run(used, len(active), len(suppressed))` (no clock — `Apply` already pre-filtered expiry into `used`); `active = append(active, gov...)`.
6. `score := scoring.Calculate(active)`.
7. `Result{Findings: active, Suppressed: suppressed, Score: score, …}`.

## `charter suppress` command (`cmd/charter/suppress.go`)

- `charter suppress <RULE-ID> --reason "…" [--path .] [--expires 90d|YYYY-MM-DD|permanent] [--approver NAME] [--dry-run]` — `--path` selects the repo (same meaning as `doctor`); resolve its root, read existing `.charter-suppress.yml` (or empty), upsert the rule-level entry (replace an existing entry for the same `rule`), marshal deterministically, write. `--expires` defaults to `90d`; a duration ⇒ absolute ISO date from `time.Now()`, a bare date stored as-is, `permanent` stored literally with a warning that it needs `--approver` and is governance-flagged. `<RULE-ID>` must be a known rule ID (validated). Print the exact entry + file path; `--dry-run` prints without writing. Exit 0 on success, 2 on error. A per-finding `--scope` flag is deferred (decision 7); hand-authored `path:` entries are still honored by the engine.
- Writes only `.charter-suppress.yml` (Charter-owned); never user source. Distinct from the diff-first `charter fix` flow (ADR-0013 decision 7).

## Architecture / ownership

- `internal/suppress/` (new) — `entry.go`/`load.go` (`.charter-suppress.yml` loader), `inline.go` (directive grammar + per-line detection), `apply.go` (`Apply` partition + clock), fuzz on both parsers. Imports `findings`, `repository`; YAML via `gopkg.in/yaml.v3`. No rule-package imports.
- `internal/rules/governance/` (new) — `governance.go` + the three checks; imports `suppress` + `findings`; pure.
- `internal/findings/` — add `Informational`. `internal/scoring/` — skip informational.
- `internal/doctor/run.go` — partition + governance + score over active; `Result.Suppressed`.
- `internal/render/{json,markdown}` + `cmd/charter/doctor.go` text path — suppressed section + informational surfacing.
- `cmd/charter/suppress.go` (new) + `root.go` (register).

Avoid: a generic policy engine, repo-wide directive scanning, profile config plumbing, or a renderer registry in v1.

## Go alignment (per golang-patterns / golang-testing)

- Pure functions take data, return concrete types; disk touch isolated in `LoadFile`/`Apply`; clock injected (`now time.Time`) so expiry is deterministic.
- Errors are values: loaders/`Apply` wrap with `%w` and fail fast; `run.go` propagates; nothing discarded with `_`.
- No package-level mutable state (non-suppressible set + directive regex are read-only).
- Deterministic output: stable finding order; entries processed in inventory/finding order; YAML written via a struct for stable key order.
- Testing: table-driven `t.Run` subtests, `t.Helper()`, `t.TempDir()`, `t.Parallel()` for independent units, fuzz targets on the YAML loader and the inline directive parser, `-race` via `moon run :test`, ≥ 85% line coverage for the new packages.

## Testing strategy

- unit: YAML entry loading (valid, missing file, malformed→error); inline directive parse (`#`/`//`/`<!-- -->`, `reason`/`expires`/`approver`, malformed→no match); `Apply` partition (honored file match, honored inline match, path-scoped match, explicit-date expired→active, permanent+approver→suppressed, permanent-without-approver→active-but-audited, malformed-date→inert, no-match→active, injected clock); governance (001 blank reason, 002 permanent+no approver, 003 rate over/under 30% + zero-denominator guard); scoring skips informational; each renderer's suppressed/info section.
- fuzz: `FuzzLoadSuppressions` and `FuzzParseInlineDirective` never panic on junk.
- integration: fixture repos through the doctor pipeline — clean (no suppressions), file-suppressed finding (excluded from score + listed), inline-suppressed finding, missing-reason→AE-SUPPRESS-001, permanent-no-approver→AE-SUPPRESS-002, high-rate→AE-SUPPRESS-003 informational, permanent-unapproved secret suppression keeps the ≤49 cap (an approved one lifts it), malformed YAML→error.
- CLI: `charter suppress` writes/updates `.charter-suppress.yml`, `--dry-run` writes nothing, `--expires 90d` stores an absolute date; round-trip (`suppress` then `doctor` shows the finding suppressed).
- dogfood: Charter's own repo has no `.charter-suppress.yml` and no inline directives ⇒ zero AE-SUPPRESS findings, score stays 100.
- `moon run :check` stays green.

## Risks

- **Inline false positives from docs mentioning the syntax** — control: per-finding detection only reads a finding's own `path:line`; repo-wide scanning is never performed.
- **Suppression as a secrets escape hatch** — control: secrets are suppressible per the docs, but only under governance — a permanent secret suppression is not honored without an `approver` (AE-SUPPRESS-002, High), so the ≤ 49 cap returns; time-bounded secret suppressions re-fire on expiry; the suppressed secret is always listed separately.
- **Score-contract regressions** — control: partition before `Calculate`; informational skipped; existing scoring tests plus new suppressed/informational tests pin the formula and caps.
- **Write command mutating unexpected files** — control: only `.charter-suppress.yml` is written; `--dry-run`; the entry is echoed; covered by CLI tests asserting no other file changes.
- **Expiry non-determinism in tests** — control: injectable `now`; CLI stores absolute dates.
- **YAML drift / key order** — control: marshal from a typed struct; round-trip test.

## Success criteria

- a finding covered by a valid `.charter-suppress.yml` entry (or an inline `charter:ignore`) is excluded from the score and listed under "suppressed"; an expired entry lets the finding re-surface as active.
- `AE-SUPPRESS-001` fires on a reason-less suppression (Medium), `AE-SUPPRESS-002` on a permanent suppression without approver (High), `AE-SUPPRESS-003` as a score-neutral informational finding when the suppression rate exceeds 30%; a permanent secret suppression without an approver is not honored (the ≤ 49 cap holds), and an approved one lifts it.
- `charter suppress AE-X --reason "…" --expires 90d` writes/echoes a `.charter-suppress.yml` entry with an absolute date; `--dry-run` writes nothing; `doctor` then shows the finding suppressed.
- the `AE-SUPPRESS-001/002/003` specs are as-built; ADR-0013 is referenced; AGENTS.md/README/ARCHITECTURE/audit reflect the 15 implemented rules.
- `moon run :check` is green and the dogfood `charter doctor` score is unchanged (100).

## References

- `docs/internal/decisions/0013-suppression-model.md`
- `docs/internal/decisions/0008-score-formula.md`, `0009-finding-location.md`, `0005-diff-first-fixes.md`, `0006-latest-docs-first.md`
- `docs/internal/architecture/charter-architecture-2026.md` (§0 Score Formula + Hard Caps, §0.2 AE-SUPPRESS rows, §1.8 `charter suppress`)
- `docs/internal/specs/AE-SUPPRESS-001.md`, `AE-SUPPRESS-002.md`, `AE-SUPPRESS-003.md`
- `docs/internal/superpowers/specs/2026-05-31-phase-1-slice-6-design.md`
- `docs/internal/superpowers/plans/2026-06-01-phase-1-slice-7.md` (derived implementation plan)
