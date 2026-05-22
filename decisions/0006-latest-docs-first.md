# ADR-0006 Latest Docs First

- Status: Accepted
- Context: Built-in model knowledge drifts faster than toolchains and APIs stabilize.
- Decision: Agents must inspect local versions and current official docs before code touching external tooling or APIs.
- Consequences: Prompting, playbooks, and reviews must all enforce freshness checks.
