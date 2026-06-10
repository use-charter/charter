# Slice 19 — Landing Page Specification

**Document type:** Specification (WHAT and WHY, not HOW)  
**Status:** Locked & Validated Against All Official Docs  
**Date:** 2026-06-10  
**Last Updated:** 2026-06-10 (post-skill validation)

## Goal

Ship a conversion-focused landing page (`use-charter.dev/`) that introduces Charter to engineers evaluating AI-agent readiness and DevOps leads managing repo compliance at scale. The page drives primary conversion (CLI install via `brew install use-charter/tap/charter`) and secondary conversion (GitHub stars, community engagement) while deferring Phase 2 dashboard signup to a waitlist.

## Problem Statement

Charter has a mature product (CLI, 18 rules, fix engine, suppression governance) and excellent reference documentation (Mintlify docs, rule pages, architecture). The missing piece is a customer-facing landing page that:
1. Articulates the problem ("Teams adopting coding agents fail not because the model is bad, but because the repo isn't ready")
2. Shows the solution in under 10 seconds (terminal card with real `charter doctor` output)
3. Makes installing trivial (copy-to-clipboard `brew install ...`)
4. Drives awareness via GitHub stars and community
5. Captures interest in the future cloud dashboard (Phase 2) without over-promising

The landing page must integrate seamlessly with the existing product identity: terminal-dashboard aesthetic, dark-first branding, semantic design tokens, and offline-first ethos.

## Design Baselines

**Validated Against High-End Visual Design & Design-Taste Skills:**

- **DESIGN_VARIANCE:** 7/10 (asymmetric layouts, intentional composition, grid-breaking)
- **MOTION_INTENSITY:** 5/10 (CSS transitions only, no heavy JS choreography)
- **VISUAL_DENSITY:** 3/10 (art-gallery mode, generous whitespace, premium feel)

