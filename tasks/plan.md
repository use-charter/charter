# Plan: Charter Product Docs Overhaul

Last updated: 2026-06-09 (rev 2 — icons strategy + Cloudflare deployment alignment)  
Scope: `docs/product/` only — no changes to `docs/internal/`, Go source, or scripts (generate-rule-pages.ts is in-scope for template update only)  
Gate: `moon run :check` must stay green throughout; validate-product-docs.ts must pass each phase

---

## Problem Statement

`docs/product/` was written during active development. It passes `moon run :check` but fails as customer-facing documentation because:

1. **Internal leakage** — ADR citations, internal file paths (`docs/internal/...`), "Generated from internal rule spec" banners, launch-gated pre-release language, internal script references baked into public pages.
2. **Missing product narrative** — no design philosophy, no "why Charter", no installation page, no troubleshooting.
3. **Navigation is reference-first, not journey-first** — drops developers into CLI commands before they understand what they're doing.
4. **Rule pages are terse stubs** — no "why this rule", no real invalid/valid examples with CLI output, no auto-fix badges.
5. **Quickstart references launch-blocked infrastructure** — still shows pre-launch "launch-gated" caveats that will confuse anyone who finds this post-launch.

---

## Target State

After this overhaul, `docs/product/` is a **four-zone, Diátaxis-aligned, internal-clean documentation site** that:

- Speaks only to customers — zero internal file paths, ADR references, ticket numbers, launch caveats, or generator banners visible
- Has a clear journey: land → understand → install → scan → fix → gate in CI
- Follows Biome/ESLint rule-page anatomy for all 18 rule pages
- Positions Charter with a real design philosophy page (the ten commitments, customer-framed)
- Uses verb-titled guide pages ("Add Charter to an existing repo", not "Adopt in Existing Repo")
- Validates clean with `moon run :docs` throughout

---

## Information Architecture (New)

### Tab 1: Docs

```
Getting Started
  ├── Introduction
  ├── Installation
  └── Quickstart

How Charter Works  (was "Concepts")
  ├── Agent Readiness Model
  ├── Scoring and Caps
  ├── The Fix Engine          ← NEW
  ├── MCP Safety Model
  └── Suppression and Governance

Guides
  ├── Add Charter to an existing repo
  ├── Run Charter in GitHub Actions
  ├── Fix findings automatically
  ├── Suppress a false positive
  ├── Investigate MCP findings
  └── Use Charter in a pre-commit hook   ← NEW

Configuration
  ├── charter.yaml reference
  └── Policy profiles   ← NEW (split from charter.yaml)

Design Philosophy   ← NEW
  └── The ten commitments

Changelog           ← NEW
  └── v0.x → v1.0
```

### Tab 2: CLI Reference (unchanged structure, cleaned content)

```
Overview
doctor | init | fix | report | explain | suppress | version
```

### Tab 3: Rules (unchanged structure, rewritten content)

```
Overview
Context | Secrets | MCP Safety | Agent Config | Environment & CI | Testing & Autonomy | Governance
```

---

## Internal Leakage Inventory (what to strip)

| File | Leakage |
|---|---|
| `rules/AE-CTX-001.mdx` (and all 17 others) | `> Generated from the internal rule spec. Edit docs/internal/specs/AE-CTX-001.md, then re-run bun scripts/generate-rule-pages.ts.` header |
| `rules/AE-CTX-001.mdx` (and several others) | `## Related ADRs` section with ADR-XXXX citations |
| `rules/overview.mdx` | "Rule pages are generated from the internal specs in `docs/internal/specs/`." + script path reference |
| `cli/doctor.mdx` | "For a repo-grounded visual reference, see... `docs/internal/architecture/charter-architecture-2026.md`" + two `docs/internal/designs/` paths |
| `quickstart.mdx` | "Current launch-state notes:", "The public Homebrew tap is still launch-gated.", "The public `go install` path is also not live yet because the vanity import host is not deployed." |
| `config/charter-yaml.mdx` | (verify for any internal refs) |
| `concepts/mcp-safety-model.mdx` | (verify clean — looks okay but check) |

---

## Dependency Graph

