package doctor

import (
	"time"

	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
	goagentconfig "go.charter.dev/charter/internal/rules/agentconfig"
	goci "go.charter.dev/charter/internal/rules/ci"
	goctx "go.charter.dev/charter/internal/rules/context"
	goenv "go.charter.dev/charter/internal/rules/environment"
	gogovernance "go.charter.dev/charter/internal/rules/governance"
	gomcp "go.charter.dev/charter/internal/rules/mcp"
	gosecrets "go.charter.dev/charter/internal/rules/secrets"
	"go.charter.dev/charter/internal/scoring"
	"go.charter.dev/charter/internal/suppress"
)

type Result struct {
	Root       string
	Threshold  int
	Passed     bool
	Findings   []findings.Finding
	Suppressed []suppress.Suppressed
	Score      scoring.Result
}

func Run(path string, threshold int) (Result, error) {
	root, err := repository.ResolveRoot(path)
	if err != nil {
		return Result{}, err
	}

	inv, err := repository.BuildInventory(root)
	if err != nil {
		return Result{}, err
	}

	all := append([]findings.Finding{}, goctx.RunCTXRules(root, inv)...)
	all = append(all, goenv.Run(root, inv)...)
	all = append(all, goci.Run(root, inv)...)

	mcpFindings, err := gomcp.Run(root, inv)
	if err != nil {
		return Result{}, err
	}
	all = append(all, mcpFindings...)

	ccFindings, err := goagentconfig.Run(root, inv)
	if err != nil {
		return Result{}, err
	}
	all = append(all, ccFindings...)

	secretFindings, err := gosecrets.RunSecretRules(root, inv)
	if err != nil {
		return Result{}, err
	}
	all = append(all, secretFindings...)

	fileEntries, err := suppress.LoadFile(root, inv)
	if err != nil {
		return Result{}, err
	}

	active, suppressed, used, err := suppress.Apply(root, all, fileEntries, time.Now())
	if err != nil {
		return Result{}, err
	}

	// len(active) is the active rule-finding count before governance findings are
	// appended — exactly the denominator AE-SUPPRESS-003 needs.
	active = append(active, gogovernance.Run(used, len(active), len(suppressed))...)

	score := scoring.Calculate(active)

	return Result{
		Root:       root,
		Threshold:  threshold,
		Passed:     score.Final >= threshold,
		Findings:   active,
		Suppressed: suppressed,
		Score:      score,
	}, nil
}
