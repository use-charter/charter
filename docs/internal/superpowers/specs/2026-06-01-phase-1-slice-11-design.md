# Phase 1 Slice 11 Design — `charter init`

## Goal

Ship `charter init` (architecture M1.4 T1.4.1): deterministically scaffold the missing agent-context files so a blank repo goes zero → a passing Charter scan, in seconds, offline, without ever overwriting or deleting a user file. Implements ADR-0019. Includes a directly-related `AE-CTX-002` false-positive fix (align its verification marker with `AE-CTX-001`).

## Audience

- coding agents implementing this slice
- maintainers reviewing the scaffold model (create-missing-only, templates, detection) and the CTX-002 alignment
- future contributors building `charter fix` (Slice 12, which owns modifying existing files)

## Scope

### In scope

- `internal/scaffold`: pure detection (language/CI/agents) + template builders + a create/skip file plan; unit-tested
- `cmd/charter/init.go` + `root.go` registration: flags, `--dry-run`, create-missing, reporting; CLI tests
- `AE-CTX-002` fix: shared verification-signal helper; CTX-002 accepts CTX-001's set; tests + spec/audit note
- `init`→`doctor` integration test on a synthesized blank Go repo (assert score ≥ 80)
- docs sync (AGENTS/README/ARCHITECTURE), architecture §1.8/T1.4.1 reconcile, roadmap

### Out of scope

- modifying existing files (e.g. appending to a present `.gitignore`, `mise.toml` for AE-ENV-001, MCP pinning) — `charter fix` (Slice 12)
- `.cursor/rules` scaffolding (AGENTS.md is the universal context); interactive prompts; `.env.example` codebase env-scanning (minimal placeholder only)
- the GitHub Action / release (Slices 9/10); cutting a tag

## Why this slice

`charter init` is the architecture's named fix for the "one-time setup trap" (a repo passes once, the user moves on / never adopts). It is pure-Go, deterministic, offline — no release dependency — so it lands and verifies fully now. It also surfaces a real cross-vendor FP in `AE-CTX-002` worth fixing while we're here.

## Grounding (verified live 2026-06-01, not from memory)

- **Rule gates (from the code):** `AE-CTX-001` (`isMeaningfulContext`) needs ≤ 600 est. tokens, ≥ 5 non-empty lines, and signals for project-summary, tech-stack, edit-boundaries, and verification-command. `AE-CC-002` needs a concrete off-limits token (`.env`/`secrets`/`.github/workflows`/`terraform`/`infra`/`db/migrations`/`credentials`/`permissions.md`). `AE-CTX-004` needs `.gitignore` agent-artifact patterns. `AE-CTX-002` (`missingRepoTruthMarkers`) currently requires literal `moon run :check`, `.env*`, `secrets/` (+ `hk.pkl`/arch-doc when present).
- **AGENTS.md** (agents.md / Linux Foundation): freeform Markdown, universal across agents (Cursor/Copilot/Claude read it); sections Project Overview, Tech Stack (versioned), Commands, Boundaries.
- **.claude/settings.json** (2026): `{$schema, permissions:{allow,ask,deny}}`; `deny: ["Read(./.env)","Read(./.env.*)","Read(./secrets/**)"]`; MCP separate (none).
- **Seams:** `cmd/charter/suppress.go` is the command+write model (cobra, `repository.ResolveRoot`, `os.WriteFile 0o644`, `commandExitError`); `internal/config` owns `charter.yaml`; `internal/rules/context` owns CTX-001/002/004; `repository.ResolveRoot(path)` resolves the root (walks to `.git`).

## `internal/scaffold` (new, pure)

- **Detection** (`Detect(root) Project`): `Project{ Languages []Lang (name+version), CI string, Agents []string }` from tracked manifests — `go.mod` (Go + `go`/`toolchain` version), `package.json` (JS/Bun), `pyproject.toml` (Python), `Cargo.toml` (Rust); CI from `.github/workflows`; agents from `.claude`/`.cursor`/copilot. Pure, deterministic, offline.
- **Templates** (pure builders returning `[]byte`): `AGENTS.md`, `charter.yaml`, `.gitignore`, `ARCHITECTURE.md`, `.env.example`, `.claude/settings.json`. The `AGENTS.md` builder takes the detected `Project` + profile and emits ≤ 600 tokens with all required signals (see below).
- **File plan** (`Plan(root, opts) []FileAction`): each candidate file → `{Path, Action: create|skip, Contents}` where `Action=skip` iff the file already exists. Pure given the on-disk presence set; the command applies it.

