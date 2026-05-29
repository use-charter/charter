package secrets

import (
	"path/filepath"
	"strings"
	"testing"

	"go.charter.dev/charter/internal/repository"
)

func TestAESEC002IgnoresEnvReferencesInConfig(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "testdata", "repos", "pass-secrets-config")
	root, err := makeTempGitRepoFromFixture(t, fixture)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := RunSecretRules(root, inv)
	for _, finding := range findings {
		if finding.RuleID == "AE-SEC-002" {
			t.Fatalf("expected no AE-SEC-002 finding, got %#v", finding)
		}
	}
}

func TestAESEC002FindsLiteralSecretInConfig(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "testdata", "repos", "pass-secrets-config")
	root, err := makeTempGitRepoFromFixture(t, fixture)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	writeFile(t, root, ".mcp.json", strings.Join([]string{
		"{",
		"  \"servers\": {",
		"    \"docs\": {",
		"      \"env\": {",
		"        \"OPENAI_API_KEY\": \"" + fakeOpenAIKey() + "\"",
		"      }",
		"    }",
		"  }",
		"}",
	}, "\n")+"\n")
	stageAndCommitAll(t, root, "fixture-update")

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	for _, finding := range RunSecretRules(root, inv) {
		if finding.RuleID != "AE-SEC-002" {
			continue
		}

		if len(finding.Evidence) == 0 {
			t.Fatalf("expected evidence for AE-SEC-002")
		}
		if strings.Contains(finding.Evidence[0], fakeOpenAIKey()) {
			t.Fatalf("expected redacted evidence, got raw secret: %#v", finding.Evidence)
		}
		return
	}

	t.Fatalf("expected AE-SEC-002 finding")
}
