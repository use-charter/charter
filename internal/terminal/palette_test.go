package terminal

import (
	"image/color"
	"strconv"
	"testing"

	"charm.land/lipgloss/v2"
)

var allTokens = []Token{
	TextPrimary, TextSecondary, TextTertiary,
	TextSuccess, TextDanger, TextWarning, TextInfo,
	BackgroundPrimary, BackgroundSecondary, BackgroundTertiary,
	BackgroundSuccess, BackgroundDanger, BackgroundWarning, BackgroundInfo,
	BorderPrimary, BorderSecondary, BorderTertiary,
	BorderSuccess, BorderDanger, BorderWarning, BorderInfo,
}

func paletteFor(tier Tier, dark bool) Palette {
	return NewPalette(Capabilities{Tier: tier}, dark)
}

func TestResolveSemanticTokenPerTier(t *testing.T) {
	t.Parallel()

	// TextSuccess exercises every tier branch with known values.
	tests := []struct {
		tier      Tier
		dark      bool
		wantColor color.Color
		bold      bool
		hasColor  bool
	}{
		{TrueColor, false, lipgloss.Color("#15803d"), false, true},
		{TrueColor, true, lipgloss.Color("#4ade80"), false, true},
		{ANSI256, false, lipgloss.Color("29"), false, true},
		{ANSI256, true, lipgloss.Color("78"), false, true},
		{ANSI16, false, lipgloss.Color("2"), false, true},
		{ANSI16, true, lipgloss.Color("10"), false, true},
		{Mono, false, lipgloss.NoColor{}, true, false},
		{Mono, true, lipgloss.NoColor{}, true, false},
	}

	for _, tc := range tests {
		got := paletteFor(tc.tier, tc.dark).Resolve(TextSuccess)
		if got.Color != tc.wantColor {
			t.Errorf("TextSuccess tier=%v dark=%v: Color = %v, want %v", tc.tier, tc.dark, got.Color, tc.wantColor)
		}
		if got.Bold != tc.bold {
			t.Errorf("TextSuccess tier=%v dark=%v: Bold = %v, want %v", tc.tier, tc.dark, got.Bold, tc.bold)
		}
		if got.HasColor() != tc.hasColor {
			t.Errorf("TextSuccess tier=%v dark=%v: HasColor = %v, want %v", tc.tier, tc.dark, got.HasColor(), tc.hasColor)
		}
	}
}

func TestMonoYieldsNoColorWithAttributes(t *testing.T) {
	t.Parallel()

	for _, dark := range []bool{false, true} {
		pal := paletteFor(Mono, dark)
		for _, tok := range allTokens {
			got := pal.Resolve(tok)
			if got.HasColor() {
				t.Errorf("Mono %v: HasColor = true, want false (color must drop in mono)", tok)
			}
			if _, ok := got.Color.(lipgloss.NoColor); !ok {
				t.Errorf("Mono %v: Color = %T, want lipgloss.NoColor", tok, got.Color)
			}
		}
	}
}

func TestMonoAttributeHierarchy(t *testing.T) {
	t.Parallel()

	pal := paletteFor(Mono, true)
	cases := []struct {
		tok     Token
		bold    bool
		faint   bool
		reverse bool
	}{
		{TextPrimary, false, false, false},
		{TextSecondary, false, true, false},
		{TextTertiary, false, true, false},
		{TextSuccess, true, false, false},
		{TextDanger, true, false, false},
		{BackgroundPrimary, false, false, false},
		{BackgroundSuccess, false, false, true},
		{BorderTertiary, false, true, false},
		{BorderDanger, true, false, false},
	}
	for _, tc := range cases {
		got := pal.Resolve(tc.tok)
		if got.Bold != tc.bold || got.Faint != tc.faint || got.Reverse != tc.reverse {
			t.Errorf("Mono %v: attrs bold=%v faint=%v reverse=%v, want bold=%v faint=%v reverse=%v",
				tc.tok, got.Bold, got.Faint, got.Reverse, tc.bold, tc.faint, tc.reverse)
		}
	}
}

func TestANSI16NeutralFallback(t *testing.T) {
	t.Parallel()

	// Neutral text has no faithful ANSI-16 color: default fg + attribute hierarchy.
	neutral := paletteFor(ANSI16, true)
	if got := neutral.Resolve(TextPrimary); got.HasColor() || got.Faint {
		t.Errorf("ANSI16 text-primary: got HasColor=%v Faint=%v, want default fg with no faint", got.HasColor(), got.Faint)
	}
	if got := neutral.Resolve(TextSecondary); got.HasColor() || !got.Faint {
		t.Errorf("ANSI16 text-secondary: got HasColor=%v Faint=%v, want default fg + faint", got.HasColor(), got.Faint)
	}
	if got := neutral.Resolve(TextTertiary); got.HasColor() || !got.Faint {
		t.Errorf("ANSI16 text-tertiary: got HasColor=%v Faint=%v, want default fg + faint", got.HasColor(), got.Faint)
	}

	// Semantic tokens keep a real ANSI-16 hue.
	if got := neutral.Resolve(TextSuccess); !got.HasColor() || got.Color != lipgloss.Color("10") {
		t.Errorf("ANSI16 text-success: Color = %v (HasColor=%v), want bright green (10)", got.Color, got.HasColor())
	}
	light := paletteFor(ANSI16, false)
	if got := light.Resolve(TextDanger); got.Color != lipgloss.Color("1") {
		t.Errorf("ANSI16 light text-danger: Color = %v, want red (1)", got.Color)
	}
}

