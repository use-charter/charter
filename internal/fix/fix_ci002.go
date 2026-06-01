package fix

import (
	"os"
	"path/filepath"
	"strings"

	"go.use-charter.dev/charter/internal/repository"
)

// charterWorkflow is the Charter CI gate workflow planted by AE-CI-002. The
// actions/checkout SHA mirrors the pin already used in .github/workflows/ci.yml
// (v6.0.2) so the generated workflow follows the repository's pinning policy.
const charterWorkflow = `name: Charter
on:
  pull_request:
  push:
    branches: [main]
permissions:
  contents: read
  security-events: write
jobs:
  charter:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
      - uses: use-charter/charter-action@v1
        with:
          threshold: "80"
`

// fixCI002 plans creation of .github/workflows/charter.yaml when no tracked
// workflow already wires up the Charter gate (via `charter doctor` or the
// first-party action). If any workflow already carries the gate, no fix is
// proposed (false).
func fixCI002(root string, inv repository.Inventory) (FilePlan, bool, error) {
	for _, p := range inv.Paths {
		if !strings.HasPrefix(p, ".github/workflows/") {
			continue
		}
		if !strings.HasSuffix(p, ".yml") && !strings.HasSuffix(p, ".yaml") {
			continue
		}
		// #nosec G304 -- p is a tracked workflow path under .github/workflows/ joined to root.
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			continue
		}
		text := string(data)
		if strings.Contains(text, "charter doctor") || strings.Contains(text, "use-charter/charter-action@") {
			return FilePlan{}, false, nil
		}
	}

	const path = ".github/workflows/charter.yaml"
	contents := []byte(charterWorkflow)

	return FilePlan{
		RuleID:   "AE-CI-002",
		Path:     path,
		Action:   Create,
		Contents: contents,
		Diff:     buildCreateDiff(path, contents),
	}, true, nil
}
