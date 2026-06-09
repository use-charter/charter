# Phase 1 Slice 18 Revamp — Implementation Plan

> **Derived from:** `docs/internal/superpowers/specs/2026-06-09-phase-1-slice-18-design-revamp.md`. Spec owns what/why; this owns how.

**Goal:** Overhaul the Slice 18 Mintlify scaffold into a genuinely customer-facing product documentation site — internal-clean, product-oriented, with a four-zone IA, Lucide icons, new pages, and properly written rule pages.

**Branch & policy:** `main`, `moon run :check` green throughout, GPG-signed commits.

**Status:** ✅ Complete — all tasks executed and pushed.

---

## Phase 1 — Strip internal leakage

Goal: zero internal engineering references visible to customers. Mechanical edits only.

- [x] **1.1** Strip `> Generated from the internal rule spec...` banner + `## Related ADRs` + `## Related Evals` from all 18 `docs/product/rules/AE-*.mdx`
- [x] **1.2** Fix `docs/product/rules/overview.mdx` — remove `docs/internal/specs/` path and `generate-rule-pages.ts` script reference
- [x] **1.3** Fix `docs/product/cli/doctor.mdx` — remove `### Visual Reference` block containing three `docs/internal/` paths
- [x] **1.4** Fix `docs/product/quickstart.mdx` — remove "Before You Start / launch-state notes" block; replace with clean install link
- [x] **1.x** Strip inline ADR citations `(per ADR-0008)` from AE-SEC-001, AE-SEC-002, AE-SUPPRESS-001, AE-SUPPRESS-003 body prose
- [x] **1.x** Fix `docs/product/how-to/run-in-github-actions.mdx` — remove "Before You Start" launch-gated section and "launch-state notes" reference

**Checkpoint verification:**
```bash
grep -r "docs/internal\|ADR-0\|launch-gated\|Generated from the internal" docs/product/
# → empty
moon run :docs
# → green
```

---

## Phase 2 — Navigation redesign + new foundation pages

Goal: implement four-zone IA; add all missing pages before touching `docs.json`.

- [x] **2.1** Create `docs/product/design-philosophy.mdx` — ten commitments in customer language, tradeoff-framed, no ADR/slice refs
- [x] **2.2** Create `docs/product/installation.mdx` — four install paths (brew, binary, go install, source) with v1.0 availability; no launch-gated prose
- [x] **2.3** Create `docs/product/concepts/fix-engine.mdx` — diff-first guarantee, four fixable rules table, why secrets are not auto-fixable, backup mechanism
- [x] **2.4** Create `docs/product/how-to/pre-commit-hook.mdx` — `hk`, `husky`, and plain shell hook configs; `--quiet --threshold` usage; exit code contract
- [x] **2.5** Create `docs/product/config/policy-profiles.mdx` — `standard`/`strict` profiles, `--threshold` override, four-level precedence
- [x] **2.6** Create `docs/product/changelog.mdx` — v1.0 entry with full feature list; no slice/ADR refs
- [x] **2.7** Rewrite `docs/product/docs.json` — Docs tab uses `anchors` with Lucide icons; CLI Reference and Rules tabs use `groups` with icons; add `"icons": {"library":"lucide"}`, `"navbar"`, `"footer"`, `"redirects": []`; set `"logo"` and `"favicon"` paths
- [x] **2.x** Copy brand assets to `docs/product/images/` — `logo-light.svg` (#0D1117 fill), `logo-dark.svg` (#FFFFFF fill), `favicon.svg`
- [x] **2.x** Fix `docs/product/DEPLOY.md` — correct grey-cloud error on optional CNAME; add Mintlify dashboard note; add root placeholder A record note

**Checkpoint verification:**
```bash
moon run :docs-product
# → 27 navigation entries resolve to MDX files
```

---

## Phase 3 — Rule page content upgrade

Goal: all 18 rule pages have customer-facing anatomy with real prose.

- [x] **3.1a** Update `scripts/generate-rule-pages.ts` — new `renderRulePage()` template; remove banner/ADR/Evals; add Why/Auto-fixable/Related rules sections; rename sections for customer audience; change `--check` to structural validation (file + title + CLI section)
- [x] **3.1b** Update all 18 `docs/internal/specs/AE-*.md` — add `Why:`, `Auto-fixable:`, `Related rules:` fields; strip `(per ADR-0008)` from AE-SEC-001/002 scoring impact fields
- [x] **3.1c** Run `bun scripts/generate-rule-pages.ts` — regenerate all 18 pages with new anatomy; verify `bun scripts/generate-rule-pages.ts --check` green
- [x] **3.2** Prose polish — AE-CTX-001/002/004/006, AE-MCP-001/002, AE-SEC-001: spec pseudocode replaced with customer-facing prose; detection descriptions, examples, and fix guidance rewritten
- [x] **3.3** Prose polish — AE-SEC-002, AE-CC-001/002, AE-ENV-001, AE-CI-002, AE-TEST-001, AE-SUPPRESS-001/002/003, AE-MCP-003, AE-AUTO-001: same treatment

**Checkpoint verification:**
```bash
grep -r "Generated from\|ADR-0\|Related Evals\|docs/internal" docs/product/rules/
# → empty
bun scripts/generate-rule-pages.ts --check
# → Rule pages OK
```

---

## Phase 4 — Polish and validation

- [x] **4.1** Rewrite `docs/product/introduction.mdx` — problem-first opening; three readiness axes (context/safety/operability); who uses Charter; what Charter is not; Cards at bottom
- [x] **4.2** Audit all remaining pages for internal leakage: `cli/`, `how-to/`, `concepts/`, `config/`, `ci/`, `docs/product/README.md`
- [x] **4.3** Final validation: `moon run :check` green; full internal-ref grep clean

**Final grep:**
```bash
grep -rn "ADR-0\|docs/internal\|launch-gated\|Generated from the internal\|Slice [0-9]\|CF-[0-9]" docs/product/
# → empty
```

---

## Commits

| SHA | Message |
|---|---|
| `ccde9d1` | `docs(rules): add customer-facing fields to rule specs and update generator anatomy` |
| `188e6b4` | `docs: Slice 18 product docs overhaul — internal-clean, four-zone IA, rule anatomy` |
| `d76353f` | `docs: expand docs/product/README — contributor guide for the Mintlify site` |

---

## Cross-task invariants

- No Go changes
- No schema changes
- `moon run :check` green at every commit
- `/rules/AE-*` URL paths unchanged — SARIF `helpUri` contract preserved
- All rule page prose edits go into `docs/internal/specs/AE-*.md` — product pages are hand-maintained but spec files remain the source for the machine fields
