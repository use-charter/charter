package terminal

import (
	"fmt"
	"strings"
)

// Tier is the resolved terminal color capability, richest to poorest. The
// palette degrades along this scale (ADR-0024): TrueColor → ANSI256 → ANSI16 →
// Mono.
type Tier uint8

const (
	// Mono is a 1-bit terminal: no color. Hierarchy is built from text
	// attributes (bold/faint/reverse) alone.
	Mono Tier = iota
	// ANSI16 is a 4-bit terminal: the 16 standard ANSI colors.
	ANSI16
	// ANSI256 is an 8-bit terminal: the 256-color xterm palette.
	ANSI256
	// TrueColor is a 24-bit terminal: full RGB.
	TrueColor
)

// String returns the lowercase tier name (e.g. "truecolor").
func (t Tier) String() string {
	switch t {
	case Mono:
		return "mono"
	case ANSI16:
		return "ansi16"
	case ANSI256:
		return "ansi256"
	case TrueColor:
		return "truecolor"
	default:
		return "unknown"
	}
}

// ColorMode is the explicit `--color` override that gates detection.
type ColorMode uint8

const (
	// ColorAuto detects color from the environment and TTY state.
	ColorAuto ColorMode = iota
	// ColorAlways forces color on, even when piped or NO_COLOR is set.
	ColorAlways
	// ColorNever forces color off.
	ColorNever
)

// String returns the lowercase mode name (e.g. "always").
func (m ColorMode) String() string {
	switch m {
	case ColorAuto:
		return "auto"
	case ColorAlways:
		return "always"
	case ColorNever:
		return "never"
	default:
		return "unknown"
	}
}

// ParseColorMode maps a `--color` flag value to a [ColorMode]. It accepts
// "auto", "always", and "never" (case-insensitive) and returns an error for
// anything else.
func ParseColorMode(s string) (ColorMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "auto":
		return ColorAuto, nil
	case "always":
		return ColorAlways, nil
	case "never":
		return ColorNever, nil
	default:
		return ColorAuto, fmt.Errorf("invalid color mode %q: want auto, always, or never", s)
	}
}

// Env is the subset of environment variables that influence color detection.
// Passing values explicitly keeps [Detect] pure and testable; the caller reads
// these from the process environment at the I/O boundary.
type Env struct {
	// NoColor is $NO_COLOR. Any non-empty value disables color (no-color.org).
	NoColor string
	// ColorTerm is $COLORTERM. "truecolor" or "24bit" implies 24-bit color.
	ColorTerm string
	// Term is $TERM, the terminal type string.
	Term string
}

// Capabilities is the immutable result of capability detection.
type Capabilities struct {
	// Tier is the resolved color tier.
	Tier Tier
	// Mode is the override that was applied.
	Mode ColorMode
	// IsTTY reports whether stdout was a terminal.
	IsTTY bool
	// Hyperlinks reports whether OSC 8 hyperlinks should be emitted.
	Hyperlinks bool
}

// ColorEnabled reports whether any color should be emitted.
func (c Capabilities) ColorEnabled() bool { return c.Tier != Mono }

// Detect resolves terminal [Capabilities] from explicit inputs. It is pure: it
// reads no globals, performs no I/O, and is safe for concurrent use.
//
// Color precedence, highest first (ADR-0024): an explicit ColorNever forces
// Mono and ColorAlways forces color (at least TrueColor); then a non-empty
// NO_COLOR, TERM=dumb, or a non-TTY stdout each force Mono; otherwise the tier
// is read from COLORTERM and TERM. OSC 8 hyperlinks are reported only when
// color is enabled, stdout is a TTY, and TERM is not "dumb" — deliberately
// conservative, since hyperlinks in a pipe or file are noise.
func Detect(env Env, isTTY bool, mode ColorMode) Capabilities {
	tier := resolveTier(env, isTTY, mode)
	return Capabilities{
		Tier:       tier,
		Mode:       mode,
		IsTTY:      isTTY,
		Hyperlinks: tier != Mono && isTTY && !isDumb(env.Term),
	}
}

// resolveTier applies the color precedence rules to produce a [Tier].
func resolveTier(env Env, isTTY bool, mode ColorMode) Tier {
	switch mode {
	case ColorNever:
		return Mono
	case ColorAlways:
		if t := rawTier(env); t != Mono {
			return t
		}
		// Forced on but no capability signal: assume the richest tier.
		return TrueColor
	}

	// ColorAuto: honor the gating signals before reading capability.
	if env.NoColor != "" {
		return Mono
	}
	if isDumb(env.Term) {
		return Mono
	}
	if !isTTY {
		return Mono
	}
	return rawTier(env)
}

// rawTier classifies the color capability implied purely by COLORTERM and TERM,
// ignoring the gating signals (TTY, NO_COLOR, dumb).
func rawTier(env Env) Tier {
	switch strings.ToLower(env.ColorTerm) {
	case "truecolor", "24bit":
		return TrueColor
	}

	term := strings.ToLower(env.Term)
	if strings.Contains(term, "256color") {
		return ANSI256
	}
	for _, p := range []string{"xterm", "screen", "tmux", "vt100", "rxvt"} {
		if strings.Contains(term, p) {
			return ANSI16
		}
	}
	return Mono
}

// isDumb reports whether TERM names a terminal with no capabilities.
func isDumb(term string) bool { return strings.EqualFold(term, "dumb") }
