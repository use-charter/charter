# Architecture Docs

Use this directory for the product authority, supporting maps, and architecture records that explain how the repo should evolve.

This content is part of the repo-internal engineering docs tree and will live under `docs/internal/architecture/` once the topology migration is complete.

Documentation authority ladder:

1. `charter-architecture-2026.md` for product behavior, command surface, rule semantics, transports, and output contracts
2. `../audit/charter-v1-audit-checklist.md` for manual rule-audit companion detail only
3. ADRs in `../decisions/` for irreversible architecture constraints
4. root companion docs for execution guidance only
5. HTML artifacts as presentation mirrors only

Normative source rule:

- `charter-architecture-2026.md` is canonical for product behavior and acceptance criteria
- HTML artifacts are presentation surfaces and must not introduce behavior-only requirements absent from the markdown source
- If markdown and HTML disagree, fix markdown first, then reconcile HTML

Read order:

- `charter-architecture-2026.md`
- relevant C4 map
- linked ADRs and RFCs
