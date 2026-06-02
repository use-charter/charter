package main

import (
	"bytes"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"go.use-charter.dev/charter/internal/doctor"
	renderjson "go.use-charter.dev/charter/internal/render/json"
	rendermarkdown "go.use-charter.dev/charter/internal/render/markdown"
	rendersarif "go.use-charter.dev/charter/internal/render/sarif"
	rendertext "go.use-charter.dev/charter/internal/render/text"
	"go.use-charter.dev/charter/internal/terminal"
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

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Scan a repository and compute a Charter score",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch format {
			case "text", "json", "markdown", "sarif":
			default:
				return commandExitError{message: "invalid format: must be text, json, markdown, or sarif", exitCode: 2}
			}

			result, err := doctor.Run(path, threshold, cmd.Flags().Changed("threshold"))
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
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
			// caps.ColorEnabled(). Detect color from the ACTUAL write target so
			// styling stays consistent with where the bytes land: --out (a file)
			// and any non-*os.File writer are treated as non-TTY, i.e. plain.
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
			}, isTTY, terminal.ColorAuto)
			// Querying the terminal background only makes sense (and only avoids
			// stray stdin reads) when color is actually enabled.
			darkBackground := false
			if caps.ColorEnabled() {
				darkBackground = lipgloss.HasDarkBackground(os.Stdin, ttyFile)
			}
			rendered := rendertext.Render(result, caps, terminal.NewPalette(caps, darkBackground))

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

	return cmd
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
