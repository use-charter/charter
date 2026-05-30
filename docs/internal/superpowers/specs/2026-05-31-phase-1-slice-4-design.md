# Phase 1 Slice 4 Design

## Goal

Give every `findings.Finding` a structured physical location (file + line) and emit it through the existing text and JSON output, and remove the duplicated agent-context-file list shared by the context and secret rules. This completes the finding model toward the documented v1 output contract and unblocks the SARIF renderer and fix engine, without yet building SARIF, `charter fix`, or new rules.

This slice implements ADR-0009 (Structured Finding Location) and resolves review findings F2 (no structured location) and M-3 (duplicated context-file list).

## Audience

- coding agents implementing the next vertical slice
- maintainers reviewing the output contract before SARIF (M1.5) and fix (M1.4) build on it
- future contributors writing the SARIF renderer and `charter fix`

## Scope

### In scope

- a canonical agent-context-file source consumed by both `internal/rules/context` and `internal/rules/secrets` (M-3)
- a `Location` type and `Locations []Location` field on `findings.Finding` (F2 / ADR-0009)
- line-number capture in the shared secret detector path and population of `Locations` in `AE-SEC-001` and `AE-SEC-002`
- population of `Locations` in the other rules where a physical site exists (`AE-CI-002`; context rules where a single file is implicated)
- `locations` in the JSON output contract and a versioned schema file under `schemas/`
- `path:line` display in the text renderer
- spec and architecture-doc alignment for the rules whose output shape changes

### Out of scope

- SARIF output (`--format sarif`) — M1.5, consumes this contract
- Markdown output — M1.1 T1.1.4 follow-up
- `charter fix` / `auto_fix` / `guidance` output fields — M1.4
- new rules (MCP, CC, SUPPRESS) — M1.2 / M1.3
- config or profile loading
- changing scoring behavior

## Why this slice

The architecture doc describes Charter as a "SARIF 2.1.0 scanner" and shows `file:line` in every documented output mode, but the implementation never captures line numbers and stores location text inside `evidence`. SARIF (M1.5) and `charter fix` (M1.4) both require a structured location. Locking the location contract now — while `schemas/` is still empty and no external consumer depends on the v1 JSON shape — is the cheapest correct time, and aligns the implementation with the already-documented contract rather than introducing a new one.

M-3 rides in the same slice because the secret rules and context rules must agree on the canonical agent-context-file set; fixing the duplication here prevents a new context file type being recognized for context quality but silently skipped by secret scanning.

## Precondition

`internal/secrets/` and `internal/rules/secrets/` are read/write blocked for the AI agent by the `fieldnation` org Cursor blocklist pattern `**/secrets/**`. Two tasks below edit those packages. Before implementation, either:

- the `fieldnation` Cursor admin narrows or removes the `**/secrets/**` pattern (preferred — also fixes the pattern wrongly applying to personal repos), or
- the secret packages are temporarily renamed to an unblocked path for the duration of the edit and renamed back afterward.

## Location Contract

```go
package findings

type Location struct {
    Path string // repo-relative, forward-slash
    Line int    // 1-based; 0 means file-level / no specific line
}

type Finding struct {
    // ...existing fields...
    Locations []Location
    Cap       int
}
```

JSON serialization (additive to the Slice 2 contract):

```json
{
  "rule_id": "AE-SEC-001",
  "severity": "BLOCKER",
  "category": "Secrets",
  "summary": "Secret detected in agent-visible context file",
  "remediation": "Remove the literal secret and reference an environment variable instead",
  "locations": [
    { "path": "AGENTS.md", "line": 14 }
  ],
  "evidence": [
    "AGENTS.md: sk-a…"
  ]
}
```

Field rules:

- `locations`: ordered array; empty for absence findings
- `path`: repo-relative, forward-slash
- `line`: 1-based; `0` rendered as file-level (no `:line` suffix in text)
- `locations` is shaped to map onto SARIF `result.locations[].physicalLocation` (`artifactLocation.uri` = `path`, `region.startLine` = `line`) so the M1.5 renderer is a thin projection
- `evidence` remains for human detail and is no longer the carrier of location data

## Per-rule location semantics

