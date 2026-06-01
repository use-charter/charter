package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.use-charter.dev/charter/internal/doctor"
	renderjson "go.use-charter.dev/charter/internal/render/json"
	rendermarkdown "go.use-charter.dev/charter/internal/render/markdown"
	rendersarif "go.use-charter.dev/charter/internal/render/sarif"
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

			var buf bytes.Buffer
			fmt.Fprintf(&buf, "charter doctor: %s\n", result.Root)
			for _, finding := range result.Findings {
				fmt.Fprintf(&buf, "%s %s %s\n", finding.RuleID, finding.Severity, finding.Summary)
				for _, loc := range finding.Locations {
					if loc.Line > 0 {
						fmt.Fprintf(&buf, "  location: %s:%d\n", loc.Path, loc.Line)
					} else {
						fmt.Fprintf(&buf, "  location: %s\n", loc.Path)
					}
				}
				for _, evidence := range finding.Evidence {
					fmt.Fprintf(&buf, "  - %s\n", evidence)
				}
				fmt.Fprintf(&buf, "  remediation: %s\n", finding.Remediation)
			}
			for _, s := range result.Suppressed {
				fmt.Fprintf(&buf, "suppressed: %s (%s)", s.Finding.RuleID, s.Source)
				if s.Reason != "" {
					fmt.Fprintf(&buf, " — %s", s.Reason)
				}
				fmt.Fprintln(&buf)
			}
			fmt.Fprintf(&buf, "score: %d (threshold %d)\n", result.Score.Final, result.Threshold)

			if out != "" {
				if err := emit(cmd, out, buf.Bytes()); err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
			} else {
				_, _ = cmd.OutOrStdout().Write(buf.Bytes())
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
