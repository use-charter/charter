package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.charter.dev/charter/internal/repository"
	"go.charter.dev/charter/internal/suppress"
)

var (
	ruleIDPattern  = regexp.MustCompile(`^AE-[A-Z]+-\d+$`)
	isoDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func newSuppressCommand() *cobra.Command {
	var path, reason, expires, approver string
	var dryRun bool

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

			w := cmd.OutOrStdout()
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
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "explicit repository path")
	cmd.Flags().StringVar(&reason, "reason", "", "human-readable reason for the suppression (required)")
	cmd.Flags().StringVar(&expires, "expires", "90d", "expiry: a duration (e.g. 90d), an ISO date (YYYY-MM-DD), or 'permanent'")
	cmd.Flags().StringVar(&approver, "approver", "", "approver handle (required for a permanent waiver to be honored)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the entry without writing the file")
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
