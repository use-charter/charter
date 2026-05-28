package doctor

import (
	"go.charter.dev/charter/internal/findings"
	"go.charter.dev/charter/internal/repository"
	goci "go.charter.dev/charter/internal/rules/ci"
	goctx "go.charter.dev/charter/internal/rules/context"
	goenv "go.charter.dev/charter/internal/rules/environment"
	"go.charter.dev/charter/internal/scoring"
)

type Result struct {
	Root     string
	Findings []findings.Finding
	Score    scoring.Result
}

func Run(path string) (Result, error) {
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

	return Result{
		Root:     root,
		Findings: all,
		Score:    scoring.Calculate(all),
	}, nil
}
