package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.charter.dev/charter/internal/doctor"
	renderjson "go.charter.dev/charter/internal/render/json"
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

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Scan a repository and compute a Charter score",
		RunE: func(cmd *cobra.Command, args []string) error {
			if format != "text" && format != "json" {
				return commandExitError{message: "invalid format: must be text or json", exitCode: 2}
			}

			result, err := doctor.Run(path, threshold)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			if format == "json" {
				data, err := renderjson.Render(result)
				if err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}

				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				if !result.Passed {
					return commandExitError{message: "score below threshold", exitCode: 1, silent: true}
				}

				return nil
			}

			if quiet {
				if result.Score.Final < threshold {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "charter: score %d, threshold %d — FAIL\n", result.Score.Final, threshold)
					return commandExitError{message: "score below threshold", exitCode: 1, silent: true}
				}
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "charter doctor: %s\n", result.Root)
			for _, finding := range result.Findings {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s\n", finding.RuleID, finding.Severity, finding.Summary)
				for _, loc := range finding.Locations {
					if loc.Line > 0 {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  location: %s:%d\n", loc.Path, loc.Line)
					} else {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  location: %s\n", loc.Path)
					}
				}
				for _, evidence := range finding.Evidence {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", evidence)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  remediation: %s\n", finding.Remediation)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "score: %d (threshold %d)\n", result.Score.Final, threshold)

			if result.Score.Final < threshold {
				return commandExitError{message: "score below threshold", exitCode: 1, silent: true}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "explicit repository path")
	cmd.Flags().IntVar(&threshold, "threshold", 80, "minimum passing score")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "suppress non-failure output")
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")

	return cmd
}