```
docs.json (navigation structure)
    │
    ├── introduction.mdx     ← depends on nothing; rewrite is self-contained
    ├── installation.mdx     ← NEW; depends on nothing
    ├── quickstart.mdx       ← depends on installation.mdx existing
    │
    ├── concepts/            ← renames to "how-charter-works/"
    │   ├── agent-readiness-model.mdx    (good, light edits)
    │   ├── scoring-and-caps.mdx         (good, light edits)
    │   ├── fix-engine.mdx               ← NEW
    │   ├── mcp-safety-model.mdx         (good, verify clean)
    │   └── suppression-governance.mdx   (good, light edits)
    │
    ├── how-to/              (keep path, rename pages titles)
    │   ├── adopt-in-existing-repo.mdx   (good, small edits)
    │   ├── run-in-github-actions.mdx    (verify clean)
    │   ├── use-charter-fix-safely.mdx   (verify clean)
    │   ├── suppress-a-finding.mdx       (verify clean)
    │   ├── investigate-mcp-findings.mdx (verify clean)
    │   └── pre-commit-hook.mdx          ← NEW
    │
    ├── config/
    │   ├── charter-yaml.mdx    (audit + possibly split)
    │   └── policy-profiles.mdx ← NEW
    │
    ├── design-philosophy.mdx   ← NEW
    ├── changelog.mdx           ← NEW
    │
    ├── cli/                    ← audit all for internal refs
    │   └── doctor.mdx          ← has confirmed internal leakage, fix
    │
    └── rules/                  ← mass-strip banner + Related ADRs from all 18 + add content
        ├── overview.mdx        ← confirmed leakage, fix
        └── AE-*.mdx × 18      ← confirmed leakage in all
```

**Implementation order follows dependencies bottom-up:**  
`docs.json` last (after all pages exist). New pages before they're added to navigation. Strip/clean before nav references them.

---

## Phases and Tasks

### Phase 1 — Strip Internal Leakage (no new content; no structural change)

**Goal:** every existing page passes the "customer who has never seen our codebase" test.

#### Task 1.1 — Strip rule page banners and ADR sections

**Description:** Remove the `> Generated from the internal rule spec...` banner and any `## Related ADRs` / `## Related Evals` sections from all 18 rule pages. These are internal provenance markers with no value to a customer.

**Acceptance criteria:**
- [ ] Zero occurrences of "Generated from the internal rule spec" in `docs/product/`
- [ ] Zero occurrences of "ADR-0" in `docs/product/`
- [ ] Zero occurrences of "Related Evals" in `docs/product/`
- [ ] All 18 rule pages still render valid MDX frontmatter

**Verification:**
- [ ] `grep -r "Generated from" docs/product/` returns empty
- [ ] `grep -r "ADR-0" docs/product/` returns empty
- [ ] `grep -r "Related Evals" docs/product/` returns empty
- [ ] `moon run :docs` passes

**Files touched:** all 18 `docs/product/rules/AE-*.mdx`  
**Scope:** Small-Medium (18 files, each 2-3 line removals)

#### Task 1.2 — Fix rules/overview.mdx

**Description:** Remove the paragraph referencing `docs/internal/specs/` and the `bun scripts/generate-rule-pages.ts` script from the public rules overview page.

**Acceptance criteria:**
- [ ] No reference to `docs/internal/` in `rules/overview.mdx`
- [ ] No reference to `generate-rule-pages.ts` or any bun script in `rules/overview.mdx`
- [ ] Page still accurately describes the rule set

**Verification:**
- [ ] `grep -n "internal" docs/product/rules/overview.mdx` returns empty
- [ ] `moon run :docs` passes

**Files touched:** `docs/product/rules/overview.mdx`  
**Scope:** XS

#### Task 1.3 — Fix cli/doctor.mdx

**Description:** Remove the "Visual Reference" note block that links to `docs/internal/architecture/charter-architecture-2026.md` and two `docs/internal/designs/` files. Replace with either nothing or a brief "See the CLI reference examples above" sentence.

**Acceptance criteria:**
- [ ] Zero references to `docs/internal/` in `cli/doctor.mdx`
- [ ] The interactive TUI section still explains what the TUI does

**Verification:**
- [ ] `grep -n "internal" docs/product/cli/doctor.mdx` returns empty

**Files touched:** `docs/product/cli/doctor.mdx`  
**Scope:** XS

#### Task 1.4 — Fix quickstart.mdx (remove launch-gated language)

**Description:** The quickstart currently reads as a pre-launch draft with a "Current launch-state notes" block acknowledging that Homebrew and `go install` aren't available. Rewrite this section to describe actual install paths cleanly, deferring the detailed install instructions to the new `installation.mdx` page (Task 2.2).

**Acceptance criteria:**
- [ ] No "launch-state" or "launch-gated" language in quickstart
- [ ] No "not live yet" or "is still launch-gated" language
- [ ] Quickstart reads as if Charter is fully shipped
- [ ] Install section points to `installation.mdx` for details

**Verification:**
- [ ] `grep -n "launch" docs/product/quickstart.mdx` returns empty
- [ ] `grep -n "not live" docs/product/quickstart.mdx` returns empty

**Files touched:** `docs/product/quickstart.mdx`  
**Scope:** Small

---

