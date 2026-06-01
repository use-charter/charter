# Charter v1.0 Launch Roadmap — Slices 9–19

Last reviewed: 2026-06-01

> **What this is:** the locked-in *execution sequence* from the current state (post–Slice 8: SARIF + policy profiles) to a public v1.0 launch. It owns ordering, dependencies, and per-slice definition-of-done — **not** product behavior.
>
> **Authority:** `docs/internal/architecture/charter-architecture-2026.md` remains canonical for behavior, command surface, and phase semantics. Each slice still gets its own ADR(s) + design spec + implementation plan before code. If this roadmap and a per-slice spec disagree, the spec wins — update this file.
>
> **Process gate (binding, per CONTRIBUTING.md):** every slice runs the grounding-and-grill pass first — inspect local manifests/lockfiles, fetch the *latest* official docs for any tool/SDK/schema touched, then critically assess the plan — before any ADR/spec/code. Spec-first: ground+grill → ADR(s) if irreversible → design spec → plan → subagent-driven execution (commit per task) → whole-slice review → push the slice's commits to `main` together once `moon run :check` is green.

## Numbering note

The source request double-used `16` and mislabeled the checklist as `1`/`17`. Canonical numbering is **9–19** below; the post-launch "go-public ops → tag v1.0.0" stays as a final ops step, not a slice.

## Sequence at a glance

| # | Slice | Roadmap ref | Type | Launch blocker | Key dependency |
|---|---|---|---|---|---|
| 9 | Release pipeline (binaries + checksums + cosign + SLSA + SBOM + Homebrew) + `charter version` | M1.5 / T1.5.2; §1.8 | CI / supply-chain | **Hard** | — (replaces `release.yml` stub) |
| 10 | GitHub Action consuming the released binary + SARIF upload + perf validation | M1.5 / T1.5.1, T1.5.3 | CI / action | **Hard** | Slice 9 (released binary) |
| 11 | `charter init` | M1.4 / T1.4.1 | Pure-Go engine | Should-have | — |
| 12 | `charter fix` | M1.4 / T1.4.2 | Pure-Go engine | Should-have | Slice 11 (templates reuse) |
| 13 | MCP Catalog v1 + FP-validation gate (≤10%) | M1.6 / T1.6.1–3 | Pure-Go engine | Should-have | — (couples to Slice 15 help pages) |
| 14 | Full codebase review + 2026 architecture/doc alignment | hardening | Review pass | Quality gate | Slices 9–13 landed |
| 15 | External docs on Mintlify (live, incl. `/rules/AE-*` pages) | distribution | Docs/web | **Hard** | rule catalog helpUris (Slice 8/13) |
| 16 | Launch website (design → build) + docs wiring | distribution | Web | Should-have | Slice 15 (docs to link) |
| 17 | Production Release Checklist (assemble the gate) | release ops | Checklist | **Hard** | Slices 9–16 |
| 18 | Close everything outstanding from the checklist | release ops | Execution | **Hard** | Slice 17 |
| 19 | Final go-live readiness (end-to-end RC dry-run + sign-off) | release ops | Verification | **Hard** | Slice 18 |
| — | go-public ops → tag `v1.0.0` | launch | Ops | — | Slice 19 |

Pure-deterministic/offline engine work (no CI secrets): **11, 12, 13**, plus the `charter version` command in 9. CI/secret/admin-side: **9 (infra), 10**. Web/content: **15, 16**. Process/verification: **14, 17, 18, 19**.

---

## Slice 9 — Release pipeline + `charter version`

- **Goal:** a real, signed, reproducible release so `@v1` resolves to a verifiable binary and `brew install` / `go install` / direct download all work. Wire the `charter version` command (6-command surface).
- **Scope:** replace the build-only `release.yml` with GoReleaser v2 (multi-OS/arch binaries + checksums), cosign v3 keyless signing, SLSA L2 provenance, SPDX 3.0.1 SBOM (e.g. syft), Homebrew tap (`use-charter/homebrew-tap`). `charter version` prints version/commit/build-date/go/platform; build metadata injected via ldflags (GoReleaser) on top of the existing `version.Version()` (`runtime/debug` fallback).
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
- **Scope:** 20–30 curated servers with pinned `stable_version` (embedded via `go:embed`, versioned); catalog schema + seed data (T1.6.1); contribution + CVE process doc (T1.6.2); **pre-ship FP gate** — ≤10% FP across 5+ real public repos, recorded in `docs/catalog-fp-validation.md` (T1.6.3).
- **Depends on:** none for the engine; help pages for catalog entries land in Slice 15.
- **Exit:** a repo pinned below the catalog stable version fires AE-MCP-001 with the upgrade message; FP rate ≤10% documented.
- **Grill / decisions:** FP validation needs real-world repos + manual classification (research effort); catalog curation source-of-truth and update cadence. Should-have (retention).

## Slice 14 — Full codebase review + 2026 architecture/doc alignment

