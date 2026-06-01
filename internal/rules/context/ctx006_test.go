package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
	"go.use-charter.dev/charter/internal/scoring"
)

func ctx006Repo(t *testing.T, agentsMD string) (string, repository.Inventory) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(agentsMD), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, repository.New([]string{"AGENTS.md"})
}

func TestCTX006FlagsOverEmphasis(t *testing.T) {
	// 5 emphatic directives in ~12 words -> ~417 per 1K, well over threshold.
	root, inv := ctx006Repo(t, "IMPORTANT NEVER MUST CRITICAL ALWAYS do the thing right now please ok")
	f, ok := checkCTX006(root, inv)
	if !ok {
		t.Fatal("expected AE-CTX-006 to fire on an over-emphasized file")
	}
	if f.RuleID != "AE-CTX-006" || !f.Informational {
		t.Fatalf("expected informational AE-CTX-006, got %+v", f)
	}
	// Informational findings must not deduct from the score.
	if got := scoring.Calculate([]findings.Finding{f}).Final; got != 100 {
		t.Fatalf("informational finding must not deduct; score = %d, want 100", got)
	}
}

func TestCTX006CleanOnDeclarativeFile(t *testing.T) {
	body := "# AGENTS.md\n\n" + strings.Repeat("This project is a deterministic Go CLI. Tests run with go test. The build uses moon. ", 20)
	root, inv := ctx006Repo(t, body)
	if _, ok := checkCTX006(root, inv); ok {
		t.Fatal("a concise declarative file should not flag AE-CTX-006")
	}
}

func TestCTX006NoContextFileNoFinding(t *testing.T) {
	root := t.TempDir()
	if _, ok := checkCTX006(root, repository.New(nil)); ok {
		t.Fatal("no context file -> AE-CTX-006 must not fire (AE-CTX-001 owns absence)")
	}
}
