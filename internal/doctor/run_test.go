package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func makeTempGitRepoFromFixture(t *testing.T, fixtureRoot string) (string, error) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	if err := copyDir(fixtureRoot, dir); err != nil {
		return "", err
	}

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Charter Test"},
		{"git", "config", "user.email", "charter@example.com"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "add", "."},
		{"git", "commit", "-m", "fixture"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("%s failed: %w\n%s", args[0], err, out)
		}
	}

	return dir, nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}

		out, err := os.Create(target)
		if err != nil {
			_ = in.Close()
			return err
		}

		if _, err := io.Copy(out, in); err != nil {
			_ = in.Close()
			_ = out.Close()
			return err
		}

		if err := in.Close(); err != nil {
			_ = out.Close()
			return err
		}

		if err := out.Close(); err != nil {
			return err
		}

		return os.Chmod(target, info.Mode())
	})
}

func mcpFindingIDs(result Result) []string {
	var ids []string
	for _, f := range result.Findings {
		if strings.HasPrefix(f.RuleID, "AE-MCP") {
			ids = append(ids, f.RuleID)
		}
	}
	return ids
}

func TestRunAgainstFixtureRepo(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "pass-slice1")
	repo, err := makeTempGitRepoFromFixture(t, root)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("expected doctor run to succeed: %v", err)
	}

	if result.Score.Final != 100 {
		t.Fatalf("expected passing score 100 for the fixture and rule set, got %d", result.Score.Final)
	}
}

func TestRunSetsThresholdAndPassed(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "repos", "pass-slice1")
	repo, err := makeTempGitRepoFromFixture(t, root)
	if err != nil {
		t.Fatalf("fixture repo setup failed: %v", err)
	}

	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("expected doctor run to succeed: %v", err)
	}

	if result.Threshold != 80 {
		t.Fatalf("expected threshold 80, got %d", result.Threshold)
	}

	if !result.Passed {
		t.Fatalf("expected run to pass")
	}
}

func TestRunMCPCleanFixtureNoMCPFindings(t *testing.T) {
	repo, err := makeTempGitRepoFromFixture(t, filepath.Join("..", "..", "testdata", "repos", "pass-mcp-clean"))
	if err != nil {
		t.Fatalf("fixture setup: %v", err)
	}
	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if ids := mcpFindingIDs(result); len(ids) != 0 {
		t.Fatalf("unexpected MCP findings: %v", ids)
	}
}

func TestRunMCPUnpinnedFixtureFlagsAEMCP001(t *testing.T) {
	repo, err := makeTempGitRepoFromFixture(t, filepath.Join("..", "..", "testdata", "repos", "fail-mcp-unpinned"))
	if err != nil {
		t.Fatalf("fixture setup: %v", err)
	}
	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if ids := mcpFindingIDs(result); len(ids) != 1 || ids[0] != "AE-MCP-001" {
		t.Fatalf("expected exactly [AE-MCP-001], got %v", ids)
	}
}

func TestRunMCPUntrustedRemoteFixtureFlagsAEMCP002(t *testing.T) {
	repo, err := makeTempGitRepoFromFixture(t, filepath.Join("..", "..", "testdata", "repos", "fail-mcp-untrusted-remote"))
	if err != nil {
		t.Fatalf("fixture setup: %v", err)
	}
	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if ids := mcpFindingIDs(result); len(ids) != 1 || ids[0] != "AE-MCP-002" {
		t.Fatalf("expected exactly [AE-MCP-002], got %v", ids)
	}
}

func TestRunMCPNoAuthFixtureFlagsAEMCP003(t *testing.T) {
	repo, err := makeTempGitRepoFromFixture(t, filepath.Join("..", "..", "testdata", "repos", "fail-mcp-noauth"))
	if err != nil {
		t.Fatalf("fixture setup: %v", err)
	}
	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if ids := mcpFindingIDs(result); len(ids) != 1 || ids[0] != "AE-MCP-003" {
		t.Fatalf("expected exactly [AE-MCP-003], got %v", ids)
	}
}

func ccFindingIDs(result Result) []string {
	var ids []string
	for _, f := range result.Findings {
		if strings.HasPrefix(f.RuleID, "AE-CC") {
			ids = append(ids, f.RuleID)
		}
	}
	return ids
}

func TestRunCCCleanFixtureNoCCFindings(t *testing.T) {
	repo, err := makeTempGitRepoFromFixture(t, filepath.Join("..", "..", "testdata", "repos", "pass-cc-clean"))
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if ids := ccFindingIDs(result); len(ids) != 0 {
		t.Fatalf("expected no AE-CC findings, got %v", ids)
	}
}

func TestRunCCDangerousHookFixture(t *testing.T) {
	repo, err := makeTempGitRepoFromFixture(t, filepath.Join("..", "..", "testdata", "repos", "fail-cc-dangerous-hook"))
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if ids := ccFindingIDs(result); len(ids) != 1 || ids[0] != "AE-CC-001" {
		t.Fatalf("expected AE-CC-001, got %v", ids)
	}
}

func TestRunCCNoScopeFixtureIsolatesAECC002(t *testing.T) {
	repo, err := makeTempGitRepoFromFixture(t, filepath.Join("..", "..", "testdata", "repos", "fail-cc-no-scope"))
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	result, err := Run(repo, 80)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	cc := ccFindingIDs(result)
	if len(cc) != 1 || cc[0] != "AE-CC-002" {
		t.Fatalf("expected exactly [AE-CC-002], got %v", cc)
	}
	for _, f := range result.Findings {
		if f.RuleID == "AE-CTX-001" {
			t.Fatalf("fixture should not trip AE-CTX-001 (it must isolate AE-CC-002); findings: %+v", result.Findings)
		}
	}
}
