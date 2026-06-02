# Phase 1 Slice 16 Design — Report Delivery (Self-Contained HTML)

## Goal

Give the already-promised `charter report` command its v1.0 purpose: a **self-contained single-file HTML report** — all data, CSS, and JS inlined — that opens from `file://`, works fully offline, and is one human's complete, shareable view of one repo's one scan. Implements **ADR-0025**; depends on **Slice 15** (reuses the finding/category model + the `charter explain` rationale + the shared design tokens). Score formula (ADR-0008) unchanged. Visual inspiration (not a copy): `docs/internal/designs/charter-html-report.html`.

## Audience

- **Primary**: the developer/owner who ran the scan (or pulled it as a CI artifact) and wants the full, navigable picture in a browser.
- **Secondary**: a teammate/lead who receives the file and must understand it **without the CLI installed**.
- **Not**: agents/CI (they consume `json`/`sarif`), nor a multi-repo fleet dashboard (future hosted/paid tier).

## Scope

### In scope
- `internal/render/html` — a pure renderer: `doctor.Result` → one self-contained `.html` (data inlined as a JSON payload; CSS/JS/template embedded via `go:embed`).
- `charter report` command: `--format html` (default), `--format markdown`, `--format json` (reuse existing renderers), `--out <file>` (default `charter-report.html`), `--open` (off by default).
- Report sections (below), client-side interactivity (filter/search/expand/copy/light-dark) — **all inlined, zero network**.
- WCAG 2.2 AA conformance; light/dark; deterministic ordering.

### Out of scope
- `charter report --serve` (deferred — ADR-0025 records the hardening checklist).
- `--format spdx` (SBOM, Phase 1.5); hosted `use-charter.dev/reports` (rejected for OSS core).
- run history/trends/diffs, multi-repo aggregation, live re-scan, fix-application, raw-source embedding.

## Grounding (verified 2026-06-02)
- **Self-contained file is the correct architecture**: Lighthouse, `go tool cover -html`, Bun standalone-html, vite-plugin-singlefile all ship one portable file; a `file://` page **cannot `fetch()` a sibling JSON**, so data must be **inlined** (the decisive constraint). Local-server viewers exist only for *dynamic* reports; Charter's is static.
- **Security** (why no server, why not hosted): the 0.0.0.0-Day class + DNS rebinding make a localhost server an unnecessary surface for a static report; hosting findings about private repos contradicts offline-first (ADR-0001). SARIF → the customer's own GitHub Security tab already covers the "hosted view" need at zero privacy cost.
- **Model reuse**: renders `doctor.Result` (findings/suppressed/summary/score/categories) + the Slice-15 `charter explain` rationale for per-rule cards — authored once, shown in CLI, TUI, and report.
- **Design/skills/fonts**: inspiration from `docs/internal/designs/charter-html-report.html`; tokens + fonts canonical in `docs/internal/designs/DESIGN-TOKENS.md`; elevated with the installed `.agents/skills/` design set (full map in `DESIGN-TOKENS.md`): `frontend-design` (lead), `web-design-guidelines` (WCAG 2.2 AA gate), `design-system-patterns` + `visual-design-foundations` (tokens/type/color), `interaction-design` (microinteractions/feedback), `design-taste-frontend`, `high-end-visual-design`, `ui-ux-pro-max` — respect the token system; screenshot-verify at build.

## Report structure (mapped to the reference, to be surpassed)
1. **Top bar + nav** — `✦ charter` + version + repo name; section nav (Score / Findings N / Suppressions).
2. **Provenance strip** — repo · git ref · commit · timestamp · `threshold · profile` · duration.
3. **Score hero** — large numeral in severity color, `/100`, PASS/FAIL pill + zone label; `role=progressbar` bar with tick labels; the score-formula line; an **active-cap alert** (what happened · why · how to fix) when a cap applies.
4. **Category scorecard** — grid of category cards (icon, name, rules count, worst severity, deduction; passed = ✓), left-border severity color.
5. **Findings** — count badge + expand/collapse-all; an **auto-fix CTA** strip when any finding is auto-fixable; severity **filter pills** (with counts; disabled at 0; `aria-pressed`) + **search**; expandable **finding cards** (severity badge, rule id, category, OWASP tag, "Fix available" badge, summary, `path:line` + copy; body: Evidence (redacted), Remediation, Auto-fix, **Why this matters** = `charter explain` rationale, rule-doc `helpUri` link); an **empty state** for over-filtering.
6. **Suppressions** — table (Rule/Reason/Expires/Status) + a governance **alert** (e.g. AE-SUPPRESS-002) with the exact resolution command.
7. **Scan summary** — metric cards (rules checked, active findings, suppressed, auto-fixable) + a severity-breakdown bar.
8. **Footer** — `✦ charter` version · "self-contained · offline · <timestamp>" · `use-charter.dev`.