### Generated `AGENTS.md` (must pass CTX-001 + CC-002 + CTX-002)

```markdown
# AGENTS.md

## Project Overview

<project-name> is a <detected-language> project. <one-line purpose placeholder the user edits>.

## Tech Stack

- <Go 1.26.x> (detected)            # one bullet per detected language+version
- CI: GitHub Actions                # when detected

## Commands

- Verify: `charter doctor`
- Build:  `go build ./...`          # best-effort per detected language
- Test:   `go test ./...`

## Edit Boundaries

- Safe for agents: application source, tests, docs
- Off-limits: `.github/workflows/`, `.env*`, `secrets/`, `db/migrations/`, `terraform/`

## Verification

- Run `charter doctor` before committing.
```

Satisfies: project-summary (Overview prose), tech-stack (`Go`/detected), edit-boundaries (`Off-limits`/`Safe for agents`), verification (`charter doctor`); off-limits tokens for CC-002 (`.env*`, `secrets/`, `.github/workflows/`); CTX-002 markers `.env*` + `secrets/` + a verification signal (after the CTX-002 fix). Comfortably < 600 tokens.

### Other templates

- `charter.yaml`: `policy:\n  profile: <profile>\n` (schema-valid).
- `.gitignore` (create-if-absent): a `# Charter / agent artifacts` block — as built includes `.hk/`, `.env*` (+ `!.env.example`), `.charter/`, `*.charter-session`, `.claude/local/`, `.cursor/cache/` — enough agent-artifact patterns to fully clear `AE-CTX-004`.
- `.claude/settings.json` (when claude): `{ "$schema": "https://json.schemastore.org/claude-code-settings.json", "permissions": { "deny": ["Read(./.env)","Read(./.env.*)","Read(./secrets/**)"] } }`.
- `ARCHITECTURE.md`, `.env.example`: minimal, useful templates.

## `cmd/charter/init.go`

- `newInitCommand()` → `init` (registered in `root.go`). Flags: `--path`, `--profile` (validate strict|standard|relaxed), `--agents` (csv), `--dry-run`, `--yes`.
- Resolve target (`--path` or CWD; tolerate a non-git dir by using the path directly). `scaffold.Detect` → `scaffold.Plan`. For each `FileAction`: `create` → ensure parent dir, `os.WriteFile(0o644)` only if still absent (re-check to avoid TOCTOU overwrite); `skip` → report. `--dry-run` → print the plan, write nothing.
- Report: `created N · skipped M`, then `› Next: charter doctor`. Exit 0 (incl. all-skipped); exit 2 on IO/flag error (`commandExitError`).
- **Never** `os.Remove`/truncate; **never** write over an existing file.

## `AE-CTX-002` fix (`internal/rules/context`)

- Extract the verification-signal check into a shared helper (e.g. `hasVerificationSignal(lower string) bool`) covering `moon run :check`, `charter doctor`, `verification command`, `verify with`. `AE-CTX-001` (`missingContextSignals`) and `AE-CTX-002` (`missingRepoTruthMarkers`) both use it.
- `AE-CTX-002` no longer hardcodes `moon run :check`; it requires *a* verification signal (any of the set), plus the unchanged `.env*` + `secrets/` + conditional `hk.pkl`/arch-doc markers. Charter's own `AGENTS.md` (has `moon run :check`) still passes; a `charter doctor`-only repo now also passes.
- Update `docs/internal/specs/AE-CTX-002.md` + the audit checklist AE-CTX-002 note to describe "a recognized verification command" rather than implying `moon run :check` specifically.

## Architecture / ownership

- `internal/scaffold/` (new) — pure detection + templates + plan; imported by `cmd/charter/init.go`. No rule-package imports.
- `cmd/charter/init.go` (new) + `root.go` (register).
- `internal/rules/context/` — shared verification-signal helper; CTX-002 generalization.
- Avoid: overwriting/deleting files; network/LLM; a generic template engine; `.cursor/rules` duplication; putting detection in the command (keep it in `scaffold`).

