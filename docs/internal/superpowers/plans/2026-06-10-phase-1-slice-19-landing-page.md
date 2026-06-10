# Slice 19 — Landing Page Implementation Plan

**Document type:** Implementation plan (HOW, not WHAT)  
**Based on:** `docs/internal/superpowers/specs/2026-06-10-phase-1-slice-19-landing-page.md`  
**Status:** Ready for execution (validated against all official docs + design skills)  
**Branch:** `main` (all commits); GPG-signed, `moon run :check` green throughout

---

## Overview

Deliver landing page in 9 phases. Each phase produces atomic commit(s) with specific verification checkpoints. No phase starts until prior phase completes.

**Total scope:** ~2000 lines (HTML/CSS/JS); zero Node.js runtime; static output to Cloudflare Pages.

**Tech Stack Validated Against Official Docs:**
- Astro v6 SSG (https://docs.astro.build/)
- Bun package manager (https://bun.sh/docs)
- Cloudflare Pages (https://developers.cloudflare.com/pages/)
- Vanilla CSS + CSS custom properties
- `<Picture />` component for image optimization

---

## Phase 1 — Project scaffold & design foundation

**Goal:** Astro project structure, Moon integration, design tokens, clean build.

### Tasks

- [ ] **1.1** Create `web/` directory at repo root
- [ ] **1.2** Create `web/moon.yml` with dev/build/check tasks (bun-based)
- [ ] **1.3** Initialize Astro: `bun create astro` (select vanilla template, skip git)
- [ ] **1.4** Install dependencies: `bun install` inside `web/`
- [ ] **1.5** Create `web/astro.config.mjs` with `output: 'static'` config
- [ ] **1.6** Create `web/src/styles/design-tokens.css` with all CSS variables:
  - Color tokens (surfaces, accents, score-zones)
  - Typography tokens (font families, sizes, line heights)
  - Spacing tokens (8-point grid system)
  - Motion tokens (easing functions, durations)
  - Extract from `docs/internal/designs/DESIGN-TOKENS.md`
- [ ] **1.7** Create `web/src/styles/global.css` with:
  - CSS resets
  - Base typography styles
  - Responsive breakpoint definitions (320/768/1440/1920)
  - Reference all design-tokens variables
- [ ] **1.8** Create `web/src/layouts/Base.astro` with:
  - Imports all CSS files
  - Meta tags template (title, description, og:*, canonical)
  - Font preload `<link rel="preload">` for Ruda 800 + Atkinson Mono regular
  - Font declarations with `font-display: swap`
  - Extract `<head>` block from `docs/internal/designs/brand/meta.html`
- [ ] **1.9** Self-host fonts in `web/src/fonts/`:
  - Ruda (400, 500, 700, 800)
  - Atkinson Hyperlegible Mono (regular, bold)
  - IBM Plex Mono (regular)
  - Share Tech (regular)
  - Copy from `docs/internal/designs/fonts/` or download from source
- [ ] **1.10** Create `.gitignore` inside `web/` (node_modules, dist, .astro, .env*)
- [ ] **1.11** Create `web/src/pages/index.astro` (empty layout, will populate in Phase 2)

### Verification

```bash
cd web
bun install
bun run build
ls dist/index.html && echo "✓ build output exists"
wc -l dist/index.html
du -sh dist/
moon run web:check && echo "✓ moon task green"
grep "Ruda" dist/index.html && echo "✓ fonts in output"
grep "preload" dist/index.html && echo "✓ font preload links present"
```

**Checkpoint:** dist/index.html exists, ≤ 200kb uncompressed, all fonts self-hosted, moon tasks green.

### Commit

```
feat(web): scaffold Astro landing page + design tokens + fonts

- Initialize Astro v6 (vanilla template) via bun
- Configure astro.config.mjs with output: 'static'
- Extract all design tokens to CSS variables (colors, typography, spacing, motion)
- Create global.css with resets and responsive breakpoints
- Create Base.astro layout with meta tags + font preload
- Self-host fonts (Ruda, Atkinson, IBM Plex Mono, Share Tech)
- Add Moon project configuration

Result: Clean Astro scaffold with zero JS, all tokens/fonts ready.
```

---

## Phase 2 — Hero section + terminal card + copy button

**Goal:** Above-fold hero with headline, subheadline, CTAs, terminal card component, working copy button.

### Tasks

- [ ] **2.1** Create `web/src/components/Hero.astro`:
  - `<header>` wrapper
  - `<h1>` "AI-agent readiness, scored." (Ruda 700/800, text-4xl md:text-5xl)
  - `<p>` subheadline (spec § Hero, Ruda 400, text-base/lg)
  - CTA group: CopyButton + "View on GitHub" link
  - Terminal card slot (right/below on desktop, below on mobile)
- [ ] **2.2** Create `web/src/components/CopyButton.astro`:
  - `<button>` with command: `brew install use-charter/tap/charter`
  - Vanilla JS (no framework): clipboard API + text selection fallback for iOS
  - States: default → hover (scale 1.02) → active (scale 0.98) → "copied" (text change, 2s timeout)
  - Icon: Phosphor Light Copy icon (NOT emoji)
- [ ] **2.3** Create `web/src/components/TerminalCard.astro`:
  - Render real `charter doctor` output as styled `<pre>` block
  - Fallback: static screenshot (WebP) from `docs/product/images/screenshots/`
  - Styling: dark terminal aesthetic with score-zone colors (green/amber/red)
  - Responsive: horizontal scroll on 320px, normal display on 768px+
  - Alt text on screenshot: "Charter doctor output showing 94/100 ship-ready score in green zone"
- [ ] **2.4** Create `web/src/styles/sections/hero.css`:
  - Hero grid layout (split on desktop, stacked on mobile)
  - Typography scale (h1 sizing)
  - Terminal card styling (dark background, rounded corners, border, shadow)
  - Score-zone colors (success #4ade80, warning #fbbf24, danger #f87171)
  - Button states: hover, active, focus-visible
  - Animations: fade-in on scroll, section-up on load
- [ ] **2.5** Create `web/src/islands/CopyButton.ts` (vanilla JS):
  - Clipboard API with fallback (text selection for iOS)
  - Success feedback: button text change "Copied!" for 2s, then revert
  - Error handling: silently fall back to text selection
  - No console errors
- [ ] **2.6** Wire Hero into `web/src/pages/index.astro` as first section
- [ ] **2.7** Test: `bun run dev` → http://localhost:3000:
  - Hero visible above fold
  - Copy button clickable, copies to clipboard
  - Terminal card renders (or screenshot fallback)
  - Responsive at 320 and 1440
  - No console errors

### Verification

```bash
cd web
bun run build
grep -o "AI-agent readiness" dist/index.html && echo "✓ headline"
grep -o "brew install" dist/index.html && echo "✓ install command"
grep -o "<header>" dist/index.html && echo "✓ semantic header"
grep -o "94/100" dist/index.html && echo "✓ terminal score"
grep -o "scale-\[1.02\]" dist/index.css && echo "✓ button hover scale"
# Manual: visit http://localhost:3000 → hero visible, copy works, terminal renders
```

**Checkpoint:** Headline, command, header element, score, button states present. Hero renders properly.

### Commit

```
feat(web/hero): hero section + terminal card + copy button

- Create Hero.astro (header, headline, subheadline, CTAs, terminal slot)
- Create CopyButton.astro (vanilla JS, clipboard API, iOS fallback, state feedback)
- Create TerminalCard.astro (styled pre or screenshot fallback)
- Create islands/CopyButton.ts (clipboard logic)
- Create sections/hero.css (grid, colors, animations, button states)
- Wire Hero into index.astro

Result: Above-fold hero with working copy button, responsive layout.
```

---

## Phase 3 — Content sections 2–5 (Problem, Solution, Value Props, Trust)

**Goal:** Mid-page content with locked copy, semantic HTML, styling.

### Tasks

- [ ] **3.1** Create `web/src/components/Section.astro`:
  - Generic section wrapper: `<section id={id} aria-labelledby={headingId}>`
  - Props: `title` (string), `id` (string), `children` (slot)
  - Optional `intro` paragraph slot
  - Enforces H2 heading inside
- [ ] **3.2** Create `web/src/components/ProblemSection.astro`:
  - Two-column layout on desktop (mobile: stacked)
  - H2: "The Problem" (or custom)
  - Copy from spec § Section 2: "Teams adopting coding agents..." (exact verbatim)
  - Styling: grid layout, responsive
- [ ] **3.3** Create `web/src/components/SolutionSection.astro`:
  - Three-step flow (horizontal on desktop, vertical on mobile)
  - H2: "How It Works"
  - Step 1/2/3: Scan, Score, Fix (copy from spec § Section 3)
  - Screenshots reused from `docs/product/images/screenshots/`:
    - doctor-overview.webp
    - fix-dry-run.webp
    - doctor-tty.webp
  - Use `<Picture />` component (AVIF/WebP/PNG fallback)
  - Explicit image dimensions
- [ ] **3.4** Create `web/src/components/ValuePropCard.astro`:
  - Card component for 4 readiness axes
  - Props: `title`, `description`, `icon`
  - Styling: consistent card design (no gratuitous elevation)
  - Four cards: Context, Safety, Operability, Governance (copy from spec § Section 4)
  - Icons: Phosphor Light (NOT emoji)
- [ ] **3.5** Create `web/src/components/TrustStrip.astro`:
  - Badge layout for four commitments
  - Copy from spec § Section 5 (Ten Commitments, select 4)
  - Styling: semantic color tokens, text-center
  - Icons (optional): Phosphor Light check icons
- [ ] **3.6** Create `web/src/styles/sections/*.css` (one per section):
  - problem.css (two-column grid)
  - solution.css (three-step flow + image styling)
  - value-props.css (four-card grid)
  - trust.css (badge layout)
- [ ] **3.7** Wire all sections into `web/src/pages/index.astro`
- [ ] **3.8** Verify copy matches spec exactly (no drift):
  - Use `grep` to confirm problem/solution/value/trust text
  - No capitalization changes, no rewording

### Verification

```bash
cd web
bun run build
grep "Teams adopting" dist/index.html && echo "✓ problem section"
grep "Scan" dist/index.html && grep "Score" dist/index.html && echo "✓ solution section"
grep "Context" dist/index.html && echo "✓ value props"
grep "Never calls an LLM" dist/index.html && echo "✓ trust strip"
# Manual: scroll through sections 2–5 → responsive at 320/768/1440, copy exact
```

**Checkpoint:** All copy present. Responsive layouts. Images load with explicit dimensions.

### Commit

```
feat(web/sections): problem, solution, value props, trust sections

- Create Section.astro (generic wrapper)
- Create ProblemSection.astro (two-column)
- Create SolutionSection.astro (three-step flow + Picture components)
- Create ValuePropCard.astro (four cards)
- Create TrustStrip.astro (four commitment badges)
- Create sections/*.css (grid, spacing, typography)
- Wire all into index.astro

Result: Mid-page content with locked copy, responsive layouts, images optimized.
```

---

## Phase 4 — Content sections 6–9 (Social Proof, CI, CTA, Footer)

**Goal:** Lower-page content, conversion mechanics, footer links.

### Tasks

- [ ] **4.1** Create `web/src/components/SocialProofSection.astro`:
  - GitHub stars (fetch at build time via GitHub API)
  - Vendor icons/logos: Claude Code, Codex, Cursor, Windsurf, Copilot, Gemini
  - Icons: use SVG files from `docs/product/images/icons/` or inline as SVG
  - Placeholder slot for real logos (flag as unfilled until permissioned)
  - NO fabricated numbers; only real GitHub stars + vendor logos
- [ ] **4.2** Create `web/src/components/CISection.astro`:
  - H2: "Gate Pull Requests on Agent-Readiness"
  - Copy: "Short band showing GitHub Action snippet + mention of SARIF 2.1.0 output"
  - CTA link to `docs/product/how-to/run-in-github-actions.mdx`
  - Code block styling (use Atkinson Mono for code)
- [ ] **4.3** Create `web/src/components/FinalCTASection.astro`:
  - Repeat install command (reuse CopyButton component)
  - Links: "Read the docs" → Mintlify, "Star on GitHub" → repo, "Notify me" → waitlist form
  - High visual hierarchy (primary/secondary CTA distinction)
- [ ] **4.4** Create `web/src/components/Footer.astro`:
  - `<footer>` semantic element
  - Link groups: Docs, Rules, GitHub, Releases, Community (Discord), License, Copyright
  - All external links: `rel="noopener noreferrer"`
  - Styling: subtle, clean (not card-based)
- [ ] **4.5** Wire GitHub API fetch (build-time):
  - Fetch star count from `https://api.github.com/repos/anthropics/charter`
  - Cache in static output (updates at next deploy)
  - Error handling: if fetch fails, use fallback string ("Join our community")
- [ ] **4.6** Wire favicon + logo into meta (from `brand/meta.html`)
- [ ] **4.7** Verify all external links have `rel="noopener noreferrer"`
- [ ] **4.8** Test keyboard navigation:
  - Tab through footer links
  - All links reachable
  - Focus visible on each link

### Verification

```bash
cd web
bun run build
grep "Gate pull requests" dist/index.html && echo "✓ CI section"
grep "href=" dist/index.html | wc -l | awk '{print "✓ links found:", $1}'
grep "noopener" dist/index.html | wc -l | awk '{print "✓ noopener links:", $1}'
grep "favicon" dist/index.html && echo "✓ favicon wired"
# Manual: Tab through footer → all links navigable, focus visible
```

**Checkpoint:** CI section present. External links secured. Footer links keyboard-navigable.

### Commit

```
feat(web/footer): social proof, CI, CTA, footer

- Create SocialProofSection.astro (GitHub stars via build-time fetch, vendor icons)
- Create CISection.astro (GitHub Action band)
- Create FinalCTASection.astro (repeat install + docs/repo/waitlist links)
- Create Footer.astro (semantic footer with all links)
- Wire GitHub API fetch at build time
- Add favicon/logo to meta

Result: Full landing page with conversion hierarchy (install → stars → waitlist).
```

---

## Phase 5 — Interactive islands & form (TDD approach)

**Goal:** Copy button, waitlist form, animations. Test-first methodology.

### TDD Checklist

**Test 1: Copy Button → Test First**
- [ ] Write test: copy button copies to clipboard (RED)
- [ ] Implement CopyButton functionality (GREEN)
- [ ] Write test: iOS fallback works (RED)
- [ ] Add fallback (GREEN)
- [ ] Write test: "copied" feedback appears for 2s (RED)
- [ ] Add feedback timeout (GREEN)

**Test 2: Form Validation → Test First**
- [ ] Write test: email validation rejects bad emails (RED)
- [ ] Implement regex validation (GREEN)
- [ ] Write test: form submits to endpoint (RED)
- [ ] Implement fetch POST (GREEN)
- [ ] Write test: form shows loading state during POST (RED)
- [ ] Add disabled + spinner state (GREEN)

**Test 3: Form Feedback → Test First**
- [ ] Write test: success toast shows on 200 (RED)
- [ ] Implement success state (GREEN)
- [ ] Write test: error toast shows on 4xx/5xx (RED)
- [ ] Implement error state (GREEN)

### Tasks

- [ ] **5.1** Create `web/src/islands/CopyButton.ts`:
  - Vanilla JS (no React/Vue/Svelte)
  - Clipboard API: `navigator.clipboard.writeText(command)`
  - Fallback: `document.execCommand('copy')` + text selection
  - Feedback: button text "Copied!" for 2000ms, then revert
  - Error handling: silent (no console errors)
- [ ] **5.2** Create `web/src/components/WaitlistForm.astro`:
  - Form with email input + submit button
  - Label above input (from web-design-guidelines skill)
  - Error text below input on validation failure
  - Form state management (vanilla JS or Astro island)
- [ ] **5.3** Create `web/src/islands/WaitlistForm.ts`:
  - Email regex validation: `/^[^\s@]+@[^\s@]+\.[^\s@]+$/`
  - POST to endpoint (TBD) with `{ email: string }`
  - Loading state: `button.disabled = true`, show spinner
  - Success: toast message "Check your email!"
  - Error: toast message from response or generic "Something went wrong"
  - No console errors
- [ ] **5.4** Create `web/src/styles/islands.css`:
  - Button states (hover, active, disabled, focus-visible)
  - Form input styling (border, focus, error states)
  - Toast styling (success/error colors)
  - Spinner animation (CSS only, respects `prefers-reduced-motion`)
- [ ] **5.5** Optional: Create `web/src/islands/TerminalAnimation.ts`:
  - CSS-only animation for terminal card (preferred)
  - Fade-in + slide-up on scroll reveal
  - Respects `prefers-reduced-motion: reduce`
  - Timing: 300-500ms easing: `cubic-bezier(0.16, 1, 0.3, 1)`
- [ ] **5.6** Test copy button:
  - Desktop: click → clipboard receives text
  - iOS: fallback text selection works
  - No console errors
- [ ] **5.7** Test form:
  - Invalid email rejected (red border + error text)
  - Valid email: submit button enabled
  - Submit: button disabled, spinner shows
  - Success: toast "Check your email!"
  - Network error: toast with error message
- [ ] **5.8** Bundle size check: islands ≤ 5kb combined JS

### Verification

```bash
cd web
bun run build
ls -lh dist/_astro/*.js | awk '{sum+=$5} END {print "Total JS:", sum, "bytes"}'
grep "prefers-reduced-motion" dist/*.css && echo "✓ reduced-motion query"
# Manual: click copy → clipboard works; submit form → loading/success/error states
```

**Checkpoint:** JS bundle ≤ 5kb. Copy button + form fully functional with all states.

### Commit

```
feat(web/islands): copy button + waitlist form (TDD)

- Create CopyButton.ts (vanilla JS, clipboard API + iOS fallback, 2s feedback)
- Create WaitlistForm.ts (email validation, POST, loading/success/error states)
- Create islands.css (button states, form styling, toast, spinner)
- Optional: TerminalAnimation.ts (CSS-only scroll reveal)
- Test copy on desktop + iOS
- Test form validation + submission
- Verify bundle ≤ 5kb

Result: Interactive islands for copy and email capture, <5kb JS.
```

---

## Phase 6 — Performance optimization

**Goal:** Hit all performance budgets; inline critical CSS; optimize images/fonts.

### Tasks

- [ ] **6.1** Audit CSS: identify above-fold styles (hero + headline + buttons)
  - Inline <4kb into `<style>` tag in Base.astro `<head>`
  - Extract below-fold CSS into deferred `<link rel="stylesheet">`
- [ ] **6.2** Optimize hero image:
  - Use `<Picture />` component (auto-generates AVIF/WebP)
  - Add explicit `width` + `height` attributes
  - Add `loading="eager" fetchpriority="high"`
  - Verify AVIF < 40kb, WebP < 50kb, PNG < 80kb
- [ ] **6.3** Optimize section images (screenshots):
  - Use existing WebP from `docs/product/images/screenshots/`
  - Use `<Picture />` with AVIF/WebP fallback
  - Add explicit dimensions
  - Add `loading="lazy"` (below-fold)
  - Verify total images ≤ 50kb
- [ ] **6.4** Verify font strategy:
  - Fonts self-hosted (done in Phase 1)
  - Preload links in `<head>` (done in Phase 1)
  - `font-display: swap` in @font-face (done in Phase 1)
  - Only Ruda 800 + Atkinson Mono regular preloaded
  - Verify fonts ≤ 30kb gzipped
- [ ] **6.5** Verify Astro minification (default enabled)
  - Check `astro.config.mjs` has minifyHTML/CSS enabled
  - Run `bun run build` and verify output is minified
- [ ] **6.6** Run Lighthouse locally:
  - `bun run dev` on http://localhost:3000
  - Run: `npx lighthouse --output-path=lighthouse.html`
  - Target: Performance ≥ 90, Accessibility ≥ 90

### Verification

```bash
cd web
bun run build

# CSS budgets
wc -c dist/index.html | awk '{sum=$1; gzip=system("gzip -c dist/index.html | wc -c"); print "HTML:", sum, "bytes; gzipped:", gzip}'

# Image budgets
find dist -name "*.webp" -o -name "*.png" -o -name "*.avif" | xargs du -c | tail -1

# Font budgets
du -sh dist/fonts/

# Total output
du -sh dist/

# Lighthouse
npx lighthouse http://localhost:3000 --output-path=lighthouse.html
# Target: Performance ≥ 90
```

**Checkpoint:** Performance ≥ 90, Accessibility ≥ 90. All budgets met.

### Commit

```
perf(web): inline critical CSS, optimize images + fonts

- Inline above-fold CSS (<4kb)
- Defer below-fold stylesheet
- Optimize hero image (AVIF + WebP + explicit dimensions)
- Optimize section images (lazy loading, explicit dimensions)
- Verify font strategy (self-hosted, preload, swap)
- Achieve Lighthouse Performance ≥ 90

Result: Page meets all performance budgets and Core Web Vitals targets.
```

---

## Phase 7 — Accessibility audit + focus states

**Goal:** WCAG 2.2 AA compliance; keyboard nav; contrast; reduced-motion.

### Tasks

- [ ] **7.1** Run axe-core automated scan:
  - `npx axe-core dist/` (or use axe DevTools extension)
  - Fix all violations (must be zero)
  - Document any intentional exclusions
- [ ] **7.2** Manual keyboard testing:
  - Tab through entire page from top to bottom
  - Verify all interactive elements reachable: buttons, links, inputs, form
  - Verify focus visible (≥3px outline or shadow) on each
  - No keyboard traps (tabbing gets stuck)
  - Tab order makes sense (left-to-right, top-to-bottom)
- [ ] **7.3** Verify heading hierarchy:
  - H1 (once, hero section): "AI-agent readiness, scored."
  - H2 (sections): "Problem", "How It Works", "Value Props", etc.
  - H3 (subsections only): use sparingly, if at all
  - Check via: `grep -o "<h[1-3]" dist/index.html | sort | uniq -c`
- [ ] **7.4** Verify all images have descriptive alt text:
  - Hero image: "Terminal showing Charter doctor output with 94/100 score"
  - Screenshots: describe what's visible (e.g., "Screenshot showing Charter fix dry-run output")
  - Icons: already semantic, no alt needed if decorative
  - Check via: `grep -c "alt=" dist/index.html`
- [ ] **7.5** Verify color contrast (4.5:1 normal, 3:1 large):
  - Use Chrome DevTools Accessibility tab
  - Check all text on dark + light backgrounds
  - Flag any violations
- [ ] **7.6** Verify `prefers-reduced-motion: reduce`:
  - Terminal animation disabled (no CSS animations)
  - Page still fully functional (no visual dependency on motion)
  - Test in Chrome DevTools: Rendering → Emulate CSS media feature prefers-reduced-motion
- [ ] **7.7** Verify semantic HTML:
  - `<header>` wrapper in hero
  - `<main>` wrapper for content sections
  - `<section>` wrappers with `aria-labelledby` for each section
  - `<footer>` wrapper at bottom
  - No `role` abuse (no `role="button"` on non-button elements)
  - Check via: `grep -o "<header>\|<main>\|<section>\|<footer>" dist/index.html`
- [ ] **7.8** Screen reader test (pick one):
  - **Mac:** VoiceOver (Cmd+F5) → read through page
  - **Windows:** NVDA (free download) → read through page
  - Verify: page structure understandable, CTAs announced, forms labeled
  - Test: keyboard-only nav through all interactive elements

### Verification

```bash
cd web
bun run build
npx axe-core dist/ && echo "✓ axe-core: zero violations"

# Heading hierarchy
grep -o "<h[1-3]" dist/index.html | sort | uniq -c

# Alt text count
grep -c "alt=" dist/index.html

# Semantic HTML
grep -c "<header>" dist/index.html && grep -c "<main>" dist/index.html && grep -c "<footer>" dist/index.html

# Manual tests:
# - Tab through entire page
# - VoiceOver/NVDA read-through
# - prefers-reduced-motion toggle in DevTools
```

**Checkpoint:** axe-core passes. All alt text present. Keyboard nav works. Screen reader usable.

### Commit

```
a11y(web): accessibility audit + focus states

- Run axe-core automated scan (zero violations)
- Manual keyboard testing (Tab through entire page)
- Verify heading hierarchy (H1/H2/H3)
- Verify all images have descriptive alt text
- Verify color contrast (4.5:1 normal, 3:1 large)
- Verify prefers-reduced-motion is respected
- Verify semantic HTML (`<header>`, `<main>`, `<section>`, `<footer>`)
- Manual screen reader test (VoiceOver/NVDA)

Result: WCAG 2.2 AA compliance achieved.
```

---

## Phase 8 — Responsive refinement + cross-browser testing

**Goal:** Perfect behavior at all breakpoints; cross-browser compatibility.

### Tasks

- [ ] **8.1** Test viewport sizes (manual or automated):
  - 320px (iPhone SE)
  - 375px (iPhone 14)
  - 768px (iPad)
  - 1024px (iPad Pro)
  - 1440px (desktop)
  - 1920px (ultra-wide)
  - Verify layout sensible at each breakpoint
- [ ] **8.2** Verify no horizontal scroll (except intentional terminal card):
  - Terminal card can scroll horizontally (containment prevents page overflow)
  - All other content fits viewport width
- [ ] **8.3** Verify touch targets ≥ 44×44px:
  - Copy button, form submit, all links
  - Use Chrome DevTools Device Mode → pointer events overlay
- [ ] **8.4** Verify images scale correctly:
  - No stretching or clipping
  - Explicit dimensions prevent layout shift
  - Responsive images via `<Picture />`
- [ ] **8.5** Test on Chrome, Firefox, Safari (desktop + mobile):
  - Text rendering consistent
  - Font loading consistent
  - Button behavior consistent
  - CSS Grid layout works
  - No vendor-specific issues
- [ ] **8.6** Run Lighthouse on mobile simulation:
  - `npx lighthouse --throttling.cpuSlowdownMultiplier=4 http://localhost:3000`
  - Target: LCP < 2.5s, FCP < 1.8s (mobile relaxed targets)
- [ ] **8.7** Check CLS (Cumulative Layout Shift):
  - Use Chrome DevTools Performance tab
  - Verify no unexpected reflows
  - Buttons don't jump when loading
  - Images don't cause shift after load
- [ ] **8.8** Test form at all breakpoints:
  - Label visible above input
  - Input readable (font size ≥ 16px on mobile)
  - Error messages visible
  - Success/error toast visible

### Verification

```bash
cd web
# Responsive testing (manual in Chrome DevTools)
# 320px, 375px, 768px, 1024px, 1440px, 1920px → no overflow

# Mobile Lighthouse
npx lighthouse --throttling.cpuSlowdownMultiplier=4 http://localhost:3000 --output-path=mobile-lighthouse.html
# Target: LCP < 2.5s, FCP < 1.8s, CLS < 0.1

# Manual: test on real device (iOS/Android) → no layout shift, touch works
```

**Checkpoint:** All breakpoints tested. No overflow. Mobile Lighthouse targets met.

### Commit

```
test(web): responsive refinement + cross-browser verification

- Test at 320, 375, 768, 1024, 1440, 1920px
- Verify no horizontal scroll (except terminal)
- Verify touch targets ≥ 44px
- Verify images scale correctly
- Test on Chrome, Firefox, Safari (desktop + mobile)
- Mobile Lighthouse: LCP < 2.5s, FCP < 1.8s, CLS < 0.1
- Manual device testing: iOS/Android

Result: Responsive and cross-browser compatible.
```

---

## Phase 9 — Deployment & Worker integration

**Goal:** Deploy to Cloudflare Pages; wire Worker routing; verify production setup.

### Tasks

- [ ] **9.1** Create Cloudflare Pages project for `web/` directory
- [ ] **9.2** Set Pages build command: `bun run build`
- [ ] **9.3** Set Pages output directory: `dist/`
- [ ] **9.4** Verify Pages deployment succeeds:
  - Visit `<project-id>.pages.dev`
  - Landing page renders
  - All links work
  - Copy button works
  - Form submits successfully
- [ ] **9.5** Update Cloudflare Worker code to add landing route:
  - Add branch for `/`:
    ```javascript
    case '/':
      return fetch(`https://<landing-pages-url>/${path}`)
    ```
  - Existing `/docs/*` and `/rules/*` routes unchanged
  - Verify no conflicts
- [ ] **9.6** Verify Worker routing end-to-end:
  - `/` routes to landing
  - `/docs/*` routes to Mintlify (unchanged)
  - `/rules/*` routes to Mintlify (unchanged)
  - Test: `curl -sI https://use-charter.dev/`
- [ ] **9.7** Update `docs/product/DEPLOY.md` with:
  - New `LANDING_ORIGIN` URL
  - Deployment steps for landing page
  - How to update Worker code for landing integration
- [ ] **9.8** Final smoke test:
  - Visit `https://use-charter.dev/` in browser
  - Hero loads
  - Copy button works
  - All links navigate
  - `/docs` redirects to Mintlify
  - `/rules` redirects to Mintlify
- [ ] **9.9** Verify SSL/TLS:
  - HTTPS enforced
  - No certificate warnings
  - `use-charter.dev` secure

### Verification

```bash
# After Pages deploy
curl -sI https://use-charter.dev/ | head -5
# → HTTP/2 200, Content-Type: text/html

curl -sI https://use-charter.dev/docs | head -5
# → HTTP/2 30x (redirect to Mintlify)

curl -sI https://use-charter.dev/rules | head -5
# → HTTP/2 30x (redirect to Mintlify)

# Manual: visit https://use-charter.dev in browser
# → hero visible, copy works, all links functional, HTTPS verified
```

**Checkpoint:** Pages deployed. Worker routing verified. Smoke test passes. Production HTTPS ready.

### Commit

```
feat(web): deploy to Cloudflare Pages + wire Worker

- Create Cloudflare Pages project for web/
- Set build command: bun run build
- Set output directory: dist/
- Update Worker code to route / → landing origin
- Verify Worker routing: / → landing, /docs/* → Mintlify, /rules/* → Mintlify
- Update docs/product/DEPLOY.md with LANDING_ORIGIN URL
- Verify production: HTTPS, all links, copy button, form submit

Result: Landing page live at use-charter.dev with proper routing.
```

---

## Commit Sequence Summary

| Phase | Commit Message | Key Files |
|-------|---|---|
| 1 | `feat(web): scaffold Astro landing page + design tokens + fonts` | `web/`, astro.config.mjs, design-tokens.css, fonts |
| 2 | `feat(web/hero): hero section + terminal card + copy button` | Hero.astro, CopyButton.ts, TerminalCard.astro, hero.css |
| 3 | `feat(web/sections): problem, solution, value props, trust sections` | Section.astro, Problem/Solution/ValueProp/Trust components |
| 4 | `feat(web/footer): social proof, CI, CTA, footer` | SocialProof/CI/FinalCTA/Footer components |
| 5 | `feat(web/islands): copy button + waitlist form (TDD)` | CopyButton.ts, WaitlistForm.ts, islands.css |
| 6 | `perf(web): inline critical CSS, optimize images + fonts` | All CSS files, Picture components |
| 7 | `a11y(web): accessibility audit + focus states` | All HTML/CSS files |
| 8 | `test(web): responsive refinement + cross-browser verification` | Responsive CSS tweaks |
| 9 | `feat(web): deploy to Cloudflare Pages + wire Worker` | docs/product/DEPLOY.md, Worker code |

Each commit includes:
```
🤖 Generated with Claude Code

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>
```

---

## Cross-Task Invariants

- **No Go changes.** Web project isolated; charter CLI unchanged.
- **`moon run :check` green at every commit.** Verified before push.
- **GPG-signed commits.** All pushes use `-S` flag.
- **`use-charter.dev/` URL stable throughout.** No DNS changes mid-implementation.
- **Copy locked.** Never drift from spec sections 1–9 (verbatim).
- **No third-party JS libraries.** Islands use vanilla JS only.
- **No analytics/tracking.** Raw HTML/CSS only.
- **Performance budgets are hard limits:** JS ≤ 150kb, CSS ≤ 30kb, images ≤ 50kb, fonts ≤ 30kb (all gzipped).
- **Tech stack locked:** Astro v6 SSG + vanilla CSS + Bun package manager. No changes mid-phase.

---

## Phase Dependencies

```
Phase 1 (scaffold)
  ↓
  ├─→ Phase 2 (hero)
  ├─→ Phase 3 (sections 2–5)
  ├─→ Phase 4 (sections 6–9)
  └─→ Phase 5 (islands)
      ↓
  Phase 6 (optimization)
      ↓
  Phase 7 (a11y) [can run parallel with Phase 6]
      ↓
  Phase 8 (responsive)
      ↓
  Phase 9 (deploy)
```

**Critical path:** 1 → (2/3/4/5 parallel) → 6 → (7 parallel) → 8 → 9

---

**Plan status:** Ready for execution (all official docs validated, all design skills consulted, all TDD patterns documented).
