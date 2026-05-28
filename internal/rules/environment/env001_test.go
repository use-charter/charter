package environment

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.charter.dev/charter/internal/repository"
)

func TestRunPassesWhenReproducibilityFilesExist(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"mise.toml":    "[tools]\ngo = \"1.26.3\"\n",
		"mise.lock":    "lock-placeholder\n",
		"go.mod":       "module example.com/passenv\n\ngo 1.26.0\n",
		"go.sum":       "example.com v0.0.0 h1:abc\n",
		"hk.pkl":       "hooks {}\n",
		"package.json": "{\n  \"name\": \"pass-env\",\n  \"private\": true,\n  \"volta\": { \"bun\": \"1.3.14\" }\n}\n",
		"bun.lock":     "lock-placeholder\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunPassesWithDevcontainerUniversalSignal(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"devcontainer.json": "{\n  \"image\": \"mcr.microsoft.com/devcontainers/base:ubuntu-24.04\"\n}\n",
		"package.json":      "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
		"package-lock.json": "{}\n",
		"pyproject.toml":    "[project]\nname = \"pass-env\"\n",
		"uv.lock":           "version = 1\n",
		"lefthook.yml":      "pre-commit:\n  commands: {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunPassesWithFlakeUniversalSignal(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"flake.nix":    "{ description = \"pass-env\"; }\n",
		"flake.lock":   "{\"version\":7}\n",
		"Cargo.toml":   "[package]\nname = \"pass-env\"\nversion = \"0.1.0\"\n",
		"Cargo.lock":   "# lock\n",
		"lefthook.yml": "pre-commit:\n  commands: {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunFindsManifestWithoutRuntimePin(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"package.json":      "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
		"package-lock.json": "{}\n",
		"hk.pkl":            "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if findings[0].RuleID != "AE-ENV-001" {
		t.Fatalf("expected AE-ENV-001, got %#v", findings[0])
	}
	if findings[0].Evidence[0] != "missing toolchain signal for active language: javascript" {
		t.Fatalf("expected missing javascript toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunPassesWithoutGoSumWhenGoHasNoDependencyState(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"go.mod":    "module example.com/passenv\n\ngo 1.26.0\n",
		"mise.toml": "[tools]\ngo = \"1.26.3\"\n",
		"hk.pkl":    "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunPassesWithPinnedVersionFiles(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".nvmrc":          "20.18.0\n",
		".python-version": "3.12.4\n",
		"rust-toolchain":  "1.78.0\n",
		".swift-version":  "6.0.3\n",
		".ruby-version":   "3.3.4\n",
		"hk.pkl":          "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunFindsFloatingVersionFiles(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".nvmrc":              "lts/*\n",
		".python-version":     "latest\n",
		"rust-toolchain":      "stable\n",
		".swift-version":      "latest\n",
		".ruby-version":       "stable\n",
		"rust-toolchain.toml": "[toolchain]\nchannel = \"stable\"\n",
		"hk.pkl":              "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: javascript") {
		t.Fatalf("expected missing javascript toolchain evidence, got %#v", findings[0].Evidence)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: python") {
		t.Fatalf("expected missing python toolchain evidence, got %#v", findings[0].Evidence)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: rust") {
		t.Fatalf("expected missing rust toolchain evidence, got %#v", findings[0].Evidence)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: swift") {
		t.Fatalf("expected missing swift toolchain evidence, got %#v", findings[0].Evidence)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: ruby") {
		t.Fatalf("expected missing ruby toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsMissingLockfile(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"package.json": "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
		"bunfig.toml":  "[install]\ncache = true\n",
		"hk.pkl":       "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if findings[0].RuleID != "AE-ENV-001" {
		t.Fatalf("expected AE-ENV-001, got %#v", findings[0])
	}
	if findings[0].Evidence[0] != "missing lockfile signal for active language: javascript" {
		t.Fatalf("expected missing bun.lock evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunTreatsRequirementsTxtAsPythonActivitySignal(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"requirements.txt": "requests==2.32.3\n",
		"hk.pkl":           "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: python") {
		t.Fatalf("expected missing python toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsUnpinnedRequirementsTxtLockfile(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".python-version":  "3.12.4\n",
		"requirements.txt": "requests>=2.0\n",
		"hk.pkl":           "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing lockfile signal for active language: python") {
		t.Fatalf("expected missing python lockfile evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunPassesWithPinnedRequirementsTxtLockfile(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".python-version":  "3.12.4\n",
		"requirements.txt": "requests==2.32.3 # pinned\nurllib3===2.2.2\n",
		"hk.pkl":           "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunPassesWithPinnedRequirementsTxtHashesAndMarkers(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".python-version":  "3.12.4\n",
		"requirements.txt": "--require-hashes\nrequests==2.32.3 --hash=sha256:abc123\nurllib3==2.2.2 ; python_version < \"3.13\"\n",
		"hk.pkl":           "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	if findings := Run(root, inv); len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRunFindsUnsafeRequirementsTxtDirectives(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{name: "editable directive", line: "--editable ."},
		{name: "constraint directive", line: "--constraint constraints.txt"},
		{name: "extra index directive", line: "--extra-index-url https://example.com/simple"},
		{name: "arbitrary directive", line: "--trusted-host example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := newEnvironmentRepo(t, map[string]string{
				".python-version":  "3.12.4\n",
				"requirements.txt": tc.line + "\nrequests==2.32.3\n",
				"hk.pkl":           "hooks {}\n",
			})

			inv, err := repository.BuildInventory(root)
			if err != nil {
				t.Fatalf("inventory failed: %v", err)
			}

			findings := Run(root, inv)
			if len(findings) != 1 {
				t.Fatalf("expected one finding, got %#v", findings)
			}
			if !containsEvidence(findings[0].Evidence, "missing lockfile signal for active language: python") {
				t.Fatalf("expected missing python lockfile evidence, got %#v", findings[0].Evidence)
			}
		})
	}
}

func TestRunFindsMalformedRequirementsTxtPin(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".python-version":  "3.12.4\n",
		"requirements.txt": "requests==\n",
		"hk.pkl":           "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing lockfile signal for active language: python") {
		t.Fatalf("expected missing python lockfile evidence, got %#v", findings[0].Evidence)
	}
}

func TestIsPinnedRequirementSpec(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{name: "accepts double equals pin", line: "requests==2.32.3", want: true},
		{name: "accepts triple equals pin", line: "urllib3===2.2.2", want: true},
		{name: "accepts hash option after pin", line: "requests==2.32.3 --hash=sha256:abc123", want: true},
		{name: "accepts marker after pin", line: "requests==2.32.3 ; python_version < \"3.13\"", want: true},
		{name: "ignores require hashes directive", line: "--require-hashes", want: true},
		{name: "rejects editable directive", line: "--editable .", want: false},
		{name: "rejects constraint directive", line: "--constraint constraints.txt", want: false},
		{name: "rejects extra index directive", line: "--extra-index-url https://example.com/simple", want: false},
		{name: "rejects arbitrary directive", line: "--trusted-host example.com", want: false},
		{name: "rejects latest", line: "requests==latest", want: false},
		{name: "rejects wildcard", line: "requests===*", want: false},
		{name: "rejects exclusion spec", line: "requests==!=2.0", want: false},
		{name: "rejects floating spec with marker", line: "requests>=2.0 ; python_version < \"3.13\"", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isPinnedRequirementSpec(tc.line); got != tc.want {
				t.Fatalf("isPinnedRequirementSpec(%q) = %t, want %t", tc.line, got, tc.want)
			}
		})
	}
}

func TestRunIgnoresInlineCommentMarkersInsideRequirementsComments(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".python-version":  "3.12.4\n",
		"requirements.txt": "requests>=2 # == note\n",
		"hk.pkl":           "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing lockfile signal for active language: python") {
		t.Fatalf("expected missing python lockfile evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsBunfigWithoutRuntimePin(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"package.json": "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
		"bunfig.toml":  "[install]\ncache = true\n",
		"bun.lock":     "lock-placeholder\n",
		"hk.pkl":       "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: javascript") {
		t.Fatalf("expected missing javascript toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsFloatingPackageJSONRuntimePins(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"package.json": "{\n  \"name\": \"pass-env\",\n  \"private\": true,\n  \"engines\": { \"node\": \">=20\" },\n  \"volta\": { \"bun\": \"^1.3.14\" }\n}\n",
		"bun.lock":     "lock-placeholder\n",
		"hk.pkl":       "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: javascript") {
		t.Fatalf("expected missing javascript toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsDevcontainerWithoutReproducibleToolchainPath(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"devcontainer.json": "{\n  \"name\": \"dev-only\"\n}\n",
		"package.json":      "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
		"package-lock.json": "{}\n",
		"hk.pkl":            "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: javascript") {
		t.Fatalf("expected missing javascript toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsFlakeWithoutLockfile(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"flake.nix":    "{ description = \"pass-env\"; }\n",
		"Cargo.toml":   "[package]\nname = \"pass-env\"\nversion = \"0.1.0\"\n",
		"Cargo.lock":   "# lock\n",
		"lefthook.yml": "pre-commit:\n  commands: {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: rust") {
		t.Fatalf("expected missing rust toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunRequiresMiseLockWhenMiseProvidesToolchain(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"mise.toml":    "[tools]\nbun = \"1.3.14\"\n",
		"package.json": "{\n  \"name\": \"pass-env\",\n  \"private\": true\n}\n",
		"bun.lock":     "lock-placeholder\n",
		"hk.pkl":       "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing mise.lock for mise toolchain declaration") {
		t.Fatalf("expected missing mise.lock evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsFloatingJavaVersionFile(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		".java-version":             "latest\n",
		"build.gradle.kts":          "java { toolchain { languageVersion = provider { JavaLanguageVersion.current() } } }\n",
		"gradle/libs.versions.toml": "[versions]\njava = \"21\"\n",
		"hk.pkl":                    "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: jvm") {
		t.Fatalf("expected missing jvm toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsFloatingGemfileRubyDeclaration(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"Gemfile":      "source \"https://rubygems.org\"\nruby \">= 3.3\"\n",
		"Gemfile.lock": "GEM\n",
		"hk.pkl":       "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: ruby") {
		t.Fatalf("expected missing ruby toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsGenericGradleToolchainTextWithoutPinnedVersion(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"build.gradle.kts":          "java { toolchain { languageVersion = provider { JavaLanguageVersion.current() } } }\n",
		"gradle/libs.versions.toml": "[versions]\njava = \"21\"\n",
		"hk.pkl":                    "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: jvm") {
		t.Fatalf("expected missing jvm toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func TestRunFindsFloatingGradleDistributionURL(t *testing.T) {
	root := newEnvironmentRepo(t, map[string]string{
		"gradle/wrapper/gradle-wrapper.properties": "distributionUrl=https\\://services.gradle.org/distributions/gradle-latest-bin.zip\n",
		"gradle/libs.versions.toml":                "[versions]\njava = \"21\"\n",
		"hk.pkl":                                   "hooks {}\n",
	})

	inv, err := repository.BuildInventory(root)
	if err != nil {
		t.Fatalf("inventory failed: %v", err)
	}

	findings := Run(root, inv)
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if !containsEvidence(findings[0].Evidence, "missing toolchain signal for active language: jvm") {
		t.Fatalf("expected missing jvm toolchain evidence, got %#v", findings[0].Evidence)
	}
}

func newEnvironmentRepo(t *testing.T, files map[string]string) string {
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
