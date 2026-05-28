package context

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"go.charter.dev/charter/internal/repository"
)

func TestAECTX001PassesForFixtureRepo(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"# Fixture Repo",
			"",
			"- uses Go",
			"- verify with `moon run :check`",
			"- off-limits: `.env*`, `secrets/`",
			"- hooks use `hk.pkl`",
			"- product truth: `docs/internal/architecture/charter-architecture-2026.md`",
		}, "\n"),
		".gitignore": strings.Join([]string{
			".charter/",
			"*.charter-session",
			".claude/local/",
			".cursor/cache/",
			".hk/",
			".env*",
		}, "\n"),
		"hk.pkl": "hooks {}\n",
		"docs/internal/architecture/charter-architecture-2026.md": "# Product Truth\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := RunCTXRules(root, inv)
	for _, finding := range findings {
		if finding.RuleID == "AE-CTX-001" {
			t.Fatalf("expected no AE-CTX-001 finding, got %#v", finding)
		}
	}
}

func TestAECTX002FindsStaleRepoTruthMarkers(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"# Fixture Repo",
			"",
			"- verify with `moon run :check`",
			"- off-limits: `.env*`, `secrets/`",
		}, "\n"),
		".gitignore": strings.Join(requiredGitignorePatterns, "\n"),
		"hk.pkl":     "hooks {}\n",
		"docs/internal/architecture/charter-architecture-2026.md": "# Product Truth\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := RunCTXRules(root, inv)
	for _, finding := range findings {
		if finding.RuleID != "AE-CTX-002" {
			continue
		}

		if !containsEvidence(finding.Evidence, "hk.pkl") {
			t.Fatalf("expected missing hk.pkl evidence, got %#v", finding.Evidence)
		}
		return
	}

	t.Fatalf("expected AE-CTX-002 finding")
}

func TestAECTX004FindsMissingIgnoresAndTrackedArtifacts(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"# Fixture Repo",
			"",
			"- verify with `moon run :check`",
			"- off-limits: `.env*`, `secrets/`",
			"- hooks use `hk.pkl`",
			"- product truth: `docs/internal/architecture/charter-architecture-2026.md`",
		}, "\n"),
		".gitignore": ".charter/\n",
		".env.local": "EXAMPLE=true\n",
		"hk.pkl":     "hooks {}\n",
		"docs/internal/architecture/charter-architecture-2026.md": "# Product Truth\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := RunCTXRules(root, inv)
	for _, finding := range findings {
		if finding.RuleID != "AE-CTX-004" {
			continue
		}

		if !containsEvidence(finding.Evidence, "missing ignore pattern: .env*") {
			t.Fatalf("expected missing ignore evidence, got %#v", finding.Evidence)
		}
		if !containsEvidence(finding.Evidence, "tracked local artifact: .env.local") {
			t.Fatalf("expected tracked artifact evidence, got %#v", finding.Evidence)
		}
		return
	}

	t.Fatalf("expected AE-CTX-004 finding")
}

func newContextRepo(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create dir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	gitInit := exec.Command("git", "init", "-q", root)
	if output, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	gitAdd := exec.Command("git", "-C", root, "add", ".")
	if output, err := gitAdd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v: %s", err, output)
	}

	return root
}

func containsEvidence(evidence []string, want string) bool {
	for _, item := range evidence {
		if item == want {
			return true
		}
	}
	return false
}