**Checkpoint 1:** After Tasks 1.1–1.4  
- [ ] `grep -r "docs/internal" docs/product/` returns empty  
- [ ] `grep -r "ADR-0" docs/product/` returns empty  
- [ ] `grep -r "launch-gated\|launch-state\|not live" docs/product/` returns empty  
- [ ] `moon run :docs` passes  
- Human review of all changed files before Phase 2

---

### Phase 2 — Navigation Redesign + New Foundation Pages

**Goal:** restructure `docs.json` to the four-zone IA; add missing foundation pages.

#### Task 2.1 — Add design-philosophy.mdx

**Description:** Create `docs/product/design-philosophy.mdx` — a customer-facing version of Charter's ten commitments and design principles. Write in "we chose X over Y because Z" language, not marketing language. Frame each commitment as a tradeoff, not a feature.

**Acceptance criteria:**
- [ ] Page covers all ten commitments from the architecture doc in customer language
- [ ] Zero internal references (no ADR numbers, no slice numbers)
- [ ] Written in tradeoff/decision language, not promotional language
- [ ] Includes the "non-goals" (what Charter is not) — these are trust-building

**Verification:**
- [ ] `grep -n "ADR\|Slice\|internal\|launch" docs/product/design-philosophy.mdx` returns empty
- [ ] Page renders clean MDX

**Files touched:** `docs/product/design-philosophy.mdx` (new)  
**Scope:** Medium

#### Task 2.2 — Add installation.mdx

**Description:** Create `docs/product/installation.mdx` covering all current and planned install paths: Homebrew, direct binary download from GitHub Releases, `go install`, and build from source. Mark paths by availability (stable / coming soon) without internal implementation language.

**Acceptance criteria:**
- [ ] Covers: Homebrew, direct binary download, `go install`, build from source
- [ ] Each path has a working command block
- [ ] Platform-specific notes (macOS, Linux, Windows) are accurate
- [ ] No "launch-gated" language — state things as they are at publish time

**Verification:**
- [ ] Page renders valid MDX
- [ ] Code blocks are syntactically correct

**Files touched:** `docs/product/installation.mdx` (new)  
**Scope:** Small

#### Task 2.3 — Add concepts/fix-engine.mdx

**Description:** Create a "The Fix Engine" concepts page that explains how `charter fix` works — diff-first, backup-then-write, what rules are auto-fixable and why others are intentionally not. This closes a notable gap: nothing in the current docs explains why certain rules can't be auto-fixed.

**Acceptance criteria:**
- [ ] Explains the diff-first guarantee in plain language
- [ ] Lists which rules have fixers and why (AE-CTX-001, AE-CTX-004, AE-CI-002, AE-MCP-001)
- [ ] Explains why secrets and dangerous-command rules are intentionally not auto-fixable
- [ ] Explains the backup mechanism (`.charter/backups/<ts>/`)
- [ ] Zero internal references

**Verification:**
- [ ] `grep -n "ADR\|internal\|fixer" docs/product/concepts/fix-engine.mdx` returns clean (no implementation terms)
- [ ] Page renders valid MDX

**Files touched:** `docs/product/concepts/fix-engine.mdx` (new)  
**Scope:** Small-Medium

#### Task 2.4 — Add how-to/pre-commit-hook.mdx

**Description:** Create a guide for using `charter doctor --quiet` in a pre-commit hook via `hk`, `husky`, or a plain shell hook. Covers the quiet-mode exit code contract and threshold override.

**Acceptance criteria:**
- [ ] Shows working hook config for `hk` (preferred) and `husky`
- [ ] Shows `--quiet --threshold` usage
- [ ] Explains the exit code contract for hooks
- [ ] No internal references to hk.pkl internals

**Files touched:** `docs/product/how-to/pre-commit-hook.mdx` (new)  
**Scope:** Small

#### Task 2.5 — Add config/policy-profiles.mdx

**Description:** Extract the policy profiles explanation from `charter-yaml.mdx` into its own page and expand it. Cover `standard`, `strict`, `relaxed` (if applicable), and how `--threshold` overrides the profile.

**Acceptance criteria:**
- [ ] Explains what profiles are and when to use each
- [ ] Shows `charter.yaml` examples for each profile
- [ ] Explains the threshold override precedence
- [ ] No internal references

**Files touched:** `docs/product/config/policy-profiles.mdx` (new), `docs/product/config/charter-yaml.mdx` (minor edit)  
**Scope:** Small

#### Task 2.6 — Add changelog.mdx

**Description:** Create a minimal `docs/product/changelog.mdx` covering the v1.0 release — what's included, the key commands, SARIF support. Keep it brief; this page exists so the site has a versioning anchor.

