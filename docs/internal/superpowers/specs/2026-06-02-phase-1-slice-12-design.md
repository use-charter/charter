# Phase 1 Slice 12 Design — `charter fix`

## Goal

Ship `charter fix` (architecture M1.4 T1.4.2): a diff-first repair engine that applies safe, reversible fixes for selected findings — previewing unified diffs, backing up before every write, never deleting, never silently mutating, never auto-fixing secrets. v1 fixers: `AE-CTX-001` (AGENTS.md), `AE-CTX-004` (.gitignore), `AE-CI-002` (Charter CI workflow). Also generalizes `AE-CI-002`'s moon-specific coverage detection (a cross-vendor FP) and exempts the first-party action from the pin check. Implements ADR-0020.

## Audience

- coding agents implementing this slice
- maintainers reviewing the fix engine's safety (backups, never-overwrite-unrelated, never-secret) and the AE-CI-002 generalization
- users running `charter init` then `charter fix` to reach a clean scan

## Scope

### In scope

- `internal/fix`: registry (RuleID→fixer), `Plan(result, root, opts) []FilePlan` (pure), unified-diff builder, applier (backup + write); unit-tested
- fixers: `AE-CTX-001` (create AGENTS.md), `AE-CTX-004` (append/create .gitignore), `AE-CI-002` (create .github/workflows/charter.yaml)
- `cmd/charter/fix.go` + `root.go`: `--rule/--dry-run/--all/--yes/--path`; CLI tests
- `AE-CI-002` generalization (direct/action coverage forms) + first-party-action pin exemption; rule tests + spec/audit update
- integration test (findings → fix → improved doctor) + docs sync

### Out of scope

- `AE-ENV-001`/`AE-MCP-001` fixers (hook-config opinionated / needs Slice 13 catalog); rewriting present-but-weak files; secret/dangerous-command auto-fix (never)
- content-aware/3-way diffs; interactive fix selection; the GitHub Action/release (Slices 9/10)

## Why this slice

`charter fix` completes the M1.4 onboarding loop (`init` scaffolds; `fix` repairs, incl. the existing-file edits `init` deferred) and powers the architecture's hero PR scenario ("`charter fix --rule AE-CTX-001` can scaffold the file"). The AE-CI-002 generalization removes a false positive that would hit nearly every external repo.

## Grounding (verified against the code)

