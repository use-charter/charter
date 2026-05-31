package main

import "github.com/spf13/cobra"

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "charter",
		Short:         "Scan repositories for AI-agent readiness",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newDoctorCommand())
	cmd.AddCommand(newSuppressCommand())
	return cmd
}
