package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/findings"
	renderjson "go.use-charter.dev/charter/internal/render/json"
	rendermarkdown "go.use-charter.dev/charter/internal/render/markdown"
	rendersarif "go.use-charter.dev/charter/internal/render/sarif"
	rendertext "go.use-charter.dev/charter/internal/render/text"
	"go.use-charter.dev/charter/internal/rules/catalog"
	"go.use-charter.dev/charter/internal/terminal"
	"go.use-charter.dev/charter/internal/tui"
)

type commandExitError struct {
	message  string
	exitCode int
	silent   bool
}

func (err commandExitError) Error() string {
	return err.message
}

func (err commandExitError) ExitCode() int {
	return err.exitCode
}

func (err commandExitError) Silent() bool {
	return err.silent
}

func newDoctorCommand() *cobra.Command {
	var path string
	var threshold int
	var quiet bool
	var format string
	var out string
	var colorFlag string
	var noColor bool
	var ruleFlag string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Scan a repository and compute a Charter score",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch format {
			case "text", "json", "markdown", "sarif":
			default:
				return commandExitError{message: "invalid format: must be text, json, markdown, or sarif", exitCode: 2}
			}

			// --quiet is a whole-repo CI gate over the 0–100 score; --rule is a
			// focused, score-free human view. They don't compose, so reject the
			// combination outright rather than silently honoring one.
			if ruleFlag != "" && quiet {
				return commandExitError{message: "--quiet cannot be combined with --rule", exitCode: 2}
			}

			mode, err := resolveColorMode(colorFlag, noColor)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			// --interactive is the opt-in TUI: it has its own gating (TTY-only,
			// text-format-only) and runs the scan itself, so it short-circuits
			// before any non-interactive rendering. When -i is unset the command
			// behaves exactly as before.
			if interactive {
				return runInteractive(cmd, path, threshold, cmd.Flags().Changed("threshold"), quiet, format, out, ruleFlag, mode)
			}

			result, err := doctor.Run(path, threshold, cmd.Flags().Changed("threshold"))
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			// --rule renders a filtered, score-free focused view (text only) and
			// exits 1 iff a named rule fired. Keep it ahead of the normal format
			// branches so the unfiltered paths stay byte-for-byte unchanged.
			// parseRuleIDs (inside runFocused) trims and rejects an all-blank
			// value, so a plain non-empty check is sufficient here.
			if ruleFlag != "" {
				return runFocused(cmd, result, ruleFlag, format, out, mode)
			}

			if format != "text" {
				var data []byte
				switch format {
				case "json":
					data, err = renderjson.Render(result)
				case "markdown":
					data, err = rendermarkdown.Render(result)
				case "sarif":
					data, err = rendersarif.Render(result)
				}
				if err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
				if err := emit(cmd, out, data); err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
				return passFail(result)
			}

			if quiet {
				if !result.Passed {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "charter: score %d, threshold %d — FAIL\n", result.Score.Final, result.Threshold)
					return commandExitError{message: "score below threshold", exitCode: 1, silent: true}
				}
				return nil
			}

			// Slice 15: route every --format text render through the single
			// canonical renderer, which branches plain vs styled internally on
			// caps.ColorEnabled(). terminalContext detects color from the ACTUAL
			// write target so styling stays consistent with where the bytes land:
			// --out (a file) and any non-*os.File writer are treated as non-TTY
			// (plain under auto). The resolved --color/--no-color mode flows in.
			caps, pal := terminalContext(cmd, out, mode)
			rendered := rendertext.Render(result, caps, pal)

			if out != "" {
				if err := emit(cmd, out, rendered); err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
			} else {
				_, _ = cmd.OutOrStdout().Write(rendered)
			}
			return passFail(result)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "explicit repository path")
	cmd.Flags().IntVar(&threshold, "threshold", 80, "minimum passing score")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "suppress non-failure output")
	cmd.Flags().StringVar(&format, "format", "text", "output format: text, json, markdown, or sarif")
	cmd.Flags().StringVar(&out, "out", "", "write output to a file instead of stdout")
	cmd.Flags().StringVar(&colorFlag, "color", "auto", "color output: auto, always, or never")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable color (equivalent to --color=never; wins over --color)")
	cmd.Flags().StringVar(&ruleFlag, "rule", "", "comma-separated rule IDs to focus on (filtered text view; omits the score)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "browse findings in an interactive TUI (requires a terminal)")

	return cmd
}

