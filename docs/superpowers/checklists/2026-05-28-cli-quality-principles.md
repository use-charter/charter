# CLI Quality Principles

- quiet mode must be intentionally terse
- findings must be specific and actionable
- machine-readable output must not depend on terminal formatting
- diff-first behavior must be preserved for fix flows
- agent-facing output must remain deterministic

## Output rules for Phase 1 slices

- use stable rule IDs, severity labels, and evidence fields
- never print raw secrets or unredacted secret-like values
- human text output must explain what failed and what to do next
- machine-readable output must be parseable without ANSI color handling
- command success and failure must be obvious from both exit code and output
