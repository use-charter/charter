// Package agentcontext is the single source of truth for the agent-visible
// context files Charter rules inspect. The context rules and the secret rules
// both consume this set so a new context file type cannot be recognized as
// context while silently escaping secret scanning.
package agentcontext

// Files are the single-file agent context candidates, in precedence order
// (the context rule treats the first matching entry as the canonical file).
var Files = []string{
	"AGENTS.md",
	"CLAUDE.md",
	".windsurfrules",
	".github/copilot-instructions.md",
	"opencode.md",
	"codex.md",
	"DESIGN.md",
	"SKILL.md",
}

// CursorRulesDir is the directory-tree context source, handled separately from
// the single-file candidates because it aggregates multiple rule files.
const CursorRulesDir = ".cursor/rules"
