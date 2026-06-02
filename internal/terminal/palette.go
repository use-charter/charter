package terminal

import (
	"image/color"
	"strconv"

	"charm.land/lipgloss/v2"
)

// Token is a semantic color token from the Charter design system. Names mirror
// the CSS custom properties in DESIGN-TOKENS.md
// (e.g. TextSuccess ↔ --color-text-success).
type Token uint8

const (
	TextPrimary Token = iota
	TextSecondary
	TextTertiary
	TextSuccess
	TextDanger
	TextWarning
	TextInfo

	BackgroundPrimary
	BackgroundSecondary
	BackgroundTertiary
	BackgroundSuccess
	BackgroundDanger
	BackgroundWarning
	BackgroundInfo

	BorderPrimary
	BorderSecondary
	BorderTertiary
	BorderSuccess
	BorderDanger
	BorderWarning
	BorderInfo
)

// String returns the design-token name (e.g. "text-success").
func (t Token) String() string {
	if int(t) >= len(tokenNames) {
		return "unknown"
	}
	return tokenNames[t]
}

var tokenNames = [...]string{
	TextPrimary:         "text-primary",
	TextSecondary:       "text-secondary",
	TextTertiary:        "text-tertiary",
	TextSuccess:         "text-success",
	TextDanger:          "text-danger",
	TextWarning:         "text-warning",
	TextInfo:            "text-info",
	BackgroundPrimary:   "background-primary",
	BackgroundSecondary: "background-secondary",
	BackgroundTertiary:  "background-tertiary",
	BackgroundSuccess:   "background-success",
	BackgroundDanger:    "background-danger",
	BackgroundWarning:   "background-warning",
	BackgroundInfo:      "background-info",
	BorderPrimary:       "border-primary",
	BorderSecondary:     "border-secondary",
	BorderTertiary:      "border-tertiary",
	BorderSuccess:       "border-success",
	BorderDanger:        "border-danger",
	BorderWarning:       "border-warning",
	BorderInfo:          "border-info",
}

// Style is a resolved token appearance: a color value plus non-color text
// attributes. Color is always non-nil; it is [lipgloss.NoColor] when the token
// should use the terminal's default color — that is, on the Mono tier, or for a
// neutral token in ANSI16 where no faithful color exists. The attributes let a
// renderer build hierarchy when color is unavailable.
type Style struct {
	Color   color.Color
	Bold    bool
	Faint   bool
	Reverse bool
}

// HasColor reports whether Color is an actual color rather than the terminal
// default ([lipgloss.NoColor]).
func (s Style) HasColor() bool {
	_, isNone := s.Color.(lipgloss.NoColor)
	return !isNone
}

// Palette resolves semantic [Token]s to concrete [Style]s for a fixed color
// [Tier] and background polarity. It holds no mutable state and is safe to copy
// and share across goroutines.
type Palette struct {
	tier Tier
	dark bool
}

// NewPalette binds a palette to the detected capabilities and the terminal
// background polarity. darkBackground selects the dark variant of each adaptive
// color (the lipgloss LightDark model); a renderer derives it from
// lipgloss.HasDarkBackground at the I/O boundary.
func NewPalette(caps Capabilities, darkBackground bool) Palette {
	return Palette{tier: caps.Tier, dark: darkBackground}
}

// Tier reports the color tier this palette resolves for.
func (p Palette) Tier() Tier { return p.tier }

// DarkBackground reports whether the dark adaptive variant is selected.
func (p Palette) DarkBackground() bool { return p.dark }

