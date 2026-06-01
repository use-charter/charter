package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"go.use-charter.dev/charter/internal/version"
)

func newVersionCommand() *cobra.Command {
	var (
		format string
		short  bool
	)
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print Charter version and build metadata",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if short {
				_, _ = fmt.Fprintln(out, version.Version())
				return nil
			}
			goVersion := strings.TrimPrefix(runtime.Version(), "go")
			platform := runtime.GOOS + "/" + runtime.GOARCH
			switch format {
			case "", "text":
				_, _ = fmt.Fprintf(out, "charter   %s\n", version.Version())
				_, _ = fmt.Fprintf(out, "commit    %s\n", version.Commit())
				_, _ = fmt.Fprintf(out, "built     %s\n", version.Date())
				_, _ = fmt.Fprintf(out, "go        %s\n", goVersion)
				_, _ = fmt.Fprintf(out, "platform  %s\n", platform)
			case "json":
				b, err := json.MarshalIndent(struct {
					Version  string `json:"version"`
					Commit   string `json:"commit"`
					Date     string `json:"date"`
					Go       string `json:"go"`
					Platform string `json:"platform"`
				}{version.Version(), version.Commit(), version.Date(), goVersion, platform}, "", "  ")
				if err != nil {
					return fmt.Errorf("render version json: %w", err)
				}
				_, _ = fmt.Fprintln(out, string(b))
			default:
				return fmt.Errorf("unknown --format %q (want text or json)", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	cmd.Flags().BoolVar(&short, "short", false, "print only the version string")
	return cmd
}
