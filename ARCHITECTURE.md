# Architecture

## Module Contract

- Single root Go module: `go.use-charter.dev/charter`
- No `go.work`
- No additional Go modules
- Charter core performs no LLM calls
- Scanner remains deterministic and offline-first
- Fixes are diff-first and never silently mutate files
- Tracked MCP server config stays absent until a pinned, reviewed, least-privilege integration exists (distinct from the AE-MCP-* rules, which scan MCP config when present)

## Root Layout

- Root contract docs stay at repo root for contributor and agent entry.
- Repo-internal engineering docs live under `docs/internal/`.
- Customer-facing docs live under `docs/product/` (the Mintlify documentation site).

- `cmd/`: binary entrypoints and command wiring (includes `cmd/charter/init.go`, the `charter init` scaffold command, `cmd/charter/fix.go`, the `charter fix` repair command, and `cmd/charter/version.go`, the `charter version` command)
- `internal/`: non-public implementation details
  - `agentcontext`: canonical agent-visible context file registry (drift guard for context and secret rules)
  - `config`: `charter.yaml` loader (MCP trusted-remote allowlist; policy profile/threshold resolution)
  - `catalog`: embedded founder-curated MCP server catalog (versions, advisories, trusted hosts) backing the AE-MCP-* rules
  - `doctor`: scan orchestration pipeline (resolves the effective threshold)
  - `explain`: rule-explanation surface reused by the CLI, TUI, and HTML report
  - `findings`: finding model with Location support (path:line)
  - `fix`: diff-first repair engine behind `charter fix` — a pure registry + `Plan` (RuleID→fixer for `AE-CTX-001` AGENTS.md, `AE-CTX-004` .gitignore, `AE-CI-002` Charter CI workflow) + unified-diff builder, with a backup-then-write applier (backs up any existing target to `.charter/backups/<ts>/` before writing; never deletes/truncates, never overwrites a Create target, never fixes secret/dangerous rules); reuses `internal/scaffold` for file contents
  - `repository`: repo resolution and file inventory
  - `rules/`: rule implementations (context, environment, ci, secrets, mcp, agentconfig, operability, governance) and `catalog` (static rule metadata for SARIF/explain)
  - `scaffold`: pure offline detection (language/CI/agents) + agent-context templates + a create/skip file plan behind `charter init` (create-missing-only, never overwrites or deletes; imported by `cmd/charter/init.go`)
  - `scoring`: score calculation and caps (skips informational findings)
  - `render/`: output formatters (styled/plain text, JSON, Markdown, SARIF 2.1.0, self-contained HTML)
  - `secrets`: secret pattern detection and redaction
  - `suppress`: suppression loading (`.charter-suppress.yml` + inline `charter:ignore`) and the active/suppressed partition
  - `version`: build version with `Commit()`/`Date()` build-stamp accessors (`runtime/debug` + ldflags fallback) for SARIF `tool.driver.version` and the `charter version` command
  - `perf`: build-tagged (`//go:build perf`) performance validation — synthesizes a ~50,000-file repo at test time and asserts `charter doctor` ≤ 2 s wall-clock / (Linux) ≤ 256 MiB peak RSS; run via the Moon `:perf` task, kept out of the default `:test`
  - `terminal`: capability detection + palette tiers for the styled TTY path
  - `tui`: interactive `charter doctor -i` Bubble Tea surface
- `api/openapi/`: future API contracts before implementation
- `schemas/`: machine-readable config and report contracts (includes `doctor-result.schema.json`, `charter-config.schema.json`)
- `docs/internal/specs/`: rule-level behavior contracts
- `docs/internal/decisions/`: accepted ADRs
- `docs/internal/rfcs/`: proposals for cross-cutting or risky changes
- `evals/`: acceptance fixtures for prompt, workflow, and future agent behavior
- `.goreleaser.yaml`: GoReleaser v2 release build (multi-platform binaries, cosign keyless bundle signing, syft SPDX-2.3 SBOMs, Homebrew cask)
- `.github/workflows/release.yml`: tag-triggered release workflow (GoReleaser + SLSA Build L3 provenance via slsa-github-generator)
- `action/`: composite GitHub Action (source of truth; seeded to `use-charter/charter-action@v1` at launch) — downloads the signed binary, verifies it (cosign keyless + sha256), runs `charter doctor --format sarif`, and uploads via `github/codeql-action/upload-sarif@v4`. Because zizmor/actionlint do not scan `action.yml`, `scripts/ci/validate-action.ts` (Bun TS) asserts the composite structure + SHA-pinned `uses:` and runs in `:check` (`action:validate`).

Release rails are mise-pinned (GoReleaser, cosign, syft) and Moon-driven via the `release`, `release-snapshot`, and `release-check` tasks (`release-check` runs in `:check`).

## Command Tree Intent

The expected CLI tree follows the product authority in `docs/internal/architecture/charter-architecture-2026.md`:

- `charter init`
- `charter doctor`
- `charter explain`
- `charter report`
- `charter fix`
- `charter suppress`
- `charter version`

Package boundaries are drawn so each command owns a vertical slice without forcing module churn elsewhere. Cobra (via Fang) is the command framework; configuration is loaded through the `internal/config` package.

## Vertical-Slice Rule

A user-facing capability keeps its parser, domain model, verification logic, and remediation logic close together instead of spreading shared behavior into generic utility packages. Shared code is introduced only when at least two slices need the same stable abstraction.

Boundary rules:

- keep external packages private by default under `internal/`
- introduce a public Go package only when a stable external integration surface is proven
- avoid catch-all `util`, `helpers`, or `common` packages
- cross-cutting behavior that changes package boundaries requires an ADR or RFC first

## Error Contract

Agent-readable errors must expose:

- stable error code
- short machine-readable summary
- human remediation guidance
- evidence fields that explain the failure
- location data: structured finding locations (path:line) per ADR-0009, with 1-based line numbers (0 = file-level absence findings)
- score cap: hard ceiling applied when finding is present
- no secret-bearing payloads

Example Finding shape:

```text
RuleID: AE-CTX-001
Severity: BLOCKER
Summary: Missing agent context file
Remediation: Create AGENTS.md with project state, tech stack, edit boundaries, and verification command
Evidence: ["no agent-visible context file found in repo root"]
Locations: [] (empty for absence findings)
Cap: 0
```

Another example with locations:

```text
RuleID: AE-SEC-001
Severity: BLOCKER
Summary: Secret detected in agent context file
Remediation: Remove the literal secret, rotate it externally, use an environment variable reference instead
Evidence: ["sk-p… (redacted token)"]
Locations: [{Path: "AGENTS.md", Line: 14}]
Cap: 49
```

Every user-facing rule or command error should be backed by pass/fail examples in `docs/internal/specs/`.
