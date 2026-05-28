# Phase 1 Slice 2 Design

## Goal

Add machine-readable JSON output to the existing `charter doctor` command for the currently implemented Slice 1 rule set, without expanding rule scope or introducing config-loading complexity.

## Audience

- coding agents implementing the next vertical slice
- maintainers reviewing output contracts
- future contributors building SARIF, GitHub Action, or report surfaces on top of a stable JSON result

## Scope

### In scope

- `--format text|json` on `charter doctor`
- stable JSON output for the currently implemented rules only
- explicit machine-readable top-level result contract
- invalid format handling
- preservation of current text output behavior as the default

### Out of scope

- new rules
- SARIF output
- Markdown output
- config or profile loading
- fix planning or apply flows
- API or MCP surface
- text output redesign beyond alignment needed to keep text and JSON semantically consistent

## Why this slice

Slice 1 proved that Charter can resolve a repo, scan it, score it, and print text. Slice 2 should stabilize the same information as a machine-readable contract before more rule surface is added.

That gives the repo a better foundation for:

- CI/reporting integrations
- future GitHub Action evolution
- SARIF and report renderers later
- regression testing on output shape

## Command Contract

`charter doctor` should support:

- `--path <repo>`
- `--threshold <n>`
- `--quiet`
- `--format text|json`

Defaults:

- `--format text`
- invalid format returns exit code `2`

### `--quiet` with JSON

For Slice 2, `--quiet` should be ignored for JSON output.

Reason:

- machine-readable output should always emit the full payload
- this avoids dual semantics for automation consumers
- text mode keeps the current quiet behavior

## JSON Result Contract

Top-level shape:

```json
{
  "repo_path": "D:/Projects/charter",
  "threshold": 80,
  "passed": true,
  "findings": [
    {
      "rule_id": "AE-CTX-001",
      "severity": "BLOCKER",
      "category": "Context",
      "summary": "Agent context file is missing",
      "remediation": "Create AGENTS.md with the required sections",
      "evidence": [
        "no supported root context file found"
      ]
    }
  ],
  "summary": {
    "blocker": 0,
    "high": 0,
    "medium": 0,
    "low": 0
  },
  "score": {
    "base": 100,
    "final": 100
  }
}
```

Field rules:

- `repo_path`: resolved repo root used for the scan
- `threshold`: effective threshold used for pass/fail
- `passed`: boolean derived from `score.final >= threshold`
- `findings`: all findings in deterministic order
- `summary`: severity counts only
- `score.base`: uncapped score from severity penalties
- `score.final`: capped score after rule-cap logic

Do not add future-facing fields in this slice.

### Result ownership

For Slice 2, it is acceptable for `doctor.Result` to grow to include the JSON-facing fields that are already first-class command facts:

- `Threshold`
- `Passed`
- score summary counts if they are not already exposed by the scoring package

Do not add fields unrelated to the current five rules.

## Ordering Contract

JSON output should be deterministic across runs for the same repo state.

Recommended ordering:

- findings sorted by severity weight descending (`BLOCKER`, `HIGH`, `MEDIUM`, `LOW`)
- then by `rule_id`

Reason:

- stable snapshots in tests
- stable consumption in CI and later action/reporting work

## Architecture

Add a narrow rendering seam, not a general renderer framework.

### Recommended ownership

- `internal/doctor/`
  - authoritative scan result structure
  - command-facing assembly of result data

- `internal/render/text/`
  - optional extraction target if the existing command rendering becomes too large

- `internal/render/json/`
  - JSON serialization only

- `cmd/charter/doctor.go`
  - choose renderer based on `--format`
  - preserve exit behavior contract

### Avoid

- generic renderer registries
- plugin-like renderer abstractions
- shared renderer “engines” with no proven second format need beyond text/json

## Text Output Compatibility

Text remains the default.

Slice 2 must not change:

- current quiet text pass behavior
- quiet text fail summary line
- invalid invocation exit code `2`
- threshold fail exit code `1`

The addition of JSON should not regress the Slice 1 text path.

## Testing Strategy

Use TDD and fixture-first proof where appropriate.

### Required tests

- one failing command-level JSON test first
- one passing JSON structure test
- one invalid `--format` test
- one text regression test to ensure default output still works
- one quiet+json test proving full JSON still emits

### Validation strategy

- decode JSON into a struct in tests instead of string-matching entire blobs where possible
- also snapshot exact output for one representative scenario to detect accidental shape drift

## Risks

### Output drift

Risk:

- text and JSON diverge semantically

Control:

- both should consume the same `doctor.Result`

### Over-abstraction

Risk:

- generic render framework lands too early

Control:

- only text + JSON in this slice

### Future incompatibility

Risk:

- unstable field names break future CI/report consumers

Control:

- freeze exact field names now and test them

## Success Criteria

Slice 2 succeeds when:

- `charter doctor --format json` works end-to-end
- JSON output is stable and documented
- current five rules serialize correctly
- text mode behavior remains intact
- invalid format returns exit `2`
- repo quality gate remains green

## References

- `README.md`
- `AGENTS.md`
- `TESTING.md`
- `docs/internal/superpowers/plans/2026-05-28-phase-1-slice-1.md`
- `docs/internal/architecture/charter-architecture-2026.md`
- `docs/internal/specs/AE-CTX-001.md`
- `docs/internal/specs/AE-CTX-002.md`
- `docs/internal/specs/AE-CTX-004.md`
- `docs/internal/specs/AE-ENV-001.md`
- `docs/internal/specs/AE-CI-002.md`
- `docs/internal/superpowers/checklists/2026-05-28-cli-quality-principles.md`
