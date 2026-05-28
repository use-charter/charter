package environment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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
	usesMiseToolchain := false

	for _, language := range activeLanguages(inv) {
		source := toolchainSignalSource(root, language, inv)
		if source == "" {
			missing = append(missing, "missing toolchain signal for active language: "+language)
		}
		if source == "mise" {
			usesMiseToolchain = true
		}
		if requiresLockfile(language, root, inv) && !hasLockfileSignal(language, root, inv) {
			missing = append(missing, "missing lockfile signal for active language: "+language)
		}
	}

	if usesMiseToolchain && !inv.Has("mise.lock") {
		missing = append(missing, "missing mise.lock for mise toolchain declaration")
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
		case "requirements.txt":
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

func toolchainSignalSource(root string, language string, inv repository.Inventory) string {
	switch language {
	case "go":
		if hasPinnedVersionFile(root, inv, ".go-version") {
			return ".go-version"
		}
		content, ok := readRepoFile(root, inv, "go.mod")
		if ok && (strings.Contains(content, "\ntoolchain go") || strings.Contains(content, "\ngo ")) {
			return "go.mod"
		}
	case "javascript":
		if hasPinnedVersionFile(root, inv, ".nvmrc") {
			return ".nvmrc"
		}
		if hasPinnedVersionFile(root, inv, ".node-version") {
			return ".node-version"
		}
		content, ok := readRepoFile(root, inv, "package.json")
		if ok && hasJavaScriptRuntimePin(content) {
			return "package.json"
		}
	case "python":
		if hasPinnedVersionFile(root, inv, ".python-version") {
			return ".python-version"
		}
		content, ok := readRepoFile(root, inv, "pyproject.toml")
		if ok && strings.Contains(strings.ToLower(content), "requires-python") {
			return "pyproject.toml"
		}
	case "rust":
		if content, ok := readRepoFile(root, inv, "rust-toolchain.toml"); ok && hasPinnedRustToolchainTOML(content) {
			return "rust-toolchain.toml"
		}
		if hasPinnedVersionFile(root, inv, "rust-toolchain") {
			return "rust-toolchain"
		}
	case "swift":
		if hasPinnedVersionFile(root, inv, ".swift-version") {
			return ".swift-version"
		}
		content, ok := readRepoFile(root, inv, "Package.swift")
		if ok && strings.Contains(strings.ToLower(content), "swift-tools-version") {
			return "Package.swift"
		}
	case "ruby":
		if hasPinnedVersionFile(root, inv, ".ruby-version") {
			return ".ruby-version"
		}
		content, ok := readRepoFile(root, inv, "Gemfile")
		if ok && hasPinnedGemfileRubyVersion(content) {
			return "Gemfile"
		}
	case "jvm":
		if hasPinnedVersionFile(root, inv, ".java-version") {
			return ".java-version"
		}
		if content, ok := readRepoFile(root, inv, "gradle/wrapper/gradle-wrapper.properties"); ok && hasPinnedGradleDistributionURL(content) {
			return "gradle/wrapper/gradle-wrapper.properties"
		}
		if content, ok := readRepoFile(root, inv, "build.gradle.kts"); ok && hasPinnedGradleToolchainSignal(content) {
			return "build.gradle.kts"
		}
		if content, ok := readRepoFile(root, inv, "build.gradle"); ok && hasPinnedGradleToolchainSignal(content) {
			return "build.gradle"
		}
	}

	if hasToolVersionsPinForLanguage(root, language, inv) {
		return ".tool-versions"
	}
	if hasDevcontainerSignal(root, inv) {
		return "devcontainer"
	}
	if hasFlakeSignal(root, inv) {
		return "flake.nix"
	}
	if hasMiseSignalForLanguage(root, language, inv) {
		return "mise"
	}

	return ""
}

var rubyVersionPattern = regexp.MustCompile(`(?m)^\s*ruby\s+["']([^"']+)["']`)

func hasMiseSignalForLanguage(root string, language string, inv repository.Inventory) bool {
	for _, path := range []string{"mise.toml", ".mise.toml"} {
		if content, ok := readRepoFile(root, inv, path); ok && hasMiseRuntimePin(content, language) {
			return true
		}
	}
	return false
}

func hasToolVersionsPinForLanguage(root string, language string, inv repository.Inventory) bool {
	content, ok := readRepoFile(root, inv, ".tool-versions")
	return ok && hasToolVersionsPin(content, language)
}

func hasDevcontainerSignal(root string, inv repository.Inventory) bool {
	for _, configPath := range []string{"devcontainer.json", ".devcontainer/devcontainer.json"} {
		content, ok := readRepoFile(root, inv, configPath)
		if !ok {
			continue
		}
		if devcontainerProvidesToolchainPath(root, inv, configPath, content) {
			return true
		}
	}
	return false
}

func hasFlakeSignal(root string, inv repository.Inventory) bool {
	if !inv.Has("flake.nix") || !inv.Has("flake.lock") {
		return false
	}
	content, ok := readRepoFile(root, inv, "flake.lock")
	return ok && strings.TrimSpace(content) != ""
}

func readRepoFile(root string, inv repository.Inventory, path string) (string, bool) {
	if !inv.Has(path) {
		return "", false
	}
	// #nosec G304 -- path is a fixed repo-relative path selected from known toolchain files.
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return "", false
	}
	return string(data), true
}

