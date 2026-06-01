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
| CF-6 | `AE-MCP-001` has no `charter fix` fixer (auto-pin `@latest`/range → exact version; rewrite a deprecated package → its successor). | Slice 13 shipped the catalog (the pin/migration target now exists), but content-aware MCP config rewrites are a `fix`-engine change. | `charter fix` v1.1 — add an `AE-MCP-001` fixer using the catalog's `stableVersion`/`successor` | Slices 12, 13 |
| CF-7 | `charter version` has no `--format json`/`--short`; the GitHub Action has no `score` output. | Minimal v1 surface; not required by current consumers. | Phase 1.5 / when a consumer needs them | Slices 9, 10 |
| CF-8 | First-party exemptions in `AE-CI-002` (the `use-charter/charter-action@<tag>` and `slsa-github-generator@<tag>` skips) accept tag pins. | Tag-pinning is the conventional/required form for these; SHA-pinning the SLSA generator is unsupported. | Revisit only if a stricter first-party-pin policy is wanted post-launch | Slices 9, 12 |
| CF-9 | SARIF `helpUri`s point at `https://use-charter.dev/rules/AE-*`, which are **dead links until the rule docs pages exist**. | Pages live in the Mintlify docs site, not yet built. | **Hard launch dependency** — must be live by/with Slice 15 (docs) | Slice 8 |
| CF-10 | `use-charter` GitHub org hardening: verify the `use-charter.dev` domain (DNS TXT) on the org and require 2FA for org members. | Recommended during the org migration; admin-side. | go-public ops / Slice 17 checklist | Org migration |
| CF-11 | macOS release binaries are cosign-signed but **not Apple-notarized** (the Homebrew cask strips `com.apple.quarantine` as the interim). | Notarization needs an Apple Developer cert. | Post-launch, when an Apple Developer cert is available | Slice 9 |
| CF-12 | The MCP catalog (`internal/catalog/data/catalog.yaml`) ships as an accurate **seed**, not the validated final dataset: package names/successors/hosts are verified, but version data is partial and **no CVE advisories ship yet**. | Curation + the T1.6.3 FP gate are `⚑ FOUNDER` launch tasks; real CVE IDs only (ADR-0021). | **Hard launch gate (Slice 17)** — refresh to ~20–30 verified servers, add any real advisories, and complete `docs/internal/catalog/fp-validation.md` (≤10% FP) before the public tag | Slice 13 |

## Phase 1.5 / post-launch product backlog

Canonical list of feature-level deferrals (supersedes the roadmap's short "Deferred to Phase 1.5" note). Not launch blockers; pulled by Phase 1 validation signals.

- **Commands:** `charter serve` (MCP server exposing `charter_doctor`/`charter_score`/`charter_fix`/`charter_explain`); `charter explain <RULE>` (reuses the rule catalog); standalone `charter report --format spdx`.
- **Output:** `--format toon`, `--format json-compact`, `--for-agent`; `--no-color`/plain-CI text variant.
- **SARIF:** 2.2 upgrade; `artifacts[]`/`invocation` enrichment; content-based (line-shift-resilient) `partialFingerprints` (today's are position-based).
- **Policy/CLI:** `charter doctor --rule` filtering; rule-level enable/disable, per-rule severity overrides, `rules.ignore`, `rules.AE-CTX-001.token_budget` (see CF-3).
- **Rules:** `AE-SEC-001` → full Gitleaks ruleset (160+ detectors); full 7-agent config parsing (T1.2.1) + deep multi-agent conflict detection (T1.2.2) — current coverage is lighter than the architecture envisions.
- **`charter fix`:** `AE-ENV-001`/`AE-MCP-001` fixers (CF-5/CF-6); present-but-weak-file rewrites; content-aware/3-way diffs; interactive selection.
- **`charter init`:** `.cursor/rules` scaffolding, `.env.example` codebase env-scanning, interactive prompts.
- **Distribution:** GHCR/container images, Scoop, Nix, apt, winget; charter-action Marketplace listing + automated monorepo→action-repo sync (today: manual seed at launch).

## Done

- (none yet)
