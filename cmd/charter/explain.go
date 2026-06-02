package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.use-charter.dev/charter/internal/explain"
	"go.use-charter.dev/charter/internal/rules/catalog"
)

func newExplainCommand() *cobra.Command {
	var format string
	var colorFlag string
	var noColor bool

	cmd := &cobra.Command{
		Use:   "explain <RULE>",
		Short: "Explain a Charter rule: ID, name, category, summary, and docs link",
		// Arg count is validated inside RunE so an arg-count problem surfaces as
		// a commandExitError{exitCode: 2} (a usage error), consistent with the
		// command's other failures.
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return commandExitError{
					message:  "explain takes exactly one rule ID argument (e.g. charter explain AE-SEC-001)",
					exitCode: 2,
				}
			}
			switch format {
			case "text", "json":
			default:
				return commandExitError{message: "invalid format: must be text or json", exitCode: 2}
			}

			// Resolve --color/--no-color before the format branch so an invalid
			// --color is rejected (exit 2) regardless of --format, matching
			// doctor. mode is only consumed by the text path below.
			mode, err := resolveColorMode(colorFlag, noColor)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			id := args[0]
			entry, ok := catalog.Lookup(id)
			if !ok {
				return commandExitError{
					message:  fmt.Sprintf("unknown rule %q; valid rule IDs: %s", id, strings.Join(catalog.IDs(), ", ")),
					exitCode: 2,
				}
			}

			if format == "json" {
				data, err := explain.JSON(entry)
				if err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			caps, pal := terminalContext(cmd, "", mode)
			_, _ = cmd.OutOrStdout().Write(explain.Text(entry, caps, pal))
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	cmd.Flags().StringVar(&colorFlag, "color", "auto", "color output: auto, always, or never")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable color (equivalent to --color=never; wins over --color)")
	return cmd
}
