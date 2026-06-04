package context

import (
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
)

func checkCTX002(root string, inv repository.Inventory) (findings.Finding, bool) {
	if !inv.Has("AGENTS.md") {
		return findings.Finding{}, false
	}

	content, ok := repository.ReadTrackedFile(root, inv, "AGENTS.md")
	if !ok {
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
	lower := strings.ToLower(content)

	if !hasVerificationSignal(lower) {
		missing = append(missing, "verification command")
	}

	requiredMarkers := []string{".env*", "secrets/"}
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