- `doctor.Run(path, threshold, thresholdSet) (Result, error)`; `Result.Findings []findings.Finding` with `RuleID`/`Severity`/`Locations`. `internal/scaffold` builds `AGENTSMarkdown`/`Gitignore`. ADR-0005: fixes diff before apply, never silent.
- `internal/rules/ci/ci002.go`: `markCoverage` sets `repo-quality` (moon :check / moon test-family), `charter-product-gate` (`isCharterDoctorCommand` / `use-charter/charter-action@`), `workflow-lint` (moon :actionlint AND :zizmor), `security` (moon :security); a missing area adds evidence → LOW finding. `unpinnedActionEvidence` flags non-SHA `uses:` except `./` and the `slsaReusableWorkflowPin`. Helpers exist: `hasExecutableCommand(runs, matcher)`, `extractWorkflowExecutables` (parses `run:`/`uses:`), `normalizeCommand` (strips `mise x --`/`env`), `hasUsePrefix`.
- Secrets/dangerous rules (AE-SEC-001/002, AE-CC-001) must NOT be fixable (manual remediation, Commitment #9).

## `internal/fix` (new, engine)

- `Action` ∈ {Create, Append}. `FilePlan{RuleID, Path, Action, NewContents []byte, Diff string}`.
- Registry: `map[string]fixer` where `fixer(root string, inv repository.Inventory) (FilePlan, bool, error)` — returns the plan for a rule, `ok=false` when nothing to do (e.g. target already satisfies). Only `AE-CTX-001`, `AE-CTX-004`, `AE-CI-002` are registered.
- `Plan(result doctor.Result, root string, inv, opts Options) ([]FilePlan, error)`: for each active finding whose RuleID has a registered fixer (and matches `opts.Rule` when set), invoke the fixer, collect plans. Pure (no writes); builds the `Diff`.
- **Unified diff** (`diff.go`): for `Create`, emit a `--- /dev/null` / `+++ b/<path>` header + one `@@ -0,0 +1,N @@` hunk of all-`+` lines; for `Append`, emit a hunk showing the appended `+` lines with a few trailing context lines of the existing tail. Hand-rolled; deterministic.
- **Applier** `Apply(root string, plans []FilePlan) ([]Applied, error)`: for each plan, if the target exists, copy it to `.charter/backups/<ts>/<path>` (MkdirAll parents) first; then write (`Create`: write NewContents to the absent path; `Append`: read existing + append the missing lines). Never `Remove`/truncate; never write outside `root`. Returns what was written + backup paths.
- The engine imports `internal/scaffold` for AGENTS.md/.gitignore contents and `internal/doctor`/`findings`/`repository`.

## Fixers

- **AE-CTX-001** (`fix_ctx001.go`): if no agent context file exists, plan `Create AGENTS.md` = `scaffold.AGENTSMarkdown(scaffold.Detect(root))`. If a context file already exists, `ok=false` (present-but-weak rewrite deferred).
- **AE-CTX-004** (`fix_ctx004.go`): compute the missing agent-artifact patterns (`.charter/`, `*.charter-session`, `.claude/local/`, `.cursor/cache/`, `.hk/`, `.env*`) vs the current `.gitignore`. If `.gitignore` absent → `Create` (`scaffold.Gitignore`); else `Append` only the missing lines (with a leading `# Charter / agent session artifacts` comment if appending). `ok=false` when all present.
- **AE-CI-002** (`fix_ci002.go`): if no workflow already provides the charter-gate, plan `Create .github/workflows/charter.yaml`:
  ```yaml
  name: Charter
  on:
    pull_request:
    push:
      branches: [main]
  permissions:
    contents: read
    security-events: write
  jobs:
    charter:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@<sha>   # SHA-pinned
        - uses: use-charter/charter-action@v1
          with:
            threshold: "80"
  ```
  (`actions/checkout` SHA-pinned with a version comment; `use-charter/charter-action@v1` is the first-party exempt ref.) This statically clears the charter-gate; the action runs doctor + uploads SARIF. `ok=false` if a charter-gate workflow already exists.

## `AE-CI-002` generalization (`internal/rules/ci/ci002.go`)

- Add matchers used inside `markCoverage` alongside the moon forms:
  - repo-quality also true when `hasExecutableCommand(runs, isTestOrBuildCommand)` — `isTestOrBuildCommand` recognizes `go test`/`go build`, `npm test`/`npm run test`, `pnpm test`/`pnpm run test`, `yarn test`, `cargo test`/`cargo build`, `pytest`/`python -m pytest`, `bun test`, `make test`/`make check` (prefix/word-boundary aware, via the existing `normalizeCommand`).
  - workflow-lint also true when (`actionlint` direct-or-action) AND (`zizmor` direct-or-action): a `run:` command starting with `actionlint`/`zizmor`, or a `uses:` containing `rhysd/actionlint` / `zizmorcore/zizmor`.
  - security also true when a recognized scanner is present: `run:` `govulncheck`/`osv-scanner`/`gitleaks`/`trivy`/`grype`, or a `uses:` containing `github/codeql-action`.
- `unpinnedActionEvidence`: add a `firstPartyActionPin` exemption (e.g. ref has prefix `use-charter/charter-action@`) next to the `slsaReusableWorkflowPin` skip.
- Charter's own repo (moon forms) keeps passing; spec `AE-CI-002.md` + audit checklist updated to describe the generic forms + the first-party exemption.

## `cmd/charter/fix.go`

- `newFixCommand()` (registered in `root.go`). Flags: `--path`, `--rule` (validate `AE-XXX-NNN`), `--dry-run`, `--all` (default behavior is all fixable; `--all` reserved/explicit), `--yes`.
- Resolve root (`repository.ResolveRoot`), build inventory, `doctor.Run` (to get findings), `fix.Plan(...)`. Print each plan's diff. If `--dry-run`, stop (write nothing). Else `fix.Apply(...)`, printing created/appended files + the backup dir. Report `N fixed` + `› Re-run: charter doctor`. Exit 0; flag/IO error → exit 2. Secret/dangerous findings without a fixer are simply not planned (and a note can list them as manual).

## Architecture / ownership

- `internal/fix/` (new): registry, `Plan`, `diff`, `Apply`, per-rule fixers. Pure planning; the applier is the only writer (backup-then-write; never delete).
- `cmd/charter/fix.go` (new) + `root.go`.
- `internal/rules/ci/` — generalized coverage + first-party exemption.
- Avoid: a diff dependency; overwriting without backup; deleting/truncating; making any secret/dangerous rule fixable; a present-but-weak-file rewrite.

## Go alignment

- Pure planning (fixers + `Plan` + diff are pure); the applier owns I/O, wraps errors `%w`, fails fast, no `_` discards, gosec-annotates fixed-path reads/writes. Backups + writes are `0o644`/`0o755`. Deterministic ordering. Tests: registry/plan/diff golden; applier backup+write (and never-overwrite-unrelated); CI-002 generalization matrix; CLI dry-run/apply/rule; `-race`; ≥85% for `internal/fix`.

## Testing & verification strategy

- **fix unit:** registry maps the 3 rules; `Plan` builds correct `FilePlan`s + diffs (golden); `Apply` backs up an existing `.gitignore` to `.charter/backups/...` then appends only missing patterns; `Create` writes only to an absent path; secret rules have no fixer.
- **CI-002 generalization unit:** a workflow running `go test` + `actionlint` + `zizmor` + `govulncheck` (no moon) satisfies all coverage; a workflow using `use-charter/charter-action@v1` is NOT flagged unpinned; Charter's own moon workflows still pass; a repo missing security still fires that sub-evidence.
- **CLI:** `fix --dry-run` writes nothing (prints diffs); `fix --rule AE-CTX-004` on a repo with a partial `.gitignore` appends the missing lines + writes a backup; `fix --rule AE-SEC-001` is a no-op (not fixable) with a clear message; invalid `--rule` → exit 2.
- **Integration:** synthesize a Go repo with a normal CI workflow (`go test` + `actionlint` + `zizmor` + `govulncheck`) but no Charter and a partial `.gitignore`; run `fix`; re-`doctor`; assert AE-CTX-004 and AE-CI-002 are cleared and the score rises. Hermetic.
- **Dogfood:** Charter's own `charter doctor` stays **100** (generalization keeps moon forms passing); `charter fix --dry-run` on Charter's repo proposes nothing for the registered rules (all already satisfied) — confirm.
- `moon run :check` green.

## Risks

- **Accidental overwrite/data loss** — control: backup-before-write to `.charter/backups/`; `Append` only adds lines; never `Remove`/truncate; `--dry-run`; tests assert backups + that unrelated bytes are untouched.
- **Secrets auto-fixed** — control: AE-SEC/AE-CC are unregistered; a test asserts `fix --rule AE-SEC-001` is a no-op.
- **CI-002 generalization regresses Charter's 100 or is over-broad** — control: matrix tests (moon still passes; direct forms pass; missing areas still fire); the first-party exemption is scoped to `use-charter/charter-action@`.
- **Scaffolded workflow doesn't clear AE-CI-002 / isn't valid** — control: integration test re-runs doctor; the workflow is actionlint-shaped (the slice's own `:actionlint`/`:zizmor` don't scan it as it's a generated string, but the integration scan validates AE-CI-002 clears). The action is unpublished pre-launch (armed-not-fired) — the workflow is correct for post-launch.

## Success criteria

- `charter fix` previews unified diffs and, on apply, backs up existing targets to `.charter/backups/` before writing; never deletes; never fixes secrets; `--rule`/`--dry-run` work; exit codes correct.
- `AE-CTX-001`/`AE-CTX-004`/`AE-CI-002` fixers produce correct files; integration test shows AE-CTX-004 + AE-CI-002 cleared after `fix` on a non-moon repo with normal CI (measured: `fix` raises a representative non-moon repo from 91 to 96).
- `AE-CI-002` recognizes direct/action CI forms + exempts the first-party action; Charter's own scan stays 100; spec/audit updated.
- ADR-0020 + this spec + the plan committed; AGENTS/README/ARCHITECTURE + architecture §1.8/T1.4.2 reflect `charter fix`; `moon run :check` green.

## References

- `docs/internal/decisions/0020-charter-fix-engine.md` (ADR-0020); `0005` (diff-first), `0019` (init/scaffold), `0016` (SLSA exemption precedent), `0008` (score)
- `docs/internal/architecture/charter-architecture-2026.md` (§1.8 `charter fix`; M1.4 T1.4.2); `internal/rules/ci/ci002.go`; `internal/scaffold`
- `docs/internal/superpowers/plans/2026-06-02-phase-1-slice-12.md` (derived implementation plan)
