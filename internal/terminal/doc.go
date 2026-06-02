// Package terminal provides offline, deterministic terminal-capability
// detection and a semantic color palette for Charter's styled terminal output.
//
// It is the foundation of the styled presentation layer described in ADR-0024.
// The package returns capability and style VALUES only: it performs no I/O,
// holds no package-level mutable state, and never writes ANSI escape codes
// itself. Higher layers (the styled renderer and the TUI) consume these values
// and own all rendering.
//
// Two concerns are exposed:
//
//   - [Detect] classifies a terminal into a color [Tier]
//     (TrueColor → ANSI256 → ANSI16 → Mono) and an OSC 8 hyperlink capability,
//     from explicitly supplied environment values, a TTY flag, and a [ColorMode]
//     override. It is pure and safe for concurrent use.
//
//   - [Palette] resolves the Charter design-system [Token]s (DESIGN-TOKENS.md)
//     to concrete [Style] values, degrading the WCAG-AA light/dark palette per
//     tier and falling back to text attributes (bold/faint/reverse) when color
//     is unavailable.
//
// Colors are expressed with charm.land/lipgloss/v2 values; the light/dark
// adaptive selection follows lipgloss's LightDark model.
package terminal
