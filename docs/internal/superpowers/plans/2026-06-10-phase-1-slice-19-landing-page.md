# Slice 19 — Landing Page Implementation Plan

**Document type:** Implementation plan (HOW, not WHAT)  
**Based on:** `docs/internal/superpowers/specs/2026-06-10-phase-1-slice-19-landing-page.md`  
**Status:** Ready for execution  
**Branch:** `main` (all commits); GPG-signed, `moon run :check` green before every push

---

## Execution Strategy

**All phases are executed by sub-agents.** The orchestrator (main Claude Code session) dispatches agents per the table below and waits for each agent's checkpoint to pass before moving forward.

### Sub-Agent Dispatch Table

| Phase | Agent prompt (copy verbatim to dispatch) | Depends on |
|-------|------------------------------------------|------------|
| 1 | `"Execute Phase 1 of Slice 19: scaffold Astro v6 project in web/, extract design tokens from docs/internal/designs/DESIGN-TOKENS.md, generate fonts.css by running bun scripts/generate-report-fonts.ts, self-host fonts in web/public/fonts/, create Base.astro layout with all meta tags from spec Critical Dependencies §5, create web/src/pages/index.astro (empty), wire moon.yml check/build/dev tasks. Verify: bun run build produces dist/index.html; moon run web:check passes. Commit: feat(web): scaffold Astro landing page + design tokens + fonts"` | None |
| 2 | `"Execute Phase 2 of Slice 19: implement Hero section (split-screen layout, H1, subheadline, CTAs) + TerminalCard component (styled pre block with 94/100 score-zone colors) + CopyButton component (vanilla JS, clipboard API, iOS fallback, 2s feedback). Wire into index.astro. Verify: bun run build, grep checks pass. Commit: feat(web/hero): hero section + terminal card + copy button"` | Phase 1 ✅ |
| 3 | `"Execute Phase 3 of Slice 19: implement content sections Problem, Solution (three-step flow + screenshots via Picture component), ValuePropCard (4 cards), TrustStrip (4 commitments). Lock copy verbatim from spec. Wire into index.astro. Verify: grep checks pass, screenshots render. Commit: feat(web/sections): problem, solution, value props, trust sections"` | Phase 1 ✅ |
| 4 | `"Execute Phase 4 of Slice 19: implement SocialProofSection (GitHub stars build-time fetch from https://api.github.com/repos/use-charter/charter, 3 confirmed SVG icons + text badges for missing icons), CISection (locked YAML snippet from spec), FinalCTASection (CopyButton reuse + links), Footer. Wire into index.astro. Verify: grep checks, noopener links. Commit: feat(web/footer): social proof, CI, CTA, footer"` | Phase 1 ✅ |
| 5 | `"Execute Phase 5 of Slice 19: implement WaitlistForm island (Vitest TDD: write failing tests first for email validation, POST loading state, success/error toasts; then implement; all tests green). Verify: bun run test, bundle ≤ 5kb. Commit: feat(web/islands): waitlist form (TDD)"` | Phase 2 ✅ (CopyButton already done in Phase 2) |
| 6 | `"Execute Phase 6 of Slice 19: performance optimization — inline above-fold CSS (<4kb), verify image budgets via Picture components, run bunx lighthouse http://localhost:4321 and confirm Performance ≥ 90. Fix any budget failures. Commit: perf(web): inline critical CSS, verify all performance budgets"` | Phases 2-5 ✅ |
| 7 | `"Execute Phase 7 of Slice 19: accessibility audit — run bunx @axe-core/cli http://localhost:4321 and fix all violations; manual keyboard Tab test; heading hierarchy grep; alt text grep; verify prefers-reduced-motion disables animations. Commit: a11y(web): accessibility audit + focus states"` | Phase 6 ✅ |
| 8 | `"Execute Phase 8 of Slice 19: responsive refinement — test at 320/375/768/1024/1440/1920px in Chrome DevTools; verify no overflow; verify touch targets ≥ 44px; run bunx lighthouse with mobile throttling; fix any layout issues. Commit: test(web): responsive refinement + cross-browser"` | Phase 7 ✅ |
| 9 | `"Execute Phase 9 of Slice 19: Cloudflare Pages deploy — create Pages project, set build=bun run build, output=dist/, verify deploy; update Worker routing for / → landing origin; update docs/product/DEPLOY.md with LANDING_ORIGIN URL; final smoke test with curl. Commit: feat(web): deploy to Cloudflare Pages + wire Worker"` | Phase 8 ✅ |

### Parallelism Rules

Phases 2, 3, and 4 can run **in parallel** after Phase 1 completes (all three are independent — different sections, different files).

Phase 5 runs **after Phase 2** only (CopyButton already implemented in Phase 2; form is the only remaining island).

Phases 6 → 7 → 8 → 9 run **strictly sequentially** (each validates output of prior phase).

```
Phase 1 (scaffold)
  ↓
  ├─→ Phase 2 (hero + CopyButton)  ─────────────────┐
  ├─→ Phase 3 (sections 2–5)                        │
  └─→ Phase 4 (sections 6–9)                        │
                                                     ↓
                                              Phase 5 (form island — after Phase 2)
                                                     ↓
                                              Phase 6 (performance)
                                                     ↓
                                              Phase 7 (a11y)
                                                     ↓
                                              Phase 8 (responsive)
                                                     ↓
                                              Phase 9 (deploy)
```

### Orchestrator Checklist (run between phases)

Before dispatching next phase:
- [ ] Prior phase committed and pushed
- [ ] `moon run :check` green (Go tests, lint, docs)
- [ ] `bun run build` from `web/` exits 0
- [ ] No regressions in prior phases (spot-check prior sections still render)

---

## Phase 1 — Project scaffold & design foundation

**Goal:** Astro project structure, Moon tasks, design tokens, self-hosted fonts, clean build.

**Pre-condition:** `web/` directory already exists with `web/moon.yml` (committed in prior session). Do NOT re-create.

### Tasks

