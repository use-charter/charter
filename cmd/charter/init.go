package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"go.use-charter.dev/charter/internal/repository"
	"go.use-charter.dev/charter/internal/scaffold"
	"go.use-charter.dev/charter/internal/terminal"
	"go.use-charter.dev/charter/internal/version"
)

func newInitCommand() *cobra.Command {
	var path, profile, agentsFlag string
	var dryRun, yes bool
	var colorFlag string
	var noColor bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold missing agent-context files (AGENTS.md, charter.yaml, ...)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			switch profile {
			case "strict", "standard", "relaxed":
			default:
				return commandExitError{message: fmt.Sprintf("invalid profile %q: must be strict, standard, or relaxed", profile), exitCode: 2}
			}

			root, err := repository.ResolveRoot(path)
			if err != nil {
				// Not inside a git repo: scaffold the literal target directory
				// (or the current directory when --path is empty).
				fallback := path
				if fallback == "" {
					fallback = "."
				}
				root, err = filepath.Abs(fallback)
				if err != nil {
					return commandExitError{message: err.Error(), exitCode: 2}
				}
			}

			proj, err := scaffold.Detect(root)
			if err != nil {
				return commandExitError{message: err.Error(), exitCode: 2}
			}

			agents := resolveAgents(agentsFlag, proj.Agents)

			actions := scaffold.Plan(proj, scaffold.Options{Profile: profile, Agents: agents}, func(rel string) bool {
				_, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
				return statErr == nil
			})

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

			w := cmd.OutOrStdout()

			if caps.ColorEnabled() {
				brand := st(terminal.TextInfo).Bold(true).Render("[C] charter")
				meta := st(terminal.TextTertiary).Render("  v" + version.Version() + "  ·  " + root)
				_, _ = fmt.Fprintln(w, brand+meta)
				_, _ = fmt.Fprintln(w)

				created, skipped := 0, 0
				for _, action := range actions {
					abs := filepath.Join(root, filepath.FromSlash(action.Path))

					switch {
					case action.Action == scaffold.Skip:
						skipped++
						_, _ = fmt.Fprintln(w, st(terminal.TextTertiary).Render("  skip  "+action.Path))
					case dryRun:
						created++
						filePart := st(terminal.TextInfo).Render(action.Path)
						statusPart := st(terminal.TextSuccess).Render("would create")
						_, _ = fmt.Fprintln(w, "  "+statusPart+"  "+filePart)
					default:
						// #nosec G301 -- scaffolded directories hold world-readable committed context files.
						if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
							return commandExitError{message: err.Error(), exitCode: 2}
						}
						// Re-stat immediately before writing so a file that appeared
						// after planning is never overwritten (TOCTOU-safe; honors
						// the never-overwrite/never-delete commitments).
						if _, statErr := os.Stat(abs); statErr == nil {
							skipped++
							_, _ = fmt.Fprintln(w, st(terminal.TextTertiary).Render("  skip  "+action.Path))
							continue
						}
						// #nosec G306 -- scaffolded context files are meant to be committed and world-readable.
						if err := os.WriteFile(abs, action.Contents, 0o644); err != nil {
							return commandExitError{message: err.Error(), exitCode: 2}
						}
						created++
						filePart := st(terminal.TextInfo).Render(action.Path)
						statusPart := st(terminal.TextSuccess).Render("create")
						_, _ = fmt.Fprintln(w, "  "+statusPart+"  "+filePart)
					}
				}

				_, _ = fmt.Fprintln(w)
				countStr := st(terminal.TextSuccess).Render(fmt.Sprintf("%d created", created))
				skipStr := st(terminal.TextTertiary).Render(fmt.Sprintf(" · %d skipped", skipped))
				_, _ = fmt.Fprintln(w, countStr+skipStr)
				arrow := st(terminal.TextInfo).Render("›")
				next := st(terminal.TextSecondary).Render(" Next: charter doctor")
				_, _ = fmt.Fprintln(w, arrow+next)
			} else {
				prefix := ""
				if dryRun {
					prefix = "would "
				}

				created, skipped := 0, 0
				for _, action := range actions {
					abs := filepath.Join(root, filepath.FromSlash(action.Path))

					switch {
					case action.Action == scaffold.Skip:
						skipped++
						_, _ = fmt.Fprintf(w, "%sskip %s\n", prefix, action.Path)
					case dryRun:
						created++
						_, _ = fmt.Fprintf(w, "%screate %s\n", prefix, action.Path)
					default:
						// #nosec G301 -- scaffolded directories hold world-readable committed context files.
						if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
							return commandExitError{message: err.Error(), exitCode: 2}
						}
						// Re-stat immediately before writing so a file that appeared
						// after planning is never overwritten (TOCTOU-safe; honors
						// the never-overwrite/never-delete commitments).
						if _, statErr := os.Stat(abs); statErr == nil {
							skipped++
							_, _ = fmt.Fprintf(w, "skip %s\n", action.Path)
							continue
						}
						// #nosec G306 -- scaffolded context files are meant to be committed and world-readable.
						if err := os.WriteFile(abs, action.Contents, 0o644); err != nil {
							return commandExitError{message: err.Error(), exitCode: 2}
						}
						created++
						_, _ = fmt.Fprintf(w, "create %s\n", action.Path)
					}
				}

				_, _ = fmt.Fprintf(w, "%d created · %d skipped\n", created, skipped)
				_, _ = fmt.Fprintln(w, "› Next: charter doctor")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "target directory to scaffold (defaults to the repository root, else the current directory)")
	cmd.Flags().StringVar(&profile, "profile", "standard", "policy profile written to charter.yaml: strict, standard, or relaxed")
	cmd.Flags().StringVar(&agentsFlag, "agents", "", "comma-separated agent surfaces to scaffold (e.g. claude,cursor); empty auto-detects")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the planned actions without writing any files")
	cmd.Flags().BoolVar(&yes, "yes", false, "accepted for scripting; init never overwrites, so this is a no-op")
	cmd.Flags().StringVar(&colorFlag, "color", "auto", "color output: auto, always, or never")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable color (equivalent to --color=never; wins over --color)")
	return cmd
}

// resolveAgents picks the agent surfaces to scaffold: an explicit --agents list
// (comma-separated, trimmed, empties dropped) wins; otherwise the detected
// surfaces are used; when neither yields anything it defaults to claude.
func resolveAgents(flag string, detected []string) []string {
	var agents []string
	if strings.TrimSpace(flag) != "" {
		for _, a := range strings.Split(flag, ",") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				agents = append(agents, trimmed)
			}
		}
	} else {
		agents = detected
	}
	if len(agents) == 0 {
		agents = []string{"claude"}
	}
	return agents
}