## Go alignment (per golang-patterns / golang-testing)

- `scaffold` is pure (templates are `func(Project, ...) []byte`; the plan is computed from an injected presence set so it's unit-testable without disk). The command owns the only disk writes; errors wrapped `%w`, fail fast, no `_` discards.
- Determinism: stable file ordering; templates are fixed given inputs; `AGENTS.md` token budget asserted in a test.
- Tests: table-driven detection; golden templates; `Plan` create/skip; CLI `--dry-run`/create/skip/`--profile`/`--agents`; the CTX-002 helper; `-race`; ≥ 85% for `scaffold`.

## Testing & verification strategy

- **scaffold unit:** detection per language/CI/agents; each template's required content (AGENTS.md has all CTX-001 signals + off-limits, ≤ 600 tokens; charter.yaml schema-valid; .claude/settings.json valid JSON, MCP-free, secret-denies); `Plan` marks existing files `skip`.
- **CTX-002 unit:** an AGENTS.md with `charter doctor` (no `moon run :check`) + `.env*` + `secrets/` does NOT fire CTX-002; one missing the verification signal still does; Charter's own repo still passes.
- **CLI:** `init` on a temp blank Go repo creates the set; re-running reports all `skip` (idempotent, nothing overwritten); `--dry-run` writes nothing; `--profile strict` writes `profile: strict`; invalid profile → exit 2.
- **Integration (the OOTB proof):** synthesize a blank Go repo (git-init, `go.mod`), run `init`, then `doctor` → assert **score ≥ 80** (measured **95** as built; residual AE-ENV-001/AE-CI-002 only, since the fixture lacks hooks/CI). Hermetic (`t.TempDir`).
- **Dogfood:** Charter's own `charter doctor` stays **100** (CTX-002 generalization keeps Charter passing; `init` is not run on Charter's already-populated repo — all-skip).
- `moon run :check` green.

## Risks

- **Generated AGENTS.md fails a gate** — control: golden + signal-assertion tests on the template; the integration test runs the real scan; token budget asserted < 600.
- **Accidental overwrite** — control: create-missing-only with a re-check before write; existing files skipped; `--dry-run`; no `Remove`/truncate anywhere; CLI test re-runs init and asserts all-skip.
- **CTX-002 generalization regresses Charter's own 100** — control: Charter's AGENTS.md still matches the shared signal set (`moon run :check`); test asserts Charter-style + charter-doctor-style both pass and a no-verification AGENTS.md still fails.
- **Detection wrong/empty** — control: Go is robust; others best-effort with safe generic fallbacks; AGENTS.md still passes with a generic stack line; non-git target tolerated.

## Success criteria

- `charter init` on a blank Go repo creates `AGENTS.md`/`charter.yaml`/`.gitignore`/`ARCHITECTURE.md`/`.env.example` (+ `.claude/settings.json` for claude), reports created/skipped, and a subsequent `charter doctor` scores **≥ 80** (integration-tested); re-run is all-skip (idempotent, zero overwrites); `--dry-run` writes nothing.
- `AE-CTX-002` no longer false-positives on a `charter doctor`-only AGENTS.md; Charter's own scan stays 100; spec/audit updated.
- ADR-0019 + this spec + the plan committed; AGENTS/README/ARCHITECTURE + architecture §1.8/T1.4.1 reflect `charter init`; `moon run :check` green.

## References

- `docs/internal/decisions/0019-charter-init-scaffold.md` (ADR-0019); `0014` (catalog), `0008` (score), `0005` (diff-first)
- `docs/internal/architecture/charter-architecture-2026.md` (§1.8 `charter init`; M1.4 T1.4.1)
- rule code: `internal/rules/context/ctx001.go`/`ctx002.go`, `internal/rules/agentconfig/cc002.go`, `internal/agentcontext`
- agents.md (Linux Foundation) standard; Claude Code `settings.json` schema (2026)
- `docs/internal/superpowers/plans/2026-06-01-phase-1-slice-11.md` (derived implementation plan)