- [ ] **1.1** Add `check` script to `web/package.json`: `"check": "astro check"` (required by `web/moon.yml` check task)
- [ ] **1.2** Verify `web/astro.config.mjs` exists; if not, create with `output: 'static'` and `site: 'https://use-charter.dev'`
- [ ] **1.3** Run `bun scripts/generate-report-fonts.ts` from repo root to regenerate `internal/render/html/assets/fonts.css`
- [ ] **1.4** Extract all `@font-face` blocks from `internal/render/html/assets/fonts.css` into `web/src/styles/fonts.css` (keep base64 data URIs intact; no external CDN URLs)
- [ ] **1.5** Copy brand assets to `web/public/`:
  - `favicon.svg` ← from `docs/internal/designs/brand/favicon.svg`
  - `og.svg` ← from `docs/internal/designs/brand/og.svg`
  - `apple-touch-icon.svg` ← from `docs/internal/designs/brand/apple-touch-icon.svg`
  - `manifest.json` ← from `docs/internal/designs/brand/manifest.json`
- [ ] **1.6** Copy all vendor icons to `web/public/icons/` (all 8 confirmed in `docs/product/images/icons/`):
  - `claude-ai.svg`, `chatgpt.svg`, `grok.svg`, `cursor.svg`, `windsurf.svg`
  - `github-copilot.svg`, `google-gemini.svg`, `codex.svg`
- [ ] **1.7** Copy screenshots to `web/public/screenshots/`:
  - `doctor-overview.webp`, `fix-dry-run.webp`, `doctor-tty.webp` ← from `docs/product/images/screenshots/`
- [ ] **1.8** Create `web/src/styles/design-tokens.css`:
  - Extract ALL `--color-*`, `--font-*`, `--space-*`, `--radius-*`, `--shadow-*` variables verbatim from `docs/internal/designs/DESIGN-TOKENS.md`
  - Add motion tokens: `--ease-enter: cubic-bezier(0.16, 1, 0.3, 1)`, `--ease-exit: cubic-bezier(0.55, 0, 1, 0.45)`, `--duration-fast: 120ms`, `--duration-normal: 250ms`, `--duration-slow: 400ms`
  - Dark mode defaults in `:root { ... }`
  - Light mode overrides in `@media (prefers-color-scheme: light) { :root { ... } }` with exact values from spec §Color Mode
  - Icon adaptation CSS classes (from spec §Vendor Icon Adaptation):
    ```css
    .icon--dark-invert  { filter: brightness(0) invert(1); }
    .icon--light-invert { /* no filter in dark mode */ }
    @media (prefers-color-scheme: light) {
      .icon--dark-invert  { filter: none; }
      .icon--light-invert { filter: brightness(0); }
    }
    ```
- [ ] **1.9** Create `web/src/styles/global.css`:
  - Import `design-tokens.css` and `fonts.css`
  - CSS reset (box-sizing, margin, padding zero)
  - Base: `background-color: var(--color-bg, #0D1117)`, `color: var(--color-text, #e6edf3)`
  - Body: `font-family: var(--font-site, "Ruda", system-ui, sans-serif)`
  - `min-height: 100dvh` on body (never `100vh`)
  - Responsive breakpoints comment map: 320/375/768/1024/1440/1920
- [ ] **1.10** Create `web/src/layouts/Base.astro` with:
  - All meta tags from spec §Critical Dependencies §5 (exact attributes)
  - `<meta name="color-scheme" content="dark light" />` (required — tells browser color mode)
  - `<link rel="preload">` for Ruda 800 font-face and Atkinson Mono 400 font-face
  - Import `global.css`
  - `<slot />` for page content
- [ ] **1.11** Create `web/src/pages/index.astro` (empty shell using Base layout):
  ```astro
  ---
  import Base from '../layouts/Base.astro';
  ---
  <Base>
    <main><!-- sections injected by phases 2–4 --></main>
  </Base>
  ```

### Verification

```bash
# From repo root
bun scripts/generate-report-fonts.ts

# From web/
cd web
bun install
bun run build
ls dist/index.html && echo "✓ build output"
grep "AI-agent readiness" dist/index.html && echo "✓ title in head"
grep "font-face" dist/index.html && echo "✓ fonts embedded"
grep "canonical" dist/index.html && echo "✓ canonical link"
grep "preload" dist/index.html && echo "✓ preload links"
grep "favicon" dist/index.html && echo "✓ favicon"
grep "color-scheme" dist/index.html && echo "✓ color-scheme meta present"
grep "prefers-color-scheme" dist/index.html && echo "✓ light mode overrides in CSS"

# Moon integration
cd ..
moon run web:check && echo "✓ moon web:check green"
moon run :check && echo "✓ full repo check green"
```

**Checkpoint:** `dist/index.html` exists; fonts embedded; meta tags present; `moon run :check` green.

### Commit

```
feat(web): scaffold Astro landing page + design tokens + fonts

- Add astro check script to web/package.json
- Create astro.config.mjs (output: static, site: use-charter.dev)
- Generate and embed brand fonts from generate-report-fonts.ts
- Copy brand assets (favicon, og.svg, apple-touch-icon, manifest) to public/
- Copy vendor icons (claude-ai, chatgpt, grok) to public/icons/
- Copy screenshots (doctor-overview, fix-dry-run, doctor-tty) to public/screenshots/
- Create design-tokens.css (verbatim from DESIGN-TOKENS.md + motion tokens)
- Create global.css (reset, base styles, min-height: 100dvh)
- Create Base.astro layout (meta tags, preload, font import)
- Create index.astro (empty shell)
```

---

## Phase 2 — Hero section + terminal card + copy button

**Goal:** Above-fold hero with headline, subheadline, CTAs, terminal card, working copy button.

**Note:** CopyButton is implemented here (Phase 2). Phase 5 does NOT re-implement it. Phase 5 implements the WaitlistForm only.

### Tasks

- [ ] **2.1** Create `web/src/components/Hero.astro`:
  - `<header>` wraps the entire hero section
  - Split-screen layout: content left, terminal card right (desktop); stacked on mobile
  - `<h1>` exact text: "AI-agent readiness, scored." (Ruda 800, large scale)
  - `<p>` subheadline: exact text from spec §Section 1
  - CTA group: `<CopyButton />` + `<a href="https://github.com/use-charter/charter" rel="noopener noreferrer">View on GitHub</a>`
  - Terminal card slot (right column desktop, below content mobile)
  - Layout: `display: grid; grid-template-columns: 1fr 1fr;` at 768px+; single column below 768px
  - `min-height: 100dvh` on hero wrapper

- [ ] **2.2** Create `web/src/components/CopyButton.astro`:
  - `<button id="copy-btn" type="button">` with visible command text: `brew install use-charter/tap/charter`
  - Phosphor Light Copy SVG icon (inline SVG, not emoji)
  - `data-command="brew install use-charter/tap/charter"` attribute
  - Default state: shows command + copy icon
  - CSS states: hover (scale 1.02, 200ms ease), active (scale 0.98, 120ms), copied (text change), focus-visible (3px outline)
  - Inline `<script>` tag with full CopyButton island logic (see task 2.3)

