# Phase 1 Slice 15 Design — Terminal Experience (Styled Output + Interactive TUI)

## Goal

Make Charter's terminal surface world-class without touching its machine contract: a **pristine styled** non-interactive default and an **opt-in interactive TUI** (`charter doctor -i`), plus the `charter explain <RULE>` rule surface and `charter doctor --rule` filtering. Offline, deterministic, no-LLM, no-network (Commitments #4/#7). Implements **ADR-0024**; the score formula (ADR-0008) is unchanged. Visual source of truth: `docs/internal/designs/charter-doctor-init-fix.html` and `docs/internal/designs/charter-supress-report-version.html`.

## Audience

- coding agents implementing the slice (subagent-driven)
- maintainers reviewing the TTY/non-TTY containment contract and dependency expansion

## Scope

### In scope
- **Charm v2 stack** (`charm.land/bubbletea/v2` v2.0.7, `charm.land/lipgloss/v2` v2.0.3, `charm.land/bubbles/v2` v2.1.0) + **`github.com/charmbracelet/fang` v1.0.0** wrapping the Cobra root (styled help/usage/errors, `--version`, man pages). `glamour` only if markdown is rendered in-TUI.
- `internal/terminal` — capability detection + the adaptive palette (the design tokens → Lipgloss).
- `internal/render/text` — the **styled** non-interactive renderer for `doctor` (TTY) + its plain/mono fallback (pipe/`NO_COLOR`/CI), matching the design references.
- `internal/tui` — the `charter doctor -i/--interactive` master-detail browser.
- `internal/explain` (over `internal/rules/catalog`) + the `charter explain <RULE>` command (`text` default, `--format json`).
- `cmd/charter`: `-i/--interactive`, `--color=auto|always|never`, `--no-color`, `--rule <ID>[,…]` on `doctor`; new `explain` command; `fang` wrapping.
- OSC 8 hyperlinks (rule id → `helpUri`, `path:line` → `file://`) gated by color/TTY.

### Out of scope
- the HTML report (`charter report`) — **Slice 16** (ADR-0025).
- always-on file-watch / live re-scan, `$EDITOR` spawn from the TUI (Phase 1.5).
- `--format toon|json-compact|for-agent`, `charter serve` (MCP), `report --format spdx` (Phase 1.5).
- any change to `json`/`sarif`/`markdown` renderers or the score formula.

## Grounding (verified 2026-06-02)
- **Versions** fetched from the module proxy (ADR-0006): `bubbletea/v2 v2.0.7`, `lipgloss/v2 v2.0.3`, `bubbles/v2 v2.1.0`, `fang v1.0.0`, `glamour v1.0.0` — all stable on `charm.land` paths.
- **Standards**: clig.dev (human-first; stdout/stderr; respect the terminal; fast start); the CLI Spec / WebCLI (structured-when-piped; **non-interactive by default**; never block without a TTY); no-color.org (`NO_COLOR` honored absolutely); terminfo.dev (color precedence + truecolor→256→16→mono; OSC 8 capability matrix — not Terminal.app).
- **`doctor.Run`** returns `Result{Findings, Suppressed, Summary, Score{Base,Final}, Categories, Root, Threshold, Passed}`; the styled renderer and TUI are **pure presentation** over this — no new scan logic.
- **Containment** (the load-bearing invariant): Charm/ANSI is reachable **only** on the interactive/TTY path. `--format {json,sarif,markdown}` and piped/`--quiet`/`NO_COLOR` output stay pure stdlib and byte-identical to today.

## Design language → Lipgloss palette

The design references define a semantic token system. Map each to a Lipgloss `AdaptiveColor` (light/dark), degrading per tier:

| Token (design) | Role | TUI usage |
|---|---|---|
| `--color-text-primary/secondary/tertiary` | foreground ramp | body / dim / faint |
| `--color-text-success / -danger / -warning / -info` | semantic fg | PASS·✓ / BLOCKER·✗ / HIGH·MEDIUM·⚠ / rule-id·accent |
| `--color-background-{success,danger,warning,info,secondary,tertiary}` | semantic bg | severity pills, callouts, panels |
| `--color-border-{…}` | borders | finding-card left border, panels, separators |