**Acceptance criteria:**
- [ ] v1.0 entry covers: 7 commands, 18 rules, SARIF, GitHub Action, HTML report
- [ ] No internal slice references ("Slice 16", "Slice 18", etc.)
- [ ] No ADR references

**Files touched:** `docs/product/changelog.mdx` (new)  
**Scope:** XS

#### Task 2.7 — Redesign docs.json navigation

**Description:** Rewrite `docs.json` to implement the four-zone IA. Split the current single "Docs" tab into proper groups: Getting Started, How Charter Works, Guides, Configuration, Design Philosophy, Changelog. Move CLI commands to their own "CLI Reference" tab (already exists). Rules tab stays.

**Acceptance criteria:**
- [ ] Navigation matches the target IA from this plan
- [ ] All referenced pages exist as files
- [ ] `validate-product-docs.ts` passes
- [ ] `moon run :docs` passes

**Verification:**
- [ ] `moon run :docs` green
- [ ] `moon run :docs-product` green

**Files touched:** `docs/product/docs.json`  
**Scope:** Small (config change, high-leverage)  
**Dependencies:** Tasks 2.1–2.6 (all new pages must exist before adding to nav)

---

**Checkpoint 2:** After Tasks 2.1–2.7  
- [ ] All new pages exist and render valid MDX  
- [ ] Navigation structure matches target IA  
- [ ] `moon run :docs` green  
- [ ] `moon run :check` green  
- Human review of navigation and new pages

---

### Phase 3 — Rule Page Content Upgrade

**Goal:** all 18 rule pages follow Biome-style anatomy with real examples and no stubs.

#### Task 3.1 — Define rule page template

**Description:** Agree on the standard structure for a rule page before touching any of the 18:

```mdx
---
title: "AE-XXX-001"
description: "[one sentence why this rule exists]"
---

**Rule ID:** AE-XXX-001  
**Severity:** [Blocker / High / Medium / Low / Informational]  
**Category:** [category]  
**Auto-fixable:** [Yes — `charter fix --rule AE-XXX-001` / No]

[One sentence explaining what this rule detects and why it matters.]

## Why this rule

[2-4 sentences. The real-world failure mode this rule prevents. Technical, honest, not promotional.]

## What triggers it

[Brief detection description in plain language. Not pseudocode. Not "inspect canonical agent-visible files".]

## Examples

### Failing

[Real-world failing scenario with the file content or command that triggers it]

### Passing

[What a passing state looks like]

## How to fix

[Concrete remediation steps. If auto-fixable, show `charter fix --rule AE-XXX-001`.]

## Edge cases

[Any gotchas or unusual inputs that affect behavior. Remove if empty.]

## Related rules

[Cross-links to closely related rules.]

## CLI

`charter explain AE-XXX-001`
```

**Acceptance criteria:**
- [ ] Template documented and agreed
- [ ] Template has no internal references or ADR fields

**Files touched:** this plan document (template captured above)  
**Scope:** XS (agreement task)

#### Task 3.2 — Rewrite high-impact rule pages first (AE-SEC-001, AE-MCP-001, AE-CTX-001, AE-CC-001)

**Description:** Apply the full template to the four rules that are most commonly triggered and most likely to be the first pages a developer reads. These four are the "trust builders" for the rule reference.

**Acceptance criteria:**
- [ ] Each page has: one-sentence description, "Why this rule", detection explanation, failing+passing examples, how to fix, related rules
- [ ] No internal references in any section
- [ ] Auto-fixable badge accurate (AE-CTX-001: yes, AE-MCP-001: yes, AE-SEC-001: no, AE-CC-001: no)
- [ ] Examples use realistic (not toy) content

**Verification:**
- [ ] `moon run :docs` passes
- [ ] Spot-check: read AE-SEC-001 cold — would a developer understand every term without context?

**Files touched:** `docs/product/rules/AE-SEC-001.mdx`, `AE-MCP-001.mdx`, `AE-CTX-001.mdx`, `AE-CC-001.mdx`  
**Scope:** Medium

#### Task 3.3 — Rewrite remaining 14 rule pages

**Description:** Apply the same template to the remaining 14 rules. Can be batched by category.

**Batch A — Context (AE-CTX-002, AE-CTX-004, AE-CTX-006)**  
**Batch B — MCP (AE-MCP-002, AE-MCP-003)**  
**Batch C — Secrets (AE-SEC-002)**  
**Batch D — Agent Config (AE-CC-002)**  
**Batch E — Env/CI/Ops (AE-ENV-001, AE-CI-002, AE-TEST-001, AE-AUTO-001)**  
**Batch F — Governance (AE-SUPPRESS-001, AE-SUPPRESS-002, AE-SUPPRESS-003)**