- [ ] **2.3** Inline `<script>` inside CopyButton.astro (vanilla JS only, no imports):
  ```javascript
  const btn = document.getElementById('copy-btn');
  const command = btn.dataset.command;
  const original = btn.querySelector('.btn-text').textContent;
  
  btn.addEventListener('click', async () => {
    try {
      await navigator.clipboard.writeText(command);
    } catch {
      // Fallback for iOS: select text then execCommand (deprecated but necessary for iOS 16-)
      const input = document.createElement('input');
      input.value = command;
      document.body.appendChild(input);
      input.select();
      document.execCommand('copy'); // intentionally used: only fallback for iOS
      document.body.removeChild(input);
    }
    btn.querySelector('.btn-text').textContent = 'Copied!';
    btn.setAttribute('aria-label', 'Command copied to clipboard');
    setTimeout(() => {
      btn.querySelector('.btn-text').textContent = original;
      btn.setAttribute('aria-label', 'Copy install command');
    }, 2000);
  });
  ```

- [ ] **2.4** Create `web/src/components/TerminalCard.astro`:
  - `<div role="img" aria-label="Charter doctor output showing 94/100 ship-ready score in green zone">`
  - `<pre>` block with real `charter doctor` output (literal ASCII terminal output with ANSI-to-CSS color classes)
  - Score zone colors from design tokens: green (#4ade80) for 94/100, amber, red
  - CSS: `overflow-x: auto` (card scrolls horizontally on small viewports); dark background from tokens; `font-family: var(--font-mono-primary)` (Atkinson Mono)
  - Fallback: `<picture>` with `doctor-overview.webp` if pre block can't render (use `<noscript>` pattern)
  - `aria-hidden="true"` on decorative terminal chrome elements (corner dots, title bar)

- [ ] **2.5** Create `web/src/styles/sections/hero.css`:
  - Hero grid layout variables
  - H1 scale: `clamp(2.5rem, 5vw + 1rem, 4.5rem)` (no fixed px)
  - Subheadline: `clamp(1rem, 1.5vw + 0.5rem, 1.25rem)`, `max-width: 65ch`
  - CTA group gap, button sizing (min 44px height)
  - Terminal card: dark bg, rounded corners, border, subtle shadow from tokens
  - Section reveal animation (keyframes: `opacity 0→1, translateY 4rem→0`)
  - `@media (prefers-reduced-motion: reduce) { .reveal { animation: none; opacity: 1; transform: none; } }`

- [ ] **2.6** Import `hero.css` in `global.css` or directly in `Hero.astro`
- [ ] **2.7** Wire `<Hero />` into `web/src/pages/index.astro` as first child of `<main>`

### Verification

```bash
cd web
bun run build
grep "AI-agent readiness" dist/index.html && echo "✓ H1 headline"
grep "brew install use-charter" dist/index.html && echo "✓ install command"
grep "<header" dist/index.html && echo "✓ semantic header element"
grep "94/100" dist/index.html && echo "✓ terminal score present"
grep "data-command" dist/index.html && echo "✓ copy button data attribute"
grep "prefers-reduced-motion" dist/index.html && echo "✓ reduced-motion CSS present"
grep "noopener" dist/index.html && echo "✓ GitHub link secured"
# Start dev server and manually verify at http://localhost:4321
bun run dev
# Check: split-screen renders, copy button clicks and shows "Copied!", terminal card renders
```

**Checkpoint:** H1, install command, header element, 94/100 score, data-command, noopener all in output. Copy button works in browser.

### Commit

```
feat(web/hero): hero section + terminal card + copy button

- Create Hero.astro (split-screen layout, H1, subheadline, CTA group)
- Create CopyButton.astro (clipboard API, iOS execCommand fallback, 2s feedback)
- Create TerminalCard.astro (styled pre block, score-zone colors, horizontal scroll)
- Create sections/hero.css (grid, typography scale, animation, reduced-motion)
- Wire Hero into index.astro

Result: Above-fold hero renders; copy button copies to clipboard.
```

---

## Phase 3 — Content sections 2–5 (Problem, Solution, Value Props, Trust)

**Goal:** Mid-page content with locked copy, semantic HTML, responsive layouts.

### Tasks

- [ ] **3.1** Create `web/src/components/Section.astro`:
  - Props interface: `{ id: string; title: string; intro?: string }`
  - Template: `<section id={id} aria-labelledby={\`${id}-heading\`}>`
  - H2 inside: `<h2 id={\`${id}-heading\`}>{title}</h2>`
  - Optional intro: `<p class="section-intro">{intro}</p>` if prop provided
  - `<slot />` for section body content
  - Note: uses Astro `<slot />`, NOT a `children` prop

- [ ] **3.2** Create `web/src/components/ProblemSection.astro`:
  - Wraps `<Section id="problem" title="The Problem">`
  - Copy: exact verbatim from spec §Section 2 (no paraphrasing, no punctuation changes)
  - Desktop: two-column (copy left, visual/emphasis right)
  - Mobile: single column

- [ ] **3.3** Create `web/src/components/SolutionSection.astro`:
  - Wraps `<Section id="solution" title="How It Works">`
  - Three-step layout: horizontal on 768px+, vertical below
  - Each step: number badge, label (Scan/Score/Fix), description, screenshot
  - Step 1 (Scan): `<Picture src="/screenshots/doctor-tty.webp" alt="Charter scan output showing all 18 rules checked" width={800} height={500} />`
  - Step 2 (Score): `<Picture src="/screenshots/doctor-overview.webp" alt="Charter score output showing 94/100 ship-ready in green zone" width={800} height={500} />`
  - Step 3 (Fix): `<Picture src="/screenshots/fix-dry-run.webp" alt="Charter fix dry-run showing unified diff of proposed changes" width={800} height={500} />`
  - All `<Picture />` imports: `import { Picture } from 'astro:assets'`
  - All images: `loading="lazy"` (below fold)

- [ ] **3.4** Create `web/src/components/ValuePropCard.astro`:
  - Props interface: `{ title: string; description: string; iconSvg: string }`
  - Renders inline SVG icon (Phosphor Light), H3 title, description paragraph
  - No emojis
  - Card has earned elevation (shadow token) — cards here DO communicate hierarchy

- [ ] **3.5** Create `web/src/components/ValuePropsSection.astro`:
  - Wraps `<Section id="value-props" title="Four Dimensions of Readiness">`
  - Renders 4 `<ValuePropCard />` instances with locked copy from spec §Section 4
  - Grid: 2×2 on 768px+, 1-column on mobile

- [ ] **3.6** Create `web/src/components/TrustStrip.astro`:
  - Wraps `<Section id="trust" title="Built on Ten Commitments">`
  - 4 items, exact text from spec §Section 5 (verbatim from architecture-2026.md):
    1. "Never send data anywhere without explicit opt-in."
    2. "Never call an LLM — all findings are deterministic."
    3. "Every finding has a rule ID, evidence, and fix suggestion."
    4. "Every release is signed (cosign) with SLSA Level 3 provenance."
  - Each item: Phosphor Light CheckCircle SVG icon + `<p>` text
  - Layout: 2-column on 768px+, 1-column on mobile

- [ ] **3.7** Create section CSS files:
  - `web/src/styles/sections/problem.css` — two-column grid
  - `web/src/styles/sections/solution.css` — three-step flow, image sizing
  - `web/src/styles/sections/value-props.css` — 2×2 card grid
  - `web/src/styles/sections/trust.css` — 2-column badge layout

- [ ] **3.8** Wire all four components into `web/src/pages/index.astro` inside `<main>`, in order: Hero → Problem → Solution → ValueProps → Trust

### Verification

```bash
cd web
bun run build
grep -c "Teams adopting coding agents" dist/index.html && echo "✓ problem copy (verbatim)"
grep -c "doctor-tty" dist/index.html && echo "✓ scan screenshot"
grep -c "doctor-overview" dist/index.html && echo "✓ score screenshot"
grep -c "fix-dry-run" dist/index.html && echo "✓ fix screenshot"
grep -c "Never send data" dist/index.html && echo "✓ trust strip commitment 1"
grep -c "Never call an LLM" dist/index.html && echo "✓ trust strip commitment 2"
grep -c "rule ID" dist/index.html && echo "✓ trust strip commitment 3"
grep -c "cosign" dist/index.html && echo "✓ trust strip commitment 4"
grep -c "aria-labelledby" dist/index.html && echo "✓ section aria attributes"
grep -c 'loading="lazy"' dist/index.html && echo "✓ lazy loading on screenshots"
```

**Checkpoint:** All verbatim copy present. All screenshots in output. Section IDs + aria-labelledby wired.

### Commit

```
feat(web/sections): problem, solution, value props, trust sections

- Create Section.astro (generic wrapper with aria-labelledby pattern)
- Create ProblemSection.astro (verbatim copy, two-column layout)
- Create SolutionSection.astro (three-step flow, Picture components for screenshots)
- Create ValuePropCard.astro + ValuePropsSection.astro (four axes, Phosphor icons)
- Create TrustStrip.astro (four verbatim commitments from architecture-2026.md)
- Create sections/*.css (grid layouts)
- Wire all sections into index.astro

Result: Mid-page content rendered with locked copy.
```

---

## Phase 4 — Content sections 6–9 (Social Proof, CI, CTA, Footer)

**Goal:** Lower-page content, conversion mechanics, footer links.

### Tasks

- [ ] **4.1** Create `web/src/components/SocialProofSection.astro`:
  - Wraps `<Section id="social-proof" title="Works with every AI coding tool">`
  - GitHub stars count: fetched in frontmatter at build time:
    ```astro
    ---
    let stars = '';
    try {
      const res = await fetch('https://api.github.com/repos/use-charter/charter');
      if (res.ok) { const data = await res.json(); stars = data.stargazers_count?.toLocaleString(); }
    } catch {}
    ---
    ```
  - Stars render: `{stars ? `⭐ ${stars} stars` : 'Star on GitHub'}` (fallback is safe; never hardcode)
  - Vendor icon grid — all 8 SVGs confirmed. Apply CSS class per icon for dark/light mode adaptation (see spec §Vendor Icon Adaptation):
    ```html
    <!-- icon--colored: no filter needed, brand colors work in both modes -->
    <img src="/icons/claude-ai.svg"      class="vendor-icon icon--colored"      alt="Claude Code" width="32" height="32" />
    <img src="/icons/google-gemini.svg"  class="vendor-icon icon--colored"      alt="Gemini"      width="32" height="32" />
    <!-- icon--light-invert: white fill; needs brightness(0) in light mode -->
    <img src="/icons/chatgpt.svg"        class="vendor-icon icon--light-invert" alt="ChatGPT"     width="32" height="32" />
    <!-- icon--dark-invert: dark/black fill; needs invert in dark mode -->
    <img src="/icons/grok.svg"           class="vendor-icon icon--dark-invert"  alt="Grok"        width="32" height="32" />
    <img src="/icons/cursor.svg"         class="vendor-icon icon--dark-invert"  alt="Cursor"      width="32" height="32" />
    <img src="/icons/windsurf.svg"       class="vendor-icon icon--dark-invert"  alt="Windsurf"    width="32" height="32" />
    <img src="/icons/github-copilot.svg" class="vendor-icon icon--dark-invert"  alt="Copilot"     width="32" height="32" />
    <img src="/icons/codex.svg"          class="vendor-icon icon--dark-invert"  alt="Codex"       width="32" height="32" />
    ```
  - Do NOT fabricate "Used by X teams at [logos]"

- [ ] **4.2** Create `web/src/components/CISection.astro`:
  - Wraps `<Section id="ci" title="Gate pull requests on agent-readiness">`
  - Intro text: "Charter's GitHub Action runs the same scan as your local workflow. Gate merges when the score falls below the threshold."
  - Code block showing locked YAML from spec §Section 7 (render inside `<pre><code>` with Atkinson Mono)
  - CTA: `<a href="https://use-charter.dev/docs/how-to/run-in-github-actions" rel="noopener noreferrer">Read the GitHub Actions guide →</a>`
  - One-line mention: "Outputs SARIF 2.1.0. Integrates with GitHub Code Scanning."

- [ ] **4.3** Create `web/src/components/FinalCTASection.astro`:
  - Wraps `<section id="final-cta" aria-labelledby="final-cta-heading">`
  - H2: "Start in 30 seconds"
  - Reuse `<CopyButton />` component (already built in Phase 2)
  - Three distinct CTA buttons with visual hierarchy:
    - Primary: "Read the docs" → `https://use-charter.dev/docs`
    - Secondary: "Star on GitHub" → `https://github.com/use-charter/charter`
    - Tertiary: `<WaitlistForm />` placeholder — renders static "Notify me" link for now; WaitlistForm island added in Phase 5
  - All external links: `rel="noopener noreferrer"`

- [ ] **4.4** Create `web/src/components/Footer.astro`:
  - `<footer>` semantic element
  - Four link columns:
    - Product: `<a href="https://use-charter.dev/docs">Documentation</a>`, `<a href="https://use-charter.dev/rules">Rules reference</a>`
    - Project: `<a href="https://github.com/use-charter/charter" rel="noopener noreferrer">GitHub</a>`, `<a href="https://github.com/use-charter/charter/releases" rel="noopener noreferrer">Releases</a>`
    - Community: Discord link (render as `<!-- TODO: Discord URL when available -->` placeholder `<span>Community</span>` until URL confirmed)
    - Legal: `<a href="https://github.com/use-charter/charter/blob/main/LICENSE" rel="noopener noreferrer">Apache-2.0</a>`, `<span>© 2026 Charter</span>`
  - Clean layout (no card boxing; use `border-top` or negative space)

- [ ] **4.5** Create `web/src/styles/sections/social-proof.css`, `ci.css`, `final-cta.css`, `footer.css`
- [ ] **4.6** Wire all four components into `web/src/pages/index.astro` in order after Trust section

### Verification

```bash
cd web
bun run build
grep "Gate pull requests" dist/index.html && echo "✓ CI section H2"
grep "use-charter/charter-action" dist/index.html && echo "✓ GitHub Action snippet"
grep "SARIF" dist/index.html && echo "✓ SARIF mention"
grep -c 'rel="noopener noreferrer"' dist/index.html | awk '{print "noopener links:", $1}' && echo "✓"
grep "Apache-2.0" dist/index.html && echo "✓ license in footer"
grep "<footer" dist/index.html && echo "✓ semantic footer element"
grep "stargazers_count\|stars" dist/index.html && echo "✓ GitHub stars wired"
```

**Checkpoint:** CI section, GitHub Action YAML, footer, noopener links, semantic footer all present.

### Commit

```
feat(web/footer): social proof, CI, CTA, footer

- Create SocialProofSection.astro (build-time GitHub stars, 3 confirmed SVG icons, text badges for unconfirmed)
- Create CISection.astro (locked YAML snippet, SARIF mention, CTA link)
- Create FinalCTASection.astro (CopyButton reuse, three-tier CTA hierarchy)
- Create Footer.astro (four link columns, semantic footer, noopener on all external links)
- Wire all sections into index.astro

Result: Full page renders end-to-end.
```

---

## Phase 5 — WaitlistForm island (TDD)

**Goal:** Waitlist email capture form. Test-first. CopyButton is already done (Phase 2).

**Test framework:** Vitest. Add as dev dependency: `bun add -D vitest jsdom @vitest/coverage-v8`. Test file: `web/src/islands/WaitlistForm.test.ts`.

### TDD Cycles (vertical slices — one test at a time)

**Cycle 1: Email validation**
- [ ] Write test: `it('rejects invalid email: "notanemail"', ...)` → RED
- [ ] Implement `validateEmail(email: string): boolean` using `/^[^\s@]+@[^\s@]+\.[^\s@]+$/` → GREEN
- [ ] Write test: `it('accepts valid email: "user@example.com"', ...)` → RED → GREEN
- [ ] Write test: `it('rejects email with spaces: "user @example.com"', ...)` → RED → GREEN

**Cycle 2: Form submission**
- [ ] Write test: `it('disables submit button during POST', ...)` using `fetch` mock → RED
- [ ] Implement loading state: `button.disabled = true` before fetch, `false` after → GREEN
- [ ] Write test: `it('shows spinner during loading state', ...)` → RED → GREEN

**Cycle 3: Response handling**
- [ ] Write test: `it('shows success toast on HTTP 200', ...)` → RED
- [ ] Implement: `showToast('Check your email!', 'success')` on `res.ok` → GREEN
- [ ] Write test: `it('shows error toast on HTTP 400', ...)` → RED
- [ ] Implement: `showToast(data.error || 'Something went wrong', 'error')` on `!res.ok` → GREEN
- [ ] Write test: `it('shows error toast on network failure', ...)` → RED
- [ ] Implement: catch block with generic error toast → GREEN

### Tasks

- [ ] **5.1** Add Vitest config: `web/vitest.config.ts` with `environment: 'jsdom'`
- [ ] **5.2** Add test script to `web/package.json`: `"test": "vitest run"`, `"test:watch": "vitest"`
- [ ] **5.3** Create `web/src/islands/WaitlistForm.ts` (vanilla JS, no frameworks):
  - Attach to `<form id="waitlist-form">` via `DOMContentLoaded`
  - Email validation using `/^[^\s@]+@[^\s@]+\.[^\s@]+$/`
  - Inline validation: show error `<span id="email-error">` on blur if invalid; clear on input
  - Submit handler: `event.preventDefault()`, disable button, show spinner
  - POST to `/api/waitlist` with `{ email }` body, `Content-Type: application/json`
  - Success (res.ok): show success toast `"Check your email!"`, clear form
  - Error (!res.ok or network): show error toast with `data.error || "Something went wrong"`
  - Toast: injected `<div role="alert" aria-live="polite">` element; auto-dismiss after 4s

- [ ] **5.4** Create `web/src/components/WaitlistForm.astro`:
  - `<form id="waitlist-form" action="#" method="post" novalidate>`
  - `<label for="waitlist-email">Email address</label>` (above input)
  - `<input type="email" id="waitlist-email" name="email" required autocomplete="email" placeholder="you@example.com" />`
  - `<span id="email-error" role="alert" aria-live="polite"></span>` (below input, empty by default)
  - `<button type="submit" id="waitlist-submit">Notify me</button>`
  - `<script>` tag importing `WaitlistForm.ts` island

- [ ] **5.5** Replace `<WaitlistForm />` placeholder in `FinalCTASection.astro` with real component

- [ ] **5.6** Create `web/src/styles/islands.css`:
  - Form layout: label above, gap between label/input/error
  - Input: border from tokens, focus transition 200ms
  - Input `[aria-invalid="true"]`: red border
  - Submit button: full states (hover scale 1.02, active scale 0.98, disabled opacity)
  - Spinner: CSS `@keyframes spin`, `border-top` trick, `animation: none` under `prefers-reduced-motion`
  - Toast: fixed bottom-right, success/error color tokens, fade animation

### Verification

```bash
cd web
bun run test && echo "✓ all Vitest tests pass"
bun run build
ls -lh dist/_astro/*.js 2>/dev/null | awk '{sum+=$5} END {print "Total JS:", sum, "bytes (target: <5120)"}' || echo "no JS bundles (islands inline)"
grep "waitlist-form" dist/index.html && echo "✓ form in output"
grep 'for="waitlist-email"' dist/index.html && echo "✓ label wired to input"
grep 'aria-live' dist/index.html && echo "✓ live region for toast"
grep "prefers-reduced-motion" dist/index.html && echo "✓ reduced-motion in CSS"
```

**Checkpoint:** All Vitest tests green. Form in output. Label above input. aria-live present. Spinner respects reduced-motion.

### Commit

```
feat(web/islands): waitlist form (TDD — Vitest)

- Add Vitest config (jsdom environment)
- Create WaitlistForm.ts island (email validation, POST, loading/success/error states)
- Create WaitlistForm.astro (label above input, error below, aria-live toast)
- Create islands.css (form states, spinner, toast, reduced-motion)
- All 10 Vitest tests passing

Result: Email capture form with full state machine.
```

---

## Phase 6 — Performance optimization

**Goal:** All performance budgets met. Lighthouse ≥ 90.

### Tasks

- [ ] **6.1** Identify above-fold CSS (hero section only: typography, layout, button, terminal card):
  - Move these styles into a `<style>` block in `Base.astro` `<head>` (inline, <4kb target)
  - Move below-fold section CSS to `<link rel="stylesheet" media="print" onload="this.media='all'">` (deferred)

- [ ] **6.2** Verify hero image optimization:
  - `<Picture />` component auto-generates AVIF + WebP
  - Verify `loading="eager" fetchpriority="high"` on hero image only
  - Verify AVIF < 40kb, WebP < 50kb in `dist/` (check actual generated files)

- [ ] **6.3** Verify all below-fold images:
  - All use `<Picture />` with explicit `width` and `height`
  - All have `loading="lazy"`
  - Verify total `dist/_astro/*.avif` + `*.webp` + `*.png` ≤ 50kb combined

- [ ] **6.4** Verify font budget:
  - Fonts are base64 in `fonts.css` — check gzip size: `gzip -c web/src/styles/fonts.css | wc -c`
  - Must be ≤ 30kb gzip
  - If over budget: subset further (Latin only, drop unused weights)

- [ ] **6.5** Verify Astro builds with minification:
  - Confirm `astro.config.mjs` does NOT disable `compressHTML` (default: enabled)
  - Run `bun run build` and check `dist/index.html` is minified (no extra whitespace between tags)

- [ ] **6.6** Run Lighthouse performance audit:
  ```bash
  # Start dev server (separate terminal)
  bun run dev &
  # Run Lighthouse (use bunx)
  bunx lighthouse http://localhost:4321 --output html --output-path ./lighthouse-desktop.html --chrome-flags="--headless"
  # Target: Performance ≥ 90, Accessibility ≥ 90
  ```

### Verification

```bash
cd web
bun run build

# CSS inline check
grep -o '<style>' dist/index.html | wc -l | awk '{print "Inline style blocks:", $1}'

# Image budgets
find dist -name "*.avif" -o -name "*.webp" -o -name "*.png" 2>/dev/null | xargs du -cb 2>/dev/null | tail -1

# Font budget
gzip -c src/styles/fonts.css | wc -c | awk '{print "Fonts gzipped:", $1, "bytes (target: <30720)"}'

# Total HTML gzip
gzip -c dist/index.html | wc -c | awk '{print "HTML gzipped:", $1, "bytes"}'

# Lighthouse
bunx lighthouse http://localhost:4321 --output json --output-path /tmp/lh.json --chrome-flags="--headless" --quiet
node -e "const r=require('/tmp/lh.json'); console.log('Perf:', Math.round(r.categories.performance.score*100), '| A11y:', Math.round(r.categories.accessibility.score*100))"
```

**Checkpoint:** Lighthouse Performance ≥ 90, Accessibility ≥ 90. All budgets within limits.

### Commit

```
perf(web): inline critical CSS, verify all performance budgets

- Inline above-fold CSS (<4kb) into Base.astro head
- Defer below-fold section CSS (media=print trick)
- Verify hero image: eager loading, explicit dimensions
- Verify below-fold images: lazy loading, Picture components
- Verify font budget: ≤30kb gzip
- Lighthouse Performance ≥ 90, Accessibility ≥ 90 confirmed
```

---

## Phase 7 — Accessibility audit

**Goal:** WCAG 2.2 AA compliance. Zero axe-core violations. Keyboard nav verified.

**Note:** axe-core CLI requires a running dev server, not a file path.

### Tasks

- [ ] **7.1** Run automated accessibility scan:
  ```bash
  # Start dev server
  bun run dev &
  sleep 3
  # Run axe-core against running server
  bunx @axe-core/cli http://localhost:4321 --exit
  ```
  Fix every reported violation to zero. Do NOT disable rules to suppress counts.

- [ ] **7.2** Verify heading hierarchy:
  ```bash
  grep -oP '<h[1-6][^>]*>' dist/index.html | sort | uniq -c
  # Must have: exactly 1 <h1>, multiple <h2>, no <h2> skipped to <h4>
  ```

- [ ] **7.3** Verify all images have descriptive alt text:
  ```bash
  # Count img elements
  grep -c '<img\|<picture' dist/index.html
  # Count alt attributes (must equal img count)
  grep -c 'alt="' dist/index.html
  # Flag any alt="" (empty) on non-decorative images
  grep 'alt=""' dist/index.html && echo "WARNING: empty alt found"
  ```

- [ ] **7.4** Verify semantic structure:
  ```bash
  grep -c '<header' dist/index.html && echo "✓ header"
  grep -c '<main' dist/index.html && echo "✓ main"
  grep -c '<footer' dist/index.html && echo "✓ footer"
  grep -c '<section' dist/index.html | awk '{print "sections:", $1}'
  grep -c 'aria-labelledby' dist/index.html | awk '{print "aria-labelledby:", $1}'
  ```

- [ ] **7.5** Verify color contrast in Chrome DevTools:
  - Open `http://localhost:4321` in Chrome
  - DevTools → Accessibility → Full page accessibility tree
  - Flag any contrast failures
  - All text on `#0D1117` background must pass 4.5:1 (normal) or 3:1 (large)

- [ ] **7.6** Verify reduced-motion:
  ```bash
  grep -c "prefers-reduced-motion" dist/index.html && echo "✓ reduced-motion query present"
  # In Chrome DevTools: Rendering → Emulate CSS prefers-reduced-motion: reduce
  # Verify: no animations play, page still fully usable
  ```

- [ ] **7.9** Verify color mode (dark + light):
  - Chrome DevTools → Rendering → Emulate `prefers-color-scheme: dark`
    - All text readable; all 8 vendor icons visible
    - `icon--dark-invert` icons appear white (grok, copilot, cursor, windsurf, codex)
    - `icon--light-invert` chatgpt icon visible as white
    - Score zone colors (green/amber/red) readable on dark bg
  - Chrome DevTools → Rendering → Emulate `prefers-color-scheme: light`
    - All text readable on light bg
    - `icon--dark-invert` icons appear dark (filter removed) ✓
    - `icon--light-invert` chatgpt icon now black (filter: brightness(0)) ✓
    - Accent color (#2563EB) readable on light bg (passes 4.5:1)
  - Verify `<meta name="color-scheme" content="dark light">` in `dist/index.html`

- [ ] **7.7** Manual keyboard test:
  - Open `http://localhost:4321` in browser
  - Tab from top of page to bottom
  - Verify: every button, link, input reachable; focus ring visible (≥3px outline); no traps; order logical
  - Test: press Enter on copy button → "Copied!" feedback
  - Test: Tab into email input → type → Tab to submit → Enter
  - Test: Shift+Tab reverses correctly

- [ ] **7.8** Screen reader test:
  - Mac VoiceOver: `Cmd+F5` → read through entire page
  - Verify: page title announced, section headings announced, CTAs have meaningful labels, form fields labeled, errors described

### Verification

```bash
bun run dev &
sleep 3
bunx @axe-core/cli http://localhost:4321 --exit && echo "✓ axe-core: zero violations"
```

**Checkpoint:** axe-core exits 0. Heading hierarchy correct. All alt text present. Keyboard nav works. Screen reader passes.

### Commit

```
a11y(web): accessibility audit — zero violations

- Fix all axe-core violations (zero remaining)
- Verify heading hierarchy: 1×H1, N×H2
- Verify all images have descriptive alt text
- Verify color contrast ≥ 4.5:1 normal / 3:1 large
- Verify prefers-reduced-motion disables animations
- Verify semantic HTML (header, main, section, footer)
- Manual keyboard + VoiceOver test passed
```

---

## Phase 8 — Responsive refinement + cross-browser

**Goal:** No layout breakage at any supported viewport. Mobile Lighthouse targets met.

### Tasks

- [ ] **8.1** Test in Chrome DevTools Device Mode at each breakpoint:
  - 320px: no horizontal overflow on body (terminal card internal scroll OK); all text readable
  - 375px: hero CTA buttons stacked correctly, not overlapping
  - 768px: hero shifts from stacked to split-screen; section grids engage
  - 1024px: no over-stretched elements
  - 1440px: max-width container contains all content; no full-bleed text
  - 1920px: content doesn't span the entire viewport; max-width constrains layout

- [ ] **8.2** Fix any overflow issues:
  ```bash
  # Check for elements wider than viewport
  # Open DevTools console and run:
  # Array.from(document.querySelectorAll('*')).filter(e => e.offsetWidth > document.body.offsetWidth)
  ```

- [ ] **8.3** Verify touch targets ≥ 44×44px:
  - Copy button: measured in DevTools
  - Email input height: ≥ 44px
  - Submit button: ≥ 44px height
  - Footer links: adequate line height + padding

- [ ] **8.4** Run mobile-throttled Lighthouse:
  ```bash
  bunx lighthouse http://localhost:4321 \
    --output html \
    --output-path ./lighthouse-mobile.html \
    --form-factor=mobile \
    --screenEmulation.mobile=true \
    --screenEmulation.width=375 \
    --screenEmulation.height=812 \
    --chrome-flags="--headless"
  # Open lighthouse-mobile.html — target: LCP < 2.5s, FCP < 1.8s, CLS < 0.1
  ```

- [ ] **8.5** Verify CLS:
  - Chrome DevTools Performance tab → record page load → inspect Layout Shifts
  - Fonts must not cause reflow (font-display: swap + preload handles this)
  - Images must not shift (explicit width+height handles this)

- [ ] **8.6** Cross-browser spot check (manual):
  - Firefox: layout renders, fonts load, copy button works
  - Safari: layout renders, copy button works (iOS clipboard fallback)
  - Verify CSS Grid properties used have cross-browser support (all target browsers support CSS Grid)

### Verification

```bash
# Mobile Lighthouse
bunx lighthouse http://localhost:4321 \
  --output json \
  --output-path /tmp/lh-mobile.json \
  --form-factor=mobile \
  --screenEmulation.mobile=true \
  --chrome-flags="--headless" \
  --quiet
node -e "const r=require('/tmp/lh-mobile.json'); const lcp=r.audits['largest-contentful-paint'].numericValue; const cls=r.audits['cumulative-layout-shift'].numericValue; console.log('LCP:', Math.round(lcp)+'ms (target:<2500) | CLS:', cls.toFixed(3), '(target:<0.1)')"
```

**Checkpoint:** LCP < 2500ms, CLS < 0.1 on mobile. No overflow at any breakpoint. Touch targets ≥ 44px.

### Commit

```
test(web): responsive refinement + cross-browser verification

- Test at 320, 375, 768, 1024, 1440, 1920px — no overflow
- Verify touch targets ≥ 44px (copy, input, submit, links)
- Mobile Lighthouse: LCP < 2.5s, CLS < 0.1
- Cross-browser: Chrome, Firefox, Safari tested
```

---

## Phase 9 — Deployment & Worker integration

**Goal:** Landing page live at `use-charter.dev/`. Worker routing correct. HTTPS verified.

### Pre-conditions

- All phases 1–8 committed and passing
- Cloudflare account access available
- `wrangler` CLI available or Cloudflare Pages dashboard used directly

### Tasks

- [ ] **9.1** Create Cloudflare Pages project:
  - Project name: `charter-landing` (or similar)
  - Connect to `use-charter/charter` GitHub repo
  - Build command: `bun run build`
  - Root directory: `web`
  - Output directory: `dist`
  - Branch to deploy: `main`

- [ ] **9.2** Verify first Pages deployment:
  - Visit `https://charter-landing.pages.dev/` (or assigned pages.dev URL)
  - Hero renders, terminal card shows, copy button works
  - No 404s on assets (fonts, screenshots, icons)

- [ ] **9.3** Locate existing Cloudflare Worker:
  - Check repo for `wrangler.toml` at root (confirmed present: `./.miserc.toml` is mise config, separate from wrangler)
  - If Worker is deployed outside repo (CF dashboard), access via `wrangler whoami` + `wrangler deploy`
  - Worker currently routes: `/docs/*` → Mintlify, `/rules/*` → Mintlify

- [ ] **9.4** Update Worker to route `/` → landing origin:
  - Add `LANDING_ORIGIN` environment variable: set to `https://charter-landing.pages.dev`
  - Add route handler for `/` (and assets like `/_astro/*`, `/fonts/*`, `/icons/*`, `/screenshots/*`, `/favicon.svg`, `/og.svg`)
  - Preserve existing Mintlify routes unchanged
  - Test locally with `wrangler dev` before deploying

- [ ] **9.5** Deploy updated Worker:
  ```bash
  wrangler deploy
  ```

- [ ] **9.6** Verify end-to-end routing:
  ```bash
  curl -sI https://use-charter.dev/ | head -10
  # → HTTP/2 200, content-type: text/html

  curl -sI https://use-charter.dev/docs | head -5
  # → HTTP/2 302 or 301 to Mintlify

  curl -sI https://use-charter.dev/rules | head -5
  # → HTTP/2 302 or 301 to Mintlify

  curl -sI https://use-charter.dev/favicon.svg | head -5
  # → HTTP/2 200, content-type: image/svg+xml
  ```

- [ ] **9.7** Update `docs/product/DEPLOY.md`:
  - Add new section: "Landing page (Cloudflare Pages)"
  - Document `LANDING_ORIGIN` environment variable
  - Document how to deploy: `bun run build` from `web/`; Pages auto-deploys on push to `main`
  - Document Worker routing: which routes go where

- [ ] **9.8** Final smoke test (browser):
  - Visit `https://use-charter.dev/`
  - Copy button: click → "Copied!" → clipboard has `brew install use-charter/tap/charter`
  - Terminal card: renders with 94/100 green
  - "View on GitHub" link: opens github.com/use-charter/charter
  - "Read the docs" link: opens use-charter.dev/docs
  - Form: type email → submit → loading spinner → response handling

- [ ] **9.9** Verify HTTPS:
  ```bash
  curl -sI https://use-charter.dev/ | grep -i "strict-transport"
  # Cloudflare enforces HSTS by default
  ```

### Verification

```bash
curl -sv https://use-charter.dev/ 2>&1 | grep -E "HTTP/|< content-type:|< server:"
# Expected: HTTP/2 200, content-type: text/html, server: cloudflare

curl -sI https://use-charter.dev/docs
# Expected: 301 or 302 to Mintlify URL

curl -sI https://use-charter.dev/rules
# Expected: 301 or 302 to Mintlify URL
```

**Checkpoint:** Landing live at use-charter.dev. Worker routing correct. HTTPS enforced. Smoke test passes.

### Commit

```
feat(web): deploy to Cloudflare Pages + wire Worker routing

- Create Cloudflare Pages project (build: bun run build, output: dist/)
- Update Cloudflare Worker: add LANDING_ORIGIN route for /
- Preserve existing /docs/* and /rules/* Mintlify routes
- Update docs/product/DEPLOY.md with LANDING_ORIGIN and Pages setup
- Verified: curl shows 200 at /, 30x at /docs and /rules
- HTTPS enforced, smoke test passed
```

---

## Commit Sequence Summary

| Phase | Commit type | Scope | Summary |
|-------|-------------|-------|---------|
| 1 | feat | web | scaffold Astro + design tokens + fonts |
| 2 | feat | web/hero | hero + terminal card + copy button |
| 3 | feat | web/sections | problem, solution, value props, trust |
| 4 | feat | web/footer | social proof, CI, CTA, footer |
| 5 | feat | web/islands | waitlist form (TDD — Vitest) |
| 6 | perf | web | inline critical CSS, verify all budgets |
| 7 | a11y | web | accessibility audit — zero violations |
| 8 | test | web | responsive refinement + cross-browser |
| 9 | feat | web | deploy to Cloudflare Pages + wire Worker |

All commits GPG-signed. Attribution:
```
Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
```

---

## Cross-Task Invariants

These hold across all phases. An agent violating any of these must stop and report before continuing.

| Invariant | Rule |
|-----------|------|
| Go codebase untouched | No changes to Go source, tests, or CLI |
| `moon run :check` green | Must pass before every push (Go tests, lint, docs all cached or green) |
| GPG-signed commits | All commits use `-S` flag |
| Bun exclusively | No `npm`, no `npx` — use `bunx` for one-off CLI tools |
| Copy verbatim | Never reword locked copy; every section's text matches spec exactly |
| No third-party JS | Islands use vanilla JS only; no React, no Vue, no libraries |
| No analytics or tracking | Pure static output; no tracking scripts |
| Performance budgets are hard limits | JS ≤ 150kb, CSS ≤ 30kb, images ≤ 50kb, fonts ≤ 30kb (all gzipped) |
| Tech stack locked | Astro v6 + vanilla CSS + Bun. No changes mid-phase |
| Dev server port | Astro default: `http://localhost:4321` (not 3000) |

---

## Phase Dependency Map

```
Phase 1 (scaffold — must complete first)
  ↓
  ├─────────────────────────────────────────────────┐
  ├─→ Phase 2 (hero + CopyButton) ←─────────────┐  │
  │                                              │  │
  ├─→ Phase 3 (sections 2–5)                    │  │
  │                                              │  │
  └─→ Phase 4 (sections 6–9)                    │  │
                                                 │  │
  [Phase 2 must complete before Phase 5]         │  │
                                                 ↓  │
                                          Phase 5   │
                                          (form)  ←─┘
                                             ↓
                                          Phase 6 (perf)
                                             ↓
                                          Phase 7 (a11y)
                                             ↓
                                          Phase 8 (responsive)
                                             ↓
                                          Phase 9 (deploy)
```

**Parallel window:** After Phase 1, dispatch Phases 2 + 3 + 4 simultaneously. Block on all three before dispatching Phase 5.

---

**Plan status:** Ready for execution. All asset paths verified. All copy locked. All dependencies documented. Sub-agent dispatch prompts ready.
