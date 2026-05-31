package agentconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/agentcontext"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
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

// checkEditScope returns an AE-CC-002 finding when an agent context source
// exists but none declares concrete off-limits paths. Absence of any context
// source is owned by AE-CTX-001 (Blocker) and is not duplicated here. An
// unreadable tracked context file fails fast with a wrapped error.
func checkEditScope(root string, inv repository.Inventory) ([]findings.Finding, error) {
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
			return nil, fmt.Errorf("read %s: %w", rel, err)
		}
		if declaresOffLimits(string(data)) {
			declared = true
		}
	}

	if cursorRules := cursorRuleFiles(inv); len(cursorRules) > 0 {
		present = append(present, agentcontext.CursorRulesDir)
		for _, rel := range cursorRules {
			// #nosec G304 -- rel is a tracked path under the .cursor/rules directory.
			data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", rel, err)
			}
			if declaresOffLimits(string(data)) {
				declared = true
			}
		}
	}

	if len(present) == 0 || declared {
		return nil, nil
	}

	return []findings.Finding{{
		RuleID:      "AE-CC-002",
		Severity:    findings.SeverityHigh,
		Category:    "Agent Config",
		Summary:     "Agent context declares no concrete off-limits paths (OWASP MCP02 Privilege Escalation via Scope Creep)",
		Remediation: "Add an 'Off-limits for agents' section listing at minimum .github/workflows/, terraform/ or infra/, db/migrations/, .env*, and secrets/ (or reference PERMISSIONS.md).",
		Evidence:    []string{"checked context sources: " + strings.Join(present, ", ")},
		Locations:   []findings.Location{{Path: present[0]}},
	}}, nil
}

// cursorRuleFiles returns the tracked files under the .cursor/rules directory,
// sorted. It enumerates from the inventory (tracked, non-ignored) rather than
// the filesystem so scanning stays tracked-only and deterministic — distinct
// from AE-CTX-001, which reads the same directory via os.ReadDir for its
// present-but-weak content check.
func cursorRuleFiles(inv repository.Inventory) []string {
	var out []string
	prefix := agentcontext.CursorRulesDir + "/"
	for _, p := range inv.Paths {
		if strings.HasPrefix(p, prefix) {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}
