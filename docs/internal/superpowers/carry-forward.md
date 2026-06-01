# Carry-Forward Ledger

Last reviewed: 2026-06-02

Durable record of items deliberately **deferred** during slice execution, so they're followed up when their trigger arrives. This is the cross-slice hygiene/debt list — distinct from:

- the **launch roadmap** (`2026-06-01-v1-launch-roadmap.md`) — the slice sequence and the Slice 17 production-release checklist (tap→public, seed `use-charter/charter-action`, first signed tag, branch protection, etc.);
- the roadmap's **"Deferred to Phase 1.5 / v1.1"** note (`charter serve`, `--format toon|json-compact`, `--for-agent`, standalone `charter report --format spdx`, AE-SEC-001 → full Gitleaks ruleset, deep multi-agent conflict detection).

When an item is resolved, strike it (or move it to a "Done" note) in the slice that closes it.

## Open items

| # | Item | Why deferred | Trigger / target | Source |
|---|---|---|---|---|
| CF-1 | §1.8 Command Gallery mockups still show illustrative `AE-MCP-001`/`AE-ENV-001` `charter fix` examples, not the shipped v1 fixer set (`AE-CTX-001`/`AE-CTX-004`/`AE-CI-002`). Authoritative set is the T1.4.2 "As built" line. | Targeted doc edits only during the slice; rewriting the stylized gallery was out of scope. | Slice 14 (full codebase/doc review) | Slice 12 |
| CF-2 | Internal HTML doc mirrors (`docs/internal/**/*.html`, e.g. `charter-architecture-2026.html`, the audit checklist `.html`) lag their `.md` after Slices 9–12 edits. | "HTML mirror-only" convention (presentation, ungated); regenerating mid-slice is risky without the gen process. | Slice 14 (review) or Slice 15 (docs) | Slices 10, 12 |
| CF-3 | `AGENTS.md` sits at ~595/600 estimated tokens; AE-CTX-001 fails at >600, so each slice touching it must trim. | Self-imposed budget; no per-rule budget override implemented. | Any slice editing AGENTS.md; consider implementing `charter.yaml → rules.AE-CTX-001.token_budget` (referenced in the audit checklist, not yet built) | Slices 10, 11, 12 |
| CF-4 | `go install go.use-charter.dev/charter/cmd/charter@…` won't resolve until `go.use-charter.dev` serves a `go-import` meta tag (vanity path is decoupled from the GitHub owner by design). | Needs the web host; binaries + Homebrew cover install meanwhile. | Slice 15/16 (web) + go-public ops | Slice 9 |
| CF-5 | `AE-ENV-001` has no `charter fix` fixer. | Its usual missing piece is an *opinionated* hook-config (not a pure file create); toolchain is normally already satisfied by `go.mod`/etc. | `charter fix` v1.1 (only if a non-opinionated default emerges) | Slice 12 |
| CF-6 | `AE-MCP-001` has no `charter fix` fixer (auto-pin `@latest`/range → exact version). | The correct pin target needs the MCP catalog. | After **Slice 13** (MCP Catalog) — add an `AE-MCP-001` fixer that pins to the catalog's stable version | Slice 12 |
| CF-7 | `charter version` has no `--format json`/`--short`; the GitHub Action has no `score` output. | Minimal v1 surface; not required by current consumers. | Phase 1.5 / when a consumer needs them | Slices 9, 10 |
| CF-8 | First-party exemptions in `AE-CI-002` (the `use-charter/charter-action@<tag>` and `slsa-github-generator@<tag>` skips) accept tag pins. | Tag-pinning is the conventional/required form for these; SHA-pinning the SLSA generator is unsupported. | Revisit only if a stricter first-party-pin policy is wanted post-launch | Slices 9, 12 |

## Done

- (none yet)
