package ci

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.charter.dev/charter/internal/repository"
)

func TestRunPassesWhenWorkflowCoverageIsPresent(t *testing.T) {
	root := newCIRepo(t, map[string]string{
		".github/workflows/ci.yml":               "name: CI\njobs:\n  check:\n    steps:\n      - run: moon run :check\n      - uses: actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955\n",
		".github/workflows/actions-security.yml": "name: Workflow Security\njobs:\n  lint:\n    steps:\n      - run: moon run :actionlint\n      - run: moon run :zizmor\n      - uses: jdx/mise-action@1648a7812b9aeae629881980618f079932869151\n",
		".github/workflows/vuln-scan.yml":        "name: Vulnerability Scan\njobs:\n  security:\n    steps:\n      - run: moon run :security\n      - uses: actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunPassesWhenWorkflowCoverageUsesYAMLFiles(t *testing.T) {
	root := newCIRepo(t, map[string]string{
		".github/workflows/ci.yaml":               "name: CI\njobs:\n  check:\n    steps:\n      - run: moon run :check\n      - uses: actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955\n",
		".github/workflows/actions-security.yaml": "name: Workflow Security\njobs:\n  lint:\n    steps:\n      - run: moon run :actionlint\n      - run: moon run :zizmor\n      - uses: jdx/mise-action@1648a7812b9aeae629881980618f079932869151\n",
		".github/workflows/vuln-scan.yaml":        "name: Vulnerability Scan\njobs:\n  security:\n    steps:\n      - run: moon run :security\n      - uses: actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunFindsUnpinnedAction(t *testing.T) {
	root := newCIRepo(t, map[string]string{
		".github/workflows/ci.yml":               "name: CI\njobs:\n  check:\n    steps:\n      - run: moon run :check\n      - uses: actions/checkout@v4\n",
		".github/workflows/actions-security.yml": "name: Workflow Security\njobs:\n  lint:\n    steps:\n      - run: moon run :actionlint\n      - run: moon run :zizmor\n",
		".github/workflows/vuln-scan.yml":        "name: Vulnerability Scan\njobs:\n  security:\n    steps:\n      - run: moon run :security\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if findings[0].RuleID != "AE-CI-002" {
		t.Fatalf("expected AE-CI-002, got %#v", findings[0])
	}
	if findings[0].Evidence[0] != "unpinned action: .github/workflows/ci.yml -> actions/checkout@v4" {
		t.Fatalf("expected unpinned action evidence, got %#v", findings[0].Evidence)
	}
}

func newCIRepo(t *testing.T, files map[string]string) string {
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
