package fix

import (
	"strings"

	"go.use-charter.dev/charter/internal/agentcontext"
	"go.use-charter.dev/charter/internal/repository"
	"go.use-charter.dev/charter/internal/scaffold"
)

// fixCTX001 plans creation of a root AGENTS.md when the repository has no
// agent-context file at all. If any single-file context candidate or any
// .cursor/rules/ entry is present, the rule is already satisfied and no fix is
// proposed (false). The template comes from scaffold so it satisfies Charter's
// own context gates out of the box.
func fixCTX001(root string, inv repository.Inventory) (FilePlan, bool, error) {
	for _, f := range agentcontext.Files {
		if inv.Has(f) {
			return FilePlan{}, false, nil
		}
	}
	for _, p := range inv.Paths {
		if p == agentcontext.CursorRulesDir || strings.HasPrefix(p, agentcontext.CursorRulesDir+"/") {
			return FilePlan{}, false, nil
		}
	}

	proj, _ := scaffold.Detect(root)
	contents := scaffold.AGENTSMarkdown(proj)

	return FilePlan{
		RuleID:   "AE-CTX-001",
		Path:     "AGENTS.md",
		Action:   Create,
		Contents: contents,
		Diff:     buildCreateDiff("AGENTS.md", contents),
	}, true, nil
}
