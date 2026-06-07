package secrets

import (
	"path/filepath"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

func TestAESEC001IgnoresSafePlaceholderFixture(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "testdata", "repos", "pass-secrets-agent")
	root, err := makeTempGitRepoFromFixture(t, fixture)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings, err := RunSecretRules(root, inv)
	if err != nil {
		t.Fatalf("run secret rules: %v", err)
	}
	for _, finding := range findings {
		if finding.RuleID == "AE-SEC-001" {
			t.Fatalf("expected no AE-SEC-001 finding, got %#v", finding)
		}
	}
}

func TestAESEC001FindsSecretInAgentFile(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "testdata", "repos", "pass-secrets-agent")
	root, err := makeTempGitRepoFromFixture(t, fixture)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	writeFile(t, root, "AGENTS.md", strings.Join([]string{
		"# pass-secrets-agent",
		"",
		"Project summary: pass-secrets-agent is a fixture repository for Charter secret-rule tests.",
		"",
		"- Tech stack: uses Go and Bun with repo automation routed through Moon.",
		"- Edit boundaries: off-limits paths include `.github/workflows/`, `.env*`, and `secrets/`.",
		"- Verify with `moon run :check` before claiming the fixture passes.",
		"- Hooks use `hk.pkl` and product truth lives in `docs/internal/architecture/charter-architecture-2026.md`.",
		"- OPENAI_API_KEY=" + fakeOpenAIKey(),
	}, "\n")+"\n")
	stageAndCommitAll(t, root, "fixture-update")

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings, err := RunSecretRules(root, inv)
	if err != nil {
		t.Fatalf("run secret rules: %v", err)
	}
	for _, finding := range findings {
		if finding.RuleID != "AE-SEC-001" {
			continue
		}

		if len(finding.Evidence) == 0 {
			t.Fatalf("expected evidence for AE-SEC-001")
		}
		if strings.Contains(finding.Evidence[0], fakeOpenAIKey()) {
			t.Fatalf("expected redacted evidence, got raw secret: %#v", finding.Evidence)
		}
		if len(finding.Locations) == 0 || finding.Locations[0].Path != "AGENTS.md" || finding.Locations[0].Line < 1 {
			t.Fatalf("expected AGENTS.md location with a line, got %#v", finding.Locations)
		}
		return
	}

	t.Fatalf("expected AE-SEC-001 finding")
}

func TestAESEC001FindsSecretInUntrackedAgentFileInInventory(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "testdata", "repos", "pass-secrets-agent")
	root, err := makeTempGitRepoFromFixture(t, fixture)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	writeFile(t, root, "CLAUDE.md", strings.Join([]string{
		"# Local Agent File",
		"",
		"Project summary: local scratch file.",
		"- verify with `moon run :check`.",
		"OPENAI_API_KEY=" + fakeOpenAIKey(),
	}, "\n")+"\n")

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}
	if !inv.Has("CLAUDE.md") {
		t.Fatalf("expected untracked CLAUDE.md to appear in inventory")
	}

	findings, err := RunSecretRules(root, inv)
	if err != nil {
		t.Fatalf("run secret rules: %v", err)
	}
	findAESEC001(t, findings)
}

func TestAESEC001FindsSecretInCursorRulesFile(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "testdata", "repos", "pass-secrets-agent")
	root, err := makeTempGitRepoFromFixture(t, fixture)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	writeFile(t, root, ".cursor/rules", strings.Join([]string{
		"---",
		"description: secret rule fixture",
		"OPENAI_API_KEY=" + fakeOpenAIKey(),
	}, "\n")+"\n")
	stageAndCommitAll(t, root, "fixture-update")

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	allFindings, err := RunSecretRules(root, inv)
	if err != nil {
		t.Fatalf("run secret rules: %v", err)
	}
	finding := findAESEC001(t, allFindings)
	if !strings.Contains(finding.Evidence[0], ".cursor/rules") {
		t.Fatalf("expected .cursor/rules evidence, got %#v", finding.Evidence)
	}
	if strings.Contains(finding.Evidence[0], fakeOpenAIKey()) {
		t.Fatalf("expected redacted evidence, got raw secret: %#v", finding.Evidence)
	}
}

func TestAESEC001FindsSecretInNestedCursorRulesFile(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "testdata", "repos", "pass-secrets-agent")
	root, err := makeTempGitRepoFromFixture(t, fixture)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	writeFile(t, root, ".cursor/rules/security/rule.mdc", strings.Join([]string{
		"---",
		"description: nested secret rule fixture",
		"OPENAI_API_KEY=" + fakeOpenAIKey(),
	}, "\n")+"\n")
	stageAndCommitAll(t, root, "fixture-update")

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	allFindings, err := RunSecretRules(root, inv)
	if err != nil {
		t.Fatalf("run secret rules: %v", err)
	}
	finding := findAESEC001(t, allFindings)
	if !strings.Contains(finding.Evidence[0], ".cursor/rules/security/rule.mdc") {
		t.Fatalf("expected nested .cursor/rules evidence, got %#v", finding.Evidence)
	}
	if strings.Contains(finding.Evidence[0], fakeOpenAIKey()) {
		t.Fatalf("expected redacted evidence, got raw secret: %#v", finding.Evidence)
	}
}

func TestAESEC001IgnoresIgnoredCursorRulesFileOutsideInventory(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "testdata", "repos", "pass-secrets-agent")
	root, err := makeTempGitRepoFromFixture(t, fixture)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	writeFile(t, root, ".gitignore", ".cursor/rules/local.mdc\n")
	writeFile(t, root, ".cursor/rules/team.mdc", strings.Join([]string{
		"---",
		"description: safe tracked rule",
		"OPENAI_API_KEY=${OPENAI_API_KEY}",
	}, "\n")+"\n")
	stageAndCommitAll(t, root, "fixture-update")

	writeFile(t, root, ".cursor/rules/local.mdc", strings.Join([]string{
		"---",
		"description: ignored local rule",
		"OPENAI_API_KEY=" + fakeOpenAIKey(),
	}, "\n")+"\n")

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings, err := RunSecretRules(root, inv)
	if err != nil {
		t.Fatalf("run secret rules: %v", err)
	}
	for _, finding := range findings {
		if finding.RuleID == "AE-SEC-001" {
			t.Fatalf("expected ignored local .cursor/rules file to stay out of inventory, got %#v", finding)
		}
	}
}

func TestRunSecretRulesEmptyInventoryIsClean(t *testing.T) {
	root := t.TempDir()

	findings, err := RunSecretRules(root, repository.Inventory{})
	if err != nil {
		t.Fatalf("expected no error on empty inventory, got %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings on empty inventory, got %#v", findings)
	}
}

func findAESEC001(t *testing.T, allFindings []findings.Finding) findings.Finding {
	t.Helper()

	for _, finding := range allFindings {
		if finding.RuleID == "AE-SEC-001" {
			if len(finding.Evidence) == 0 {
				t.Fatalf("expected evidence for AE-SEC-001")
			}
			if len(finding.Locations) == 0 || finding.Locations[0].Line < 1 {
				t.Fatalf("expected a 1-based location for AE-SEC-001, got %#v", finding.Locations)
			}
			return finding
		}
	}

	t.Fatalf("expected AE-SEC-001 finding")
	return findings.Finding{}
}
