# Slice 19 — Landing Page Specification

**Document type:** Specification (WHAT and WHY, not HOW)  
**Status:** Locked & Validated  
**Date:** 2026-06-10  
**Last validated:** 2026-06-10 (all asset paths and source references verified against live repo)

---

## Goal

Ship a conversion-focused landing page (`use-charter.dev/`) that introduces Charter to engineers evaluating AI-agent readiness and DevOps leads managing repo compliance at scale. The page drives primary conversion (CLI install via `brew install use-charter/tap/charter`) and secondary conversion (GitHub stars, community engagement) while deferring Phase 2 dashboard signup to a waitlist.

---

## Problem Statement

Charter has a mature product (CLI, 18 rules, fix engine, suppression governance) and excellent reference documentation (Mintlify docs, rule pages, architecture). The missing piece is a customer-facing landing page that:

1. Articulates the problem: "Teams adopting coding agents fail not because the model is bad, but because the repo isn't ready"
2. Shows the solution in under 10 seconds (terminal card with real `charter doctor` output)
3. Makes installing trivial (copy-to-clipboard `brew install ...`)
4. Drives awareness via GitHub stars and community
5. Captures interest in the future cloud dashboard (Phase 2) without over-promising

The landing page must integrate seamlessly with the existing product identity: terminal-dashboard aesthetic, dark-first branding, semantic design tokens, and offline-first ethos.

---

## Design Baselines

Validated against high-end visual design, design-taste, and frontend design skills.

| Dial | Value | Meaning |
|------|-------|---------|
| DESIGN_VARIANCE | 7/10 | Asymmetric layouts, intentional composition, grid-breaking |
| MOTION_INTENSITY | 5/10 | CSS transitions only; no heavy JS choreography |
| VISUAL_DENSITY | 3/10 | Art-gallery mode; generous whitespace; premium feel |

**Critical anti-patterns — these are banned:**

