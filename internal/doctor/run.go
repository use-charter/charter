package doctor

import (
	"time"

	"go.use-charter.dev/charter/internal/config"
	"go.use-charter.dev/charter/internal/findings"
	"go.use-charter.dev/charter/internal/repository"
	goagentconfig "go.use-charter.dev/charter/internal/rules/agentconfig"
	goci "go.use-charter.dev/charter/internal/rules/ci"
	goctx "go.use-charter.dev/charter/internal/rules/context"
	goenv "go.use-charter.dev/charter/internal/rules/environment"
	gogovernance "go.use-charter.dev/charter/internal/rules/governance"
	gomcp "go.use-charter.dev/charter/internal/rules/mcp"
	gooperability "go.use-charter.dev/charter/internal/rules/operability"
	gosecrets "go.use-charter.dev/charter/internal/rules/secrets"
	"go.use-charter.dev/charter/internal/scoring"
	"go.use-charter.dev/charter/internal/suppress"
)

type Result struct {
	Root       string
	Threshold  int
	Passed     bool
	Findings   []findings.Finding
	Suppressed []suppress.Suppressed
	Score      scoring.Result
	// PathsScanned is the number of repository paths Charter inventoried for this
	// run — a deterministic measure of scan breadth surfaced in the text summary.
	PathsScanned int
}

func Run(path string, threshold int, thresholdSet bool) (Result, error) {
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
	all = append(all, gooperability.Run(root, inv)...)

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

	policy, err := config.LoadPolicy(root, inv)
	if err != nil {
		return Result{}, err
	}
	effective, err := config.ResolveThreshold(policy, threshold, thresholdSet)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Root:         root,
		Threshold:    effective,
		Passed:       score.Final >= effective,
		Findings:     active,
		Suppressed:   suppressed,
		Score:        score,
		PathsScanned: len(inv.Paths),
	}, nil
}
