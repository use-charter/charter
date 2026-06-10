package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"go.use-charter.dev/charter/internal/repository"
	"go.use-charter.dev/charter/internal/suppress"
	"go.use-charter.dev/charter/internal/terminal"
)

var (
	ruleIDPattern  = regexp.MustCompile(`^AE-[A-Z]+-\d+$`)
	isoDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func newSuppressCommand() *cobra.Command {
	var path, reason, expires, approver string
	var dryRun bool
	var colorFlag string
	var noColor bool

	cmd := &cobra.Command{
		Use:   "suppress <RULE-ID>",
		Short: "Record a suppression in .charter-suppress.yml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rule := strings.TrimSpace(args[0])
			if !ruleIDPattern.MatchString(rule) {
				return commandExitError{message: fmt.Sprintf("invalid rule id %q: expected form AE-XXX-NNN", rule), exitCode: 2}
			}
			if strings.TrimSpace(reason) == "" {
				return commandExitError{message: "a --reason is required", exitCode: 2}
			}

			root, err := repository.ResolveRoot(path)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			storedExpires, warn, err := resolveExpires(expires, approver, time.Now())
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			entry := suppress.FileEntry{Rule: rule, Reason: strings.TrimSpace(reason), Expires: storedExpires, Approver: strings.TrimSpace(approver)}
			out, err := suppress.UpsertFile(root, entry)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			mode, merr := resolveColorMode(colorFlag, noColor)
			if merr != nil {
				return commandExitError{message: merr.Error(), exitCode: 2}
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

			w := cmd.OutOrStdout()

			if caps.ColorEnabled() {
				if warn != "" {
					_, _ = fmt.Fprintln(w, "warning: "+warn)
				}

				brand := st(terminal.TextInfo).Bold(true).Render("[C] charter")
				sup := st(terminal.TextTertiary).Render("suppress")
				ruleStyled := st(terminal.TextInfo).Bold(true).Render(entry.Rule)
				divider := st(terminal.TextTertiary).Render(strings.Repeat("─", 52))
				dot := st(terminal.TextTertiary).Render("  ·  ")

				label := func(name string) string {
					return "  " + st(terminal.TextTertiary).Render(fmt.Sprintf("%-10s", name))
				}
				val := func(v string) string { return st(terminal.TextSecondary).Render(v) }

				_, _ = fmt.Fprintln(w, brand+"  "+sup+"  "+ruleStyled)
				_, _ = fmt.Fprintln(w, divider)
				_, _ = fmt.Fprintln(w)
				_, _ = fmt.Fprintln(w, label("reason")+val(entry.Reason))
				_, _ = fmt.Fprintln(w, label("expires")+val(entry.Expires))
				if entry.Approver != "" {
					_, _ = fmt.Fprintln(w, label("approver")+val(entry.Approver))
				}
				_, _ = fmt.Fprintln(w)
				_, _ = fmt.Fprintln(w, divider)
				if dryRun {
					_, _ = fmt.Fprintln(w, "  "+
						st(terminal.TextWarning).Render("▸")+" "+
						st(terminal.TextWarning).Render("dry run")+
						dot+
						st(terminal.TextTertiary).Render(".charter-suppress.yml not written")+
						dot+
						st(terminal.TextTertiary).Render("remove --dry-run to apply"))
				} else {
					// #nosec G306 -- governance config is meant to be world-readable and committed.
					if err := os.WriteFile(filepath.Join(root, suppress.File), out, 0o644); err != nil {
						return commandExitError{message: err.Error(), exitCode: 2}
					}
					_, _ = fmt.Fprintln(w, "  "+
						st(terminal.TextSuccess).Render("✓")+" "+
						st(terminal.TextSuccess).Render("written")+
						st(terminal.TextTertiary).Render("  .charter-suppress.yml"))
				}
			} else {
				if warn != "" {
					_, _ = fmt.Fprintln(w, "warning: "+warn)
				}
				_, _ = fmt.Fprintf(w, "suppress %s\n  reason:   %s\n", entry.Rule, entry.Reason)
				_, _ = fmt.Fprintf(w, "  expires:  %s\n", entry.Expires)
				if entry.Approver != "" {
					_, _ = fmt.Fprintf(w, "  approver: %s\n", entry.Approver)
				}
				if dryRun {
					_, _ = fmt.Fprintf(w, "(dry run — %s not written)\n", suppress.File)
					return nil
				}
				// #nosec G306 -- governance config is meant to be world-readable and committed.
				if err := os.WriteFile(filepath.Join(root, suppress.File), out, 0o644); err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
				_, _ = fmt.Fprintf(w, "written %s\n", suppress.File)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "explicit repository path")
	cmd.Flags().StringVar(&reason, "reason", "", "human-readable reason for the suppression (required)")
	cmd.Flags().StringVar(&expires, "expires", "90d", "expiry: a duration (e.g. 90d), an ISO date (YYYY-MM-DD), or 'permanent'")
	cmd.Flags().StringVar(&approver, "approver", "", "approver handle (required for a permanent waiver to be honored)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the entry without writing the file")
	cmd.Flags().StringVar(&colorFlag, "color", "auto", "color output: auto, always, or never")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable color (equivalent to --color=never; wins over --color)")
	return cmd
}

// resolveExpires turns the --expires flag into the stored value: a duration like
// "90d" becomes an absolute YYYY-MM-DD from now; an ISO date passes through;
// "permanent" passes through, warning when no approver is set (it will not be honored).
func resolveExpires(expires, approver string, now time.Time) (stored, warn string, err error) {
	x := strings.TrimSpace(expires)
	switch {
	case strings.EqualFold(x, "permanent"):
		if strings.TrimSpace(approver) == "" {
			return "permanent", "permanent waiver without --approver is not honored and is flagged by AE-SUPPRESS-002", nil
		}
		return "permanent", "", nil
	case isoDatePattern.MatchString(x):
		if _, perr := time.Parse("2006-01-02", x); perr != nil {
			return "", "", fmt.Errorf("invalid --expires date %q: %w", expires, perr)
		}
		return x, "", nil
	default:
		days, derr := parseDayDuration(x)
		if derr != nil {
			return "", "", fmt.Errorf("invalid --expires %q: use a duration like 90d, an ISO date YYYY-MM-DD, or permanent", expires)
		}
		return now.AddDate(0, 0, days).Format("2006-01-02"), "", nil
	}
}

// parseDayDuration parses a "<n>d" day window.
func parseDayDuration(s string) (int, error) {
	if !strings.HasSuffix(s, "d") {
		return 0, fmt.Errorf("not a day duration")
	}
	n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid day duration %q", s)
	}
	return n, nil
}