| Rule | Location populated | Line captured |
|---|---|---|
| AE-SEC-001 | file containing the secret | yes (matching line) |
| AE-SEC-002 | MCP/config file containing the secret | yes (matching line) |
| AE-CI-002 | implicated workflow file when one specific file is at fault | optional |
| AE-CTX-001 | the resolved context file when present-but-weak; none when absent | no (file-level) |
| AE-CTX-002 | the context file | no (file-level) |
| AE-CTX-004 | `.gitignore` when present | no (file-level) |
| AE-ENV-001 | none (absence finding) | n/a |

Findings with no physical site emit `"locations": []`.

## Canonical agent-context list (M-3)

Introduce a leaf package holding the single source of truth, consumed by both rule packages:

```go
package agentcontext

// Files are the single-file agent context candidates, in precedence order.
var Files = []string{
    "AGENTS.md", "CLAUDE.md", ".windsurfrules",
    ".github/copilot-instructions.md", "opencode.md",
    "codex.md", "DESIGN.md", "SKILL.md",
}

// CursorRulesDir is the directory-tree context source handled separately
// from single-file candidates.
const CursorRulesDir = ".cursor/rules"
```

- `internal/rules/context` uses `Files` + `CursorRulesDir` for context detection (replacing `supportedContextFiles` + the inline `.cursor/rules` handling).
- `internal/rules/secrets` uses `Files` + `CursorRulesDir` for the secret-scan target set (replacing `agentVisibleFileTargets`).
- The package is a dependency-free leaf; no rule package imports another rule package.

## Architecture / ownership

- `internal/findings/` — owns `Location` and the `Locations` field; no behavior.
- `internal/agentcontext/` (new leaf) — canonical context-file set.
- `internal/secrets/` — detector returns the matched line; redaction unchanged.
- `internal/rules/*` — populate `Locations`; secrets and context consume `agentcontext`.
- `internal/render/json/` — add `locations` to `findingDTO`.
- text renderer — display `path:line` (or `path` when `line == 0`).
- `schemas/` — versioned result/finding schema including `locations`, linked to ADR-0009.

Avoid: a generic location framework, multi-region SARIF detail beyond `{path,line}`, or any renderer registry.

## Line capture approach

The secret rules already iterate `for _, line := range strings.Split(data, "\n")`. Switch to `for i, line := range ...` and record `i + 1` on a match. The shared detector returns the match; the rule owns the line index. No detector signature change is strictly required, but the rule must thread the 1-based index into the `Location`.

## Testing strategy

- decode JSON into a struct and assert `locations[0].path` / `locations[0].line` for a secret finding
- assert `locations` is empty for an absence finding (for example `AE-ENV-001` with no toolchain)
- table test for the text renderer: `path:line` when line set, `path` when `line == 0`
- a single canonical-list test proving a file added to `agentcontext.Files` is both recognized as context and scanned for secrets (closes the M-3 drift gap)
- snapshot one representative JSON payload to detect shape drift
- the repo quality gate (`moon run :check`) stays green

## Risks

- Output drift between text and JSON — control: both consume the same `findings.Finding`.
- Over-shaping the location type — control: `{path,line}` only; richer SARIF regions are projected at render time in M1.5, not stored.
- Schema/spec divergence — control: land the `schemas/` file and update the affected `AE-*` specs and the architecture-doc examples in the same slice.
- Blocked secret packages — control: the precondition above is satisfied before implementation starts.

## Success criteria

- `findings.Finding` carries structured `Locations`; absence findings emit `[]`.
- `charter doctor --format json` emits `locations` matching the schema; `jq '.findings[0].locations[0].line'` returns a number for a secret finding.
- text output shows `path:line` for located findings.
- a single `agentcontext` source feeds both context and secret rules; the drift test passes.
- ADR-0009 is referenced; a `schemas/` contract file exists.
- `moon run :check` is green and the dogfood `charter doctor` score is unchanged.

## References

- `docs/internal/decisions/0009-finding-location.md`
- `docs/internal/decisions/0004-contract-first-interfaces.md`
- `docs/internal/architecture/charter-architecture-2026.md` (§ output examples, M1.1 T1.1.4, M1.5 T1.5.1)
- `docs/internal/superpowers/specs/2026-05-29-phase-1-slice-2-design.md`
- `docs/internal/specs/AE-SEC-001.md`, `AE-SEC-002.md`, `AE-CI-002.md`
- `schemas/README.md`
