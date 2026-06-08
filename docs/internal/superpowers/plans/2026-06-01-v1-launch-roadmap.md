# Charter v1.0 Launch Roadmap — Slices 9–22

Last reviewed: 2026-06-02

> **What this is:** the locked-in *execution sequence* from the current state (post–Slice 8: SARIF + policy profiles) to a public v1.0 launch. It owns ordering, dependencies, and per-slice definition-of-done — **not** product behavior.
>
> **Authority:** `docs/internal/architecture/charter-architecture-2026.md` remains canonical for behavior, command surface, and phase semantics. Each slice still gets its own ADR(s) + design spec + implementation plan before code. If this roadmap and a per-slice spec disagree, the spec wins — update this file.
>
> **Process gate (binding, per CONTRIBUTING.md):** every slice runs the grounding-and-grill pass first — inspect local manifests/lockfiles, fetch the *latest* official docs for any tool/SDK/schema touched, then critically assess the plan — before any ADR/spec/code. Spec-first: ground+grill → ADR(s) if irreversible → design spec → plan → subagent-driven execution (commit per task) → whole-slice review → push the slice's commits to `main` together once `moon run :check` is green.

## Numbering note

Canonical numbering is **9–22** below; the post-launch "go-public ops → tag v1.0.0" stays as a final ops step, not a slice. (Slice 14 — the agent-operability rule expansion — was inserted after Slice 13 shipped. Slices **15** (Terminal Experience — styled output + interactive TUI, ADR-0024) and **16** (HTML report — self-contained single file, ADR-0025) were then inserted as the final feature work before launch hardening; the former Slices 15–20 shifted to **17–22**.)

## Sequence at a glance

| # | Slice | Roadmap ref | Type | Launch blocker | Key dependency |
|---|---|---|---|---|---|
| 9 | Release pipeline (binaries + checksums + cosign + SLSA + SBOM + Homebrew) + `charter version` | M1.5 / T1.5.2; §1.8 | CI / supply-chain | **Hard** | — (replaces `release.yml` stub) |
| 10 | GitHub Action consuming the released binary + SARIF upload + perf validation | M1.5 / T1.5.1, T1.5.3 | CI / action | **Hard** | Slice 9 (released binary) |
| 11 | `charter init` | M1.4 / T1.4.1 | Pure-Go engine | Should-have | — |
| 12 | `charter fix` | M1.4 / T1.4.2 | Pure-Go engine | Should-have | Slice 11 (templates reuse) |
| 13 | MCP Catalog v1 + FP-validation gate (≤10%) | M1.6 / T1.6.1–3 | Pure-Go engine | Should-have | — |
| 14 | Agent-operability rule expansion (AE-TEST-001, AE-AUTO-001, AE-CTX-006 + category scorecard) | ADR-0023 | Pure-Go engine | Should-have (top-tier) | Slices 9–13 landed |
| 15 | Terminal experience — styled output + interactive TUI (`doctor -i`) + `charter explain` + `doctor --rule` | ADR-0024 | Pure-Go engine | Should-have (top-tier) | Slices 9–14 landed |
| 16 | HTML report — self-contained single file (`charter report --format html`) — ✅ **shipped** | ADR-0025 | Pure-Go engine | Should-have | Slice 15 |
| 17 | Full codebase review + 2026 architecture/doc alignment | hardening | Review pass | Quality gate | Slices 9–16 landed |
| 18 | External docs on Mintlify (live, incl. `/rules/AE-*` pages) | distribution | Docs/web | **Hard** | rule catalog helpUris (8/13/14) + new command surface (15/16) |
| 19 | Launch website (design → build) + docs wiring | distribution | Web | Should-have | Slice 18 (docs to link) |
| 20 | Production Release Checklist (assemble the gate) | release ops | Checklist | **Hard** | Slices 9–19 |
| 21 | Close everything outstanding from the checklist | release ops | Execution | **Hard** | Slice 20 |
| 22 | Final go-live readiness (end-to-end RC dry-run + sign-off) | release ops | Verification | **Hard** | Slice 21 |
| — | go-public ops → tag `v1.0.0` | launch | Ops | — | Slice 22 |