// runInteractive launches the `charter doctor -i` browser. It is the only path
// to the TUI, and it is deliberately narrow: the full-screen, human-only view
// is incompatible with every machine-readable or headless mode (all of which
// carry the score by contract), so those combinations are rejected with exit 2.
// The TUI must NEVER run on a non-TTY (piped, redirected, CI, or --out) — there
// it would emit control sequences into a file or hang waiting on input — so the
// command gates on a real terminal first and fails fast. None of this touches
// the non-interactive output contract.
func runInteractive(cmd *cobra.Command, path string, threshold int, thresholdSet, quiet bool, format, out, ruleFlag string, mode terminal.ColorMode) error {
	switch {
	case format != "text":
		return commandExitError{message: "--interactive cannot be combined with --format " + format, exitCode: 2}
	case quiet:
		return commandExitError{message: "--interactive cannot be combined with --quiet", exitCode: 2}
	case out != "":
		return commandExitError{message: "--interactive cannot be combined with --out", exitCode: 2}
	case ruleFlag != "":
		return commandExitError{message: "--interactive cannot be combined with --rule", exitCode: 2}
	}

	ttyFile, ok := stdoutTTY(cmd)
	if !ok {
		return commandExitError{message: "--interactive requires a terminal", exitCode: 2}
	}

	result, err := doctor.Run(path, threshold, thresholdSet)
	if err != nil {
		return commandExitError{message: err.Error(), exitCode: 2}
	}

	// Detect color/background from the real TTY (out is "" here: the TUI always
	// renders to stdout). scan re-runs the identical scan for the `r` rescan key.
	caps, pal := terminalContext(cmd, "", mode)
	scan := func() (doctor.Result, error) { return doctor.Run(path, threshold, thresholdSet) }
	if err := tui.Run(result, scan, caps, pal, ttyFile); err != nil {
		return commandExitError{message: err.Error(), exitCode: 2}
	}
	return nil
}

// stdoutTTY returns the command's stdout as an *os.File when it is a character
// device (a real terminal), reporting false for a pipe, file, or any non-file
// writer. It is the single TTY gate for the interactive browser.
func stdoutTTY(cmd *cobra.Command) (*os.File, bool) {
	f, ok := cmd.OutOrStdout().(*os.File)
	if !ok {
		return nil, false
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, false
	}
	if fi.Mode()&os.ModeCharDevice == 0 {
		return nil, false
	}
	return f, true
}

// runFocused renders the filtered `--rule` view: the full scan result projected
// down to the named rules only, with no score or scorecard (a partial view must
// not imply a 0–100 verdict). It is text-only; combining --rule with a
// machine-readable format is rejected since those carry the score by contract.
// Exit code: 1 iff any named rule produced a finding, else 0.
func runFocused(cmd *cobra.Command, result doctor.Result, ruleFlag, format, out string, mode terminal.ColorMode) error {
	if format != "text" {
		return commandExitError{message: "--rule is only supported with --format text", exitCode: 2}
	}
	ids, err := parseRuleIDs(ruleFlag)
	if err != nil {
		return commandExitError{message: err.Error(), exitCode: 2}
	}

	caps, pal := terminalContext(cmd, out, mode)
	rendered := rendertext.RenderFocused(result, ids, caps, pal)

	if out != "" {
		if err := emit(cmd, out, rendered); err != nil {
			return commandExitError{message: err.Error(), exitCode: 2}
		}
	} else {
		_, _ = cmd.OutOrStdout().Write(rendered)
	}

	if focusedFired(result.Findings, ids) {
		return commandExitError{message: "findings present for filtered rules", exitCode: 1, silent: true}
	}
	return nil
}

// parseRuleIDs splits a comma-separated --rule value into validated, de-duped
// rule IDs (order preserved). Each ID must exist in the catalog; an unknown ID
// is an error that lists the valid IDs. An empty list (all blank) is rejected.
func parseRuleIDs(flag string) ([]string, error) {
	var ids []string
	seen := make(map[string]bool)
	for _, part := range strings.Split(flag, ",") {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		if _, ok := catalog.Lookup(id); !ok {
			return nil, fmt.Errorf("unknown rule %q; valid rule IDs: %s", id, strings.Join(catalog.IDs(), ", "))
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("--rule requires at least one rule ID (e.g. --rule AE-SEC-001)")
	}
	return ids, nil
}

// focusedFired reports whether any finding's RuleID is in ids. It is kept
// separate from internal/render/text.filterFindings on purpose: that helper
// lives in the render package and projects findings to a slice for display,
// whereas this only needs a boolean any-match to choose the exit code. Sharing
// would couple the CLI to a render-package export for a tiny membership test,
// so the two-line rule-set map is duplicated rather than abstracted.
func focusedFired(all []findings.Finding, ids []string) bool {
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	for _, f := range all {
		if want[f.RuleID] {
			return true
		}
	}
	return false
}

// emit writes data to outPath (with a one-line stderr summary) or to stdout.
func emit(cmd *cobra.Command, outPath string, data []byte) error {
	if outPath == "" {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return err
	}
	if !bytes.HasSuffix(data, []byte{'\n'}) {
		data = append(data, '\n')
	}
	// #nosec G306 -- report output is non-sensitive and meant to be shareable.
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "charter: wrote %s\n", outPath)
	return nil
}

func passFail(result doctor.Result) error {
	if !result.Passed {
		return commandExitError{message: "score below threshold", exitCode: 1, silent: true}
	}
	return nil
}
