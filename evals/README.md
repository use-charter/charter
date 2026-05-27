# Evals

Use this directory for prompt, workflow, and future agent evaluation artifacts.

Principles:

- eval-driven development over vibe-based confidence
- capture edge cases and regressions as they appear
- link evals to RFCs, specs, and test fixtures

Add or update an eval when a non-trivial agent-facing workflow, prompt contract, or machine-judged behavior changes.

First-slice expectation:

- if the first rule slice changes agent-facing output, add an eval artifact or a documented reason why deterministic Go tests are enough
- if no eval is needed, the slice should still cite the proof model and CLI quality principles explicitly

References:

- `docs/internal/superpowers/checklists/2026-05-28-first-slice-proof-model.md`
- `docs/internal/superpowers/checklists/2026-05-28-cli-quality-principles.md`
