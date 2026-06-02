package terminal

import "testing"

func TestDetectTierMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		env   Env
		isTTY bool
		mode  ColorMode
		want  Tier
	}{
		// Explicit overrides win, highest precedence.
		{"never forces mono even with truecolor", Env{ColorTerm: "truecolor", Term: "xterm-256color"}, true, ColorNever, Mono},
		{"never forces mono when not a tty", Env{}, false, ColorNever, Mono},
		{"always truecolor from colorterm", Env{ColorTerm: "truecolor"}, false, ColorAlways, TrueColor},
		{"always 256 from term", Env{Term: "xterm-256color"}, false, ColorAlways, ANSI256},
		{"always 16 from term", Env{Term: "xterm"}, false, ColorAlways, ANSI16},
		{"always falls back to truecolor with no signal", Env{}, false, ColorAlways, TrueColor},
		{"always overrides NO_COLOR", Env{NoColor: "1", ColorTerm: "truecolor"}, true, ColorAlways, TrueColor},
		{"always overrides dumb", Env{Term: "dumb"}, true, ColorAlways, TrueColor},

		// Auto gating: each gate forces mono.
		{"auto NO_COLOR forces mono", Env{NoColor: "1", ColorTerm: "truecolor", Term: "xterm-256color"}, true, ColorAuto, Mono},
		{"auto NO_COLOR any value forces mono", Env{NoColor: "0", Term: "xterm-256color"}, true, ColorAuto, Mono},
		{"auto dumb forces mono", Env{Term: "dumb", ColorTerm: "truecolor"}, true, ColorAuto, Mono},
		{"auto non-tty forces mono", Env{Term: "xterm-256color", ColorTerm: "truecolor"}, false, ColorAuto, Mono},

		// Auto capability detection (gates passed).
		{"auto colorterm truecolor", Env{ColorTerm: "truecolor", Term: "xterm"}, true, ColorAuto, TrueColor},
		{"auto colorterm 24bit", Env{ColorTerm: "24bit", Term: "xterm"}, true, ColorAuto, TrueColor},
		{"auto term 256color", Env{Term: "xterm-256color"}, true, ColorAuto, ANSI256},
		{"auto term screen-256color", Env{Term: "screen-256color"}, true, ColorAuto, ANSI256},
		{"auto term xterm", Env{Term: "xterm"}, true, ColorAuto, ANSI16},
		{"auto term screen", Env{Term: "screen"}, true, ColorAuto, ANSI16},
		{"auto term tmux", Env{Term: "tmux"}, true, ColorAuto, ANSI16},
		{"auto term vt100", Env{Term: "vt100"}, true, ColorAuto, ANSI16},
		{"auto term rxvt", Env{Term: "rxvt-unicode"}, true, ColorAuto, ANSI16},
		{"auto unknown term is mono", Env{Term: "foobar"}, true, ColorAuto, Mono},
		{"auto empty env is mono", Env{}, true, ColorAuto, Mono},

		// Case-insensitivity and precedence within capability detection.
		{"colorterm is case-insensitive", Env{ColorTerm: "TrueColor"}, true, ColorAuto, TrueColor},
		{"term is case-insensitive", Env{Term: "XTERM-256COLOR"}, true, ColorAuto, ANSI256},
		{"colorterm beats term tier", Env{ColorTerm: "truecolor", Term: "xterm-256color"}, true, ColorAuto, TrueColor},
		{"256color substring beats 16-color match", Env{Term: "tmux-256color"}, true, ColorAuto, ANSI256},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Detect(tc.env, tc.isTTY, tc.mode).Tier
			if got != tc.want {
				t.Fatalf("Detect(%+v, tty=%v, %v).Tier = %v, want %v",
					tc.env, tc.isTTY, tc.mode, got, tc.want)
			}
		})
	}
}

func TestDetectHyperlinks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		env   Env
		isTTY bool
		mode  ColorMode
		want  bool
	}{
		{"auto truecolor tty", Env{ColorTerm: "truecolor"}, true, ColorAuto, true},
		{"auto 256 tty", Env{Term: "xterm-256color"}, true, ColorAuto, true},
		{"auto ansi16 tty", Env{Term: "xterm"}, true, ColorAuto, true},
		{"auto non-tty has no links", Env{Term: "xterm-256color"}, false, ColorAuto, false},
		{"auto NO_COLOR has no links", Env{NoColor: "1", Term: "xterm-256color"}, true, ColorAuto, false},
		{"auto dumb has no links", Env{Term: "dumb"}, true, ColorAuto, false},
		{"never has no links", Env{ColorTerm: "truecolor"}, true, ColorNever, false},
		{"always tty has links", Env{ColorTerm: "truecolor"}, true, ColorAlways, true},
		{"always non-tty has no links", Env{ColorTerm: "truecolor"}, false, ColorAlways, false},
		{"always dumb has no links despite color", Env{Term: "dumb"}, true, ColorAlways, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Detect(tc.env, tc.isTTY, tc.mode).Hyperlinks
			if got != tc.want {
				t.Fatalf("Detect(%+v, tty=%v, %v).Hyperlinks = %v, want %v",
					tc.env, tc.isTTY, tc.mode, got, tc.want)
			}
		})
	}
}

func TestCapabilitiesColorEnabled(t *testing.T) {
	t.Parallel()

	if got := Detect(Env{}, true, ColorNever); got.ColorEnabled() {
		t.Errorf("ColorEnabled() = true for mono, want false")
	}
	if got := Detect(Env{ColorTerm: "truecolor"}, true, ColorAuto); !got.ColorEnabled() {
		t.Errorf("ColorEnabled() = false for truecolor, want true")
	}
}

func TestDetectRecordsModeAndTTY(t *testing.T) {
	t.Parallel()

	caps := Detect(Env{ColorTerm: "truecolor"}, true, ColorAlways)
	if caps.Mode != ColorAlways {
		t.Errorf("Mode = %v, want always", caps.Mode)
	}
	if !caps.IsTTY {
		t.Errorf("IsTTY = false, want true")
	}
}

func TestParseColorMode(t *testing.T) {
	t.Parallel()

	ok := []struct {
		in   string
		want ColorMode
	}{
		{"auto", ColorAuto},
		{"always", ColorAlways},
		{"never", ColorNever},
		{"AUTO", ColorAuto},
		{" Always ", ColorAlways},
	}
	for _, tc := range ok {
		got, err := ParseColorMode(tc.in)
		if err != nil {
			t.Errorf("ParseColorMode(%q) unexpected error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("ParseColorMode(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}

	for _, bad := range []string{"", "on", "yes", "256"} {
		if _, err := ParseColorMode(bad); err == nil {
			t.Errorf("ParseColorMode(%q) = nil error, want error", bad)
		}
	}
}

func TestTierString(t *testing.T) {
	t.Parallel()

	cases := map[Tier]string{
		Mono:      "mono",
		ANSI16:    "ansi16",
		ANSI256:   "ansi256",
		TrueColor: "truecolor",
		Tier(99):  "unknown",
	}
	for tier, want := range cases {
		if got := tier.String(); got != want {
			t.Errorf("Tier(%d).String() = %q, want %q", tier, got, want)
		}
	}
}

func TestColorModeString(t *testing.T) {
	t.Parallel()

	cases := map[ColorMode]string{
		ColorAuto:     "auto",
		ColorAlways:   "always",
		ColorNever:    "never",
		ColorMode(99): "unknown",
	}
	for mode, want := range cases {
		if got := mode.String(); got != want {
			t.Errorf("ColorMode(%d).String() = %q, want %q", mode, got, want)
		}
	}
}
