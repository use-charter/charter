package context

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

var requiredGitignorePatterns = []string{
	".charter/",
	"*.charter-session",
	".claude/local/",
	".cursor/cache/",
	".hk/",
	".env*",
}

func checkCTX004(root string, inv repository.Inventory) (findings.Finding, bool) {
	// #nosec G304 -- .gitignore is a fixed repo-relative contract path.
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return findings.Finding{
			RuleID:      "AE-CTX-004",
			Severity:    findings.SeverityMedium,
			Category:    "Context",
			Summary:     "Repository is missing the tracked .gitignore contract for local agent artifacts",
			Remediation: "Add a root .gitignore with explicit local agent, cache, and env ignore patterns.",
			Evidence:    []string{".gitignore"},
		}, true
	}

	var evidence []string
	patterns := parseGitignorePatterns(string(data))
	for _, pattern := range requiredGitignorePatterns {
		if !slices.Contains(patterns, pattern) {
			evidence = append(evidence, "missing ignore pattern: "+pattern)
		}
	}

	for _, path := range inv.Paths {
		if isTrackedLocalArtifact(path) {
			evidence = append(evidence, "tracked local artifact: "+path)
		}
	}

	if len(evidence) == 0 {
		return findings.Finding{}, false
	}

	sort.Strings(evidence)
	evidence = slices.Clip(evidence)
	return findings.Finding{
		RuleID:      "AE-CTX-004",
		Severity:    findings.SeverityMedium,
		Category:    "Context",
		Summary:     ".gitignore does not fully exclude local agent artifacts",
		Remediation: "Add the missing ignore patterns and stop tracking local agent or env artifacts.",
		Evidence:    evidence,
		Locations:   []findings.Location{{Path: ".gitignore"}},
	}, true
}

func parseGitignorePatterns(content string) []string {
	var patterns []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		patterns = append(patterns, trimmed)
	}
	return patterns
}

func isTrackedLocalArtifact(path string) bool {
	if strings.HasPrefix(path, ".charter/") || strings.HasPrefix(path, ".claude/local/") || strings.HasPrefix(path, ".cursor/cache/") || strings.HasPrefix(path, ".hk/") {
		return true
	}
	if strings.HasSuffix(path, ".charter-session") {
		return true
	}
	if strings.HasPrefix(path, ".env") && path != ".env.example" {
		return true
	}
	return false
}