- **Goal:** a top-to-bottom hardening pass before we point the world at it — pristine, idiomatic, drift-free.
- **Scope:** whole-codebase review against every ADR/spec/architecture clause and latest Go idioms/patterns; reconcile any doc drift; run review subagents (e.g. `go-reviewer`, `security-reviewer`, `ce-code-review`). Not a feature slice.
- **Depends on:** Slices 9–13 landed.
- **Exit:** review report with findings resolved; `moon run :check` green; Charter dogfoods to 100; zero stale doc references.
- **Grill / decisions:** scope discipline — fix real issues, resist speculative refactors (Hard Constraint: no speculative refactors). Quality gate.

## Slice 15 — External docs on Mintlify

- **Goal:** live customer-facing docs, including the `/rules/AE-*` pages the SARIF `helpUri`s already point at (currently dead links until this ships).
- **Scope:** Mintlify site (latest `docs.json` schema — fetch fresh, Mintlify moved off `mint.json`); quickstart, CLI reference, CI integration guide, full rule reference (one page per AE-* rule resolving `use-charter.dev/rules/AE-XXX`), config (`charter.yaml`) reference, suppression governance.
- **Depends on:** rule catalog + helpUri scheme (Slices 8/13).
- **Exit:** docs deployed at the chosen domain; every `helpUri` resolves; quickstart reproduces a real scan.
- **Grill / decisions:** hosting + domain (`docs.use-charter.dev` vs `use-charter.dev/docs`); latest Mintlify config schema and navigation model. **Hard blocker** (helpUris must resolve at launch).

## Slice 16 — Launch website + docs wiring

- **Goal:** the marketing/landing surface for `use-charter.dev`, wired to the docs.
- **Scope:** design + implementation (lean on `frontend-design`); hero around the PR-comment scenario, install paths, live examples, CTA into docs/quickstart; link Mintlify docs.
- **Depends on:** Slice 15 (docs to link).
- **Exit:** site live; install instructions accurate; navigation into docs works.
- **Grill / decisions:** stack choice (static/Astro/Next), hosting, and keeping it honest with as-built behavior. Should-have.

## Slice 17 — Production Release Checklist (assemble)

- **Goal:** assemble the single gate document that must be all-green to launch.
- **Scope:** extend `docs/internal/runbooks/release.md` into a v1.0 launch checklist: signed release verified (cosign/SLSA/SBOM), perf budgets met, all 6 commands working, Action verified end-to-end, docs live + helpUris resolve, website live, GitHub Discussions enabled, alerts configured (Signals 1/3/4), branch protection + private vuln reporting + CodeQL/Scorecard on public, LICENSE/NOTICE/CHANGELOG + versioning policy, trademark ADR-0010 = CLEARED, demo asset.
- **Depends on:** Slices 9–16.
- **Exit:** checklist authored, every item with an owner and a verifiable check. **Hard blocker.**

## Slice 18 — Close outstanding checklist items

- **Goal:** drive every open checklist item to done.
- **Scope:** execute whatever Slice 17 surfaced as not-yet-green (admin settings, missing assets, doc fixes, etc.).
- **Depends on:** Slice 17.
- **Exit:** checklist fully green except the final tag. **Hard blocker.**

## Slice 19 — Final go-live readiness

- **Goal:** an end-to-end release-candidate dry-run and sign-off.
- **Scope:** tag an RC; verify the full pipeline on the about-to-be-public repo; smoke-test every install path (brew, `go install`, the Action); confirm SARIF lands in Code Scanning; final sign-off.
- **Depends on:** Slice 18.
- **Exit:** RC verified across all surfaces; go/no-go recorded. **Hard blocker.**

→ **go-public ops → tag `v1.0.0`** (repo public, Discussions live, alerts armed, then the release tag).

---

## Cross-cutting (applies to every slice)

- **Hard constraints hold throughout:** no LLM calls in core; deterministic, offline-first core; diff-first, no silent mutation; fail fast; never log/print secrets; no speculative refactors.
- **Latest-docs-first (ADR-0006):** Slices 9, 10, 15 touch fast-moving external tooling (GoReleaser, cosign, SLSA, codeql-action, Mintlify) — fetch current official docs during each grill; do not rely on training data for config schemas or action versions.
- **Deferred to Phase 1.5 / v1.1 (not in this roadmap):** `charter serve` (MCP server), `--format toon`, `--format json-compact`, `--for-agent`, `charter report --format spdx` as a standalone command, AE-SEC-001 expansion to the full Gitleaks ruleset, deep multi-agent conflict detection (T1.2.2).
- **Validation ≠ launch:** §1.7 Phase 1 exit signals (organic CI adoption, stranger issues, mentions, community self-help) are measured *after* launch and decide Phase 2 — they are not pre-launch gates.

## Critical path

```
9 ─▶ 10 ─┐
         ├─▶ 14 ─▶ 17 ─▶ 18 ─▶ 19 ─▶ (public + v1.0.0)
11 ─▶ 12 ┤              ▲
13 ──────┘              │
        13 ─▶ 15 ─▶ 16 ─┘
```

11/12/13 (pure-Go) can proceed in parallel with 9/10 (CI/infra). 15 depends on the rule set + catalog; 16 depends on 15; 17 gates on everything.