Pure-deterministic/offline engine work (no CI secrets): **11, 12, 13, 14, 15, 16**, plus the `charter version` command in 9. CI/secret/admin-side: **9 (infra), 10**. Web/content: **18, 19**. Process/verification: **17, 20, 21, 22**.

---

## Slice 9 — Release pipeline + `charter version`

- **Goal:** a real, signed, reproducible release so `@v1` resolves to a verifiable binary and `brew install` / `go install` / direct download all work. Wire the `charter version` command (6-command surface).
- **Scope:** replace the build-only `release.yml` with GoReleaser v2 (multi-OS/arch binaries + checksums), cosign v3 keyless signing, SLSA L3 provenance, SPDX 2.3 SBOM (via syft), Homebrew tap (`use-charter/homebrew-tap`). `charter version` prints version/commit/build-date/go/platform; build metadata injected via ldflags (GoReleaser) on top of the existing `version.Version()` (`runtime/debug` fallback).
- **Depends on:** nothing in-repo; needs the `use-charter` org + a tap repo + GitHub OIDC (`id-token: write`).
- **Exit:** pushing a `v0.x` tag produces signed binaries + checksums + provenance + SBOM on GitHub Releases; `cosign verify-blob` passes; `charter version` output matches the architecture doc; Charter still scores 100 on itself (dogfood AE-CI-002).
- **Grill / decisions:** keyless cosign (Fulcio/Rekor) verification is cleanest on a **public** repo — confirm whether we sign now (private) or gate full keyless verification on go-public; GoReleaser v2 config schema, cosign v3, and the SLSA generator action versions must be fetched fresh (do not trust training data); Homebrew tap is admin-side. **Hard blocker.**

## Slice 10 — GitHub Action + SARIF upload + perf validation

- **Goal:** the primary product surface — `uses: use-charter/charter-action@v1` — runs `charter doctor`, emits SARIF, uploads to Code Scanning, and gates on threshold. Validate performance budgets.
- **Scope:** composite action that downloads the Slice 9 binary pinned by version+checksum (fast, no per-run build); inputs `threshold`, `path`, `version`, `fail-below`, SARIF on by default; wires `github/codeql-action/upload-sarif@v4`. Perf validation (T1.5.3): synthetic 50k-file fixture, assert ≤2s wall / ≤256 MB RSS, race-clean.
- **Depends on:** Slice 9 (released, checksummed binary).
- **Exit:** sample repo using the action shows Charter alerts in the GitHub Security tab; workflow exits non-zero below threshold; perf budgets met in CI.
- **Grill / decisions:** action distribution model — `use-charter/charter-action` implies a **dedicated repo** vs the in-repo `action/` subdir referenced as `owner/repo/path@ref`; composite vs Docker vs JS (composite preferred); SARIF upload requires `security-events: write`. **Hard blocker.**

## Slice 11 — `charter init`

- **Goal:** zero → ≥80 out-of-the-box in under 2 minutes; kill the onboarding/“one-time setup” friction.
- **Scope:** detect language/CI/agents; scaffold `AGENTS.md`, `ARCHITECTURE.md`, `.claude/settings.json` (no MCP bootstrap), `.env.example`, `charter.yaml` (`profile: standard`). Flags `--yes --profile --agents`. Idempotent; never overwrite without consent; report created/skipped/overwritten counts.
- **Depends on:** none (pure-Go, deterministic, offline).
- **Exit:** `charter init` then `charter doctor` scores ≥80 for a Go repo; generated `AGENTS.md` is within the ≤600-token AE-CTX-001 budget; templates themselves pass Charter’s rules (dogfood).
- **Grill / decisions:** template content must not trip AE-* rules; overwrite/consent semantics vs Commitment #3 (never delete) and #4 (diff-first). Should-have for adoption.

## Slice 12 — `charter fix`

