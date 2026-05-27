# AE-CTX-002

- Severity: Medium
- Category: Context
- Description: Agent context must remain current with repo reality.
- Detection logic: compare stated stack, setup path, verification command, off-limits paths, and hook/tooling references against actual repo manifests, workflow files, and directory structure.
- Pass example: `AGENTS.md` points at `moon run :check`, mentions `hk.pkl`, and matches the current toolchain and repo boundaries.
- Fail example: the context file references stale runtimes, missing task paths, or MCP/config behavior that no longer matches the repo.
- Evidence expectations: last-reviewed date if present, plus exact contradictions between the file and repo state.
- Edge cases: an old date alone is not enough to fail the rule if the content is still factually accurate.
- Remediation: update the context file to reflect current repo truth and keep review dates fresh.
- Related ADRs: ADR-0006
- Related evals: None yet
