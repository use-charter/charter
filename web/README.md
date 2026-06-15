# web/

Charter's public web presence — the marketing site, the founder dashboard, and
the blog — built with [Astro](https://astro.build) and deployed to Cloudflare
Pages at [use-charter.dev](https://use-charter.dev). (The product documentation
is a separate Mintlify site under [`docs/product/`](../docs/product/); the apex
router that stitches the two together lives in [`infra/`](../infra/).)

Static-first: every page prerenders to HTML. Interactivity is small vanilla-TS
"islands" hydrated per page — no UI framework.

## Pages

| Route | Source | What it is |
|-------|--------|------------|
| `/` | `src/pages/index.astro` | Marketing landing — hero, the doctor session, rule + lifecycle sections, waitlist. |
| `/dashboard` | `src/pages/dashboard.astro` | Founder mission-control (stars, traffic, releases, community). Gated by Cloudflare Access; stats come from the apex router's `/dashboard/api/stats`. |
| `/blog` · `/blog/<slug>` | `src/pages/blog/` | Editorial blog from Markdown in `src/content/blog/`, with reading progress + TOC scroll-spy. |
| `/legal/{privacy,terms,license}` | `src/pages/legal/` | Legal pages via the shared `LegalDoc` component. |
| `/blog/rss.xml` · `/og/<slug>.png` | `src/pages/blog/rss.xml.ts` · `src/pages/og/[...slug].png.ts` | Hand-rolled RSS feed; per-post Open Graph cards rendered at build (satori → SVG, resvg → PNG). |
| `/llms.txt` · `/robots.txt` · sitemap | `src/pages/*.ts`, `@astrojs/sitemap` | Machine-readable surfaces for crawlers and agents. |

## Layout

```
src/
├── pages/        route files (.astro pages + .ts endpoints)
├── components/   reusable UI — SiteNav, SiteFooter, ThemeSwitch, LegalDoc
├── islands/      hydrated vanilla-TS behavior — landing, dashboard, blog, legal, theme, footer, WaitlistForm
├── layouts/      Base.astro — <head>, SEO/OG meta, JSON-LD, theme bootstrap
├── content/      blog Markdown + content.config.ts (content-layer schema)
├── styles/       design-tokens.css + per-surface CSS (marketing, legal, blog, dashboard)
├── assets/       fonts (incl. static Ruda instances for OG rendering)
└── lib/          shared helpers
functions/api/    Cloudflare Pages Functions — waitlist.ts (waitlist signup endpoint)
```

`SiteNav` and `SiteFooter` are the shared chrome for secondary pages (legal,
blog); each consuming page calls `initThemeSwitch()` / `initFooterGlow()`. The
home page has its own primary nav. Theme is a three-state switcher
(system / light / dark) resolved pre-paint in `Base.astro`.

## Develop

```bash
bun install
bun run dev       # astro dev (localhost:4321)
bun run build     # static build → dist/
bun run check     # astro check (type-check .astro + .ts)
bun run test      # vitest
```

From the repo root these run under Moon (`web:build`, `web:test`, …) and fold
into `moon run :check`. The type gate is `astro check`; tests are Vitest
(`src/**/*.test.ts`).

## Conventions

- **Design tokens, not hardcoded values.** Colors/space/type live in `design-tokens.css` as `--color-*` / `--ck-*`; both light and dark are first-class.
- **Compositor-friendly motion.** Animate `transform`/`opacity`; gate entrances behind `prefers-reduced-motion`.
- **Self-contained SEO.** Canonical, OG/Twitter, JSON-LD, RSS, and per-post OG images are generated at build — new blog posts are zero-config.
- **Deploy.** Cloudflare Pages builds on changes under `web/`; `wrangler.toml` holds the Pages config.
