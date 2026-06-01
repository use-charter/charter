// Package scaffold is the pure, deterministic, offline engine behind
// `charter init`: it detects a repository's languages/CI/agent surfaces,
// builds agent-context file templates, and computes a create-or-skip file
// plan. It performs no disk writes and no network access — the command layer
// owns all I/O. Everything here is unit-testable without touching disk.
package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Lang is a detected language and its version. Version is best-effort and may
// be empty when it cannot be parsed trivially.
type Lang struct{ Name, Version string }

// Project is the deterministic summary of a repository used to populate
// scaffold templates.
type Project struct {
	Name   string   // base name of the repo dir
	Langs  []Lang   // detected languages, stable order
	CI     string   // "GitHub Actions" or ""
	Agents []string // subset of {"claude","cursor","copilot"} detected
}

// Detect inspects manifests directly under root and returns a best-effort
// Project. It is deterministic and offline: missing manifests are simply not
// detected and never produce an error. A non-nil error is reserved for a
// genuine I/O failure worth surfacing; detection today is best-effort and
// returns nil.
func Detect(root string) (Project, error) {
	p := Project{Name: filepath.Base(root)}

	// #nosec G304 -- fixed manifest filename joined to the scan root, not user-controlled.
	if data, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
		p.Langs = append(p.Langs, Lang{Name: "Go", Version: goModVersion(string(data))})
	}
	// #nosec G304 -- fixed manifest filename joined to the scan root, not user-controlled.
	if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		p.Langs = append(p.Langs, Lang{Name: "JavaScript/TypeScript", Version: packageJSONNodeVersion(data)})
	}
	// #nosec G304 -- fixed manifest filename joined to the scan root, not user-controlled.
	if data, err := os.ReadFile(filepath.Join(root, "pyproject.toml")); err == nil {
		p.Langs = append(p.Langs, Lang{Name: "Python", Version: pyProjectRequiresPython(string(data))})
	}
	if _, err := os.Stat(filepath.Join(root, "Cargo.toml")); err == nil {
		p.Langs = append(p.Langs, Lang{Name: "Rust", Version: ""})
	}

	p.CI = detectCI(root)
	p.Agents = detectAgents(root)

	return p, nil
}

// goModVersion returns the version declared by the `go` directive, falling
// back to the `toolchain` line (with its leading "go" stripped) when no `go`
// directive is present.
func goModVersion(content string) string {
	var goVersion, toolchain string
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "go":
			goVersion = fields[1]
		case "toolchain":
			toolchain = strings.TrimPrefix(fields[1], "go")
		}
	}
	if goVersion != "" {
		return goVersion
	}
	return toolchain
}

// packageJSONNodeVersion shallow-parses engines.node; anything unparseable or
// absent yields an empty version rather than an error.
func packageJSONNodeVersion(data []byte) string {
	var pkg struct {
		Engines struct {
			Node string `json:"node"`
		} `json:"engines"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}
	return pkg.Engines.Node
}

// pyProjectRequiresPython finds a trivially-formatted requires-python value
// without a full TOML parse.
func pyProjectRequiresPython(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "requires-python") {
			continue
		}
		_, value, found := strings.Cut(trimmed, "=")
		if !found {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), "\"'")
	}
	return ""
}

func detectCI(root string) string {
	entries, err := os.ReadDir(filepath.Join(root, ".github", "workflows"))
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if name := entry.Name(); strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			return "GitHub Actions"
		}
	}
	return ""
}

func detectAgents(root string) []string {
	var agents []string
	if isDir(filepath.Join(root, ".claude")) {
		agents = append(agents, "claude")
	}
	if isDir(filepath.Join(root, ".cursor")) {
		agents = append(agents, "cursor")
	}
	if isFile(filepath.Join(root, ".github", "copilot-instructions.md")) {
		agents = append(agents, "copilot")
	}
	return agents
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
