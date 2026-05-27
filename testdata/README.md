# Testdata

Use this directory for deterministic fixtures used by rules, renderers, fixes, and evals.

First-slice expectation:

- the first implemented rule should add at least one pass fixture and one fail fixture here, or document why a direct code-level failing test is sufficient
- fixture names should describe the rule and scenario, not implementation detail
- fixtures must remain secret-safe and fully reviewable in git
- each fixture should map back to a spec or rule contract

Proof model reference:

- `docs/superpowers/checklists/2026-05-28-first-slice-proof-model.md`
