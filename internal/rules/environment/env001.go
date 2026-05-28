package environment

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
)

func Run(root string, inv repository.Inventory) []findings.Finding {
	missing := requiredEnvironmentFiles(root, inv)
	if len(missing) == 0 {
		return nil
	}

	sort.Strings(missing)
	return []findings.Finding{{
		RuleID:      "AE-ENV-001",
		Severity:    findings.SeverityMedium,
		Category:    "Environment",
		Summary:     "Repository reproducibility surface is incomplete",
		Remediation: "Commit the missing toolchain, lockfile, or hook config required by the active repo stack.",
		Evidence:    missing,
	}}
}

func requiredEnvironmentFiles(root string, inv repository.Inventory) []string {
	var missing []string

	for _, language := range activeLanguages(inv) {
		if !hasToolchainSignal(language, inv) {
			missing = append(missing, "missing toolchain signal for active language: "+language)
		}
		if requiresLockfile(language, root, inv) && !hasLockfileSignal(language, inv) {
			missing = append(missing, "missing lockfile signal for active language: "+language)
		}
	}

	if !hasHookSignal(root, inv) {
		missing = append(missing, "missing committed hook configuration")
	}

	return missing
}

func activeLanguages(inv repository.Inventory) []string {
	languages := map[string]bool{}
	for _, path := range inv.Paths {
		switch path {
		case "go.mod":
			languages["go"] = true
		case "package.json", "bunfig.toml", ".nvmrc", ".node-version":
			languages["javascript"] = true
		case "pyproject.toml", ".python-version", "uv.toml":
			languages["python"] = true
		case "rust-toolchain.toml", "rust-toolchain", "Cargo.toml":
			languages["rust"] = true
		case ".swift-version", "Package.swift":
			languages["swift"] = true
		case ".ruby-version", "Gemfile":
			languages["ruby"] = true
		case "gradle/wrapper/gradle-wrapper.properties", "build.gradle.kts", "build.gradle":
			languages["jvm"] = true
		}
	}

	var out []string
	for language := range languages {
		out = append(out, language)
	}
	sort.Strings(out)
	return out
}

func hasToolchainSignal(language string, inv repository.Inventory) bool {
	if inv.Has("mise.toml") || inv.Has(".mise.toml") || inv.Has(".tool-versions") || inv.Has("devcontainer.json") || inv.Has("flake.nix") {
		return true
	}

	switch language {
	case "go":
		return inv.Has("go.mod") || inv.Has(".go-version")
	case "javascript":
		return inv.Has(".nvmrc") || inv.Has(".node-version") || inv.Has("bunfig.toml") || inv.Has("package.json")
	case "python":
		return inv.Has("pyproject.toml") || inv.Has(".python-version") || inv.Has("uv.toml")
	case "rust":
		return inv.Has("rust-toolchain.toml") || inv.Has("rust-toolchain")
	case "swift":
		return inv.Has(".swift-version") || inv.Has("Package.swift")
	case "ruby":
		return inv.Has(".ruby-version") || inv.Has("Gemfile")
	case "jvm":
		return inv.Has("gradle/wrapper/gradle-wrapper.properties") || inv.Has(".java-version") || inv.Has("build.gradle.kts") || inv.Has("build.gradle")
	default:
		return false
	}
}

func requiresLockfile(language string, root string, inv repository.Inventory) bool {
	switch language {
	case "go":
		if !inv.Has("go.mod") {
			return false
		}
		// #nosec G304 -- go.mod is read from the resolved repository root and the file name is fixed.
		data, err := os.ReadFile(filepath.Join(root, "go.mod"))
		if err != nil {
			return true
		}
		text := string(data)
		return strings.Contains(text, "require ") || strings.Contains(text, "replace ") || strings.Contains(text, "exclude ")
	case "javascript":
		return inv.Has("package.json")
	case "python":
		return inv.Has("pyproject.toml")
	case "rust":
		return inv.Has("Cargo.toml")
	case "swift":
		return inv.Has("Package.swift")
	case "ruby":
		return inv.Has("Gemfile")
	case "jvm":
		return inv.Has("build.gradle.kts") || inv.Has("build.gradle") || inv.Has("gradle/wrapper/gradle-wrapper.properties")
	default:
		return false
	}
}

func hasLockfileSignal(language string, inv repository.Inventory) bool {
	switch language {
	case "go":
		return inv.Has("go.sum")
	case "javascript":
		return inv.Has("bun.lock") || inv.Has("bun.lockb") || inv.Has("package-lock.json") || inv.Has("yarn.lock") || inv.Has("pnpm-lock.yaml")
	case "python":
		return inv.Has("uv.lock") || inv.Has("poetry.lock") || inv.Has("requirements.txt")
	case "rust":
		return inv.Has("Cargo.lock")
	case "swift":
		return inv.Has("Package.resolved")
	case "ruby":
		return inv.Has("Gemfile.lock")
	case "jvm":
		return inv.Has("gradle/verification-metadata.xml") || inv.Has("gradle/libs.versions.toml")
	default:
		return false
	}
}

func hasHookSignal(root string, inv repository.Inventory) bool {
	if inv.Has("hk.pkl") || inv.Has("lefthook.yml") || inv.Has("lefthook.toml") || inv.Has(".pre-commit-config.yaml") || inv.Has(".overcommit.yml") {
		return true
	}
	for _, path := range inv.Paths {
		if strings.HasPrefix(path, ".husky/") || strings.HasPrefix(path, ".cargo-husky/hooks/") {
			return true
		}
	}
	if !inv.Has("package.json") {
		return false
	}
	// #nosec G304 -- package.json is read from the resolved repository root and the file name is fixed.
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return false
	}
	text := string(data)
	return strings.Contains(text, "simple-git-hooks") || strings.Contains(text, "lint-staged")
}