Reproduce, from the references: the `✦ charter` brand mark + dim version; `❯` prompt; braille spinner `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`; scan bars `██████████` (done/scanning) / `··········` (pending); per-category rows (`label · bar · ✓ · N rules · Nms`); severity badge pills; finding cards (severity left-border, Evidence/Fix/Auto-fix rows, OWASP tag); the score hero (large numeral in severity color + bar + PASS/FAIL pill + cap line); the per-screen key-hint footer. Glyphs (Tabler-style ✓/✗/⚠) **always** have an ASCII fallback for mono/no-Unicode terminals; severity is always a **text label**, never color alone.

**Fonts are terminal-inherited** — a TUI sets no typeface; we control only color, layout, and glyphs. The canonical font system (`docs/internal/designs/DESIGN-TOKENS.md`) governs the HTML report (Slice 16) + web/docs, not the terminal. Palette tiers and contrast (light/dark, AA) are grounded with the installed **`ui-ux-pro-max`** + **`design-taste-frontend`** + **`visual-design-foundations`** skills (color / accessibility / type-scale / metric-based discipline), and **`interaction-design`** for spinner / rescan / state feedback (full map in `DESIGN-TOKENS.md`).

## Command surface (v1.0, post-slice)
- `charter doctor` — styled when TTY, plain when piped/`NO_COLOR`/CI (unchanged JSON/SARIF/markdown).
- `charter doctor -i/--interactive` — TUI (errors exit 2 if not a TTY).
- `charter doctor --rule AE-XXX[,AE-YYY]` — scope the scan to named rules.
- `charter doctor --color auto|always|never` / `--no-color`.
- `charter explain <RULE>` — rule rationale/remediation/refs (`--format text|json`).

## Interactive TUI (`doctor -i`)
Master-detail over a single scan result (Bubble Tea MVU; Mode-2026 synchronized output):
- **Header**: `✦ charter` + repo/ref + score + the category scorecard.
- **List** (`bubbles/list` or `table`): findings groupable/sortable by severity & category; suppressed + informational in their own filters.
- **Detail pane** (`bubbles/viewport`): id, severity, category, summary, evidence (redacted), remediation, `path:line`, auto-fix hint, OWASP tag, and the `charter explain` rationale.
- **Interactions**: filter by severity/category; `/` search; vim+arrow nav; `tab` switch panes; `y` copy `path:line` (OSC 52); `r` rescan; `?` help; `q` quit.
- **Footer**: `bubbles/help` key bar.

## Architecture / ownership
- New: `internal/terminal`, `internal/render/text`, `internal/tui`, `internal/explain`. `cmd/charter` gains flags + the `explain` command + `fang` wrapping.
- Avoid: any Charm/ANSI on non-TTY paths; new scan logic in the renderer/TUI; nondeterministic ordering.

## Testing & verification
- **Contract test (load-bearing)**: assert `--format {json,sarif,markdown}`, piped stdout, `--quiet`, and `NO_COLOR=1` emit **zero ANSI** and identical bytes to the pre-slice baseline (golden), proving containment.
- **Styled renderer**: snapshot per color tier (truecolor/256/16/mono) via a forced-tier flag; ASCII-fallback snapshot.
- **`internal/terminal`**: precedence matrix (`--no-color` > `NO_COLOR` > `TERM=dumb` > `!isatty` > `COLORTERM`/`TERM`).
- **TUI**: model-level unit tests (Bubble Tea `Update` transitions: filter, search, select, rescan) using `teatest`/message injection — **no** golden of the live screen.
- **`explain`**: every catalog rule resolves; `--format json` shape; unknown rule → exit 2 with guidance.
- **Dogfood**: `charter doctor` (styled) still **100**; `moon run :check` green; `-race`/`gosec`/`golangci-lint` clean; ≥85% coverage on new pure packages (TUI excluded from the threshold, model logic covered).

## Success criteria
- Styled `doctor` + `doctor -i` TUI + `explain` + `--rule` + `--no-color` shipped, matching `docs/internal/designs/charter-doctor-init-fix.html` / `charter-supress-report-version.html`.
- Containment contract test green; non-TTY output byte-identical to baseline; dogfood 100; `moon run :check` green.
- ADR-0024 + this spec + plan committed; architecture §1.8 + command surface amended; HTML mirror regenerated (CF-2 gate); no stale references.

## References
- `docs/internal/decisions/0024-interactive-tui-and-terminal-output.md`; `0008` (score, unchanged); `0014` (catalog); `0006` (latest-docs-first)
- `docs/internal/designs/charter-doctor-init-fix.html`, `docs/internal/designs/charter-supress-report-version.html`, `docs/internal/designs/DESIGN-TOKENS.md`
- `docs/internal/superpowers/plans/2026-06-02-phase-1-slice-15.md`
