// Package catalog holds static, per-rule metadata (name, category, description,
// help URI) used to enrich SARIF tool.driver.rules[] and, later, charter explain.
// Severity is intentionally absent: a finding's severity is the single source of
// truth for its SARIF level. The catalog ID set is bound to the behavioral specs
// by a drift test (catalog_test.go).
package catalog

import "sort"

// Entry is the static metadata for one Charter rule.
type Entry struct {
	ID               string
	Name             string
	Category         string
	ShortDescription string
	HelpURI          string
}

func help(id string) string { return "https://use-charter.dev/rules/" + id }

var entries = map[string]Entry{
	"AE-CTX-001":      {ID: "AE-CTX-001", Name: "AgentContextFilePresent", Category: "Context", ShortDescription: "Agent context file is present and meaningful for agents.", HelpURI: help("AE-CTX-001")},
	"AE-CTX-002":      {ID: "AE-CTX-002", Name: "AgentContextConsistent", Category: "Context", ShortDescription: "Agent context matches the actual repository state.", HelpURI: help("AE-CTX-002")},
	"AE-CTX-004":      {ID: "AE-CTX-004", Name: "AgentArtifactsGitignored", Category: "Context", ShortDescription: "Local agent session artifacts are git-ignored.", HelpURI: help("AE-CTX-004")},
	"AE-CTX-006":      {ID: "AE-CTX-006", Name: "InstructionEmphasisRestrained", Category: "Context", ShortDescription: "Agent instructions avoid over-emphasis that degrades adherence (informational).", HelpURI: help("AE-CTX-006")},
	"AE-TEST-001":     {ID: "AE-TEST-001", Name: "AutomatedTestsPresent", Category: "Testing", ShortDescription: "Automated tests are present for each active language so an agent can self-verify.", HelpURI: help("AE-TEST-001")},
	"AE-AUTO-001":     {ID: "AE-AUTO-001", Name: "VerificationCommandDiscoverable", Category: "Autonomy", ShortDescription: "The project's test command is discoverable via a task runner or conventional toolchain.", HelpURI: help("AE-AUTO-001")},
	"AE-SEC-001":      {ID: "AE-SEC-001", Name: "NoRawSecretsInContext", Category: "Secrets", ShortDescription: "No raw secret patterns in agent-visible files.", HelpURI: help("AE-SEC-001")},
	"AE-SEC-002":      {ID: "AE-SEC-002", Name: "NoSecretsInMCPConfig", Category: "Secrets", ShortDescription: "No secret-like values in MCP server config.", HelpURI: help("AE-SEC-002")},
	"AE-MCP-001":      {ID: "AE-MCP-001", Name: "MCPServerPinned", Category: "MCP Safety", ShortDescription: "Every MCP server is pinned to an exact, current, non-deprecated version per the MCP catalog.", HelpURI: help("AE-MCP-001")},
	"AE-MCP-002":      {ID: "AE-MCP-002", Name: "MCPRemoteTrusted", Category: "MCP Safety", ShortDescription: "Remote MCP origins are in the MCP catalog or the trusted allowlist.", HelpURI: help("AE-MCP-002")},
	"AE-MCP-003":      {ID: "AE-MCP-003", Name: "MCPRemoteAuthDeclared", Category: "MCP Safety", ShortDescription: "Remote MCP servers declare authorization metadata.", HelpURI: help("AE-MCP-003")},
	"AE-CC-001":       {ID: "AE-CC-001", Name: "NoDangerousHookCommands", Category: "Agent Config", ShortDescription: "Agent hook configs contain no dangerous shell commands.", HelpURI: help("AE-CC-001")},
	"AE-CC-002":       {ID: "AE-CC-002", Name: "ExplicitAgentEditScope", Category: "Agent Config", ShortDescription: "Agent context declares explicit off-limits paths.", HelpURI: help("AE-CC-002")},
	"AE-ENV-001":      {ID: "AE-ENV-001", Name: "ReproducibleToolchain", Category: "Environment", ShortDescription: "A reproducible toolchain declaration is present.", HelpURI: help("AE-ENV-001")},
	"AE-CI-002":       {ID: "AE-CI-002", Name: "CharterInCI", Category: "CI", ShortDescription: "CI runs charter doctor and workflow linters cleanly.", HelpURI: help("AE-CI-002")},
	"AE-SUPPRESS-001": {ID: "AE-SUPPRESS-001", Name: "SuppressionHasReason", Category: "Governance", ShortDescription: "Every suppression includes a reason.", HelpURI: help("AE-SUPPRESS-001")},
	"AE-SUPPRESS-002": {ID: "AE-SUPPRESS-002", Name: "PermanentSuppressionApproved", Category: "Governance", ShortDescription: "Permanent suppressions name an approver.", HelpURI: help("AE-SUPPRESS-002")},
	"AE-SUPPRESS-003": {ID: "AE-SUPPRESS-003", Name: "SuppressionRate", Category: "Governance", ShortDescription: "Suppression rate is within a healthy range (informational).", HelpURI: help("AE-SUPPRESS-003")},
}

// Lookup returns the catalog entry for a rule ID.
func Lookup(id string) (Entry, bool) {
	e, ok := entries[id]
	return e, ok
}

// IDs returns all catalog rule IDs, sorted.
func IDs() []string {
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// categoryOrder is the canonical display order for the nine rule categories. It
// drives the readiness scorecard so the category list is stable and matches the
// product's mental model (context first, governance last) rather than the
// alphabetical ID order.
var categoryOrder = []string{
	"Context", "Secrets", "MCP Safety", "Agent Config",
	"Environment", "CI", "Testing", "Autonomy", "Governance",
}

// Categories returns the rule categories in canonical display order, restricted
// to categories that actually have at least one catalog rule.
func Categories() []string {
	counts := RuleCountByCategory()
	out := make([]string, 0, len(categoryOrder))
	for _, c := range categoryOrder {
		if counts[c] > 0 {
			out = append(out, c)
		}
	}
	// Append any category present in the catalog but missing from the canonical
	// order (defensive: a new category should be added to categoryOrder, but it
	// must never silently vanish from the scorecard).
	seen := map[string]bool{}
	for _, c := range out {
		seen[c] = true
	}
	extra := make([]string, 0)
	for c := range counts {
		if !seen[c] {
			extra = append(extra, c)
		}
	}
	sort.Strings(extra)
	return append(out, extra...)
}

// RuleCountByCategory returns the number of catalog rules in each category.
func RuleCountByCategory() map[string]int {
	counts := make(map[string]int, len(categoryOrder))
	for _, e := range entries {
		counts[e.Category]++
	}
	return counts
}
