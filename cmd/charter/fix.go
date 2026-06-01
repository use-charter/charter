package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/fix"
	"go.use-charter.dev/charter/internal/repository"
)

func newFixCommand() *cobra.Command {
	var path, rule string
	var dryRun, all, yes bool

	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Preview unified diffs for fixable findings and apply them (originals are backed up)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rule = strings.TrimSpace(rule)
			if rule != "" && !ruleIDPattern.MatchString(rule) {
				return commandExitError{message: fmt.Sprintf("invalid rule id %q: expected form AE-XXX-NNN", rule), exitCode: 2}
			}

			root, err := repository.ResolveRoot(path)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			inv, err := repository.BuildInventory(root)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			result, err := doctor.Run(path, 80, false)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			plans, err := fix.Plan(result, root, inv, fix.Options{Rule: rule})
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			w := cmd.OutOrStdout()

			if len(plans) == 0 {
				// A targeted rule with no registered fixer (notably the secret and
				// dangerous-config rules) is reported as manual-only so the caller
				// knows the no-op is by design, not a missed finding.
				if rule != "" && !fix.Fixable(rule) {
					_, _ = fmt.Fprintf(w, "%s is not auto-fixable; remediate manually.\n", rule)
				} else {
					_, _ = fmt.Fprintln(w, "nothing to fix.")
				}
				return nil
			}

			for _, plan := range plans {
				_, _ = fmt.Fprintf(w, "%s  %s\n", plan.RuleID, plan.Path)
				_, _ = fmt.Fprint(w, plan.Diff)
			}

			if dryRun {
				_, _ = fmt.Fprintln(w, "(dry run — no files written)")
				return nil
			}

			written, backupDir, err := fix.Apply(root, plans)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}
			for _, p := range written {
				_, _ = fmt.Fprintf(w, "wrote %s\n", p)
			}
			if backupDir != "" {
				_, _ = fmt.Fprintf(w, "backups: %s\n", backupDir)
			}
			_, _ = fmt.Fprintf(w, "%d fixed\n", len(written))
			_, _ = fmt.Fprintln(w, "› Re-run: charter doctor")
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "explicit repository path")
	cmd.Flags().StringVar(&rule, "rule", "", "limit the repair to a single rule id (e.g. AE-CTX-004)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the diffs without writing any files")
	cmd.Flags().BoolVar(&all, "all", false, "plan every fixable rule (already the default)")
	cmd.Flags().BoolVar(&yes, "yes", false, "accepted for scripting; fix is non-interactive, so this is a no-op")
	return cmd
}
