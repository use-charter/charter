package main

import (
	"os"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"go.use-charter.dev/charter/internal/terminal"
)

// resolveColorMode maps the --color and --no-color flags to a
// terminal.ColorMode. --no-color is equivalent to --color=never and WINS over
// --color, so a conflicting `--no-color --color=always` resolves to never (the
// conservative choice). Any other --color value that is not auto/always/never
// is an error.
func resolveColorMode(colorFlag string, noColor bool) (terminal.ColorMode, error) {
	if noColor {
		return terminal.ColorNever, nil
	}
	return terminal.ParseColorMode(colorFlag)
}

// terminalContext detects the color capabilities and palette for the writer the
// command will actually write to. out is the --out path ("" means stdout); when
// it is set, or the command's stdout is not an *os.File, the destination is
// treated as non-TTY (plain under auto). Environment signals (NO_COLOR,
// COLORTERM, TERM) are read here at the I/O boundary and combined with mode by
// terminal.Detect, so NO_COLOR is still honored under auto and overridden under
// --color=always exactly as terminal.Detect documents.
//
// Background polarity is queried only for a real, color-enabled TTY; a forced
// color mode on a pipe or file defaults to the light variant rather than reading
// stdin, keeping the path deterministic and free of stray terminal queries.
func terminalContext(cmd *cobra.Command, out string, mode terminal.ColorMode) (terminal.Capabilities, terminal.Palette) {
	var ttyFile *os.File
	isTTY := false
	if out == "" {
		if f, ok := cmd.OutOrStdout().(*os.File); ok {
			ttyFile = f
			if fi, statErr := f.Stat(); statErr == nil {
				isTTY = fi.Mode()&os.ModeCharDevice != 0
			}
		}
	}

	caps := terminal.Detect(terminal.Env{
		NoColor:   os.Getenv("NO_COLOR"),
		ColorTerm: os.Getenv("COLORTERM"),
		Term:      os.Getenv("TERM"),
	}, isTTY, mode)

	darkBackground := false
	if caps.ColorEnabled() && caps.IsTTY && ttyFile != nil {
		darkBackground = lipgloss.HasDarkBackground(os.Stdin, ttyFile)
	}
	return caps, terminal.NewPalette(caps, darkBackground)
}