// Resolve returns the [Style] for a token at the palette's tier and polarity.
//
// Degradation per tier: TrueColor uses the 24-bit hex; ANSI256 uses the nearest
// xterm-256 index to that hex; ANSI16 uses the token's standard ANSI color, or
// falls back to the terminal default plus attributes for neutral tokens that
// have no ANSI-16 equivalent; Mono drops all color and relies on attributes.
func (p Palette) Resolve(tok Token) Style {
	if int(tok) >= len(tokens) {
		return Style{Color: lipgloss.NoColor{}}
	}
	d := tokens[tok]

	switch p.tier {
	case TrueColor:
		return Style{Color: p.adaptive(d.light, d.dark)}
	case ANSI256:
		return Style{Color: lipgloss.Color(strconv.Itoa(nearestANSI256(p.hex(d))))}
	case ANSI16:
		code := d.ansi16Light
		if p.dark {
			code = d.ansi16Dark
		}
		if code == "" {
			// Neutral token: ANSI-16 has no faithful gray here, so use the
			// terminal default foreground and lean on attributes for hierarchy.
			return d.mono.style()
		}
		return Style{Color: lipgloss.Color(code)}
	default: // Mono
		return d.mono.style()
	}
}

// adaptive picks the light or dark hex via the lipgloss LightDark model.
func (p Palette) adaptive(light, dark string) color.Color {
	choose := lipgloss.LightDark(p.dark)
	return choose(lipgloss.Color(light), lipgloss.Color(dark))
}

// hex returns the truecolor hex selected for the palette's polarity.
func (p Palette) hex(d tokenDef) string {
	if p.dark {
		return d.dark
	}
	return d.light
}

// attrs captures the non-color text attributes used to convey hierarchy when
// color is reduced or absent.
type attrs struct {
	bold    bool
	faint   bool
	reverse bool
}

func (a attrs) style() Style {
	return Style{
		Color:   lipgloss.NoColor{},
		Bold:    a.bold,
		Faint:   a.faint,
		Reverse: a.reverse,
	}
}

// tokenDef is the per-token color data across tiers.
//
//   - light/dark are the canonical 24-bit hexes (WCAG-AA on a light/dark
//     terminal background respectively); they also seed the nearest-ANSI256
//     downsample.
//   - ansi16Light/ansi16Dark are standard ANSI color indices ("" means the
//     token has no faithful ANSI-16 color and should use the terminal default).
//   - mono holds the attribute fallback used on Mono (and for neutral ANSI-16).
type tokenDef struct {
	light, dark             string
	ansi16Light, ansi16Dark string
	mono                    attrs
}

// Attribute fallbacks, named for intent. Never mutated after package init;
// treated as constants (Go has no const struct values).
var (
	attrNone    = attrs{}
	attrFaint   = attrs{faint: true}
	attrBold    = attrs{bold: true}
	attrReverse = attrs{reverse: true}
)

// tokens is the canonical Charter palette. The concrete hex values, ANSI-16
// fallbacks, and contrast rationale are documented in DESIGN-TOKENS.md.
var tokens = [...]tokenDef{
	// Text — neutral ramp uses the terminal default in ANSI-16 (no faithful
	// gray) and conveys hierarchy with faint; semantic text carries a hue plus
	// bold for Mono emphasis.
	TextPrimary:   {light: "#111827", dark: "#f9fafb", mono: attrNone},
	TextSecondary: {light: "#4b5563", dark: "#d1d5db", mono: attrFaint},
	TextTertiary:  {light: "#646b78", dark: "#9ca3af", mono: attrFaint},
	TextSuccess:   {light: "#15803d", dark: "#4ade80", ansi16Light: "2", ansi16Dark: "10", mono: attrBold},
	TextDanger:    {light: "#b91c1c", dark: "#f87171", ansi16Light: "1", ansi16Dark: "9", mono: attrBold},
	TextWarning:   {light: "#b45309", dark: "#fbbf24", ansi16Light: "3", ansi16Dark: "11", mono: attrBold},
	TextInfo:      {light: "#1d4ed8", dark: "#60a5fa", ansi16Light: "4", ansi16Dark: "12", mono: attrBold},

	// Background — neutral surfaces are subtle tints (no fill in ANSI-16/Mono);
	// semantic surfaces invert to reverse-video in Mono.
	BackgroundPrimary:   {light: "#ffffff", dark: "#0d1117", mono: attrNone},
	BackgroundSecondary: {light: "#f9fafb", dark: "#161b22", mono: attrNone},
	BackgroundTertiary:  {light: "#f3f4f6", dark: "#1f2937", mono: attrNone},
	BackgroundSuccess:   {light: "#f0fdf4", dark: "#052e16", ansi16Light: "2", ansi16Dark: "10", mono: attrReverse},
	BackgroundDanger:    {light: "#fef2f2", dark: "#450a0a", ansi16Light: "1", ansi16Dark: "9", mono: attrReverse},
	BackgroundWarning:   {light: "#fffbeb", dark: "#422006", ansi16Light: "3", ansi16Dark: "11", mono: attrReverse},
	BackgroundInfo:      {light: "#eff6ff", dark: "#172554", ansi16Light: "4", ansi16Dark: "12", mono: attrReverse},

	// Border — neutral dividers are intentionally subtle (faint for the
	// hairline tertiary divider); semantic borders use bold in Mono.
	BorderPrimary:   {light: "#d1d5db", dark: "#374151", mono: attrNone},
	BorderSecondary: {light: "#e5e7eb", dark: "#2b3240", mono: attrNone},
	BorderTertiary:  {light: "#eceef1", dark: "#21262d", mono: attrFaint},
	BorderSuccess:   {light: "#86efac", dark: "#166534", ansi16Light: "2", ansi16Dark: "10", mono: attrBold},
	BorderDanger:    {light: "#fca5a5", dark: "#991b1b", ansi16Light: "1", ansi16Dark: "9", mono: attrBold},
	BorderWarning:   {light: "#fcd34d", dark: "#92400e", ansi16Light: "3", ansi16Dark: "11", mono: attrBold},
	BorderInfo:      {light: "#93c5fd", dark: "#1e40af", ansi16Light: "4", ansi16Dark: "12", mono: attrBold},
}

