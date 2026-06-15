# internal/

Charter's implementation packages. Everything here is private to the module
(`go.use-charter.dev/charter`) — nothing under `internal/` is a public API, which
keeps the import surface free to evolve. The CLI in [`cmd/charter/`](../cmd/charter/)
wires these packages together; it holds no business logic of its own.

## The pipeline

A scan flows through these packages in order:

```
repository → rules/* → findings → suppress → scoring → render/*
   (inventory)  (checks)  (model)   (waivers)  (0–100)   (output)
```

`doctor` orchestrates the whole run; the renderers, `fix`, and `tui` consume its
`Result`.

## Packages

| Package | Role |
|---------|------|
| `repository` | Resolves the git root and builds the tracked-file inventory; gates all content reads (10 MiB cap, symlink containment, inventory-only). |
| `agentcontext` | Single source of truth for agent-visible context file types (`AGENTS.md`, `CLAUDE.md`, `.cursor/rules`, …) so context and secret rules never disagree. |
| `config` | Loads `charter.yaml` (MCP trusted remotes, policy profile/threshold) and resolves the effective passing threshold (flag > explicit > profile > default). |
| `catalog` | Embeds the founder-curated MCP server catalog (versions, advisories, trusted hosts); powers the MCP rules. |
| `doctor` | Orchestrates the scan→score→report run: executes rules, applies suppressions, emits governance findings, computes the score, applies the policy threshold. |
| `findings` | The `Finding` model (rule ID, severity, category, summary, remediation, evidence, locations, cap) and the severity weights. |
| `rules/*` | One subpackage per rule family — `context`, `secrets`, `mcp`, `agentconfig`, `environment`, `ci`, `operability`, `governance` — plus `catalog` (static per-rule metadata for SARIF/`explain`). |
| `suppress` | Loads `.charter-suppress.yml` and inline `charter:ignore` directives; partitions findings into active/suppressed with expiry, approver, and scope validation. |
| `scoring` | Computes the final score (deduct Blocker 20 / High 10 / Medium 4 / Low 1), applies hard caps, and rolls up per-category readiness. |
| `render/*` | Output formatters: `text` (styled TTY + plain), `json`, `markdown`, `sarif` (2.1.0), and `html` (self-contained, fonts/CSS/JS inlined). |
| `fix` | Diff-first repair engine behind `charter fix`: plans changes, renders unified-diff previews, and applies them behind backup + repo-containment + registered-fixer gates. Never touches secret/dangerous rules. |
| `scaffold` | Pure offline engine behind `charter init`: detects languages/CI/agent surfaces, generates context-file templates, and computes a create-or-skip plan (no disk writes of its own). |
| `explain` | Thin projection over `rules/catalog` for `charter explain <RULE>` (JSON or styled/plain text). |
| `terminal` | Offline capability detection (color tier, hyperlinks) and the semantic palette that maps design tokens to WCAG-AA styles. |
| `tui` | Bubble Tea master-detail browser for `charter doctor -i` (filterable finding list, detail pane, in-place rescan). |
| `perf` | Build-tagged (`//go:build perf`) performance assertion: a synthesized ~50k-file repo must scan in ≤ 2 s / ≤ 256 MiB. Run via `moon run :perf`, kept out of the default test run. |
| `version` | Build-stamp accessors (commit, date) for `charter version` and SARIF `tool.driver.version`. |

## Conventions

- **Offline and deterministic.** No network, no LLM, no clock reads outside injected timestamps — same repo, same output.
- **Safe I/O.** All file content is read through `repository`'s gated reader; renderers, `fix.Plan`, and `scaffold.Detect` are pure (no hidden state or writes).
- **Fail fast.** Errors stop the run; no silent degradation.
- **No catch-all packages.** No `util`/`helpers`/`common`. Shared code appears only when two callers need the same stable abstraction (see [`ARCHITECTURE.md`](../ARCHITECTURE.md)).

Rule behavior contracts live in [`docs/internal/specs/`](../docs/internal/specs/);
the canonical product behavior is [`docs/internal/architecture/charter-architecture-2026.md`](../docs/internal/architecture/charter-architecture-2026.md).