func hasMiseRuntimePin(content string, language string) bool {
	for _, key := range miseToolKeys(language) {
		if hasPinnedAssignment(content, key) {
			return true
		}
	}
	return false
}

func miseToolKeys(language string) []string {
	switch language {
	case "go":
		return []string{"go"}
	case "javascript":
		return []string{"node", "bun"}
	case "python":
		return []string{"python"}
	case "rust":
		return []string{"rust"}
	case "swift":
		return []string{"swift"}
	case "ruby":
		return []string{"ruby"}
	case "jvm":
		return []string{"java", "gradle"}
	default:
		return nil
	}
}

func hasPinnedAssignment(content string, key string) bool {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, "=") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		lhs := strings.TrimSpace(parts[0])
		rhs := strings.TrimSpace(parts[1])
		lhs = strings.Trim(lhs, `"`)
		if lhs != key {
			continue
		}
		if looksPinnedVersion(strings.Trim(rhs, `"'`)) {
			return true
		}
	}
	return false
}

func hasToolVersionsPin(content string, language string) bool {
	keys := toolVersionsKeys(language)
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		for _, key := range keys {
			if fields[0] == key && looksPinnedVersion(fields[1]) {
				return true
			}
		}
	}
	return false
}

func toolVersionsKeys(language string) []string {
	switch language {
	case "go":
		return []string{"golang", "go"}
	case "javascript":
		return []string{"nodejs", "node", "bun"}
	case "python":
		return []string{"python"}
	case "rust":
		return []string{"rust"}
	case "swift":
		return []string{"swift"}
	case "ruby":
		return []string{"ruby"}
	case "jvm":
		return []string{"java", "gradle"}
	default:
		return nil
	}
}

func hasJavaScriptRuntimePin(content string) bool {
	var manifest struct {
		Engines        map[string]string `json:"engines"`
		Volta          map[string]string `json:"volta"`
		PackageManager string            `json:"packageManager"`
	}
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return false
	}

	for _, key := range []string{"node", "bun"} {
		if looksPinnedVersion(manifest.Engines[key]) || looksPinnedVersion(manifest.Volta[key]) {
			return true
		}
	}

	if name, version, ok := strings.Cut(manifest.PackageManager, "@"); ok {
		if (name == "bun" || name == "node") && looksPinnedVersion(version) {
			return true
		}
	}

	return false
}

func hasPinnedVersionFile(root string, inv repository.Inventory, path string) bool {
	content, ok := readRepoFile(root, inv, path)
	if !ok {
		return false
	}

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return looksPinnedVersion(trimmed)
	}

	return false
}

var rustToolchainChannelPattern = regexp.MustCompile(`(?m)^\s*channel\s*=\s*["']([^"']+)["']`)

var pinnedRequirementsPattern = regexp.MustCompile(`^(.+?)(===|==)\s*(.+)$`)

func hasPinnedRustToolchainTOML(content string) bool {
	match := rustToolchainChannelPattern.FindStringSubmatch(content)
	return len(match) == 2 && looksPinnedVersion(match[1])
}

func hasPinnedGemfileRubyVersion(content string) bool {
	match := rubyVersionPattern.FindStringSubmatch(content)
	return len(match) == 2 && looksPinnedVersion(match[1])
}

var gradleJVMPinPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)jvmToolchain\s*\(\s*([0-9]+(?:\.[0-9]+)*)\s*\)`),
	regexp.MustCompile(`(?i)JavaLanguageVersion\.of\s*\(\s*([0-9]+(?:\.[0-9]+)*)\s*\)`),
	regexp.MustCompile(`(?i)JavaVersion\.VERSION_([0-9_]+)`),
}

func hasPinnedGradleToolchainSignal(content string) bool {
	for _, pattern := range gradleJVMPinPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			version := strings.ReplaceAll(match[1], "_", ".")
			if looksPinnedVersion(version) {
				return true
			}
		}
	}
	return false
}

var (
	gradleDistributionPattern        = regexp.MustCompile(`(?m)^\s*distributionUrl\s*=\s*([^\s#]+)\s*$`)
	gradleDistributionVersionPattern = regexp.MustCompile(`(?i)gradle-([a-z0-9._-]+?)-(?:bin|all)\.zip`)
)

func hasPinnedGradleDistributionURL(content string) bool {
	match := gradleDistributionPattern.FindStringSubmatch(content)
	if len(match) != 2 {
		return false
	}
	versionMatch := gradleDistributionVersionPattern.FindStringSubmatch(match[1])
	if len(versionMatch) != 2 {
		return false
	}
	return looksPinnedVersion(versionMatch[1])
}

func looksPinnedVersion(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	lower := strings.ToLower(value)
	if strings.ContainsAny(lower, "*^~<>") || strings.Contains(lower, "latest") || strings.Contains(lower, "||") {
		return false
	}
	if strings.Contains(lower, ".x") || strings.HasSuffix(lower, "x") {
		return false
	}
	for _, token := range []string{"stable", "nightly", "beta", "canary", "current", "lts/"} {
		if strings.Contains(lower, token) {
			return false
		}
	}

	return strings.IndexFunc(value, func(r rune) bool { return r >= '0' && r <= '9' }) >= 0
}

func devcontainerProvidesToolchainPath(root string, inv repository.Inventory, configPath string, content string) bool {
	type buildConfig struct {
		Dockerfile string `json:"dockerfile"`
		DockerFile string `json:"dockerFile"`
	}
	type config struct {
		Image             string          `json:"image"`
		DockerFile        string          `json:"dockerFile"`
		Build             buildConfig     `json:"build"`
		DockerComposeFile json.RawMessage `json:"dockerComposeFile"`
	}

	var parsed config
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return false
	}

	if devcontainerImagePinned(parsed.Image) {
		return true
	}

	for _, candidate := range []string{parsed.DockerFile, parsed.Build.DockerFile, parsed.Build.Dockerfile} {
		if candidate != "" && devcontainerReferencedFileExists(configPath, candidate, inv) {
			return true
		}
	}

	for _, candidate := range devcontainerComposeFiles(parsed.DockerComposeFile) {
		if devcontainerReferencedFileExists(configPath, candidate, inv) {
			return true
		}
	}

	_ = root
	return false
}

func devcontainerComposeFiles(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil && single != "" {
		return []string{single}
	}

	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		return many
	}

	return nil
}

func devcontainerImagePinned(image string) bool {
	image = strings.TrimSpace(image)
	if image == "" {
		return false
	}
	if strings.Contains(image, "@sha256:") {
		return true
	}
	lastColon := strings.LastIndex(image, ":")
	lastSlash := strings.LastIndex(image, "/")
	if lastColon <= lastSlash {
		return false
	}
	tag := strings.ToLower(strings.TrimSpace(image[lastColon+1:]))
	return tag != "" && tag != "latest"
}

func devcontainerReferencedFileExists(configPath string, referenced string, inv repository.Inventory) bool {
	referenced = strings.TrimSpace(referenced)
	if referenced == "" {
		return false
	}

	base := filepath.Dir(filepath.FromSlash(configPath))
	resolved := filepath.Clean(filepath.Join(base, filepath.FromSlash(referenced)))
	resolved = filepath.ToSlash(resolved)
	if resolved == "." {
		return false
	}
	return inv.Has(resolved)
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
		return inv.Has("pyproject.toml") || inv.Has("requirements.txt")
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

func hasLockfileSignal(language string, root string, inv repository.Inventory) bool {
	switch language {
	case "go":
		return inv.Has("go.sum")
	case "javascript":
		return inv.Has("bun.lock") || inv.Has("bun.lockb") || inv.Has("package-lock.json") || inv.Has("yarn.lock") || inv.Has("pnpm-lock.yaml")
	case "python":
		return inv.Has("uv.lock") || inv.Has("poetry.lock") || hasPinnedRequirementsLockfile(inv, root)
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

func hasPinnedRequirementsLockfile(inv repository.Inventory, root string) bool {
	content, ok := readRepoFile(root, inv, "requirements.txt")
	if !ok {
		return false
	}

	hasRequirement := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(stripInlineComment(line))
		if trimmed == "" {
			continue
		}

		hasRequirement = true
		if !isPinnedRequirementSpec(trimmed) {
			return false
		}
	}

	return hasRequirement
}

func stripInlineComment(line string) string {
	before, _, _ := strings.Cut(line, "#")
	return before
}

func isPinnedRequirementSpec(line string) bool {
	if strings.HasPrefix(line, "--") {
		return true
	}

	match := pinnedRequirementsPattern.FindStringSubmatch(line)
	if len(match) != 4 {
		return false
	}

	name := strings.TrimSpace(match[1])
	version := normalizeRequirementVersion(match[3])
	return name != "" && !strings.ContainsAny(version, "=!") && looksPinnedVersion(version)
}

func normalizeRequirementVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return ""
	}

	return strings.Fields(version)[0]
}