// cubeLevels are the six channel intensities of the xterm-256 color cube.
var cubeLevels = [6]int{0x00, 0x5f, 0x87, 0xaf, 0xd7, 0xff}

// nearestANSI256 maps a 24-bit hex color to the closest xterm-256 palette index
// (always in 16-255), choosing between the 6×6×6 color cube and the 24-step
// grayscale ramp by squared-RGB distance.
func nearestANSI256(hex string) int {
	r, g, b := parseHexRGB(hex)

	ri, gi, bi := cubeIndex(r), cubeIndex(g), cubeIndex(b)
	cubeCode := 16 + 36*ri + 6*gi + bi
	cr, cg, cb := cubeLevels[ri], cubeLevels[gi], cubeLevels[bi]

	avg := (r + g + b) / 3
	grayIdx := 0
	switch {
	case avg > 238:
		// Conservative early clamp to the last ramp step (23). The unclamped
		// (avg-3)/10 first overflows 23 at avg=243, and its output is identical
		// across avg in [239,242], so clamping at >238 changes nothing.
		grayIdx = 23
	case avg > 3:
		grayIdx = (avg - 3) / 10
	}
	grayVal := 8 + 10*grayIdx
	grayCode := 232 + grayIdx

	if dist(r, g, b, grayVal, grayVal, grayVal) < dist(r, g, b, cr, cg, cb) {
		return grayCode
	}
	return cubeCode
}

// cubeIndex returns the 0-5 color-cube index nearest to channel value v (0-255).
func cubeIndex(v int) int {
	switch {
	case v < 48:
		return 0
	case v < 115:
		return 1
	default:
		if i := (v - 35) / 40; i < 5 {
			return i
		}
		return 5
	}
}

// dist is the squared Euclidean distance between two RGB triples.
func dist(r1, g1, b1, r2, g2, b2 int) int {
	dr, dg, db := r1-r2, g1-g2, b1-b2
	return dr*dr + dg*dg + db*db
}

// parseHexRGB extracts the 0-255 channels of a "#rrggbb" string. Inputs are the
// package's own validated constants, so malformed input degrades to black
// rather than erroring.
func parseHexRGB(hex string) (r, g, b int) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0
	}
	return nibble(hex[1])<<4 | nibble(hex[2]),
		nibble(hex[3])<<4 | nibble(hex[4]),
		nibble(hex[5])<<4 | nibble(hex[6])
}

// nibble converts a single hex digit to its 0-15 value (0 for non-hex bytes).
func nibble(b byte) int {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0')
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10
	default:
		return 0
	}
}
