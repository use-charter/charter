# evals/

Prompt, workflow, and agent-behavior evaluation artifacts.

Principles:

- eval-driven development over vibe-based confidence
- capture edge cases and regressions as they appear
- link evals to RFCs, specs, and test fixtures

Add or update an eval when a non-trivial agent-facing workflow, prompt contract,
or machine-judged behavior changes. If deterministic Go tests fully cover the
behavior, note that instead of adding a redundant eval.

Run them through `moon run :eval`.
