# Phase 1 Slice 6 Design

## Goal

Implement the two Agent Config Intelligence rules — `AE-CC-001` (dangerous shell commands in agent hook configs; Blocker; OWASP MCP05) and `AE-CC-002` (explicit agent edit-scope; High; OWASP MCP02) — and add a `--format markdown` renderer for PR-comment output. With AE-CC, the **12 core v1 rules are complete**. Implements ADR-0012; the Markdown renderer completes M1.1 T1.1.4.

This slice does not build policy profiles (deferred to its own slice), `charter fix`, SARIF, or the AE-SUPPRESS governance rules.

## Audience

- coding agents implementing this vertical slice
- maintainers reviewing the agent-config detection contract and the third output format
- future contributors extending hook-config coverage (Pkl/YAML/shell managers) and the operator-chaining detector

## Scope

### In scope

- a new `internal/rules/agentconfig` package: JSON hook-config discovery/parse + `AE-CC-001` (dangerous commands) and `AE-CC-002` (edit scope), wired into the doctor pipeline (error-propagating)
- a new `internal/render/markdown` package and `--format markdown` in `charter doctor`
- structured `path:line` (or file-level) locations on AE-CC findings (ADR-0009)
- fixtures under `testdata/repos/`, as-built alignment of the `AE-CC-001/002` specs, the audit checklist rows, AGENTS.md/README, and the architecture-doc rule list

### Out of scope

- policy profiles (`charter.yaml → policy.profile`) — separate slice
- operator-chaining / command-substitution injection detection (`&&`, `;`, `$(…)`, backticks) — deferred (FP-prone)
- Pkl/YAML/shell hook managers (`hk.pkl`, `.pre-commit-config.yaml`, `lefthook.yml`, `.husky/`) — JSON hook configs only in v1
- `charter fix` (M1.4), `--format sarif` (M1.5), AE-SUPPRESS governance rules
- changing scoring behavior (AE-CC-001 uses the existing blocker cap; AE-CC-002 is High, no cap)

## Why this slice

AE-CC-001/002 are the last two core rules; landing them completes the 12-rule core and closes OWASP MCP05/MCP02 coverage. The Markdown renderer is small, additive, and reuses the finding model + `Locations` exactly like the JSON renderer, giving PR-comment output the same slice the new rules ship. Both reuse the JSON config-scan pattern (MCP) and the `agentcontext` registry.

## Grounding (verified live 2026-05-31, not from memory)

