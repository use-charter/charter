# Charter Brand Assets

Last reviewed: 2026-06-02

Canonical brand source for the Charter wordmark, icons, and web metadata. Visual identity (colors + fonts) is governed by [`../DESIGN-TOKENS.md`](../DESIGN-TOKENS.md); this folder holds the deployable source assets. All SVGs are valid XML (no `--` in comments) and render in strict engines (resvg / takumi / browsers).

## The mark

Charter's mark is a geometric **`[C]`** — a left bracket, a 300° `C` arc, a right bracket — drawn on a 64-grid, font-independent, and symmetric. `mark.svg` uses `fill="currentColor"` so it inherits the surrounding text color (site, docs, README).

| File | Size | Purpose |
|---|---|---|
| `mark.svg` | 64×64, `currentColor` | canonical wordmark glyph (site/docs/inline) |
| `favicon.svg` | 32×32, white-on-dark square | browser tab (modern) |
| `apple-touch-icon.svg` | 180×180, white-on-dark | iOS/iPad home screen → export `apple-touch-icon.png` |
| `icon-maskable.svg` | 512×512, padded safe zone | Android PWA maskable → export `icon-maskable-512.png` |
| `og.svg` | 1200×630 | social card design source → render `og.png` |
| `manifest.json` | — | PWA web app manifest |
| `meta.html` | — | `<head>` snippet (OG/Twitter/icons/fonts/tokens) |

## Colors

Charter is **dark-first**. Surface `#0D1117` (Charter Dark), mark/text `#FFFFFF`, accent blue `#2563EB`. Full token set in `../DESIGN-TOKENS.md`. The favicon is an intentional **solid dark square** rather than a transparent dark-mode-adaptive glyph — Safari ignores `prefers-color-scheme` in SVG favicons (as of 2026), so owning a brand-colored square is the robust, cross-platform choice.

## Fonts

Per the canonical system: **Ruda** (wordmark/headings), **Atkinson Hyperlegible Mono** (CLI/code), **IBM Plex Mono** (rule IDs / metadata), **Share Tech** (accents). All SIL OFL 1.1. The OG wordmark uses Ruda 900.

## Generating raster assets

No Satori. Use **[takumi](https://github.com/kane50613/takumi)** — a fast Rust, `next/og`-compatible engine — as the render path:

- **`og.png`** — port `og.svg`'s layout to a takumi template (next/og-compatible JSX) and render at 1200×630; keep ≤ 300 KB. (Or rasterize `og.svg` with Playwright so the Google-Fonts `@import` resolves.)
- **`apple-touch-icon.png`** (180×180), **`icon-192.png`** / **`icon-512.png`** (PWA, `purpose:"any"`), **`icon-maskable-512.png`** (from `icon-maskable.svg`, `purpose:"maskable"`) — rasterize with resvg or any SVG→PNG tool.
- **`favicon.ico`** (16/32/48) — derive from `favicon.svg` for legacy fallback.

## Tagline

Primary: **“AI-agent readiness, scored.”** Hero motif (command triad): **Scan · Score · Fix**. Alternatives on the table: *“The agent-readiness score for your repo.”*, *“Make every repo agent-ready.”*

## Consistency note

The terminal UI wordmark should use the canonical **`[C]` charter** lockup (the `[C]` renders natively as ASCII in any terminal), superseding the placeholder `✦` glyph used in early TUI mockups.
