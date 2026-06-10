package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"go.use-charter.dev/charter/internal/terminal"
	"go.use-charter.dev/charter/internal/version"
)

func newVersionCommand() *cobra.Command {
	var (
		format    string
		short     bool
		colorFlag string
		noColor   bool
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
				mode, err := resolveColorMode(colorFlag, noColor)
				if err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
				caps, pal := terminalContext(cmd, "", mode)

				st := func(tok terminal.Token) lipgloss.Style {
					resolved := pal.Resolve(tok)
					s := lipgloss.NewStyle()
					if resolved.HasColor() {
						s = s.Foreground(resolved.Color)
					}
					if resolved.Bold {
						s = s.Bold(true)
					}
					if resolved.Faint {
						s = s.Faint(true)
					}
					if resolved.Reverse {
						s = s.Reverse(true)
					}
					return s
				}

				commit := version.Commit()
				if len(commit) > 8 {
					commit = commit[:8]
				}

				if caps.ColorEnabled() {
					brand := st(terminal.TextInfo).Bold(true).Render("[C] charter")
					ver := st(terminal.TextPrimary).Bold(true).Render(version.Version())
					_, _ = fmt.Fprintln(out, brand+"  "+ver)
					_, _ = fmt.Fprintln(out)
					dot := st(terminal.TextTertiary).Render("  ·  ")
					built := version.Date()
					if len(built) > 10 {
						built = built[:10]
					}
					meta := st(terminal.TextTertiary).Render("  go "+goVersion) +
						dot + st(terminal.TextTertiary).Render(platform) +
						dot + st(terminal.TextTertiary).Render("commit "+commit) +
						dot + st(terminal.TextTertiary).Render(built)
					_, _ = fmt.Fprintln(out, meta)
				} else {
					_, _ = fmt.Fprintf(out, "charter   %s\n", version.Version())
					_, _ = fmt.Fprintf(out, "commit    %s\n", commit)
					_, _ = fmt.Fprintf(out, "built     %s\n", version.Date())
					_, _ = fmt.Fprintf(out, "go        %s\n", goVersion)
					_, _ = fmt.Fprintf(out, "platform  %s\n", platform)
				}
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
	cmd.Flags().StringVar(&colorFlag, "color", "auto", "color output: auto, always, or never")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable color (equivalent to --color=never; wins over --color)")
	return cmd
}
