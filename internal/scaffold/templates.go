package scaffold

import (
	"fmt"
	"strings"
)

// AGENTSMarkdown builds the root agent-context file. The output is designed to
// satisfy Charter's own rules out of the box: it stays under the AE-CTX-001
// ~600-token budget with >= 5 non-empty lines, names the tech stack and a
// `charter doctor` verification command, and declares concrete off-limits
// paths (`.github/workflows/`, `.env*`, `secrets/`, ...) for AE-CC-002 and the
// `.env*`/`secrets/` markers AE-CTX-002 requires.
func AGENTSMarkdown(p Project) []byte {
	primary := primaryLanguage(p)

	var b strings.Builder
	b.WriteString("# AGENTS.md\n\n")

	b.WriteString("## Project Overview\n\n")
	fmt.Fprintf(&b, "%s is a %s project. Describe what it does and who uses it (edit this line).\n\n", p.Name, primary)

	b.WriteString("## Tech Stack\n\n")
	for _, line := range techStackLines(p) {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')

	b.WriteString("## Commands\n\n")
	b.WriteString("- Verify: `charter doctor`\n")
	for _, line := range buildTestCommands(primary) {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')

	b.WriteString("## Edit Boundaries\n\n")
	b.WriteString("- Safe for agents: application source, tests, docs\n")
	b.WriteString("- Off-limits: `.github/workflows/`, `.env*`, `secrets/`, `db/migrations/`, `terraform/`\n\n")

	b.WriteString("## Verification\n\n")
	b.WriteString("- Run `charter doctor` before committing.\n")

	return []byte(b.String())
}

func primaryLanguage(p Project) string {
	if len(p.Langs) > 0 && strings.TrimSpace(p.Langs[0].Name) != "" {
		return p.Langs[0].Name
	}
	return "software"
}

func techStackLines(p Project) []string {
	var lines []string
	for _, lang := range p.Langs {
		entry := "- " + lang.Name
		if strings.TrimSpace(lang.Version) != "" {
			entry += " " + lang.Version
		}
		lines = append(lines, entry)
	}
	if len(lines) == 0 {
		lines = append(lines, "- Primary language: (edit)")
	}
	if p.CI != "" {
		lines = append(lines, "- CI: "+p.CI)
	}
	return lines
}

// buildTestCommands returns best-effort build/test command bullets for the
// primary language. Unknown or absent languages get none (the `charter doctor`
// verification line still stands on its own).
func buildTestCommands(primary string) []string {
	switch primary {
	case "Go":
		return []string{"- Build: `go build ./...`", "- Test: `go test ./...`"}
	case "JavaScript/TypeScript":
		return []string{"- Build: `npm run build`", "- Test: `npm test`"}
	case "Python":
		return []string{"- Test: `pytest`"}
	case "Rust":
		return []string{"- Build: `cargo build`", "- Test: `cargo test`"}
	default:
		return nil
	}
}

// CharterYAML builds the minimal gate config selecting a policy profile.
func CharterYAML(profile string) []byte {
	return []byte("policy:\n  profile: " + profile + "\n")
}

// Gitignore builds an ignore block covering the agent-artifact, hook-state, and
// env patterns AE-CTX-004 requires (.charter/, *.charter-session, .claude/local/,
// .cursor/cache/, .hk/, .env*). The trailing negation keeps the committed
// .env.example trackable while .env and friends stay ignored.
func Gitignore() []byte {
	return []byte("# Charter / agent session artifacts\n" +
		".charter/\n" +
		"*.charter-session\n" +
		".claude/local/\n" +
		".cursor/cache/\n" +
		".hk/\n" +
		".env*\n" +
		"!.env.example\n")
}

// ClaudeSettings builds an MCP-free Claude Code settings file that denies reads
// of secret-bearing paths.
func ClaudeSettings() []byte {
	return []byte(`{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "permissions": {
    "deny": ["Read(./.env)", "Read(./.env.*)", "Read(./secrets/**)"]
  }
}
`)
}

// ArchitectureMarkdown builds a minimal ARCHITECTURE.md stub for the project.
func ArchitectureMarkdown(p Project) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s Architecture\n\n", p.Name)
	b.WriteString("## Overview\n\n")
	b.WriteString("Describe the system's purpose and major components (edit this section).\n\n")
	b.WriteString("## Layout\n\n")
	b.WriteString("Document the key directories and their responsibilities (edit this section).\n")
	return []byte(b.String())
}

// EnvExample builds a placeholder .env.example with a header comment and one
// example variable.
func EnvExample() []byte {
	return []byte("# Example environment variables. Copy to .env (git-ignored) and fill in real values.\n" +
		"# Never commit real secrets.\n" +
		"# EXAMPLE_VAR=\n")
}
