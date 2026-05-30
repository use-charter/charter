package agentcontext

import "testing"

func TestFilesAreTheCanonicalContextSet(t *testing.T) {
	want := []string{
		"AGENTS.md",
		"CLAUDE.md",
		".windsurfrules",
		".github/copilot-instructions.md",
		"opencode.md",
		"codex.md",
		"DESIGN.md",
		"SKILL.md",
	}

	if len(Files) != len(want) {
		t.Fatalf("expected %d context files, got %d: %v", len(want), len(Files), Files)
	}
	for i, f := range want {
		if Files[i] != f {
			t.Errorf("Files[%d] = %q, want %q (precedence order matters)", i, Files[i], f)
		}
	}
}

func TestCursorRulesDirIsStable(t *testing.T) {
	if CursorRulesDir != ".cursor/rules" {
		t.Fatalf("CursorRulesDir = %q, want .cursor/rules", CursorRulesDir)
	}
}
