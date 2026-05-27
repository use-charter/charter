# Phase 0 Closure Ledger

## Gate Status

- Gate 0: Complete
- Gate 1: Complete
- Gate 2: Complete
- Gate 3: Not started
- Gate 4: Not started

## Hard Rules

- No gate may be marked complete without repo evidence.
- No score increase may be claimed without a linked evidence entry.
- No contradiction may be deferred silently.
- Any drift found must be either fixed in-gate or explicitly block the gate.

## Evidence Index

- Pending

## Gate 0 Contradiction Matrix

| Concern | Files | Current conflict | Canonical owner | Resolution status |
|---|---|---|---|---|
| Bootstrap MCP policy | `SECURITY.md`, `.github/copilot-instructions.md`, `docs/architecture/charter-architecture-2026.md` | Bootstrap docs forbid `.mcp.json`; architecture examples create it | `docs/architecture/charter-architecture-2026.md` after reconciliation | Resolved |
| Documentation authority ladder | root docs, architecture docs, audit docs, HTML mirrors | authority was implied, not explicit everywhere | `docs/architecture/README.md` plus canonical product markdown | Resolved |

## Gate 0 Evidence

- Canonical bootstrap MCP rule moved into `docs/architecture/charter-architecture-2026.md` and aligned across root trust surfaces.
- Documentation authority ladder added and cross-referenced from root docs and audit companion surfaces.
- HTML mirror role remains presentation-only and subordinate to markdown.

## Gate 1 Task Path Matrix

| Surface | Setup path | Verify path | Notes |
|---|---|---|---|
| Local root | `mise install` | `mise x -- moon run :check` | Verified green on Windows worktree |
| Hooks | `./scripts/install-hooks.sh` | hk-managed `moon run` tasks | Same root task family preserved |
| CI | `jdx/mise-action` | workflow `moon run` tasks | Still rooted in repo-wide Moon tasks |
| Project task surfaces | project `moon.yml` files | `mise x -- moon run cmd:test cmd:build docs:lint web:lint` | Verified green after script normalization |

## Gate 1 Evidence

- `.miserc.toml` no longer breaks Windows hook/commit parsing.
- Root Moon tasks no longer depend on shell patterns that fail on Windows (`true`, `test -f`, `mkdir -p`, wildcard positional arguments).
- Helper scripts are tracked and used consistently by root and project Moon surfaces.
- `mise x -- moon run :check` passes from the isolated Windows worktree.

## Gate 2 Drift Ledger

| File | Section | Drift type | Resolution |
|---|---|---|---|
| `docs/architecture/charter-architecture-2026.md` | tool/version references | stale canonical content | Resolved |
| `docs/architecture/charter-architecture-2026.html` | init scaffold and tool/version references | stale mirror content | Resolved |
| `docs/audit/charter-v1-audit-checklist.md` | version examples and CI rule references | stale canonical content | Resolved |
| `docs/audit/charter-v1-audit-checklist.html` | source-of-truth banner and CI version references | stale mirror content | Resolved |

## Gate 2 Evidence

- Markdown was reconciled before HTML where stale versions existed.
- HTML init scaffold no longer reintroduces the bootstrap `.mcp.json` contradiction.
- Architecture and audit mirrors now reflect current pinned versions for Bun, Moon, hk, gofumpt, zizmor, and OSV-Scanner.
- Mirror-only role is explicit in both architecture and audit presentation surfaces.