- Emojis anywhere (copy, decoration, icons) — use Phosphor Light SVG icons instead
- More than 1 accent color (#2563EB blue; saturation <80%)
- Inter font — banned. Ruda exclusively for sans
- `height: 100vh` — use `min-height: 100dvh` for iOS Safari stability
- Generic cards without earned elevation (card used only where depth communicates hierarchy)
- Centered hero with gradient blob and uniform three-card row — banned

---

## Scope

### In scope

- **Landing page:** Single long-form HTML page served at `/` on `use-charter.dev`
- **Sections:** 9 sections (Hero → Footer) with locked copy
- **Visual design:** Terminal-dashboard aesthetic (dark-first, Ruda + Atkinson Mono, score-zone colors)
- **Technology:** Astro v6 SSG + vanilla CSS + Cloudflare Pages as `LANDING_ORIGIN`
- **Performance:** LCP < 1.5s, FCP < 1s, INP < 200ms, CLS < 0.1; JS ≤ 150kb gzip, CSS ≤ 30kb gzip
- **Accessibility:** WCAG 2.2 AA (semantic HTML, keyboard nav, reduced-motion, contrast)
- **Conversion:** Copy-to-clipboard button, GitHub link, waitlist email capture
- **Integration:** Wire as Cloudflare Pages app; update Worker `LANDING_ORIGIN` in `docs/product/DEPLOY.md`

### Out of scope

- Multi-page marketing site — use landing + link to Mintlify
- Cloud dashboard signup — Phase 2; waitlist is deferred, non-blocking capture
- User authentication
- Fabricated social proof logos — only permissioned logos or explicit placeholder markup
- Analytics, tracking, or third-party scripts

---

## Information Architecture & Copy

### Section 1: Hero (above fold)

**Purpose:** Outcome-first headline + instant install

| Element | Content |
|---------|---------|
| H1 | "AI-agent readiness, scored." |
| Subheadline | "Charter is an offline-first CLI that audits any repo against 18 rules and returns a deterministic 0–100 score — in under 2 seconds, with no data leaving your machine." |
| Primary CTA | Copy-to-clipboard button: `brew install use-charter/tap/charter` |
| Secondary CTA | `View on GitHub` → `https://github.com/use-charter/charter` |
| Hero visual | Terminal card showing real `charter doctor` output → `94/100 Ship-ready` (green zone) |
| Mobile behavior | Terminal card reflows with horizontal scroll; copy button stacks below headline |

Hero layout: split-screen on desktop (copy left, terminal right); stacked on mobile. NOT centered hero.

### Section 2: Problem

**Purpose:** Name the pain; establish credibility

**Locked copy (verbatim from `docs/product/introduction.mdx`):**

> "Teams adopting coding agents don't fail because the model is bad — they fail because the repo isn't ready.
> Unpinned MCP servers. Secrets visible to agents. No verification command. Outdated AGENTS.md. These aren't policy violations; they're operational gaps that block agent adoption."

### Section 3: How It Works

**Purpose:** Scan → Score → Fix triad visual

**Structure:** Three-step horizontal flow (mobile: vertical stack)

| Step | Label | Content |
|------|-------|---------|
| 1 | Scan | Deterministic, offline, no LLM, no network calls |
| 2 | Score | Formula: `max(0, 100 − B×20 − H×10 − M×4 − L×1)` + four score zones |
| 3 | Fix | Diff-first, never auto-fixes secrets, human review always required |

**Screenshots to use** (existing files in `docs/product/images/screenshots/`):
- `doctor-overview.webp` (scoring visual — Step 2)
- `fix-dry-run.webp` (diff output — Step 3)
- `doctor-tty.webp` (full scan output — Step 1)

### Section 4: Value Props

**Purpose:** Differentiation via readiness axes (not CLI commands)

Four cards. Each names the axis, explains why it matters to agent-adopting teams, and mentions the corresponding CLI command as an implementation detail only.

| Card | Axis | Copy direction |
|------|------|----------------|
| 1 | Context | "Can agents see secrets? Outdated AGENTS.md? Missing .env.example?" |
| 2 | Safety | "Are credentials in agent-visible paths? Policy enforcement?" |
| 3 | Operability | "Is there a verify command? Pre-commit integration? CI gating?" |
| 4 | Governance | "Can teams waive findings? Are waivers tracked and audited?" |

### Section 5: Trust / Determinism

**Purpose:** Earn developer credibility

**Locked copy — verbatim from `docs/internal/architecture/charter-architecture-2026.md` lines 57–62 (Ten Commitments):**

Select these four (exact text):

1. "Never send data anywhere without explicit opt-in."
2. "Never call an LLM — all findings are deterministic."
3. "Every finding has a rule ID, evidence, and fix suggestion."
4. "Every release is signed (cosign) with SLSA Level 3 provenance."

### Section 6: Social Proof

**Purpose:** Adoption signal

**Components:**
- GitHub stars count (fetched at build time from `https://api.github.com/repos/use-charter/charter`; cached in static output)
- Vendor compatibility icons: Claude Code, ChatGPT, Grok (SVGs confirmed in `docs/product/images/icons/`), plus Cursor, Windsurf, Copilot, Gemini, Codex (placeholders — these SVGs do NOT exist yet; must render as text-only badges until SVGs are sourced and permissioned)
- **Strict rule:** Do NOT fabricate "Used by X teams at [Company logos]". GitHub stars + confirmed vendor icons only.

### Section 7: CI / SARIF Strip

**Purpose:** Reach DevOps-lead persona

**Headline:** "Gate pull requests on agent-readiness"

**Locked snippet** (exact text from `docs/product/how-to/run-in-github-actions.mdx`):

```yaml
name: Charter
on:
  pull_request:
  push:
    branches: [main]
permissions:
  actions: read
  contents: read
  security-events: write
jobs:
  charter:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
      - uses: use-charter/charter-action@v1
        with:
          threshold: "80"
```

**CTA:** "Read the GitHub Actions guide" → `https://use-charter.dev/docs/how-to/run-in-github-actions`

Mention SARIF 2.1.0 output + GitHub Code Scanning integration.

### Section 8: Final CTA

**Purpose:** Convert

| Element | Content |
|---------|---------|
| Install command | Repeat copy button: `brew install use-charter/tap/charter` |
| Primary CTA | "Read the docs" → `https://use-charter.dev/docs` |
| Secondary CTA | "Star on GitHub" → `https://github.com/use-charter/charter` |
| Tertiary (deferred) | "Notify me about the dashboard" → waitlist email capture form (endpoint TBD — see Critical Dependencies) |

### Section 9: Footer

**Purpose:** Links + community

| Group | Items |
|-------|-------|
| Product | Docs, Rules reference |
| Project | GitHub repo, Releases |
| Community | Discord (mark as placeholder if link unavailable) |
| Legal | Apache-2.0 license, Copyright notice |

All external links: `rel="noopener noreferrer"`.

---

## Visual Direction

### Design System

**Source of truth:** `docs/internal/designs/DESIGN-TOKENS.md` and `docs/internal/designs/brand/README.md`. Do NOT re-invent; extract exactly.

**Color palette:**

| Token | Value | Use |
|-------|-------|-----|
| Primary surface | `#0D1117` | Page background (dark-first) |
| Primary accent | `#2563EB` | Buttons, links, active states |
| Success (score ≥80) | `#4ade80` | Terminal score green zone |
| Warning (score 60–79) | `#fbbf24` | Terminal score amber zone |
| Danger (score <60) | `#f87171` | Terminal score red zone |

**Typography (weights confirmed in `DESIGN-TOKENS.md`):**

| Use | Family | Weight |
|-----|--------|--------|
| H1, H2, wordmark | Ruda | 700 / 800 (max is 800) |
| Body, subheads | Ruda | 400 / 500 |
| Code, terminal, commands | Atkinson Hyperlegible Mono | 400 / 500 |
| Metadata, rule IDs (small) | IBM Plex Mono | 400 / 500 |
| Status labels (tiny, accent) | Share Tech | 400 |

**Fonts sourcing:** No woff2 files exist in the repo directly. They are Latin-subset woff2 base64-embedded, generated by `bun scripts/generate-report-fonts.ts`. Run this script to generate `internal/render/html/assets/fonts.css`, then extract the `@font-face` blocks for self-hosting in `web/src/styles/fonts.css`. Do NOT link Google Fonts CDN (offline-first ethos).

**Hero visual:** Terminal card showing real `charter doctor` output (OUTCOME = 94/100 Ship-ready). CSS-only animation: fade-in + slide-up on scroll reveal. Static screenshot fallback: `docs/product/images/screenshots/doctor-overview.webp`.

### Interaction Timings

| Interaction | Duration | Easing |
|-------------|----------|--------|
| Button scale feedback (`active` state, scale 0.98) | 100–150ms | `cubic-bezier(0.55, 0, 1, 0.45)` |
| Copy button "Copied!" text change | 100–150ms instant, revert after 2000ms | — |
| Form input focus border transition | 200–300ms | `cubic-bezier(0.16, 1, 0.3, 1)` |
| Section fade-up scroll reveal | 300–500ms | `cubic-bezier(0.16, 1, 0.3, 1)` |
| Hover scale up (scale 1.02) | 200ms | `cubic-bezier(0.16, 1, 0.3, 1)` |

### Component States (mandatory)

**Copy button:**
- `default` → `hover` (scale 1.02, 200ms) → `active` (scale 0.98, 100ms) → `copied` (text "Copied!", icon change, 2s timeout, then revert)

**Waitlist form:**
- `empty` → `focus` (input border transition, 200ms) → `loading` (submit disabled, CSS spinner) → `success` (toast "Check your email!") / `error` (toast with error message)

**Section reveals:**
- `invisible` (opacity: 0, transform: translateY(4rem)) → `visible` (opacity: 1, transform: translateY(0), 300–500ms)
- Disabled entirely under `prefers-reduced-motion: reduce`

### Responsive

- **Mobile-first breakpoints:** 320px, 375px, 768px, 1024px, 1440px, 1920px
- **Terminal card:** `overflow-x: auto` inside card; `overflow-x: hidden` on `<body>` — card scrolls, page doesn't
- **Touch targets:** All buttons and links ≥ 44×44px
- **Hero full-height:** `min-height: 100dvh` (never `height: 100vh`)
- **Layout:** CSS Grid + media queries; never flexbox percentage math

### Accessibility (WCAG 2.2 AA)

| Requirement | Specifics |
|-------------|-----------|
| Semantic HTML | `<header>`, `<main>`, `<section aria-labelledby="…">`, `<footer>` |
| Heading hierarchy | H1 (once, hero) → H2 (each section) → H3 (subsections only when needed) |
| Images | Every `<img>` and `<picture>` has descriptive `alt=""` |
| Focus states | Minimum 3px outline OR 2px shadow with 2px offset; contrast ≥ 3:1 against background |
| Reduced motion | `prefers-reduced-motion: reduce` disables all CSS animations; page fully functional without motion |
| Color contrast | Normal text ≥ 4.5:1; large text ≥ 3:1 |
| Keyboard nav | All interactive elements (buttons, links, inputs) reachable via Tab; no traps; logical order |

### Anti-Template Checklist

Before marking any component done:
- [ ] Does NOT look like a default Tailwind or generic template?
- [ ] Does it have intentional hover/focus/active states?
- [ ] Does it use hierarchy rather than uniform emphasis?
- [ ] Would this look credible in a real product screenshot?

---

## Technology Stack

**All decisions validated against official docs:**
- Astro v6: https://docs.astro.build/
- Bun: https://bun.sh/docs
- Cloudflare Pages: https://developers.cloudflare.com/pages/

### Framework

**Astro v6 (SSG)**
- `output: 'static'` in `astro.config.mjs`
- Zero JS shipped by default; islands add minimal JS where needed
- `<Picture />` from `astro:assets` for automatic AVIF/WebP/fallback image optimization
- Rationale: ships zero JS to browser by default; ideal for static landing page

**Vanilla CSS + CSS Custom Properties**
- All tokens extracted verbatim from `docs/internal/designs/DESIGN-TOKENS.md`
- CSS architecture: `design-tokens.css` → `global.css` → `sections/*.css` → `islands.css`
- Responsive: CSS Grid + media queries
- No Tailwind, no PostCSS, no CSS-in-JS
- Rationale: 4–8kb gzip vs 12–18kb Tailwind; direct token reuse; no duplication

**Package manager: Bun exclusively**
- `bun create astro` to scaffold
- `bun install` to install dependencies
- `bun run build` to produce static output
- `bun run dev` to run dev server (Astro default: `http://localhost:4321`)
- No npm, no npx for project commands (use `bunx` instead)

### Hosting

**Cloudflare Pages**
- Build command: `bun run build` (from `web/package.json`)
- Output directory: `dist/`
- Zero Node.js runtime (pure static)

**Worker routing:** Existing Cloudflare Worker (`action/` dir or root wrangler config) updated to proxy `/` to `LANDING_ORIGIN` (Cloudflare Pages URL). Existing `/docs/*` and `/rules/*` routes to Mintlify remain unchanged.

**Monorepo:** `web/` is a Moon project (`web/moon.yml` already committed). Add `dev`, `build`, `check` tasks.

### Performance Budgets (hard limits — all gzipped)

| Asset type | Budget |
|------------|--------|
| JavaScript | ≤ 150kb (Astro default: 0kb; islands add minimal) |
| CSS | ≤ 30kb total; ≤ 4kb inline critical CSS in `<head>` |
| Images | ≤ 50kb total (all hero + section images combined) |
| Fonts | ≤ 30kb |

**Core Web Vitals targets:**

| Metric | Target |
|--------|--------|
| LCP | < 1.5s |
| FCP | < 1.0s |
| INP | < 200ms |
| CLS | < 0.1 |
| TBT | < 200ms |

**Image strategy:**
- Hero image: `loading="eager" fetchpriority="high"` (above fold)
- Below-fold images: `loading="lazy"`
- All images: explicit `width` + `height` attributes (prevents CLS)
- Format: AVIF primary, WebP fallback, PNG last resort
- Screenshot reuse: use existing WebP from `docs/product/images/screenshots/` as-is

**Font strategy:**
- Self-hosted (no Google Fonts CDN)
- `font-display: swap`
- Preload: Ruda 800 only (heading weight) + Atkinson Mono 400 (code weight)
- Latin subset only

---

## Content Rules

All copy is **locked before implementation begins.** No drift from these sources.

| Section | Source | Lock status |
|---------|--------|-------------|
| H1 | Brand tagline (DESIGN-TOKENS.md) | Locked |
| Hero subheadline | architecture-2026.md commitment language | Locked |
| Problem copy | `docs/product/introduction.mdx` verbatim | Locked |
| Solution steps | Brand motif "Scan · Score · Fix" + architecture-2026.md score formula | Locked |
| Value props axes | Docs IA: Context, Safety, Operability, Governance | Locked |
| Trust strip | architecture-2026.md lines 57–62 (Ten Commitments #1, 2, 5, 6) verbatim | Locked |
| CI snippet | `docs/product/how-to/run-in-github-actions.mdx` exact YAML | Locked |
| CTA copy | "Get started", "View on GitHub", "Read the docs", "Star on GitHub" | Locked |

**Claim rules:**
- No benchmark numbers beyond "<2 seconds" and "50k-file budget" (both verified in architecture docs)
- Every claim derivable from `charter-architecture-2026.md`, `introduction.mdx`, or `quickstart.mdx`
- Severity/score language must match docs exactly
- No fabricated social proof

---

## Critical Dependencies

### 1. Waitlist endpoint (TBD — blocker for Section 8 form)

**Contract:**
- `POST /api/waitlist`
- Request body: `{ "email": "string" }`
- Success: `HTTP 200` + `{ "success": true, "message": "string" }`
- Failure: `HTTP 400/500` + `{ "error": "string" }`
- Must validate + sanitize email server-side
- Must handle duplicate email gracefully (idempotent or 409)
- Must store securely (ops/privacy responsibility)

**Implementation plan behavior while TBD:** Form renders. Submit button works. POST fires to placeholder URL `/api/waitlist`. 4xx response shows generic error toast. No fake success state. Uncomment real endpoint URL when ops provides it.

### 2. GitHub API (build-time fetch)

- URL: `https://api.github.com/repos/use-charter/charter`
- Field: `stargazers_count`
- No auth required (public repo)
- Failure fallback: render "⭐ Star on GitHub" with no count (never hardcode a number)
- Caches in static output; updates at next Pages deploy

### 3. Existing assets (all paths verified against live repo)

| Asset | Verified path | Status |
|-------|--------------|--------|
| Design tokens | `docs/internal/designs/DESIGN-TOKENS.md` | ✅ Exists |
| Brand meta.html | `docs/internal/designs/brand/meta.html` | ✅ Exists |
| og:image | `docs/internal/designs/brand/og.svg` | ✅ Exists |
| Favicon | `docs/internal/designs/brand/favicon.svg` | ✅ Exists |
| Screenshots | `docs/product/images/screenshots/*.webp` (13 files) | ✅ Exists |
| claude-ai icon | `docs/product/images/icons/claude-ai.svg` | ✅ Exists |
| chatgpt icon | `docs/product/images/icons/chatgpt.svg` | ✅ Exists |
| grok icon | `docs/product/images/icons/grok.svg` | ✅ Exists |
| Cursor icon | `docs/product/images/icons/cursor.svg` | ❌ Missing — render as text badge |
| Windsurf icon | `docs/product/images/icons/windsurf.svg` | ❌ Missing — render as text badge |
| Copilot icon | `docs/product/images/icons/copilot.svg` | ❌ Missing — render as text badge |
| Gemini icon | `docs/product/images/icons/gemini.svg` | ❌ Missing — render as text badge |
| Codex icon | `docs/product/images/icons/codex.svg` | ❌ Missing — render as text badge |
| Font woff2 files | Generated by `bun scripts/generate-report-fonts.ts` (output: `internal/render/html/assets/fonts.css`) | ✅ Script exists; run to extract @font-face blocks |

### 4. Cloudflare Worker file

The Worker file location must be confirmed before Phase 9 (deployment). Check repo root for `wrangler.toml` or `action/` for the composite action. Worker routing update in Phase 9 modifies the `LANDING_ORIGIN` route.

### 5. Meta tags (required in Base.astro)

```html
<title>Charter — AI-agent readiness, scored.</title>
<meta name="description" content="Charter is an offline-first CLI that audits any repo against 18 rules and returns a deterministic 0–100 score in under 2 seconds. No data leaves your machine." />
<meta property="og:title" content="Charter — AI-agent readiness, scored." />
<meta property="og:description" content="Offline-first CLI. Deterministic 0–100 score. 18 rules. No LLM calls. No network. Works with Claude Code, Codex, Cursor, Windsurf, Copilot, Gemini." />
<meta property="og:image" content="/og.svg" />
<meta property="og:type" content="website" />
<link rel="canonical" href="https://use-charter.dev/" />
<link rel="icon" href="/favicon.svg" type="image/svg+xml" />
```

---

## Acceptance Criteria

All items must pass before marking "Done."

**Responsive & Performance**
- [ ] Page renders at 320, 375, 768, 1024, 1440, 1920px without horizontal scroll (except intentional terminal card overflow)
- [ ] All CTA buttons and links ≥ 44×44px touch targets
- [ ] Hero image: `loading="eager" fetchpriority="high"` present in output HTML
- [ ] Lighthouse Performance ≥ 90, Accessibility ≥ 90 (`bunx lighthouse http://localhost:4321`)
- [ ] LCP < 1.5s, FCP < 1.0s, INP < 200ms, CLS < 0.1 (Lighthouse report)
- [ ] Gzipped budgets verified: CSS ≤ 30kb, JS ≤ 150kb, images ≤ 50kb, fonts ≤ 30kb

**Accessibility (WCAG 2.2 AA)**
- [ ] `bun run check` (Astro type-check) exits 0
- [ ] `bunx @axe-core/cli http://localhost:4321` — zero violations
- [ ] Manual Tab test: all interactive elements reachable, no keyboard traps, logical order
- [ ] Focus visible: ≥ 3px outline or shadow on all buttons/links/inputs
- [ ] Heading hierarchy confirmed: exactly one H1, H2 per section, H3 only as subsections
- [ ] Color contrast: ≥ 4.5:1 normal text, ≥ 3:1 large text (Chrome DevTools verified)
- [ ] All `<img>` and `<picture>` elements have non-empty, descriptive `alt` attributes
- [ ] Form label positioned above input, error text below input, helper text in markup
- [ ] `@media (prefers-reduced-motion: reduce)` disables animations; page is fully functional without motion
- [ ] VoiceOver (Mac) or NVDA (Windows): page structure readable, CTAs announced, form labels read correctly

**SEO & Metadata**
- [ ] All required meta tags present in `<head>` (title, description, og:*, canonical, favicon)
- [ ] All external links have `rel="noopener noreferrer"`
- [ ] GitHub stars count rendered (or fallback text, never hardcoded number)
- [ ] Zero console errors or warnings in `bun run build` output and browser devtools

**Assets & Brand**
- [ ] Fonts self-hosted (served from `web/public/fonts/`); no CDN link in HTML
- [ ] All images use `<Picture />` with AVIF primary, WebP fallback
- [ ] All images have explicit `width` + `height` attributes
- [ ] Terminal card uses CSS animation only (no JS animation library)
- [ ] Icons are SVG or Phosphor Light (zero emojis in markup or copy)

**Functionality**
- [ ] Copy button: copies command to clipboard on desktop; falls back to text selection on iOS (no `execCommand` errors surfacing to user)
- [ ] Form validation: invalid email shows red border + error message below input; submit disabled until valid
- [ ] Form submission: POST fires, button disabled + spinner shows during request, success/error toast shows on response
- [ ] All links navigate: GitHub, Mintlify docs, CI guide, license

**Build & Deployment**
- [ ] `bun run build` from `web/` outputs clean `dist/` with no errors
- [ ] Cloudflare Pages deployment succeeds; `dist/index.html` served at root
- [ ] Worker routing verified: `curl -sI https://use-charter.dev/` → 200; `/docs` → 30x to Mintlify; `/rules` → 30x to Mintlify
- [ ] HTTPS enforced; no certificate warnings on `use-charter.dev`

---

## Delivery Model

Delivered as a series of atomic, GPG-signed commits on `main`. Implementation is sub-agent driven — see implementation plan for agent dispatch table and phase execution strategy.

**Spec status:** Locked (2026-06-10)  
**Ready for implementation:** Yes — all asset paths verified, all copy locked, all dependencies documented
