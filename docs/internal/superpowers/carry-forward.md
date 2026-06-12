# Carry-Forward Ledger

Last reviewed: 2026-06-02 (carry-forward cleanup pass — all code-resolvable items closed)

Durable record of items deliberately **deferred** during slice execution, so they're followed up when their trigger arrives. This is the cross-slice hygiene/debt list — distinct from:

- the **launch roadmap** (`2026-06-01-v1-launch-roadmap.md`) — the slice sequence and the Slice 20 production-release checklist (tap→public, seed `use-charter/charter-action`, first signed tag, branch protection, etc.);
- the roadmap's **"Deferred to Phase 1.5 / v1.1"** note (`charter serve`, `--format toon|json-compact`, `--for-agent`, `charter report --serve` + `charter report --format spdx`, AE-SEC-001 → full Gitleaks ruleset, deep multi-agent conflict detection, always-on TUI watch mode).

When an item is resolved, strike it (or move it to a "Done" note) in the slice that closes it.

## Open items

All **code-resolvable** carry-forward debt was cleared in the 2026-06-02 cleanup pass (see Done). The items below are **not closable by writing code** — they need an external resource (a deployed web host, an Apple Developer cert), a GitHub org-admin action, or ongoing human curation. Each is correctly scheduled to its triggering slice/event.

| # | Item | Why it can't be code-closed now | Trigger / target | Source |
|---|---|---|---|---|
| CF-4 | ~~`go install go.use-charter.dev/charter/cmd/charter@…` won't resolve until `go.use-charter.dev` serves a `go-import` meta tag.~~ **Code-closed (Slice 19, ADR-0026):** the `charter-go-vanity` worker (`infra/go-vanity/`) serves the `go-import` meta and type-checks. Resolution now needs only `wrangler deploy` + the `go.` DNS (auto-created by the `custom_domain` route). | Worker + config committed; live deploy remaining. Binaries + Homebrew still cover install meanwhile. | Slice 19 deploy (`infra/`) | Slice 9 |
| CF-9 | SARIF `helpUri`s point at `https://use-charter.dev/rules/AE-*`, dead until the rule docs pages are **deployed**. | The page *content* can be built in Slice 18; the links only resolve once the docs site is live at the domain. | **Hard launch dependency** — Slice 18 build + deploy | Slice 8 |
| CF-10 | `use-charter` GitHub org hardening: verify the `use-charter.dev` domain (DNS TXT) and require 2FA for members. | Pure GitHub org-admin UI action; no repo code involved. | go-public ops / Slice 20 checklist | Org migration |
| CF-11 | macOS release binaries are cosign-signed but **not Apple-notarized** (the Homebrew cask strips `com.apple.quarantine` interim). | Notarization needs an Apple Developer cert we don't have. | Post-launch, when an Apple Developer cert is available | Slice 9 |
| CF-12 | MCP catalog curation is an ongoing founder duty. **Done so far:** T1.6.3 FP-validation (11 real repos, **0% FP**), ~24 servers + 60+ vendor hosts, real `mcp-server-git` CVEs. **Ongoing:** broaden the FP run, keep advisories/versions current; behind-stable version data is only seeded for `filesystem`. | Human curation against a fast-moving ecosystem (real CVE IDs only, ADR-0021). | Refresh + re-run FP validation at the **Slice 20** release gate; ongoing via T1.6.2 | Slice 13 |

## Phase 1.5 / post-launch product backlog

Canonical list of feature-level deferrals (supersedes the roadmap's short "Deferred to Phase 1.5" note). Not launch blockers; pulled by Phase 1 validation signals.

- **Commands:** `charter serve` (MCP server exposing `charter_doctor`/`charter_score`/`charter_fix`/`charter_explain`); `charter report --serve` (hardened loopback report viewer — `127.0.0.1` + Host-allowlist + token + auto-shutdown, ADR-0025) and `charter report --format spdx` (SBOM). *(`charter explain` and the HTML `charter report` shipped in Slices 15/16.)*
- **Output:** `--format toon`, `--format json-compact`, `--for-agent`. *(`--no-color` + styled/plain TTY-aware `doctor` output shipped in Slice 15.)*
- **Terminal/report:** always-on TUI watch / live re-scan + `$EDITOR` spawn from the `doctor -i` TUI (Slice 15 deferred).
- **SARIF:** 2.2 upgrade; `artifacts[]`/`invocation` enrichment; content-based (line-shift-resilient) `partialFingerprints` (today's are position-based).
- **Policy/CLI:** `charter doctor --rule` filtering; rule-level enable/disable, per-rule severity overrides, `rules.ignore`, `rules.AE-CTX-001.token_budget` (see CF-3).
- **Rules:** `AE-SEC-001` → full Gitleaks ruleset (160+ detectors); full 7-agent config parsing (T1.2.1) + deep multi-agent conflict detection (T1.2.2) — current coverage is lighter than the architecture envisions; `AE-TEST-001` active-language detection now ignores `//go:embed`'d assets (the embedded HTML-report fonts/templates) — **non-embedded** asset-dir web files served from disk by a Go server still activate a language, so revisit the heuristic if real false-positive reports arise.
- **`charter fix`:** `AE-ENV-001`/`AE-MCP-001` fixers (CF-5/CF-6); present-but-weak-file rewrites; content-aware/3-way diffs; interactive selection.
- **`charter init`:** `.cursor/rules` scaffolding, `.env.example` codebase env-scanning, interactive prompts.
- **Distribution:** GHCR/container images, Scoop, Nix, apt, winget; charter-action Marketplace listing + automated monorepo→action-repo sync (today: manual seed at launch).

## Done

**2026-06-02 carry-forward cleanup pass** (every code-resolvable item closed, code + docs green, dogfood 100):

- **CF-1** — §1.8 command gallery rewritten to the shipped fixer set (`AE-CTX-001`/`AE-CTX-004`/`AE-CI-002`/`AE-MCP-001`); dropped the non-existent `AE-ENV-001` fixer examples; corrected the v1 `jq` example and added `charter version --short`/`--format json`.
- **CF-2** — added `scripts/generate-doc-html.ts` (deterministic, `marked`-based) + a `moon docs-html` drift gate wired into `:docs`; regenerated both HTML mirrors from their `.md`. They can no longer drift.
- **CF-3** — corrected the audit checklist's false "configurable via `charter.yaml → rules.AE-CTX-001.token_budget`" claim; the v1 budget is a fixed 600 (per-rule override remains a Phase 1.5 item).
- **CF-5** — *decision (not built):* `AE-ENV-001` has no `charter fix` fixer because its usual gap is an *opinionated* hook config, which conflicts with Charter's no-opinionated-defaults stance. Stays in the Phase 1.5 backlog; revisit only if a non-opinionated default emerges.
- **CF-6** — shipped a safe `AE-MCP-001` fixer in `charter fix`: advisory pin → fixed version, unpinned/behind cataloged package → catalog stable, via an exact-token in-place bump (new `Replace` action + multi-plan fixer). Never rewrites a deprecated package (no drop-in successor).
- **CF-7** — `charter version --short` and `--format json`; the GitHub Action exposes a `score` output.
- **CF-8** — *decision (intended behavior):* the `AE-CI-002` first-party (`use-charter/charter-action@<tag>`) and SLSA-generator tag-pin exemptions are correct — tag-pinning is the required/conventional form for those actions (ADR-0016/0020). Not debt.
- **CF-13** — `AE-MCP-003` now exempts catalog-known OAuth vendor hosts (Sentry/Atlassian/Context7/…), eliminating the OAuth false-positive class surfaced by the FP-validation run.