func TestResolveTrueColorPolarity(t *testing.T) {
	t.Parallel()

	light := paletteFor(TrueColor, false).Resolve(TextPrimary)
	dark := paletteFor(TrueColor, true).Resolve(TextPrimary)
	if light.Color != lipgloss.Color("#111827") {
		t.Errorf("light text-primary = %v, want #111827", light.Color)
	}
	if dark.Color != lipgloss.Color("#f9fafb") {
		t.Errorf("dark text-primary = %v, want #f9fafb", dark.Color)
	}
}

func TestResolveEveryTokenEveryTier(t *testing.T) {
	t.Parallel()

	for _, tier := range []Tier{Mono, ANSI16, ANSI256, TrueColor} {
		for _, dark := range []bool{false, true} {
			pal := paletteFor(tier, dark)
			for _, tok := range allTokens {
				got := pal.Resolve(tok)
				if got.Color == nil {
					t.Fatalf("Resolve(%v) tier=%v dark=%v returned nil Color", tok, tier, dark)
				}
				switch tier {
				case TrueColor, ANSI256:
					if !got.HasColor() {
						t.Errorf("Resolve(%v) tier=%v: HasColor=false, want true", tok, tier)
					}
				case Mono:
					if got.HasColor() {
						t.Errorf("Resolve(%v) tier=%v: HasColor=true, want false", tok, tier)
					}
				}
			}
		}
	}
}

func TestResolveOutOfRangeToken(t *testing.T) {
	t.Parallel()

	got := paletteFor(TrueColor, true).Resolve(Token(250))
	if got.HasColor() {
		t.Errorf("out-of-range token: HasColor=true, want false (safe default)")
	}
}

func TestPaletteAccessors(t *testing.T) {
	t.Parallel()

	pal := NewPalette(Capabilities{Tier: ANSI256}, true)
	if pal.Tier() != ANSI256 {
		t.Errorf("Tier() = %v, want ansi256", pal.Tier())
	}
	if !pal.DarkBackground() {
		t.Errorf("DarkBackground() = false, want true")
	}
}

func TestNearestANSI256(t *testing.T) {
	t.Parallel()

	cases := map[string]int{
		"#000000": 16,
		"#ffffff": 231,
		"#ff0000": 196,
		"#00ff00": 46,
		"#0000ff": 21,
		"#808080": 244,
		"#4ade80": 78,
		"#15803d": 29,
		"#b91c1c": 124,
		"#1d4ed8": 26,
	}
	for hex, want := range cases {
		if got := nearestANSI256(hex); got != want {
			t.Errorf("nearestANSI256(%s) = %d, want %d", hex, got, want)
		}
		if want < 16 || want > 255 {
			t.Errorf("nearestANSI256(%s) expectation %d out of 16-255 range", hex, want)
		}
	}
}

func TestNearestANSI256MalformedHex(t *testing.T) {
	t.Parallel()

	// Wrong length fails the parse guard and degrades to black (index 16).
	if got := nearestANSI256("nope"); got != 16 {
		t.Errorf("nearestANSI256(%q) = %d, want 16", "nope", got)
	}

	// A correctly-shaped string with a non-hex byte ('!') exercises nibble's
	// default branch (treated as 0): "#abc!ef" -> rgb(171,192,239) -> 147.
	if got := nearestANSI256("#abc!ef"); got != 147 {
		t.Errorf("nearestANSI256(%q) = %d, want 147", "#abc!ef", got)
	}
}

func TestANSI256ResolvesToIndexedColor(t *testing.T) {
	t.Parallel()

	got := paletteFor(ANSI256, true).Resolve(TextInfo)
	want := lipgloss.Color(strconv.Itoa(nearestANSI256("#60a5fa")))
	if got.Color != want {
		t.Errorf("ANSI256 text-info = %v, want %v", got.Color, want)
	}
}

func TestTokenString(t *testing.T) {
	t.Parallel()

	cases := map[Token]string{
		TextPrimary:       "text-primary",
		TextInfo:          "text-info",
		BackgroundSuccess: "background-success",
		BorderInfo:        "border-info",
		Token(200):        "unknown",
	}
	for tok, want := range cases {
		if got := tok.String(); got != want {
			t.Errorf("Token(%d).String() = %q, want %q", tok, got, want)
		}
	}
}

func TestStyleHasColor(t *testing.T) {
	t.Parallel()

	if (Style{Color: lipgloss.NoColor{}}).HasColor() {
		t.Errorf("NoColor style HasColor = true, want false")
	}
	if !(Style{Color: lipgloss.Color("#ffffff")}).HasColor() {
		t.Errorf("colored style HasColor = false, want true")
	}
}
