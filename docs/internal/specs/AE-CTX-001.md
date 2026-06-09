# AE-CTX-001

- Severity: Blocker
- Category: Context
- Description: Agent context file must exist, be meaningful, and fit budget.
- Detection logic: inspect canonical agent-visible files in repo root, confirm at least one valid context file exists, confirm it is non-empty, contains project summary + tech stack + edit boundaries + verification command, and flag files that exceed the configured token budget.
- Pass example: `AGENTS.md` exists with project state, commands, hard constraints, edit scope, and context loading.
- Fail example: no agent context file, empty placeholder text, or a context file that is so large it risks truncation.
- Evidence expectations: file path, detected format, first substantive lines, and note whether the file appears over budget.
- Edge cases: a repo using `CLAUDE.md`, `DESIGN.md`, or `SKILL.md` instead of `AGENTS.md` can pass if the file satisfies the same content requirements.
- Remediation: create or tighten the canonical context file and keep it small enough to survive agent context windows.
- Why: Agents operate from what they can read in the repo. A missing or empty context file means the agent starts every session with no orientation, guesses at the stack, and may edit files it should leave alone.
- Auto-fixable: Yes — `charter fix --rule AE-CTX-001`
- Related rules: AE-CTX-002, AE-CTX-004, AE-CTX-006, AE-CC-002
- Related ADRs: ADR-0006
- Related evals: None yet
