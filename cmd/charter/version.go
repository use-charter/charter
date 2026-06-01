package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"go.use-charter.dev/charter/internal/version"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Charter version and build metadata",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "charter   %s\n", version.Version())
			_, _ = fmt.Fprintf(out, "commit    %s\n", version.Commit())
			_, _ = fmt.Fprintf(out, "built     %s\n", version.Date())
			_, _ = fmt.Fprintf(out, "go        %s\n", strings.TrimPrefix(runtime.Version(), "go"))
			_, _ = fmt.Fprintf(out, "platform  %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
}
