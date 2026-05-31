package agentconfig

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.charter.dev/charter/internal/agentcontext"
	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
)

// offLimitsTokens are concrete sensitive-path signals that count as an explicit
// edit-scope declaration. A reference to PERMISSIONS.md also counts.
var offLimitsTokens = []string{
	".env", "secrets", ".github/workflows", "terraform", "infra", "db/migrations", "credentials", "permissions.md",
}

func declaresOffLimits(text string) bool {
	lower := strings.ToLower(text)
	for _, tok := range offLimitsTokens {
		if strings.Contains(lower, tok) {
			return true
		}
	}
	return false
}

// checkEditScope returns an AE-CC-002 finding when no agent context file
// declares concrete off-limits paths. Absence of any context file is owned by
// AE-CTX-001 (Blocker) and is not duplicated here.
func checkEditScope(root string, inv repository.Inventory) []findings.Finding {
	var present []string
	declared := false

	candidates := append(append([]string{}, agentcontext.Files...), "PERMISSIONS.md")
	sort.Strings(candidates)
	for _, rel := range candidates {
		if !inv.Has(rel) {
			continue
		}
		present = append(present, rel)
		// #nosec G304 -- rel is a fixed context/permissions path from the tracked inventory.
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		if declaresOffLimits(string(data)) {
			declared = true
		}
	}

	if len(present) == 0 || declared {
		return nil
	}

	return []findings.Finding{{
		RuleID:      "AE-CC-002",
		Severity:    findings.SeverityHigh,
		Category:    "Agent Config",
		Summary:     "Agent context declares no concrete off-limits paths (OWASP MCP02 Privilege Escalation via Scope Creep)",
		Remediation: "Add an 'Off-limits for agents' section listing at minimum .github/workflows/, terraform/ or infra/, db/migrations/, .env*, and secrets/ (or reference PERMISSIONS.md).",
		Evidence:    []string{"checked context files: " + strings.Join(present, ", ")},
		Locations:   []findings.Location{{Path: present[0]}},
	}}
}
