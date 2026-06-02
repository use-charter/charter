# Charter Design Tokens & Visual Identity

Last reviewed: 2026-06-02

Canonical **visual-identity** reference for the terminal experience (Slice 15, ADR-0024) and the HTML report (Slice 16, ADR-0025), and future web/docs. Product *behavior* authority stays `docs/internal/architecture/charter-architecture-2026.md`; this owns visual tokens only. Visual mockups: `docs/internal/designs/*.html` (inspiration to surpass, not copy).

## Color tokens (semantic)

Light/dark via `prefers-color-scheme`; in the TUI each maps to a Lipgloss `AdaptiveColor`, degrading truecolor→256→16→mono (ADR-0024).

- **text**: `--color-text-{primary, secondary, tertiary, success, danger, warning, info}`
- **background**: `--color-background-{primary, secondary, tertiary, success, danger, warning, info}`
- **border**: `--color-border-{primary, secondary, tertiary, success, danger, warning, info}`
- **radius**: `--border-radius-{md, lg}`

Severity mapping: `BLOCKER → danger`, `HIGH → warning`, `MEDIUM → warning (dimmed)`, `LOW → info`, `PASS/✓ → success`, rule-id/accent → `info`. **Severity always carries a text label — never color alone** (WCAG 1.4.1 / 1.3.1).

## Concrete palette

Canonical concrete values for the semantic tokens above, defined by the terminal layer (Slice 15, `internal/terminal`) and reused by the HTML report (Slice 16). The mockups (`*.html`) reference token *names*; these are the values they resolve to. Each token carries a light and a dark 24-bit hex (the Lipgloss v2 LightDark pair), an ANSI-16 fallback, and a Mono attribute fallback. ANSI-256 is **not** enumerated — it is computed at runtime as the nearest xterm-256 index to the selected hex (`internal/terminal.nearestANSI256`).

**Text**

| Token | Light | Dark | ANSI-16 (light / dark) | Mono |
|---|---|---|---|---|
| `text-primary` | `#111827` | `#f9fafb` | terminal default fg | — |
| `text-secondary` | `#4b5563` | `#d1d5db` | terminal default fg | faint |
| `text-tertiary` | `#646b78` | `#9ca3af` | terminal default fg | faint |
| `text-success` | `#15803d` | `#4ade80` | green `2` / `10` | bold |
| `text-danger` | `#b91c1c` | `#f87171` | red `1` / `9` | bold |
| `text-warning` | `#b45309` | `#fbbf24` | yellow `3` / `11` | bold |
| `text-info` | `#1d4ed8` | `#60a5fa` | blue `4` / `12` | bold |

**Background** (surfaces; subtle tints in truecolor/256, vivid hue only in the lossy ANSI-16)

| Token | Light | Dark | ANSI-16 (light / dark) | Mono |
|---|---|---|---|---|
| `background-primary` | `#ffffff` | `#0d1117` | terminal default (no fill) | — |
| `background-secondary` | `#f9fafb` | `#161b22` | terminal default (no fill) | — |
| `background-tertiary` | `#f3f4f6` | `#1f2937` | terminal default (no fill) | — |
| `background-success` | `#f0fdf4` | `#052e16` | green `2` / `10` | reverse |
| `background-danger` | `#fef2f2` | `#450a0a` | red `1` / `9` | reverse |
| `background-warning` | `#fffbeb` | `#422006` | yellow `3` / `11` | reverse |
| `background-info` | `#eff6ff` | `#172554` | blue `4` / `12` | reverse |

**Border** (intentionally subtle; severity is always carried by a text label per WCAG 1.4.1, so borders are never the sole state indicator)

| Token | Light | Dark | ANSI-16 (light / dark) | Mono |
|---|---|---|---|---|
| `border-primary` | `#d1d5db` | `#374151` | terminal default | — |
| `border-secondary` | `#e5e7eb` | `#2b3240` | terminal default | — |
| `border-tertiary` | `#eceef1` | `#21262d` | terminal default | faint |
| `border-success` | `#86efac` | `#166534` | green `2` / `10` | bold |
| `border-danger` | `#fca5a5` | `#991b1b` | red `1` / `9` | bold |
| `border-warning` | `#fcd34d` | `#92400e` | yellow `3` / `11` | bold |
| `border-info` | `#93c5fd` | `#1e40af` | blue `4` / `12` | bold |

Notes:

- **ANSI-16 codes** are Lipgloss/ANSI indices: `1` red, `2` green, `3` yellow, `4` blue (+8 for the bright variant). The light variant uses the base index; the dark variant uses the bright index. Neutral tokens have no faithful ANSI-16 gray, so they fall back to the terminal default and rely on attributes for hierarchy.
- **ANSI-256** examples (nearest to the hex): `text-success` → `29` (light) / `78` (dark); `text-danger` → `124` (light); `text-info` → `26` (light).
- **Mono** drops all color (terminal default) and builds hierarchy with attributes only: primary plain, secondary/tertiary faint, semantic bold, semantic surfaces reverse-video.
- **Contrast (WCAG 2.2 AA):** verified against light surfaces (`#ffffff`, `#f9fafb`, `#f3f4f6`) and dark surfaces (`#0d1117`, `#1e1e1e`, `#1f2937`). Every text token clears 4.5:1 on its mode's surfaces, and each semantic text token also clears 4.5:1 on its own tinted background.

## Font system (canonical — product-wide)

Confirmed `:root` tokens — the single source of truth for every Charter surface (site, docs, and the audit report):