**Critical Anti-Patterns:**
- No emojis (use Phosphor Light icons or SVG instead)
- Max 1 accent color (#2563EB, <80% saturation)
- No Inter font (Ruda exclusively)
- No h-screen (use min-h-[100dvh] for iOS Safari stability)
- No generic cards without elevation justification
- No center-biased hero (asymmetric/split-screen preferred)

## Scope

### In scope

- **Landing page:** Single long-form HTML page served at `/` on `use-charter.dev`
- **Page sections:** 9 sections (Hero → Footer) with locked copy, wireframes at 320/768/1440
- **Visual design:** Terminal-dashboard aesthetic (dark-first, Ruda/Atkinson Mono, score-zone colors)
- **Technology:** Astro (SSG) + vanilla CSS + Cloudflare Pages as `LANDING_ORIGIN`
- **Performance:** LCP < 1.5s, FCP < 1s, INP < 200ms, CLS < 0.1; JS ≤ 150kb gz, CSS ≤ 30kb gz
- **Accessibility:** WCAG 2.2 AA (semantic HTML, keyboard nav, reduced-motion, contrast)
- **Conversion mechanics:** Copy-to-clipboard button, GitHub link, waitlist email capture
- **Integration:** Wire as Cloudflare Pages app; update Worker `LANDING_ORIGIN` in DEPLOY.md

### Out of scope

- **Multi-page marketing site:** No pricing, no blog, no solutions pages — use landing + link to Mintlify
- **Cloud dashboard signup:** Dashboard is Phase 2 (unbuilt); waitlist is deferred, non-blocking capture
- **User authentication:** No login, no account creation
- **Real social proof logos:** Use GitHub stars + vendor trust badges until logos are permissioned
- **Performance optimization beyond CSS/image:** No third-party scripts, no analytics, no tracking

## Information Architecture & Copy

Single scroll, 9 sections. Copy is drawn from approved brand language and existing product docs (introduction.mdx, quickstart.mdx, architecture-2026.md).

### Section 1: Hero (above fold)

**Purpose:** Outcome-first headline + instant install  
**Headline:** "AI-agent readiness, scored."  
**Subheadline:** "Charter is an offline-first CLI that audits any repo against 18 rules and returns a deterministic 0–100 score — in under 2 seconds, with no data leaving your machine."  
**Primary CTA:** Copy-to-clipboard button with `brew install use-charter/tap/charter`  
**Secondary CTA:** `View on GitHub` (linking to repo)  
**Right/below:** Animated terminal showing real `charter doctor` output → `94/100 Ship-ready` with green zone label  
**Mobile:** Terminal card reflows horizontally; copy button stacks below headline  

### Section 2: Problem

**Purpose:** Name the pain; establish credibility  
**Copy direction:**  
> "Teams adopting coding agents don't fail because the model is bad — they fail because the repo isn't ready.  
> Unpinned MCP servers. Secrets visible to agents. No verification command. Outdated AGENTS.md. These aren't policy violations; they're operational gaps that block agent adoption."

*[Lifted verbatim from introduction.mdx]*

### Section 3: Solution / How It Works

**Purpose:** Scan → Score → Fix triad visual  
**Structure:** Three-step horizontal flow (mobile: vertical stack)  
- **Step 1: Scan** — Deterministic, offline, no LLM, no network calls
- **Step 2: Score** — Formula: `max(0, 100 − B×20 − H×10 − M×4 − L×1)` + four score zones (success/warning/danger + metadata)
- **Step 3: Fix** — Diff-first, never auto-fixes secrets, human review always required

**Visual:** Use existing WebP screenshots from `docs/product/images/screenshots/`:
- doctor-overview.webp (scoring visual)
- fix-dry-run.webp (diff output)
- doctor-tty.webp (full scan output)

### Section 4: Value Props (3–4 cards)

**Purpose:** Differentiation via outcome axes (not CLI commands)  
**Four readiness axes** (from docs IA):
1. **Context** — Agent visibility: "Can agents see secrets? Outdated AGENTS.md? Missing .env.example?"
2. **Safety** — Secret management: "Are credentials in agent-visible paths? Policy enforcement?"
3. **Operability** — Automation readiness: "Is there a verify command? Pre-commit integration? CI gating?"
4. **Governance** — Suppression & approval: "Can teams waive findings? Are waivers tracked and audited?"

**Copy:** Each card names the axis, explains why it matters to teams adopting agents, and mentions the corresponding CLI commands (doctor/fix/suppress/explain) *as implementation details*, not primary selling points.

### Section 5: Trust / Determinism

**Purpose:** Earn developer credibility — why Charter is different  
**Copy:** Short trust strip pulling four of the Ten Commitments:
- "Never calls an LLM — all scoring is deterministic, offline, rule-based"
- "Never sends data without opt-in — your repo stays on your machine"
- "Every finding has a rule ID, evidence, and a fix — no black boxes"
- "Every release is cosign-signed with SLSA L3 provenance — supply-chain verified"

*(Source: architecture-2026.md lines 55–79)*

### Section 6: Social Proof

**Purpose:** Adoption signal  
**Components:**
- GitHub stars count (fetched at build time, cached)
- "Works with Claude Code, Codex, Cursor, Windsurf, Copilot, Gemini" (vendor icons from `docs/product/images/icons/`)
- **Placeholder:** Logo slot for real adoptions — flag as unfilled until permissioned logos exist; spec says "Do NOT fabricate 'Used by X teams at [logos]'"

### Section 7: CI / SARIF Strip

**Purpose:** Reach DevOps-lead persona  
**Headline:** "Gate pull requests on agent-readiness"  
**Copy:** Short band showing GitHub Action snippet + mention of SARIF 2.1.0 output  
**CTA:** Link to `docs/product/how-to/run-in-github-actions.mdx`

### Section 8: Final CTA

**Purpose:** Convert  
**Components:**
- Repeat install command (copy button)
- `Read the docs` → Mintlify (`/docs`)
- `Star on GitHub` → repo URL
- **Deferred:** `Notify me about the dashboard` → waitlist email capture (POST to TBD endpoint)

### Section 9: Footer

**Purpose:** Links + community  
**Components:** Docs, Rules reference, GitHub repo, releases, community/Discord (flag if placeholder), license (Apache-2.0), copyright

---

## Visual Direction

### Design System Alignment

**Mandatory:** Use existing DESIGN-TOKENS.md + brand/README.md exactly. Do NOT re-invent.

**Color palette:**
- **Primary surface:** `#0D1117` (dark background) / `#ffffff` (light, for optional light mode fallback)
- **Primary accent:** `#2563EB` (blue — docs primary)
- **Score zones:** 
  - Success: `#4ade80` (green, 80+ score)
  - Warning: `#fbbf24` (amber, 60–79 score)
  - Danger: `#f87171` (red, <60 score)
- **Semantic tokens:** success/danger/warning/info colors from DESIGN-TOKENS.md; use for badges, charts, state indicators

**Typography:**
- **Headings (H1, H2):** Ruda 700/800 (wordmark uses Ruda 800/900; match it)
- **Body & subheads:** Ruda 400/500 (inherit size scale from tokens)
- **Code/terminal/command blocks:** Atkinson Hyperlegible Mono (canonical primary mono)
- **Metadata/rule IDs/status labels:** IBM Plex Mono (secondary mono, small sizes only)
- **Status-accent tiny labels:** Share Tech (small labels only, ~10px)

**Signature visual:** Terminal card in hero section showing real `charter doctor` output (OUTCOME = 94/100 Ship-ready). Shows the OUTCOME, not the product. Animation (CSS only): fade-in + fade-up on scroll reveal. Fallback: static screenshot.

**Interaction Timings (validated against interaction-design skill):**
- Button scale feedback (`active:scale-[0.98]`): 100-150ms
- Copy button "copied" feedback: 100-150ms  
- Form input focus transition: 200-300ms
- Section fade-up on scroll reveal: 300-500ms
- Easing: Use `cubic-bezier(0.16, 1, 0.3, 1)` for enter (out), `cubic-bezier(0.55, 0, 1, 0.45)` for exit (in)

**Component States (mandatory per design-taste skills):**
- Copy button: default → hover (scale 1.02) → active (scale 0.98) → "copied" (text change, 2s)
- Form: empty → focus (input focus transition) → loading (disabled + spinner) → success/error (toast)
- Section reveals: invisible (opacity 0, translate-y-16) → visible (opacity 1, translate-y-0)

### Responsive Design

- **Mobile-first:** 320px (min width), 375px (typical), 768px (tablet), 1024px (laptop), 1440px (desktop), 1920px (ultra-wide)
- **Terminal card behavior:** Horizontal scroll inside card on 320px; never overflow page
- **Touch-friendly:** Button/link targets ≥ 44×44px
- **Breakpoint strategy:** Prefer CSS Grid + media queries; no wrapper divs solely for layout

### Accessibility (WCAG 2.2 AA)

- **Semantic HTML:** `<header>`, `<main>`, `<section aria-labelledby="...">`, `<footer>`
- **Heading hierarchy:** H1 (once, hero) → H2 (sections) → H3 (subsections only where needed)
- **Images:** Every image has descriptive alt text; screenshots have alt text explaining the output
- **Links:** Descriptive text; avoid "click here", "learn more" vagueness
- **Focus states:** Visible ≥ 3px outline or 2px shadow, 2px offset, contrast ≥ 3:1 against background
- **Reduced motion:** `prefers-reduced-motion: reduce` disables terminal animation, hero parallax, smooth scrolling
- **Color contrast:** All text meets 4.5:1 (normal) or 3:1 (large). Design tokens are pre-verified by DESIGN-TOKENS.md
- **Keyboard nav:** All interactive elements (buttons, links, copy button, form inputs) reachable via Tab; no keyboard traps; logical tab order

### Anti-Template Checklist

Per global design-quality rules, avoid:
- Default centered hero + gradient blob + three uniform cards
- Oversized padding destroying hierarchy
- Generic stock photos (use terminal screenshot)
- Uniform radius/shadow everywhere (vary by depth)
- Safe gray-on-white (use the dark theme intentionally)
- Emoji in copy or decoration

**Required qualities:** hierarchy via scale contrast, intentional rhythm in spacing, terminal card for depth, Ruda + mono pairing with character, semantic color use (not decorative), designed hover/focus states, asymmetric/editorial composition where appropriate, subtle motion (only where it serves).

---

## Technology Stack Decisions

**Official Documentation Sources:**
- Astro v6 Docs: https://docs.astro.build/
- Bun Docs: https://bun.sh/docs
- Cloudflare Pages: https://developers.cloudflare.com/pages/

### Framework & Build

**Astro v6 (Static Site Generation)**
- **Decision:** Astro v6 with vanilla template via `bun create astro`
- **Rationale:** Ships zero JS by default (pure HTML/CSS); Astro v6 auto-renders components to static HTML; SSG fits landing page exactly
- **Astro Output Config:** Set `output: 'static'` in `astro.config.mjs`
- **Image Optimization:** Use `<Picture />` component from `astro:assets` (auto-generates AVIF/WebP with PNG fallback)
- **Why not Next.js:** Hydration overhead for 95%-static page conflicts with 150kb budget
- **Why not Vite Plus:** Pre-1.0 alpha, no Bun support, redundant with Astro

**Vanilla CSS + CSS Custom Properties**
- **Decision:** Vanilla CSS (no frameworks)
- **Rationale:** Direct reuse of DESIGN-TOKENS.md semantic tokens; zero token duplication friction
- **CSS Architecture:** Extract all tokens to `design-tokens.css` (--color-*, --font-*, --space-*, --line-height-*)
- **Responsive:** CSS Grid + media queries (never flexbox percentage math like `w-[calc(33%-1rem)]`)
- **Why not Tailwind v4:** 12–18kb gzipped vs 4–8kb vanilla; exceeds 30kb budget; maintenance burden

**Build Tool:** Vite (automatic via Astro 6)
- No custom Vite config needed
- No PostCSS config needed (vanilla CSS custom properties work natively)

**Package Manager:** Bun (exclusively)
- `bun create astro` initializes project
- `bun install` installs dependencies
- `bun run build` builds for production
- `bun run dev` runs dev server

### Hosting & Deployment

**Cloudflare Pages**
- Build command: `bun run build` (from web/package.json)
- Output directory: `dist/`
- Static HTML/CSS/images output (zero Node.js runtime)
- Domain: `use-charter.dev` (via Worker proxy)
- **wrangler.toml:** Optional for static sites (Pages dashboard UI config sufficient)

**Worker Routing:** Update existing Cloudflare Worker to route `/` to landing origin while `/docs/*` and `/rules/*` stay on Mintlify.

**Monorepo:** Astro project lives in `web/` as a new Moon project with dev/build/check tasks.

### Performance Budgets (Hard Limits)

**CSS:** ≤ 30kb gzipped
- Inline critical above-fold CSS (<4kb)
- Defer non-critical section styling
- No unused CSS; vanilla CSS aids here

**JavaScript:** ≤ 150kb gzipped
- Astro outputs zero JS by default (pure HTML/CSS)
- Island: copy-to-clipboard button (native browser APIs, <1kb)
- Island: terminal animation (if needed; <5kb if using CSS animation instead of JS)
- Waitlist form: vanilla fetch, no libraries (<2kb)

**Images:** ≤ 50kb total (all hero + section images)
- AVIF primary format (highest compression)
- WebP fallback
- PNG fallback (screenshots)
- Explicit `width` and `height` on all images
- `loading="eager" fetchpriority="high"` for hero image only
- `loading="lazy"` for below-fold content
- Screenshot reuse: don't re-optimize; use existing WebP from `docs/product/images/screenshots/`

**Fonts:** ≤ 30kb gzipped (all weights/styles)
- Preload critical weight only (Ruda 800 for headings; Atkinson Mono regular for code)
- Self-host (don't use Google Fonts CDN — offline-first ethos)
- `font-display: swap` (let text render while font loads)
- Subset if feasible (Latin + symbols only, no CJK)

**Core Web Vitals targets:**
- LCP < 1.5s (Largest Contentful Paint)
- FCP < 1s (First Contentful Paint)
- INP < 200ms (Interaction to Next Paint)
- CLS < 0.1 (Cumulative Layout Shift)
- TBT < 200ms (Total Blocking Time)

---

## Content & Copy

All copy is **locked before implementation begins.** Copy sources:

| Section | Source | Status |
|---------|--------|--------|
| Hero headline | Brand tagline "AI-agent readiness, scored." | Locked (DESIGN-TOKENS.md) |
| Hero subheadline | architecture-2026.md (Commitment: deterministic, offline, rule-based) | Locked |
| Problem | introduction.mdx ("Teams adopting agents fail not because the model is bad...") | Locked (verbatim) |
| Solution headline | Brand motif: "Scan · Score · Fix" | Locked |
| Value props | Readiness axes: Context, Safety, Operability, Governance | Locked (from docs IA) |
| Trust strip | Ten Commitments (architecture-2026.md lines 55–79) | Locked (select 4) |
| CTA copy | "Get started", "View on GitHub", "Read the docs", "Star on GitHub" | Locked |

**Content rules:**
- No benchmark numbers beyond "<2s" and "50k-file budget" (both verified in architecture docs)
- Every claim must be derivable from architecture-2026.md, introduction.mdx, or quickstart.mdx
- Severity/score language must match docs exactly (no drift between landing page and docs)
- No fabricated social proof; use real GitHub stars + vendor logos only if permissioned

---

## Success Criteria

**Technical:**
- ✅ Lighthouse scores ≥ 90 (Performance, Accessibility, Best Practices)
- ✅ Core Web Vitals all green (LCP < 1.5s, FCP < 1s, INP < 200ms, CLS < 0.1, TBT < 200ms)
- ✅ Performance budgets met (JS ≤ 150kb, CSS ≤ 30kb, images ≤ 50kb, fonts ≤ 30kb gzipped)
- ✅ Build output is pure HTML/CSS + static assets (no Node.js runtime)
- ✅ Astro `astro check` passes (type safety)

**Accessibility:**
- ✅ WCAG 2.2 AA compliance (axe-core automated + manual keyboard testing)
- ✅ Semantic HTML validation (no role abuse, proper heading hierarchy, meaningful alt text)
- ✅ Focus visible on all interactive elements
- ✅ Reduced-motion respected (no animation flicker)
- ✅ Color contrast ≥ 4.5:1 (normal) / 3:1 (large) on all text

**Design:**
- ✅ Matches terminal-dashboard aesthetic (dark theme, Ruda + mono, semantic colors)
- ✅ Consistent with existing brand assets (DESIGN-TOKENS.md, brand/README.md, meta.html)
- ✅ Responsive at 320 / 768 / 1440 (no overflow, sensible reflows, touch targets ≥ 44px)
- ✅ Avoids template aesthetic; demonstrates intentional composition and interaction states

**Content:**
- ✅ Copy is locked, sourced from approved product docs, never contradicts Mintlify docs
- ✅ No fabricated social proof; GitHub stars are real, logos are permissioned or placeholders flagged
- ✅ Conversion hierarchy clear (primary: install; secondary: GitHub; deferred: waitlist)

**Functional:**
- ✅ Copy-to-clipboard button works (includes fallback text selection)
- ✅ All links functional (GitHub repo, Mintlify docs, etc.)
- ✅ Waitlist email capture posts to TBD endpoint (spec as placeholder; endpoint provided by operations)
- ✅ Cloudflare Pages deployment successful; Worker routing verified (`/` → landing, `/docs/*` → Mintlify, `/rules/*` → Mintlify)

---

## Acceptance Criteria

**Before marking "Done":**

**Responsive & Performance:**
- [ ] Page renders at 320, 375, 768, 1024, 1440, 1920 without horizontal scroll (except intentional terminal card)
- [ ] All CTA buttons ≥ 44×44px touch targets (copy, form submit, links)
- [ ] Hero image with `loading="eager" fetchpriority="high"` loads before LCP
- [ ] Lighthouse Performance ≥ 90, Accessibility ≥ 90
- [ ] Core Web Vitals: LCP < 1.5s, FCP < 1s, INP < 200ms, CLS < 0.1
- [ ] CSS ≤ 30kb gzipped, JS ≤ 150kb gzipped, images ≤ 50kb total, fonts ≤ 30kb gzipped

**Accessibility (WCAG 2.2 AA):**
- [ ] `astro check` exits 0 (type safety)
- [ ] axe-core automated scan: zero violations
- [ ] Manual keyboard test: Tab through entire page, all interactive elements reachable
- [ ] Focus visible ≥ 3px outline or shadow on all buttons/links/inputs
- [ ] Heading hierarchy: H1 (once, hero) → H2 (sections) → H3 (subsections only)
- [ ] Color contrast ≥ 4.5:1 (normal text), ≥ 3:1 (large text) — Chrome DevTools verify
- [ ] All images + screenshots have descriptive alt text
- [ ] Form has proper labels (above input), error messages (below), helper text (optional)
- [ ] `prefers-reduced-motion: reduce` disables animations; page still fully functional
- [ ] Screen reader test (Mac VoiceOver or Windows NVDA): page structure readable, CTAs announced

**SEO & Metadata:**
- [ ] Meta tags present: title, description, og:title, og:description, og:image, canonical
- [ ] All external links semantic + `rel="noopener noreferrer"`
- [ ] GitHub stars count fetched at build time and cached
- [ ] No console errors or warnings in production build

**Assets & Security:**
- [ ] Fonts self-hosted (Ruda, Atkinson Mono, IBM Plex Mono, Share Tech); no Google CDN
- [ ] Images use `<Picture />` with AVIF primary, WebP fallback, PNG fallback
- [ ] All images explicit dimensions (`width` + `height` attributes)
- [ ] Hero terminal card: CSS animation only (no heavy JS)
- [ ] Icons: Phosphor Light or SVG (no emojis)

**Functionality:**
- [ ] Copy-to-clipboard button: works on desktop, has text fallback for iOS
- [ ] Form validation: email regex validates, shows error state with message
- [ ] Form submission: POST to endpoint, shows loading state (disabled + spinner), success/error toast
- [ ] All links functional: GitHub, docs, Mintlify paths, waitlist endpoint

**Build & Deployment:**
- [ ] `bun run build` outputs clean `dist/` directory
- [ ] Cloudflare Pages deployment succeeds
- [ ] Worker routing verified: `/` → landing, `/docs/*` → Mintlify, `/rules/*` → Mintlify
- [ ] HTTPS verified: no cert warnings on use-charter.dev

---

## Critical Dependencies

1. **Waitlist endpoint (TBD):** Implementation requires a POST endpoint. Endpoint contract:
   - Accept: `{ email: string }`
   - Response: `200 + { success: true, message: string }` (on success) or `400/500 + { error: string }` (on failure)
   - Must handle spam/duplicates gracefully
   - Must sanitize email input server-side
   - Must store emails securely (ops/privacy responsibility)
   - Form must show loading state during POST (disabled button + spinner)
   - Form must show success toast on 200, error toast on 400+

2. **GitHub API:** Fetch star count at build time via GitHub API (no auth needed for public repos). Cache in static output; updates at next Pages deploy. Source: https://api.github.com/repos/anthropics/charter

3. **Existing assets:** 
   - Design tokens in `docs/internal/designs/DESIGN-TOKENS.md` (extract to CSS variables exactly)
   - Brand meta.html in `docs/internal/designs/brand/` (extract `<head>` block)
   - Screenshots in `docs/product/images/screenshots/` (use as-is; reuse existing WebP, no re-optimization)
   - Fonts: Ruda, Atkinson Hyperlegible Mono, IBM Plex Mono, Share Tech (self-host in `web/src/fonts/`)
   - Icons: Phosphor Light (NOT emojis)

4. **Cloudflare Worker code:** Existing Worker updated to route `/` → landing origin. Implementation plan coordinates this. Source: Cloudflare Worker in repo root.

5. **Meta Tags Template:** Required in Base.astro:
   - `<title>` with keyword "AI-agent readiness, scored"
   - `<meta name="description">` (160 chars)
   - `<meta property="og:title">` (Open Graph)
   - `<meta property="og:description">`
   - `<meta property="og:image">` (hero image AVIF/WebP with PNG fallback)
   - `<link rel="canonical">` (https://use-charter.dev/)

---

## Delivery Model

Delivered as a series of atomic, GPG-signed commits on `main`. See implementation plan for commit structure and sequencing.

**Spec status:** Locked (2026-06-10)  
**Ready for implementation:** Yes
