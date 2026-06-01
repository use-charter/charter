package scaffold

import "slices"

// Action is whether a candidate file should be created or left untouched.
type Action int

const (
	// Create writes the template contents because the file is absent.
	Create Action = iota
	// Skip leaves an already-present file untouched (never overwritten).
	Skip
)

// FileAction is a single planned file outcome. Contents is populated only for
// Create actions.
type FileAction struct {
	Path     string
	Action   Action
	Contents []byte
}

// Options configures the plan: the policy profile written to charter.yaml and
// the agent surfaces to scaffold for.
type Options struct {
	Profile string
	Agents  []string
}

// Plan computes the create-or-skip actions for the candidate scaffold files in
// a stable order. The exists predicate reports whether a repo-relative path is
// already present; it is injected so the plan is unit-testable without disk.
// A candidate is Skip when it already exists and Create (with template
// contents) otherwise. Plan never plans an overwrite or deletion.
func Plan(p Project, opts Options, exists func(string) bool) []FileAction {
	type candidate struct {
		path     string
		contents []byte
	}

	candidates := []candidate{
		{"AGENTS.md", AGENTSMarkdown(p)},
		{"charter.yaml", CharterYAML(opts.Profile)},
		{".gitignore", Gitignore()},
		{"ARCHITECTURE.md", ArchitectureMarkdown(p)},
		{".env.example", EnvExample()},
	}
	if slices.Contains(opts.Agents, "claude") {
		candidates = append(candidates, candidate{".claude/settings.json", ClaudeSettings()})
	}

	actions := make([]FileAction, 0, len(candidates))
	for _, c := range candidates {
		if exists(c.path) {
			actions = append(actions, FileAction{Path: c.path, Action: Skip})
			continue
		}
		actions = append(actions, FileAction{Path: c.path, Action: Create, Contents: c.contents})
	}
	return actions
}