```css
:root {
  --font-site:   "Ruda", system-ui, sans-serif;                         /* site / marketing / UI / audit body + headings */
  --font-audit:  "Ruda", system-ui, sans-serif;                         /* audit report body (same family as site) */
  --font-code:   "Atkinson Hyperlegible Mono", ui-monospace, monospace; /* PRIMARY code / CLI font */
  --font-meta:   "IBM Plex Mono", ui-monospace, monospace;              /* optional formal metadata layer */
  --font-accent: "Share Tech", system-ui, sans-serif;                   /* system labels / status accents only */
}
```

| Surface / element | Token | Family |
|---|---|---|
| Site + audit body; headings (700/800); finding titles (700); table text (500); badges (700 uppercase) | `--font-site` / `--font-audit` | **Ruda** |
| **CLI blocks, commands, terminal output, code snippets, config examples, inline code** | `--font-code` | **Atkinson Hyperlegible Mono** |
| Dense identifiers — rule IDs (`AE-SEC-001`), rule paths (`repo.rules.security.no-env-leak`), file paths (`src/config/auth.ts`), `confidence: 0.92` | `--font-meta` | **IBM Plex Mono** (optional) |
| System labels / status accents | `--font-accent` | **Share Tech** (accent only) |

Rules of thumb:
- **`--font-code` (Atkinson Hyperlegible Mono) is the primary monospace** for all code/CLI/inline — legibility-first, fitting a correctness tool. **Do not default code/rules to IBM Plex Mono.**
- `--font-meta` (IBM Plex Mono) is **optional**, applied only for a more formal audit-metadata layer (rule IDs / namespaces / file paths / confidence).
- **Ruda** (Omnibus-Type) is the single sans across site + audit — no separate display face (**Space Grotesk dropped, too common**); the `✦ charter` wordmark uses Ruda 700/800. **Share Tech** is accents only.
- All families are **SIL OFL 1.1** → embeddable offline. The **TUI is terminal-inherited** (sets no typeface); this system governs the HTML report + web/docs.

## Offline / self-contained constraint (ADR-0001, ADR-0025)

The HTML report embeds **Latin-subset `woff2`** of the required families/weights directly in the inlined CSS (base64) — **no CDN, no external fetch**. Keep the payload bounded: minimal weights (Ruda 400/500/700/800; Atkinson Hyperlegible Mono 400/500 — primary code/CLI; IBM Plex Mono 400/500 — optional metadata; Share Tech 400 — accent only, embed only if used), Latin subset only. Every stack ends in a system fallback:

- sans → `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif`
- mono → `ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace`

If the embed budget is ever exceeded, fall back to the system stacks **before** adding any network dependency — offline-first wins over brand fidelity.

**TUI (Slice 15) — fonts are terminal-inherited.** A TUI cannot set typefaces; Charter controls only color, layout, and glyphs (with ASCII fallback). The font system above applies to the **HTML report + web/docs**, not the terminal.

## Design skills to apply (installed — verified on disk 2026-06-02)

Project-level design skills live in **`.agents/skills/`** (gitignored — local-only, fully available to the agent at build time); user-level + plugin skills are also available. Read the relevant ones at implementation.

**Token & system architecture (governs this file):**
- **`design-system-patterns`** — `.agents/skills/` — design tokens, semantic token hierarchy (primitive → semantic → component), light/dark via CSS custom properties. The skill behind this token system.
- **`visual-design-foundations`** — `.agents/skills/` — typography scale, color theory, spacing, iconography, hierarchy, dark mode.

**Report build (Slice 16):**
- **`frontend-design`** — `.agents/skills/` — distinctive, production-grade frontend; avoids generic AI aesthetics. **Lead.**
- **`web-design-guidelines`** — `.agents/skills/` — Web Interface Guidelines / accessibility audit. **The WCAG 2.2 AA gate.**
- **`design-taste-frontend`** — `.agents/skills/` — senior UI/UX: metric-based rules, component architecture, hardware-accelerated CSS. (Prefer over `design-taste-frontend-v1`.)
- **`high-end-visual-design`** — `.agents/skills/` — premium-agency polish; blocks cheap defaults.
- **`interaction-design`** — `.agents/skills/` — microinteractions, transitions, loading/empty states, hover/focus feedback (the report's client-side interactivity).
- **`ui-ux-pro-max`** — `.agents/skills/` — palettes, font pairings, typography, color, accessibility, charts.
- **`design-md`** — `.agents/skills/` — synthesize the semantic system into a DESIGN.md if helpful.
- **`web-component-design`** — `.agents/skills/` — composition/structure thinking only; the report is **vanilla single-file HTML/CSS/JS**, so this applies as guidance, not framework code.

> All skills are sourced from **`.agents/skills/`** only (no `~/.claude/` or plugin skills). Screenshot-verify the rendered report against the design refs + `web-design-guidelines` as a build step (no dedicated screenshot skill required).

**TUI (Slice 15):** `ui-ux-pro-max` + `design-taste-frontend` + `visual-design-foundations` (palette / contrast / type-scale, terminal-adapted) + `interaction-design` (spinner / rescan / state feedback).

**Other surfaces (later slices):** `landing-page-design` → launch website (Slice 19). Out of scope for 15/16: `redesign-existing-projects`, `sleek-design-mobile-apps`, `ckm-banner-design`, `api-and-interface-design`, `design-taste-frontend-v1`.

## References
- `docs/internal/designs/*.html` (mockups: doctor/init/fix, suppress/report/version, html-report)
- `docs/internal/decisions/0024-interactive-tui-and-terminal-output.md` (tokens → Lipgloss)
- `docs/internal/decisions/0025-report-delivery-self-contained-html.md` (report, offline font embedding)
