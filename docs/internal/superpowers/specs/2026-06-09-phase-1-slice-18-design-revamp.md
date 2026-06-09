# Phase 1 Slice 18 Design Revamp — Product Docs Overhaul

## Goal

Transform the Slice 18 Mintlify scaffold from a developer-internal draft into a genuinely customer-facing product documentation site. The original scaffold was written during active development and contained engineering artifacts visible to customers: ADR citations, internal file paths, generator banners, launch-gated language, and spec pseudocode. This revamp strips all of that and rebuilds the documentation as a product.

## Problem statement

The Slice 18 scaffold passed `moon run :check` and fulfilled the technical contract (all 18 `/rules/AE-*` pages resolve, `helpUri`s work) but failed as customer-facing documentation for five reasons:

1. **Internal leakage** — `> Generated from the internal rule spec. Edit docs/internal/specs/...` banners on every rule page; `## Related ADRs: ADR-XXXX` sections; `docs/internal/` paths in `doctor.mdx`; generator script references in `rules/overview.mdx`; inline `(per ADR-0008)` citations in rule body prose.
2. **Launch-frozen language** — `quickstart.mdx` and `run-in-github-actions.mdx` contained pre-launch caveats ("launch-gated", "not live yet") written at implementation time and never cleaned up.
3. **Reference-first navigation** — the original flat tab structure dropped users into CLI commands with no orientation, no product narrative, and no journey.
4. **Missing content** — no installation page, no design philosophy, no fix engine explanation, no pre-commit guide, no policy profiles page, no changelog.
5. **Rule pages as stubs** — `## Detection Logic: inspect canonical agent-visible files...` was spec pseudocode, not customer documentation.

## Scope

### In scope

- Strip all internal engineering artifacts from `docs/product/` (banners, ADR refs, internal paths, launch-gated language)
- Redesign `docs.json` navigation: four-zone IA using Mintlify anchors/groups with Lucide icons
- Add six new pages: `installation.mdx`, `design-philosophy.mdx`, `concepts/fix-engine.mdx`, `how-to/pre-commit-hook.mdx`, `config/policy-profiles.mdx`, `changelog.mdx`
- Rewrite `introduction.mdx` problem-first
- Rewrite all 18 rule pages with customer-facing anatomy (Why / What triggers it / Examples / How to fix / Score impact / Edge cases / Related rules)
- Update `scripts/generate-rule-pages.ts` template to emit new anatomy; update `--check` to structural validation
- Add `Why`, `Auto-fixable`, `Related rules` fields to all 18 `docs/internal/specs/AE-*.md`
- Wire brand assets: `logo-light.svg` (#0D1117), `logo-dark.svg` (#FFFFFF), `favicon.svg`
- Fix `DEPLOY.md`: grey-cloud error on CNAME, Mintlify dashboard note, root A placeholder record, step-by-step setup guide
- Update root `README.md`: remove launch-gated language, clean install section
- Update `docs/product/README.md`: contributor guide for the Mintlify site

### Out of scope

- No Go source changes
- No schema changes
- No new rules
- No deployment execution (Phase B setup is documented, not performed)
- No Slice 19 landing site work

## Information architecture decision

**Decision: four-zone IA with Mintlify anchors for the Docs tab.**

| Zone | Mintlify element | Icon |
|---|---|---|
| Getting Started | anchor | `rocket` |
| How Charter Works | anchor | `compass` |
| Guides | anchor | `route` |
| Configuration | anchor | `sliders-horizontal` |
| CI & GitHub Action | anchor | `git-branch` |
| Design Philosophy | anchor | `lightbulb` |
| Changelog | anchor | `history` |

CLI Reference and Rules tabs use `groups` with icons (reference-style density, not zone differentiation).

**Why anchors for Docs, groups for CLI/Rules:** Mintlify's `anchor` type renders with prominent sidebar icons — right for the user-journey Docs tab where each zone is semantically distinct. `group` renders as a lighter section header — right for reference tabs where the goal is scannable density. Using the same element everywhere loses the visual hierarchy.

**Icon library:** Lucide (`"icons": { "library": "lucide" }`). Modern, consistent with developer tools, has all needed icons.

## Generator architecture decision

**Decision: update generator template + spec files; structural check mode.**

The original generator fully overwrote rule pages on every run and used full content equality in `--check`. This made hand-maintaining prose impossible — any edit would be wiped or fail CI.

New approach:
- `renderRulePage()` emits the new customer anatomy from spec fields
- Spec files add three new fields: `Why`, `Auto-fixable`, `Related rules`
- `--check` mode verifies structural integrity (file exists, correct title, CLI section present) — not content equality
- Rule pages are bootstrapped by the generator and hand-maintained thereafter

**Why not keep full equality check:** spec one-liners (e.g., `Detection logic: inspect canonical agent-visible files...`) cannot carry the paragraph-level customer prose that makes good documentation. Full equality check would either force spec files to become customer-doc files (polluting the internal engineering record) or force constant sync between specs and product pages (high friction, low value).

## Cloudflare deployment decision

**Primary path:** Cloudflare Worker proxies `/docs/*` and `/rules/*` on the `use-charter.dev` zone to `charter.mintlify.dev`. Mintlify custom domain is set to `use-charter.dev` in the dashboard.

**DNS records:**
- `A use-charter.dev → 192.0.2.1` (proxied, orange cloud) — placeholder so Worker routes fire
- `CNAME docs → cname.mintlify.builders` (DNS only, grey cloud) — optional direct subdomain access

**Error in original DEPLOY.md:** the optional CNAME was marked "(proxied — orange cloud)". This is wrong — orange-clouded CNAME to Mintlify terminates TLS at Cloudflare, breaking Mintlify's certificate provisioning. Fixed in this revamp.

## Success criteria (all met)

- `grep -r "docs/internal\|ADR-0\|launch-gated\|Generated from the internal" docs/product/` → empty
- All 27 `docs.json` navigation entries resolve to real `.mdx` files
- `moon run :check` green
- Six new pages exist and contain no internal references
- All 18 rule pages have the new anatomy with customer-facing prose
- `bun scripts/generate-rule-pages.ts --check` passes (structural check)
- `logo-light.svg` (#0D1117), `logo-dark.svg` (#FFFFFF), `favicon.svg` present in `docs/product/images/`
- `DEPLOY.md` is a step-by-step setup guide with correct DNS guidance

## Commits

- `ccde9d1` — `docs(rules): add customer-facing fields to rule specs and update generator anatomy`
- `188e6b4` — `docs: Slice 18 product docs overhaul — internal-clean, four-zone IA, rule anatomy`
- `d76353f` — `docs: expand docs/product/README — contributor guide for the Mintlify site`

## References

- `tasks/plan.md` — detailed phase/task breakdown with decisions and tradeoffs
- `docs/internal/superpowers/plans/2026-06-09-phase-1-slice-18-revamp.md` — execution record
- `docs/internal/superpowers/specs/2026-06-09-phase-1-slice-18-design.md` — original Slice 18 design (what was built before this revamp)
- `docs/internal/superpowers/plans/2026-06-09-phase-1-slice-18.md` — original Slice 18 implementation plan