- **Claude Code hooks:** JSON in `.claude/settings.json`, `.claude/settings.local.json` (and plugin `hooks/hooks.json`). Shape: `hooks.<Event>[].hooks[].command` / `args` (`{type:"command", command, args}`); a flatter `hooks.<Event>[].command` form also appears. Events: PreToolUse, PostToolUse, Stop, Notification, …
- **Cursor hooks:** `.cursor/hooks.json` (`version:1`), shape `hooks.<event>[].command` (`beforeShellExecution`, `preToolUse`, `afterFileEdit`, …).
- **OWASP MCP Top 10 (beta):** MCP05 Command Injection & Execution (→ AE-CC-001); MCP02 Privilege Escalation via Scope Creep (→ AE-CC-002; the doc's "Permissioning Failures" is the older label).
- **CLI:** `cmd/charter/doctor.go` currently allows `--format text|json`; markdown must be added to the allowed set and dispatched. The JSON renderer (`internal/render/json`) sorts findings by severity-weight desc then rule_id asc — the markdown renderer reuses that ordering.

## Precondition

None blocking. This slice lives in new packages (`internal/rules/agentconfig`, `internal/render/markdown`) plus `internal/doctor/run.go` and `cmd/charter/doctor.go`. It does **not** touch the admin-blocked `internal/secrets/` or `internal/rules/secrets/` paths.

Cross-rule note: `AE-CTX-001` already requires a generic "edit boundaries" content signal for a context file to count as meaningful. `AE-CC-002` is deliberately stricter (concrete off-limits **paths**), so the two are complementary, not duplicative.

## AE-CC-001 — dangerous hook commands

- **Scan set:** `.claude/settings.json`, `.claude/settings.local.json`, `.cursor/hooks.json` (tracked only).
- **Command extraction:** parse JSON; locate the top-level `hooks` object; walk it recursively collecting every `command` (string) and `args` (string slice) value. This handles both the Claude nested and Cursor flat shapes without hard-coding event names.
- **Dangerous set (high-confidence, low-FP):** destructive — `rm -rf`, `git reset --hard`, `git clean -fd`, `dd `, `mkfs`, `truncate`; privilege escalation — `sudo `, `chmod 777`, `chown -R `. Match is case-insensitive, whitespace-normalized.
- **Deferred (not flagged in v1):** operator chaining / command substitution (`&&`, `||`, `;`, `$(…)`, backticks) — FP-prone; documented future refinement.
- **Finding:** Blocker, Category "Agent Config", `path:line` of the matched command, evidence quoting the offending command.

## AE-CC-002 — explicit edit scope

- **Inputs:** the `agentcontext` files + `PERMISSIONS.md` when present (tracked only).
- **Pass:** a concrete off-limits-path declaration exists — a recognized sensitive-path token presented as off-limits (`.env`, `secrets`, `.github/workflows`, `terraform`, `infra`, `db/migrations`, `credentials`) or a reference to `PERMISSIONS.md`.
- **Fail (High):** no concrete off-limits-path declaration in any context file.
- **Finding:** High, Category "Agent Config", file-level location (the context file evaluated), evidence listing the files checked. No finding when no context file exists at all (AE-CTX-001 owns that absence, Blocker).

## Markdown renderer

- `internal/render/markdown.Render(doctor.Result) ([]byte, error)`: a GitHub-PR-comment-friendly projection — a heading (`# Charter` + score/threshold/pass-fail), a severity summary line, and a findings table (Rule | Severity | Location | Summary) ordered like JSON (severity-weight desc, rule_id asc); a clean "no findings" body when empty. Reuses `findings.Finding` + `Locations`; no new data.
- `cmd/charter/doctor.go`: add `markdown` to the allowed `--format` values and dispatch to the renderer; exit-code semantics match the other formats.

## Architecture / ownership

- `internal/rules/agentconfig/` (new) — `hooks.go` (JSON hook discovery/parse → command list, pure), `cc001.go` (dangerous-command check), `cc002.go` (edit-scope check over agentcontext + PERMISSIONS.md), `agentconfig.go` (`Run(root, inv) ([]findings.Finding, error)` — discover, parse, aggregate, fail-fast). No rule package imports another rule package.
- `internal/render/markdown/` (new) — `Render(doctor.Result) ([]byte, error)`; depends only on `doctor` + `findings`, mirroring `internal/render/json`.
- `internal/doctor/run.go` — call `goagentconfig.Run` after the MCP rules, before secrets; propagate the error.
- `cmd/charter/doctor.go` — `--format text|json|markdown`.
- `internal/findings/`, `internal/scoring/` — unchanged (AE-CC-001 reuses the blocker cap; AE-CC-002 is High).

Avoid: a generic hook-config framework, YAML/Pkl parsing, operator-chaining heuristics, or a renderer registry in v1.

## Go alignment (per golang-patterns / golang-testing)

- Pure functions take data and return concrete types; the filesystem touch is isolated in `Run`, so the checks and the renderer unit-test without disk.
- Errors are values: `Run` and the JSON parse wrap with `%w` and fail fast; `run.go` propagates; nothing discarded with `_`.
- No package-level mutable state (the dangerous-pattern and sensitive-path sets are read-only).
- Deterministic output: each check sorts findings; the markdown renderer reuses the JSON severity/rule_id ordering.
- Testing: table-driven `t.Run` subtests, `t.Helper()`, `t.TempDir()`, `t.Parallel()` for independent units, a fuzz target on the hook-config JSON parser, `-race` via `moon run :test`, ≥85% line coverage for the new packages.

## Testing strategy

- unit: command extraction (Claude nested + Cursor flat shapes), the dangerous-command matcher (each pattern + safe negatives like `cd app && npm test`), the off-limits-path detector (pass with `.env`/`secrets`/PERMISSIONS.md ref; fail with no declaration), markdown rendering (findings table + empty case)
- parser fuzz: `FuzzParseHookConfig` (never panics on junk JSON)
- integration: fixture repos through `agentconfig.Run` (inline temp git repo, env001 pattern) — clean, dangerous-hook → AE-CC-001, no-edit-scope → AE-CC-002, malformed → error
- doctor-level: `makeTempGitRepoFromFixture` fixtures exercised through `Run`
- CLI: `--format markdown` smoke (golden-ish substring assertions)
- dogfood: Charter's own repo — `.claude/settings.json` hooks (if any) are safe; AGENTS.md declares off-limits paths (`.env*`, `secrets/`) and PERMISSIONS.md exists, so AE-CC-002 passes; score stays 100
- `moon run :check` stays green

## Risks

- **AE-CC-001 false positives / negatives** — control: high-confidence destructive/escalation set only; operator-chaining deferred; evidence quotes the exact command; Blocker severity reserved for unambiguous patterns.
- **AE-CC-002 over-firing on small repos** — control: documented FP risk; presence-based on concrete sensitive-path tokens or a PERMISSIONS.md reference; High (not Blocker).
- **Overlap with AE-CTX-001** — control: AE-CC-002 requires concrete off-limits paths, stricter than the generic edit-boundary signal.
- **Markdown injection / formatting** — control: render trusted finding fields only; no raw HTML; evidence already secret-redacted upstream.
- **Hook-config shape drift across clients** — control: recursive `command`/`args` walk rather than hard-coded event names.

## Success criteria

- `charter doctor` on a repo with a dangerous hook command emits `AE-CC-001` (Blocker, score ≤ 59) with a `path:line` location; a repo whose context declares no off-limits paths emits `AE-CC-002` (High).
- `charter doctor --format markdown` prints a PR-comment-ready report with the findings table and the score line; `--format` rejects unknown values with exit 2.
- the `AE-CC-001/002` specs are as-built; ADR-0012 is referenced; fixtures and the architecture-doc rule list reflect the 12 implemented rules.
- `moon run :check` is green and the dogfood `charter doctor` score is unchanged (100).

## References

- `docs/internal/decisions/0012-agent-config-scanning.md`
- `docs/internal/decisions/0009-finding-location.md`, `0011-mcp-config-scanning.md`, `0006-latest-docs-first.md`
- `docs/internal/architecture/charter-architecture-2026.md` (§ rule catalog, M1.1 T1.1.4, M1.2)
- `docs/internal/specs/AE-CC-001.md`, `AE-CC-002.md`
- `docs/internal/superpowers/specs/2026-05-31-phase-1-slice-5-design.md`
- `docs/internal/superpowers/plans/2026-05-31-phase-1-slice-6.md` (derived implementation plan)