## Typography & fonts (canonical — `DESIGN-TOKENS.md`)
- **Ruda** (`--font-site`/`--font-audit`) — body (400/500), headings (700/800), finding titles (700), table text (500), badges (700 uppercase); the `✦ charter` wordmark uses Ruda 700/800.
- **Atkinson Hyperlegible Mono** (`--font-code`) — **primary** mono: CLI blocks, commands, terminal output, code snippets, config examples, inline code (400/500).
- **IBM Plex Mono** (`--font-meta`) — *optional* formal metadata layer only: rule IDs, rule paths, file paths, confidence. Do **not** default code/rules to it.
- **Share Tech** (`--font-accent`) — system labels / status accents only.
- All SIL OFL 1.1 → **embedded as Latin-subset `woff2` inline** (base64), minimal weights, **no CDN/fetch** (ADR-0001/0025). Every stack ends in a system fallback; if the embed budget is exceeded, fall back to system fonts before any network use.

## Accessibility (WCAG 2.2 AA — minimum, per the reference)
- 2.1.1 keyboard-operable finding headers (`role=button`, `tabindex`, Enter/Space); 2.4.7 visible focus; 2.5.5/2.5.8 target size; 3.3.2 input labels (search `aria-label`); 4.1.2 name/role/value (`role=progressbar`, `aria-expanded`, `aria-pressed`); 1.3.1 semantic structure (`header`/`main`/`section`/`article`/`footer`, `aria-label`s). Severity always carries a text label; never color-only. Verified against WCAG 2.2 AA with `web-design-guidelines` (audit) + a screenshot pass at build.

## Architecture / ownership
- New: `internal/render/html` (+ embedded `assets/` via `go:embed`). `cmd/charter` gains `report`. Reuses `internal/explain` + the finding/category model from Slice 15.
- Avoid: external assets/network of any kind; non-determinism (stable ordering); embedding raw source (redacted evidence + `path:line` only); duplicating the renderers (markdown/json reuse existing).

## Testing & verification
- **Self-containment test (load-bearing)**: assert the rendered HTML references **no** `http(s)://` asset and no external `src`/`href` (links to `helpUri` are allowed as anchors but not as loaded resources); a known `Result` fixture renders deterministically (golden, timestamp normalized).
- **Content**: every finding/suppression/category from a fixture appears; redaction preserved; cap notice present when capped; empty-findings repo renders a clean PASS report.
- **Renderer reuse**: `--format markdown|json` from `report` matches the existing renderers byte-for-byte.
- **A11y**: structural assertions (roles/labels present) in the golden; manual WCAG 2.2 AA pass via `web-design-guidelines` (screenshots).
- **Dogfood**: `charter report` on Charter itself produces a valid, self-contained PASS report; `moon run :check` green; ≥85% coverage on `internal/render/html`.

## Success criteria
- `charter report --format html` emits one self-contained, offline, WCAG-2.2-AA file matching the elevated design; `--out`/`--open`/markdown/json work; self-containment + golden tests green.
- ADR-0025 + this spec + plan committed; architecture `charter report` purpose aligned to html (spdx labeled Phase 1.5); HTML mirror regenerated; dogfood green; no stale references.

## References
- `docs/internal/decisions/0025-report-delivery-self-contained-html.md`; `0024` (TUI/tokens + explain); `0001` (offline-first); `0008` (score, unchanged); `0014` (catalog)
- `docs/internal/designs/charter-html-report.html`; `docs/internal/designs/DESIGN-TOKENS.md` (canonical colors + fonts + installed design skills)
- `docs/internal/superpowers/plans/2026-06-02-phase-1-slice-16.md`
