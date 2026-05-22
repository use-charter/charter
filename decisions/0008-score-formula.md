# ADR-0008 Score Formula

- Status: Accepted
- Context: Trust requires a stable and explainable scoring model.
- Decision: Use `max(0, 100 - B×20 - H×10 - M×4 - L×1)` with documented hard-cap rules.
- Consequences: Rule severities and public docs must stay aligned with the scoring contract.