**Acceptance criteria:**
- [ ] All 14 pages follow the template
- [ ] No internal references remain
- [ ] Each has at minimum: why, detection, examples (fail+pass), fix

**Verification:**
- [ ] `moon run :docs` passes
- [ ] `grep -r "Generated from\|ADR-0\|Related Evals\|docs/internal" docs/product/rules/` returns empty

**Files touched:** 14 rule pages  
**Scope:** Large (14 files; can be parallelized by category batch)

---

**Checkpoint 3:** After Tasks 3.1–3.3  
- [ ] All 18 rule pages follow the template  
- [ ] Zero internal leakage in rules/  
- [ ] `moon run :docs` green  
- Human spot-check: read 3 random rule pages cold as a new user

---

### Phase 4 — Polish and Validation

#### Task 4.1 — Introduction rewrite

**Description:** Rewrite `introduction.mdx` with a stronger product narrative. Lead with the problem Charter solves (teams adopt agents, but repos aren't ready for them). Introduce the three axes (context, safety, operability). Cover who Charter is for. Move the link cards to the bottom.

**Acceptance criteria:**
- [ ] Opens with the customer problem, not a product description
- [ ] Mentions all three readiness axes
- [ ] "For whom" is explicit (individual dev, platform team, security team)
- [ ] "What Charter is not" present (not a code reviewer, not a linter, not a secrets vault)
- [ ] Zero internal references

**Scope:** Small-Medium

#### Task 4.2 — Audit remaining cli/ and how-to/ pages

**Description:** Systematic pass over all remaining pages not yet touched. Look for any internal leakage, stale "launch-gated" language, or unusable placeholder text.

**Files to audit:**
- `cli/init.mdx`, `cli/fix.mdx`, `cli/report.mdx`, `cli/explain.mdx`, `cli/suppress.mdx`, `cli/version.mdx`, `cli/overview.mdx`
- `how-to/run-in-github-actions.mdx`, `how-to/use-charter-fix-safely.mdx`, `how-to/suppress-a-finding.mdx`, `how-to/investigate-mcp-findings.mdx`
- `concepts/suppression-governance.mdx`
- `config/charter-yaml.mdx`
- `ci/github-action.mdx`

**Acceptance criteria:**
- [ ] No `docs/internal/` path references in any file
- [ ] No ADR-XXXX references in any file
- [ ] No "launch-gated" / "not live yet" language
- [ ] All code examples use valid Charter command syntax

**Verification:**
- [ ] `grep -r "docs/internal\|ADR-0\|launch-gated\|not live" docs/product/` returns empty (final check)

**Scope:** Medium

#### Task 4.3 — Final validation pass

**Description:** Run the full docs gate, dogfood the navigation, verify all links in docs.json are real files.

**Acceptance criteria:**
- [ ] `moon run :check` green
- [ ] `moon run :docs-product` green
- [ ] `moon run :docs-product-rules` green
- [ ] All nav entries in `docs.json` resolve to real `.mdx` files
- [ ] `grep -r "docs/internal\|ADR-0\|launch-gated\|Generated from the internal" docs/product/` → empty

**Scope:** XS

---

**Checkpoint 4: Done**  
- [ ] All acceptance criteria above green  
- [ ] Human walks through the navigation top-to-bottom in a browser  
- [ ] Commit group: `docs: product docs overhaul — internal-clean, four-zone IA, rule page content upgrade`

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| `validate-product-docs.ts` has rules we haven't read | Medium — could break `:docs` when adding new pages | Read `scripts/validate-product-docs.ts` before Phase 2 to understand all constraints |
| `generate-rule-pages.ts` overwrites our manual rule page rewrites | High — wipes Phase 3 work | Verify the script's behavior; if it fully overwrites, we need to decouple Phase 3 from the generator or update the generator templates |
| New pages added to nav before they exist | Medium — breaks `validate-product-docs.ts` | Always create the file before adding to `docs.json` |
| Internal references reintroduced by future generate-rule-pages.ts runs | Medium — re-leaks ADR refs | After Phase 3, update the generator template to not emit ADR sections |
| docs.json schema constraints from Mintlify | Low — might reject tab/group restructure | Check Mintlify docs.json schema before Task 2.7 |

---

---

## Icons Strategy

**Decision: anchors with icons for Docs tab; groups with icons for CLI Reference and Rules tabs.**

Rationale:
- Mintlify's `anchor` type renders as a prominent sidebar section with a visible icon — the "premium" look, right for the user-journey Docs tab where each zone is meaningfully distinct.
- `group` with icon renders as a lighter section header — right for reference tabs (CLI, Rules) where the goal is density and scannability, not zone differentiation.
- Top-level `"icons": { "library": "lucide" }` locks the icon set. Lucide is the modern default on Mintlify; it has all the icons we need and is visually consistent with developer tools.

**Icon map — Lucide names:**

| Anchor / Group | Icon name | Rationale |
|---|---|---|
| Getting Started | `rocket` | Standard "launch" signal |
| How Charter Works | `compass` | Orientation / understanding |
| Guides | `route` | Task paths / how-to flows |
| Configuration | `sliders-horizontal` | Tuning/settings affordance |
| CI & GitHub Action | `git-branch` | CI/version-control visual |
| Design Philosophy | `lightbulb` | Principle / rationale |
| Changelog | `history` | Time / versioning |
| CLI Reference overview | `terminal` | Shell / CLI |
| CLI commands group | `code-2` | Command syntax |
| Rules overview | `layout-grid` | Grid of rule items |
| Context rules | `file-text` | Context files |
| Secrets rules | `lock` | Credential safety |
| MCP Safety rules | `shield` | Security boundary |
| Agent Config rules | `bot` | Agent configuration |
| Environment & CI | `wrench` | Toolchain setup |
| Testing & Autonomy | `flask-conical` | Test/verification |
| Governance rules | `scale` | Balance / oversight |

**URL paths do not change.** Pages under `concepts/`, `how-to/`, `config/`, `rules/` keep their current slugs — only sidebar display names change. This means **zero redirects needed** for the restructure. The `/rules/AE-*` SARIF helpUri contract is preserved automatically.

---

## Cloudflare + Mintlify Deployment Alignment

Source: `docs/product/DEPLOY.md` (already exists, mostly correct — one error noted below).

### Architecture (confirmed correct)

```
use-charter.dev  (Cloudflare Registrar + DNS)
     │
     ├── /docs/*  ─── CF Worker (docs-proxy) ──► charter.mintlify.dev
     ├── /rules/* ─── CF Worker (docs-proxy) ──► charter.mintlify.dev
     └── /*       ─── CF Worker ──► LANDING_ORIGIN (Slice 19, TBD)
```

Mintlify serves from `charter.mintlify.dev`. The Worker proxies `/docs/*` and `/rules/*` to it, setting `X-Forwarded-Host: use-charter.dev` so Mintlify knows its public hostname.

### Required Cloudflare DNS (for Worker to fire on the zone)

Cloudflare Workers on a route only fire when the request hits Cloudflare's network. For the root domain `use-charter.dev` to route through the Worker, a proxied (orange cloud) placeholder record must exist:

```
Type   Name             Content      Proxy
A      use-charter.dev  192.0.2.1    ✓ Proxied (orange cloud)
```

`192.0.2.1` is a non-routable IP (RFC 5737 TEST-NET-1). The Worker intercepts before the IP is ever used.

### Error in DEPLOY.md (fix needed)

> "Optional: docs subdomain — CNAME `docs.use-charter.dev` → `charter.mintlify.dev` (proxied — orange cloud)"

**This is wrong.** A CNAME pointing directly to `charter.mintlify.dev` must be **DNS only (grey cloud)**. Orange cloud makes Cloudflare terminate TLS, breaking Mintlify's SSL certificate provisioning. Mintlify cannot issue a Let's Encrypt cert for a hostname whose DNS is orange-clouded through a different CDN.

Correct:
```
Type    Name   Content                  Proxy
CNAME   docs   charter.mintlify.dev     ✗ DNS only (grey cloud)
```

This is optional — the Worker route is the primary path. But if the subdomain is ever added, it must be grey-cloud.

### Mintlify dashboard setup (out of docs.json scope)

In the Mintlify project settings:
- **Custom domain**: `use-charter.dev`
- Mintlify will use `X-Forwarded-Host: use-charter.dev` from the Worker headers to recognize its public URL.
- No SSL challenge needed for `use-charter.dev` — the Worker terminates TLS at Cloudflare's edge, Mintlify sees the forwarded hostname.

### docs.json additions (owned by Task 2.7)

`docs.json` itself does not have a custom domain field — domain is Mintlify-dashboard only. However Task 2.7 must also add:

1. `"icons": { "library": "lucide" }` — enables the full Lucide icon set globally.
2. `"logo"` block — light/dark SVGs pointing to assets in `docs/product/images/` (create placeholders if not yet designed).
3. `"favicon"` — `/favicon.svg` placeholder.
4. `"navbar"` — GitHub link + "Get Started" CTA button.
5. `"footer"` — GitHub social link.
6. `"redirects": []` — empty array now; populated if any path changes during Phase 3–4 restructure.
7. Convert Docs tab `groups` → `anchors` (each zone becomes an anchor with icon).
8. Add `"icon"` to all groups in CLI Reference and Rules tabs.

### Concrete docs.json target (Task 2.7 output)

```json
{
  "$schema": "https://mintlify.com/docs.json",
  "name": "Charter",
  "theme": "mint",
  "colors": {
    "primary": "#2563EB",
    "light": "#60A5FA",
    "dark": "#1D4ED8"
  },
  "icons": {
    "library": "lucide"
  },
  "logo": {
    "light": "/images/logo-light.svg",
    "dark": "/images/logo-dark.svg",
    "href": "https://use-charter.dev"
  },
  "favicon": "/images/favicon.svg",
  "description": "AI-agent readiness, scored. Charter is an offline-first scanner that gives every repo an AI-agent readiness score and shows you how to fix it.",
  "seo": {
    "indexing": "navigable",
    "metatags": {
      "og:site_name": "Charter",
      "og:title": "Charter — AI-Agent Readiness Scanner",
      "og:description": "Offline-first scanner that scores a repo's AI-agent readiness and shows you how to fix it.",
      "twitter:card": "summary_large_image"
    }
  },
  "search": {
    "prompt": "Search Charter docs..."
  },
  "navbar": {
    "links": [
      { "label": "GitHub", "href": "https://github.com/use-charter/charter" }
    ],
    "primary": {
      "type": "button",
      "label": "Get Started",
      "href": "/quickstart"
    }
  },
  "footer": {
    "socials": {
      "github": "https://github.com/use-charter/charter"
    }
  },
  "redirects": [],
  "navigation": {
    "tabs": [
      {
        "tab": "Docs",
        "anchors": [
          {
            "anchor": "Getting Started",
            "icon": "rocket",
            "pages": ["introduction", "installation", "quickstart"]
          },
          {
            "anchor": "How Charter Works",
            "icon": "compass",
            "pages": [
              "concepts/agent-readiness-model",
              "concepts/scoring-and-caps",
              "concepts/fix-engine",
              "concepts/mcp-safety-model",
              "concepts/suppression-governance"
            ]
          },
          {
            "anchor": "Guides",
            "icon": "route",
            "pages": [
              "how-to/adopt-in-existing-repo",
              "how-to/run-in-github-actions",
              "how-to/use-charter-fix-safely",
              "how-to/suppress-a-finding",
              "how-to/investigate-mcp-findings",
              "how-to/pre-commit-hook"
            ]
          },
          {
            "anchor": "Configuration",
            "icon": "sliders-horizontal",
            "pages": ["config/charter-yaml", "config/policy-profiles"]
          },
          {
            "anchor": "CI & GitHub Action",
            "icon": "git-branch",
            "pages": ["ci/github-action"]
          },
          {
            "anchor": "Design Philosophy",
            "icon": "lightbulb",
            "pages": ["design-philosophy"]
          },
          {
            "anchor": "Changelog",
            "icon": "history",
            "pages": ["changelog"]
          }
        ]
      },
      {
        "tab": "CLI Reference",
        "groups": [
          {
            "group": "Overview",
            "icon": "terminal",
            "pages": ["cli/overview"]
          },
          {
            "group": "Commands",
            "icon": "code-2",
            "pages": [
              "cli/doctor",
              "cli/init",
              "cli/fix",
              "cli/report",
              "cli/explain",
              "cli/suppress",
              "cli/version"
            ]
          }
        ]
      },
      {
        "tab": "Rules",
        "groups": [
          {
            "group": "Overview",
            "icon": "layout-grid",
            "pages": ["rules/overview"]
          },
          {
            "group": "Context",
            "icon": "file-text",
            "pages": [
              "rules/AE-CTX-001",
              "rules/AE-CTX-002",
              "rules/AE-CTX-004",
              "rules/AE-CTX-006"
            ]
          },
          {
            "group": "Secrets",
            "icon": "lock",
            "pages": ["rules/AE-SEC-001", "rules/AE-SEC-002"]
          },
          {
            "group": "MCP Safety",
            "icon": "shield",
            "pages": ["rules/AE-MCP-001", "rules/AE-MCP-002", "rules/AE-MCP-003"]
          },
          {
            "group": "Agent Config",
            "icon": "bot",
            "pages": ["rules/AE-CC-001", "rules/AE-CC-002"]
          },
          {
            "group": "Environment & CI",
            "icon": "wrench",
            "pages": ["rules/AE-ENV-001", "rules/AE-CI-002"]
          },
          {
            "group": "Testing & Autonomy",
            "icon": "flask-conical",
            "pages": ["rules/AE-TEST-001", "rules/AE-AUTO-001"]
          },
          {
            "group": "Governance",
            "icon": "scale",
            "pages": [
              "rules/AE-SUPPRESS-001",
              "rules/AE-SUPPRESS-002",
              "rules/AE-SUPPRESS-003"
            ]
          }
        ]
      }
    ]
  }
}
```

### DEPLOY.md patch needed (Task 2.7 scope)

Fix the one error in `docs/product/DEPLOY.md`:
- Line ~99: change `(proxied — orange cloud)` → `(DNS only — grey cloud)`
- Add note: "Mintlify custom domain must be set to `use-charter.dev` in the Mintlify project dashboard (not in docs.json)."
- Add note: "Root `use-charter.dev` A record must exist as proxied placeholder (192.0.2.1, orange cloud) for the Worker route to fire."

---

## Resolved Decisions

### Q1 — `generate-rule-pages.ts` approach

**Decision: update the generator + spec files to emit the new customer-facing anatomy.**

The script fully overwrites rule pages on every run (`writeFileSync`, no merge). The `--check` mode does an exact content diff — any manual edit that diverges from the generator output fails `:docs-product-rules`. Any hand-written prose added in Phase 3 gets wiped on the next `moon run :docs`.

**Approach: update both `renderRulePage()` and the 18 spec files.** Spec files become the single source of truth for all rule content — internal and customer-facing.

Changes to `renderRulePage()` (Task 3.1a):
- Remove `> Generated from the internal rule spec...` banner
- Remove `## Related ADRs` and `## Related Evals` output blocks entirely
- Add `**Auto-fixable:**` in the metadata header line (new spec field)
- Rename/reframe section mappings:
  - `## Detection Logic` → `## What triggers it`
  - `## Pass Example` + `## Fail Example` → `## Examples` with `### Passing` / `### Failing` subsections
  - `## Remediation` → `## How to fix`
  - Add `## Why this rule` (new spec field)
  - Add `## Related rules` (new spec field — rule IDs, not ADR numbers)
- Keep `## Edge Cases` (hidden if empty) and `## CLI`
- **Check mode**: verify frontmatter correctness and required heading names exist — not full-content equality (too brittle for prose)

Changes to spec files (Task 3.1b — all 18 `docs/internal/specs/AE-*.md`):
- Add `- Why: <1–2 sentence customer-facing rationale>` field
- Add `- Auto-fixable: Yes — charter fix --rule AE-XXX / No` field
- Add `- Related rules: AE-XXX-001, AE-XXX-002` field (replaces ADR cross-links)
- Rephrase `Detection Logic` in plain customer language (remove pseudocode)
- Enhance `Pass Example` / `Fail Example` with realistic content and real CLI output
- Keep `Related ADRs` / `Related Evals` fields in spec files (internal record) — they just won't be rendered to product pages any more

After Task 3.1: run `bun scripts/generate-rule-pages.ts` to regenerate all 18 pages with the new anatomy. Tasks 3.2–3.3 then do manual prose polish where the spec content needs further customer-facing refinement.

### Q2 — Install paths

**Decision: show all four real paths; no "launch-gated" language.**

From GoReleaser config and the release pipeline:

| Path | Availability |
|---|---|
| GitHub Releases binary (darwin/linux/windows, amd64+arm64) | v1.0 tag |
| `brew install use-charter/tap/charter` | v1.0 — tap goes public at Slice 20 |
| `go install go.use-charter.dev/charter/cmd/charter@latest` | v1.0 — requires vanity host (CF-4, Slice 18/19 deploy) |
| Build from source | Always |

`installation.mdx` (Task 2.2) shows all four paths marked by availability. No pre-launch caveats in prose.

### Q3 — Changelog location

**Decision: changelog lives in docs** at `/changelog`, not the landing site. Already reflected as an anchor in the target `docs.json`.

### Q4 — Design Philosophy location

**Decision: top-level anchor** in the Docs tab. Already reflected in the target `docs.json`. Mintlify supports single-page anchors — `"pages": ["design-philosophy"]` is valid schema.

---

## What This Plan Changes (vs. original scope)

- `scripts/generate-rule-pages.ts` — `renderRulePage()` template updated (in-scope, Task 3.1a)
- `docs/internal/specs/AE-*.md` — all 18 spec files gain new customer-facing fields (in-scope, Task 3.1b; spec files are the source of truth for product pages)
- `docs/product/DEPLOY.md` — one-line error fix (proxied → grey cloud for optional CNAME) + two notes added (Task 2.7 scope)

---

## What This Plan Does NOT Change

- No changes to `docs/internal/` architecture docs, ADRs, RFCs, playbooks, or runbooks
- No changes to Go source, schemas, or test fixtures
- No changes to `scripts/validate-product-docs.ts` behavior
- No deployment infrastructure changes — Worker code in `DEPLOY.md` is correct as-is
