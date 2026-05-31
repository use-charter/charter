package context

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

func checkCTX002(root string, inv repository.Inventory) (findings.Finding, bool) {
	if !inv.Has("AGENTS.md") {
		return findings.Finding{}, false
	}

	// #nosec G304 -- AGENTS.md is a fixed repo-relative contract path.
	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		return findings.Finding{
			RuleID:      "AE-CTX-002",
			Severity:    findings.SeverityMedium,
			Category:    "Context",
			Summary:     "AGENTS.md could not be read to verify repo truth",
			Remediation: "Restore a readable AGENTS.md that matches current repo behavior.",
			Evidence:    []string{"AGENTS.md"},
			Locations:   []findings.Location{{Path: "AGENTS.md"}},
		}, true
	}

	content := string(data)
	missing := missingRepoTruthMarkers(content, inv)
	if len(missing) == 0 {
		return findings.Finding{}, false
	}

	sort.Strings(missing)
	return findings.Finding{
		RuleID:      "AE-CTX-002",
		Severity:    findings.SeverityMedium,
		Category:    "Context",
		Summary:     "AGENTS.md does not match current repository truth",
		Remediation: "Update AGENTS.md so its commands, hook references, and boundaries match the tracked repo state.",
		Evidence:    missing,
		Locations:   []findings.Location{{Path: "AGENTS.md"}},
	}, true
}

func missingRepoTruthMarkers(content string, inv repository.Inventory) []string {
	var missing []string

	requiredMarkers := []string{
		"moon run :check",
		".env*",
		"secrets/",
	}

	if inv.Has("hk.pkl") {
		requiredMarkers = append(requiredMarkers, "hk.pkl")
	}
	if inv.Has("docs/internal/architecture/charter-architecture-2026.md") {
		requiredMarkers = append(requiredMarkers, "docs/internal/architecture/charter-architecture-2026.md")
	}

	for _, marker := range requiredMarkers {
		if !strings.Contains(content, marker) {
			missing = append(missing, marker)
		}
	}

	return missing
}