- **Goal:** the diff-first repair engine behind the hero PR-comment scenario.
- **Scope (M1.4-scoped):** `--dry-run` shows unified diffs; `--all`/`--rule`/`--yes`; backups to `.charter/backups/`; **never** silent mutation, **never** auto-fix secrets (Commitments #3/#4, ADR-0005). Initial fixers limited to safe rewrites: `AGENTS.md`/`.gitignore`/Charter workflow scaffolding, `mise.toml` creation (AE-ENV-001), MCP pin (AE-MCP-001). Rescan + score-delta after apply.
- **Depends on:** Slice 11 (reuse init templates/detection).
- **Exit:** `--dry-run` mutates nothing and prints diffs; apply is reversible via backups; rescan shows the score delta; complex rewrites explicitly deferred.
- **Grill / decisions:** atomic writes, diff fidelity, which rules are genuinely “safe” to auto-fix in v1. Should-have.

## Slice 13 — MCP Catalog v1 + FP-validation gate

- **Goal:** the recurring-engagement loop — make AE-MCP-001/002 reference a versioned catalog so findings re-fire when a server’s stable version advances or a CVE lands, even on repos that previously passed.
- **Scope:** 20–30 curated servers with pinned `stable_version` (embedded via `go:embed`, versioned); catalog schema + seed data (T1.6.1); contribution + CVE process doc (T1.6.2); **pre-ship FP gate** — ≤10% FP across 5+ real public repos, recorded in `docs/internal/catalog/fp-validation.md` (T1.6.3).
- **Depends on:** none for the engine; help pages for catalog entries land in Slice 18.
- **Exit:** a repo pinned below the catalog stable version fires AE-MCP-001 with the upgrade message; FP rate ≤10% documented.
- **Grill / decisions:** FP validation needs real-world repos + manual classification (research effort); catalog curation source-of-truth and update cadence. Should-have (retention).

## Slice 14 — Agent-operability rule expansion

- **Goal:** take Charter from an agent-*setup* linter to an agent-*operability* scorecard by scoring the third axis of agent-readiness — can the agent verify its work and run the project? — alongside context and safety.
- **Scope:** `AE-TEST-001` (tests present, High), `AE-AUTO-001` (verification command discoverable, Medium), `AE-CTX-006` (instruction over-emphasis, informational), and a per-category readiness scorecard in the report. Offline/deterministic/no-LLM; score formula (ADR-0008) unchanged. Implements ADR-0023.
- **Depends on:** Slices 9–13 landed.
- **Exit:** rule set 15→18; Charter dogfoods to 100; FP ≤10% on real repos (recorded); `moon run :check` green.
- **Grill / decisions:** "active language" = non-test source outside tooling dirs (FP guard); conventional toolchains satisfy AE-AUTO-001; scorecard is reporting-only. Grounded in the AGENTS.md standard + instruction-following research.

## Slice 15 — Terminal experience (styled output + interactive TUI)

- **Goal:** make Charter's terminal surface world-class **without** touching its machine contract — a pristine styled non-interactive default and an opt-in `charter doctor -i` browser — and add the `charter explain` rule surface + `doctor --rule` filtering.
- **Scope:** adopt the `charm.land` v2 stack (`bubbletea/v2` v2.0.7, `lipgloss/v2` v2.0.3, `bubbles/v2` v2.1.0) + `fang` v1.0.0, **confined to the TTY path**; styled `doctor` output (color-tier/`NO_COLOR`/OSC 8, scorecard, finding cards, score hero) with a pure-stdlib non-TTY fallback; the interactive TUI (filter/search/drill-in/rescan); `charter explain <RULE>` (`text`/`json`); `charter doctor --rule`, `--color`, `--no-color`. Implements ADR-0024; matches `docs/internal/designs/*.html` + `DESIGN-TOKENS.md`. Score formula (ADR-0008) unchanged.
- **Depends on:** Slices 9–14 landed.
- **Exit:** styled `doctor` + `-i` TUI + `explain` + `--rule` shipped; the **containment contract test** proves `--format {json,sarif,markdown}` / piped / `NO_COLOR` output is byte-identical to baseline; dogfood 100; `moon run :check` green.
- **Grill / decisions:** non-interactive default preserved (clig.dev / CLI Spec); Charm only on the TTY path; `charm.land` v2 versions fetched fresh (ADR-0006); fonts are terminal-inherited; uses only **installed** design skills (`ui-ux-pro-max`). Should-have (top-tier).

## Slice 16 — HTML report (self-contained single file)

- **Status:** ✅ shipped on `main` (ADR-0025) — new `internal/render/html` renderer + `charter report` command (`--format html|markdown|json`, `--out`, `--open`), brand fonts embedded, and AE-TEST-001 refined so `go:embed`'d assets no longer activate a language. Charter dogfoods to 100.
- **Goal:** give the already-promised `charter report` command its v1.0 purpose — one portable, offline HTML file a human can read, keep, and share.
- **Scope:** `charter report --format html` (default) → a self-contained single file (data/CSS/JS/fonts **inlined**, opens from `file://`, zero network); `--out`, `--open`; `--format markdown|json` reuse the existing renderers. New `internal/render/html`. Implements ADR-0025; inspiration `docs/internal/designs/charter-html-report.html`, built with the installed `.agents/skills/` design set (see `DESIGN-TOKENS.md`), WCAG 2.2 AA, fonts per `DESIGN-TOKENS.md`. `--serve` (hardening checklist recorded), `--format spdx`, and hosted reports are **out of scope** (Phase 1.5 / rejected).
- **Depends on:** Slice 15 (finding/category model + `charter explain` + design tokens).
- **Exit:** a self-contained, offline, WCAG-2.2-AA report matching the elevated design; self-containment + golden tests green; dogfood `charter report` on Charter itself opens cleanly; `moon run :check` green.
- **Grill / decisions:** single-file (not server/hosted) per ADR-0025; offline font embedding (SIL OFL/Apache, Latin-subset woff2); redacted evidence only. Should-have.

## Slice 17 — Full codebase review + 2026 architecture/doc alignment

- **Goal:** a top-to-bottom hardening pass before we point the world at it — pristine, idiomatic, drift-free.
- **Scope:** whole-codebase review against every ADR/spec/architecture clause and latest Go idioms/patterns; reconcile any doc drift; run review subagents (e.g. `go-reviewer`, `security-reviewer`, `ce-code-review`). Now also covers the Slice 15/16 terminal/TUI/report code (incl. the Charm dependency surface + the containment contract). Not a feature slice.
- **Depends on:** Slices 9–16 landed.
- **Exit:** review report with findings resolved; `moon run :check` green; Charter dogfoods to 100; zero stale doc references.
- **Grill / decisions:** scope discipline — fix real issues, resist speculative refactors (Hard Constraint: no speculative refactors). Quality gate.

## Slice 18 — External docs on Mintlify

- **Goal:** live customer-facing docs, including the `/rules/AE-*` pages the SARIF `helpUri`s already point at (currently dead links until this ships). *(Foundation drafted under `web/docs/` — ADR-0022, currently stashed; the page set now includes the Slice 14 rules and the Slice 15/16 command surface — TUI, `charter report`, `charter explain`.)*
- **Scope:** Mintlify site (latest `docs.json` schema — fetch fresh, Mintlify moved off `mint.json`); quickstart, CLI reference (incl. `report`/`explain`/`-i`), CI integration guide, full rule reference (one page per AE-* rule resolving `use-charter.dev/rules/AE-XXX`), config (`charter.yaml`) reference, suppression governance.
- **Depends on:** rule catalog + helpUri scheme (Slices 8/13/14) + the new command surface (Slices 15/16); content reviewed against the as-built CLI after Slice 17.
- **Exit:** docs deployed at the chosen domain; every `helpUri` resolves; quickstart reproduces a real scan.
- **Grill / decisions:** hosting + domain (`docs.use-charter.dev` vs `use-charter.dev/docs`); latest Mintlify config schema and navigation model. **Hard blocker** (helpUris must resolve at launch).

## Slice 19 — Launch website + docs wiring

- **Goal:** the marketing/landing surface for `use-charter.dev`, wired to the docs.
- **Scope:** design + implementation (lean on the `.agents/skills/` design set — `landing-page-design` + `frontend-design` + `high-end-visual-design` + `visual-design-foundations` + `ui-ux-pro-max`); hero around the PR-comment scenario, install paths, live examples (the TUI + HTML report make strong demo assets), CTA into docs/quickstart; link Mintlify docs.
- **Depends on:** Slice 18 (docs to link).
- **Exit:** site live; install instructions accurate; navigation into docs works.
- **Grill / decisions:** stack choice (static/Astro/Next), hosting, and keeping it honest with as-built behavior. Should-have.

## Slice 20 — Production Release Checklist (assemble)

- **Goal:** assemble the single gate document that must be all-green to launch.
- **Scope:** extend `docs/internal/runbooks/release.md` into a v1.0 launch checklist: signed release verified (cosign/SLSA/SBOM), perf budgets met, all 7 commands working (incl. `charter report` + the `doctor -i` TUI + `charter explain`), Action verified end-to-end, docs live + helpUris resolve, website live, GitHub Discussions enabled, alerts configured (Signals 1/3/4), branch protection + private vuln reporting + CodeQL/Scorecard on public, LICENSE/NOTICE/CHANGELOG + versioning policy, trademark ADR-0010 = CLEARED, demo asset, flip `use-charter/homebrew-tap` to **public** (kept private until launch so `brew install` works), create + seed `use-charter/charter-action` from `action/` (tag `v1`) and switch Charter's own CI to dogfood `use-charter/charter-action@`. **MCP catalog T1.6.3 refresh + FP re-validation (CF-12).**
- **Depends on:** Slices 9–19.
- **Exit:** checklist authored, every item with an owner and a verifiable check. **Hard blocker.**

## Slice 21 — Close outstanding checklist items

- **Goal:** drive every open checklist item to done.
- **Scope:** execute whatever Slice 20 surfaced as not-yet-green (admin settings, missing assets, doc fixes, etc.).
- **Depends on:** Slice 20.
- **Exit:** checklist fully green except the final tag. **Hard blocker.**

## Slice 22 — Final go-live readiness

- **Goal:** an end-to-end release-candidate dry-run and sign-off.
- **Scope:** tag an RC; verify the full pipeline on the about-to-be-public repo; smoke-test every install path (brew, `go install`, the Action); confirm SARIF lands in Code Scanning; final sign-off.
- **Depends on:** Slice 21.
- **Exit:** RC verified across all surfaces; go/no-go recorded. **Hard blocker.**

→ **go-public ops → tag `v1.0.0`** (repo public, Discussions live, alerts armed, then the release tag).

---

## Cross-cutting (applies to every slice)

- **Hard constraints hold throughout:** no LLM calls in core; deterministic, offline-first core; diff-first, no silent mutation; fail fast; never log/print secrets; no speculative refactors. The Slice 15 TUI keeps the non-interactive `--format`/piped/`NO_COLOR` contract byte-stable; the Slice 16 report is self-contained and network-free.
- **Latest-docs-first (ADR-0006):** Slices 9, 10, 15, 18 touch fast-moving external tooling (GoReleaser, cosign, SLSA, codeql-action, the Charm `charm.land` v2 stack, Mintlify) — fetch current official docs during each grill; do not rely on training data for config schemas or action/module versions.
- **Deferred to Phase 1.5 / v1.1:** the canonical list lives in `carry-forward.md` (Phase 1.5 backlog) — `charter serve`, `--format toon|json-compact`, `--for-agent`, `charter report --serve` + `charter report --format spdx`, AE-SEC-001 → full Gitleaks ruleset, deep multi-agent parsing/conflict detection, SARIF 2.2/enrichment, richer policy/CLI, an always-on TUI watch mode, and more distribution channels.
- **Validation ≠ launch:** §1.7 Phase 1 exit signals (organic CI adoption, stranger issues, mentions, community self-help) are measured *after* launch and decide Phase 2 — they are not pre-launch gates.
- **Carry-forward ledger:** smaller cross-slice follow-ups live in `carry-forward.md`. As of 2026-06-02 all code-resolvable items are closed; the open items (CF-4/9/10/11/12) are external/ops, scheduled to Slices 18/20/22 + go-public. Review it when starting each slice.

## Critical path

```
9 ─▶ 10 ─┐
11 ─▶ 12 ┼─▶ 15 ─▶ 16 ─▶ 17 ─▶ 20 ─▶ 21 ─▶ 22 ─▶ (public + v1.0.0)
13 ─▶ 14 ┘                     ▲
              18 ─▶ 19 ────────┘
```

11/12/13 (pure-Go) can proceed in parallel with 9/10 (CI/infra). 15→16 (the experience pair) are pure-Go/offline; 17 (review) depends on the full feature set 9–16 and covers it; 18 depends on the rule set + catalog + the new command surface; 19 depends on 18; 20 gates on everything.
