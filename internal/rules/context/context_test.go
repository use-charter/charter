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
			"Charter fixture repo used to prove context rule behavior.",
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

func TestAECTX001PassesForCursorRulesContext(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		".cursor/rules/charter.mdc": strings.Join([]string{
			"# Fixture Repo",
			"",
			"Charter fixture repo used to prove Cursor rule context behavior.",
			"- tech stack: Go and Bun",
			"- verify with `moon run :check`",
			"- off-limits: `.env*`, `secrets/`",
		}, "\n"),
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

func TestAECTX001FindsWeakContextContent(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"# Fixture Repo",
			"",
			"- verify with `moon run :check`",
			"- off-limits: `.env*`, `secrets/`",
		}, "\n"),
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := RunCTXRules(root, inv)
	for _, finding := range findings {
		if finding.RuleID != "AE-CTX-001" {
			continue
		}

		if !containsEvidence(finding.Evidence, "missing content signal: tech stack") {
			t.Fatalf("expected missing tech stack evidence, got %#v", finding.Evidence)
		}
		if !containsEvidence(finding.Evidence, "missing content signal: project summary") {
			t.Fatalf("expected missing project summary evidence, got %#v", finding.Evidence)
		}
		if !containsEvidencePrefix(finding.Evidence, "first substantive line: # Fixture Repo") {
			t.Fatalf("expected first substantive line evidence, got %#v", finding.Evidence)
		}
		for _, item := range finding.Evidence {
			if strings.Contains(item, "verify with") {
				t.Fatalf("expected no raw context excerpt evidence, got %#v", finding.Evidence)
			}
		}
		return
	}

	t.Fatalf("expected AE-CTX-001 finding")
}

func TestAECTX001FailsOverBudgetContext(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"# Fixture Repo",
			"",
			"Charter fixture repo used to prove over-budget context fails even when content is otherwise meaningful.",
			"- tech stack: Go 1.26.3 CLI with Bun tooling and Moonrepo tasks.",
			"- off-limits: `.env*`, `secrets/`, signing keys, production infra.",
			"- verify with `moon run :check`.",
			strings.Repeat("This context sentence exists only to push the file beyond the token budget while staying meaningful and explicit. ", 30),
		}, "\n"),
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := RunCTXRules(root, inv)
	for _, finding := range findings {
		if finding.RuleID != "AE-CTX-001" {
			continue
		}

		if !containsEvidencePrefix(finding.Evidence, "context appears over budget: ~") {
			t.Fatalf("expected over-budget evidence, got %#v", finding.Evidence)
		}
		return
	}

	t.Fatalf("expected AE-CTX-001 finding for over-budget context")
}

func TestAECTX001PrefersAGENTSOverSecondaryContextFiles(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"# Fixture Repo",
			"",
			"- verify with `moon run :check`",
			"- off-limits: `.env*`, `secrets/`",
		}, "\n"),
		".github/copilot-instructions.md": strings.Join([]string{
			"# Secondary Context",
			"",
			"Charter fixture repo used to prove secondary context files cannot mask a weak AGENTS.md.",
			"- tech stack: Go and Bun",
			"- verify with `moon run :check`",
			"- off-limits: `.env*`, `secrets/`",
		}, "\n"),
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := RunCTXRules(root, inv)
	for _, finding := range findings {
		if finding.RuleID != "AE-CTX-001" {
			continue
		}

		if !containsEvidence(finding.Evidence, "context location: AGENTS.md") {
			t.Fatalf("expected AGENTS.md evidence, got %#v", finding.Evidence)
		}
		if containsEvidence(finding.Evidence, "context location: .github/copilot-instructions.md") {
			t.Fatalf("expected canonical AGENTS.md to take precedence, got %#v", finding.Evidence)
		}
		return
	}

	t.Fatalf("expected AE-CTX-001 finding")
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

func TestAECTX001RedactsSecretLikeFirstSubstantiveLineEvidence(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"OPENAI_API_KEY=sk-proj-abcdefghijklmnopqrstuvwxyz1234567890",
			"",
			"- verify with `moon run :check`",
		}, "\n"),
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := RunCTXRules(root, inv)
	for _, finding := range findings {
		if finding.RuleID != "AE-CTX-001" {
			continue
		}

		if !containsEvidencePrefix(finding.Evidence, "first substantive line: [redacted]") {
			t.Fatalf("expected redacted first substantive line evidence, got %#v", finding.Evidence)
		}
		for _, item := range finding.Evidence {
			if strings.Contains(item, "sk-proj-") {
				t.Fatalf("expected secret-like content to be redacted, got %#v", finding.Evidence)
			}
		}
		return
	}

	t.Fatalf("expected AE-CTX-001 finding")
}

func TestAECTX004FindsMissingIgnoresAndTrackedArtifacts(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"# Fixture Repo",
			"",
			"Charter fixture repo used to prove gitignore rule behavior.",
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

func TestAECTX004IgnoresCommentedPatterns(t *testing.T) {
	root := newContextRepo(t, map[string]string{
		"AGENTS.md": strings.Join([]string{
			"# Fixture Repo",
			"",
			"Charter fixture repo used to prove gitignore comment handling.",
			"- verify with `moon run :check`",
			"- off-limits: `.env*`, `secrets/`",
			"- hooks use `hk.pkl`",
			"- product truth: `docs/internal/architecture/charter-architecture-2026.md`",
		}, "\n"),
		".gitignore": strings.Join([]string{
			"# .charter/",
			"# *.charter-session",
			"# .claude/local/",
			"# .cursor/cache/",
			"# .hk/",
			"# .env*",
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
		if finding.RuleID != "AE-CTX-004" {
			continue
		}

		if !containsEvidence(finding.Evidence, "missing ignore pattern: .charter/") {
			t.Fatalf("expected commented .charter pattern not to count, got %#v", finding.Evidence)
		}
		if !containsEvidence(finding.Evidence, "missing ignore pattern: .env*") {
			t.Fatalf("expected commented .env pattern not to count, got %#v", finding.Evidence)
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

func containsEvidencePrefix(evidence []string, want string) bool {
	for _, item := range evidence {
		if strings.HasPrefix(item, want) {
			return true
		}
	}
	return false
}
