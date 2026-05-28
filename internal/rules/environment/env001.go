package environment

import (
	"sort"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
)

func Run(root string, inv repository.Inventory) []findings.Finding {
	missing := requiredEnvironmentFiles(inv)
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

func requiredEnvironmentFiles(inv repository.Inventory) []string {
	required := []string{"mise.toml", "mise.lock", "hk.pkl"}

	if inv.Has("go.mod") {
		required = append(required, "go.sum")
	}
	if inv.Has("package.json") {
		required = append(required, "bun.lock")
	}

	var missing []string
	for _, path := range required {
		if !inv.Has(path) {
			missing = append(missing, path)
		}
	}

	return missing
}
