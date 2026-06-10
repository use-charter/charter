package main

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/fix"
	"go.use-charter.dev/charter/internal/repository"
	"go.use-charter.dev/charter/internal/terminal"
)

func newFixCommand() *cobra.Command {
	var path, rule string
	var dryRun, all, yes bool
	var colorFlag string
	var noColor bool

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

			if caps.ColorEnabled() {
				for i, plan := range plans {
					if i > 0 {
						_, _ = fmt.Fprintln(w)
					}
					_, _ = fmt.Fprintln(w, st(terminal.TextInfo).Bold(true).Render(plan.RuleID)+"  "+st(terminal.TextTertiary).Render(plan.Path))
					for _, line := range strings.Split(plan.Diff, "\n") {
						if line == "" {
							continue
						}
						var rendered string
						switch {
						case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
							rendered = st(terminal.TextTertiary).Render(line)
						case strings.HasPrefix(line, "@@"):
							rendered = st(terminal.TextTertiary).Render(line)
						case strings.HasPrefix(line, "+"):
							rendered = st(terminal.TextSuccess).Render(line)
						case strings.HasPrefix(line, "-"):
							rendered = st(terminal.TextDanger).Render(line)
						default:
							rendered = st(terminal.TextTertiary).Render(line)
						}
						_, _ = fmt.Fprintln(w, rendered)
					}
				}

				divider := st(terminal.TextTertiary).Render(strings.Repeat("─", 52))
				dot := st(terminal.TextTertiary).Render("  ·  ")

				if dryRun {
					_, _ = fmt.Fprintln(w, divider)
					_, _ = fmt.Fprintln(w, "  "+
						st(terminal.TextWarning).Render("▸")+" "+
						st(terminal.TextWarning).Render("dry run")+
						dot+
						st(terminal.TextTertiary).Render(fmt.Sprintf("%d fix(es) ready", len(plans)))+
						dot+
						st(terminal.TextInfo).Render("charter fix")+
						st(terminal.TextTertiary).Render(" to apply"))
					return nil
				}

				written, backupDir, err := fix.Apply(root, plans)
				if err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
				for _, p := range written {
					_, _ = fmt.Fprintln(w, "  "+
						st(terminal.TextSuccess).Render("✓")+" "+
						st(terminal.TextSuccess).Render("written")+
						st(terminal.TextTertiary).Render("  "+p))
				}
				if backupDir != "" {
					_, _ = fmt.Fprintf(w, "backups: %s\n", backupDir)
				}
			} else {
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
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "explicit repository path")
	cmd.Flags().StringVar(&rule, "rule", "", "limit the repair to a single rule id (e.g. AE-CTX-004)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the diffs without writing any files")
	cmd.Flags().BoolVar(&all, "all", false, "plan every fixable rule (already the default)")
	cmd.Flags().BoolVar(&yes, "yes", false, "accepted for scripting; fix is non-interactive, so this is a no-op")
	cmd.Flags().StringVar(&colorFlag, "color", "auto", "color output: auto, always, or never")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable color (equivalent to --color=never; wins over --color)")
	return cmd
}
