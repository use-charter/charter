# ADR-0007 No LLM Calls in Core

- Status: Accepted
- Context: Charter audits AI-readiness and must itself be predictable, cheap, and inspectable.
- Decision: Core scanner, scorer, and renderers do not call LLMs.
- Consequences: Any future agent-facing features must remain outside the trusted scan path.
