# Charter Docs Overhaul — Task List

Last updated: 2026-06-09 (rev 2 — all open questions resolved, icons + CF aligned)  
Full plan: `tasks/plan.md`  
Gate: `moon run :check` green throughout

---

## Phase 1 — Strip Internal Leakage
*No structural changes. Mechanical edits. Safe to do first.*

- [x] **1.1** Strip `> Generated from the internal rule spec...` banner + `## Related ADRs` + `## Related Evals` from all 18 `docs/product/rules/AE-*.mdx`
- [x] **1.2** Fix `docs/product/rules/overview.mdx` — remove `docs/internal/specs/` path + `generate-rule-pages.ts` script reference
- [x] **1.3** Fix `docs/product/cli/doctor.mdx` — remove "Visual Reference" block with `docs/internal/` paths
- [x] **1.4** Fix `docs/product/quickstart.mdx` — remove "Current launch-state notes" block + all "launch-gated" / "not live yet" language
- [x] **1.x** Strip inline ADR citations from AE-SEC-001, AE-SEC-002, AE-SUPPRESS-001, AE-SUPPRESS-003 body prose
- [x] **1.x** Fix `how-to/run-in-github-actions.mdx` — remove "Before You Start" launch-gated section

**Checkpoint 1:**
```
grep -r "docs/internal\|ADR-0\|launch-gated\|Generated from the internal" docs/product/ → empty
moon run :docs → green
```

---

## Phase 2 — Navigation Redesign + New Foundation Pages
*All new pages must be created before Task 2.7 touches docs.json.*

- [x] **2.1** Create `docs/product/design-philosophy.mdx`
- [x] **2.2** Create `docs/product/installation.mdx`
- [x] **2.3** Create `docs/product/concepts/fix-engine.mdx`
- [x] **2.4** Create `docs/product/how-to/pre-commit-hook.mdx`
- [x] **2.5** Create `docs/product/config/policy-profiles.mdx`
- [x] **2.6** Create `docs/product/changelog.mdx`
- [x] **2.7** Redesign `docs/product/docs.json` — anchors + Lucide icons; fix DEPLOY.md grey-cloud error + add Mintlify dashboard + root placeholder A record notes

**Checkpoint 2:**
```
moon run :docs → green
moon run :docs-product → green
# verify nav matches target IA from plan.md
# human: walk navigation top-to-bottom
```

---

## Phase 3 — Rule Page Content Upgrade
*Must complete Q1 generator decision before this phase (resolved — see plan.md).*

- [x] **3.1a** Update `scripts/generate-rule-pages.ts` — new anatomy template; structural check mode (file+title+CLI section, not full content equality)
- [x] **3.1b** All 18 `docs/internal/specs/AE-*.md` — added Why/Auto-fixable/Related rules; cleaned scoring impact ADR citations
- [x] **3.1c** Regenerated all 18 pages; check passes
- [x] **3.2 + 3.3a–f** All 18 rule pages prose-polished — spec pseudocode replaced with customer-facing prose

**Checkpoint 3:**
```
grep -r "Generated from\|ADR-0\|Related Evals\|docs/internal" docs/product/rules/ → empty
moon run :docs → green
# human: read 3 random rule pages cold as a new user
```

---

## Phase 4 — Polish and Validation

- [x] **4.1** Rewrote `docs/product/introduction.mdx` — problem-first, three axes, for-whom, what-it-is-not, start-here cards
- [x] **4.2** Full audit: zero internal refs in cli/, how-to/, concepts/, config/, ci/ — clean; fixed stale README.md line
- [x] **4.3** `moon run :check` green; all 27 nav entries resolve; full internal-ref grep clean

**Checkpoint 4 (Done):**
```
moon run :check → green
grep -r "docs/internal\|ADR-0\|launch-gated\|Generated from the internal" docs/product/ → empty
# human: browser walk-through of full nav
# commit group ready
```

---

## Dependency order

```
1.1–1.4  (independent, batch together)
    ↓
2.1–2.6  (independent of each other, parallel OK)
    ↓
2.7      (depends on all new pages existing)
    ↓
3.1a–3.1c (must be sequential: script → specs → regenerate)
    ↓
3.2, 3.3a–f (independent of each other, can parallelize by batch)
    ↓
4.1–4.3
```
