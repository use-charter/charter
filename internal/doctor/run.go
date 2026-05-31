package doctor

import (
	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
	goci "go.charter.dev/charter/internal/rules/ci"
	goctx "go.charter.dev/charter/internal/rules/context"
	goenv "go.charter.dev/charter/internal/rules/environment"
	gomcp "go.charter.dev/charter/internal/rules/mcp"
	gosecrets "go.charter.dev/charter/internal/rules/secrets"
	"go.charter.dev/charter/internal/scoring"
)

type Result struct {
	Root      string
	Threshold int
	Passed    bool
	Findings  []findings.Finding
	Score     scoring.Result
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

	secretFindings, err := gosecrets.RunSecretRules(root, inv)
	if err != nil {
		return Result{}, err
	}
	all = append(all, secretFindings...)

	score := scoring.Calculate(all)

	return Result{
		Root:      root,
		Threshold: threshold,
		Passed:    score.Final >= threshold,
		Findings:  all,
		Score:     score,
	}, nil
}
