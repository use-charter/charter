# docs/

Two trees, two audiences:

- [`internal/`](./internal/) — repository engineering knowledge (ADRs, specs, runbooks, architecture, design system).
- [`product/`](./product/) — the customer-facing Mintlify site, deployed at [use-charter.dev/docs](https://use-charter.dev/docs) and `/rules`.

## Authority model

Behavior is owned by source and contract docs, never by presentation:

1. **`internal/architecture/charter-architecture-2026.md`** — canonical product behavior (command surface, rule semantics, output contracts).
2. **`internal/decisions/`** (ADRs) — irreversible constraints and cross-cutting decisions.
3. **`internal/specs/`** (`AE-*`) — rule-level behavior contracts.
4. **Root contract docs** ([`ARCHITECTURE.md`](../ARCHITECTURE.md), [`TESTING.md`](../TESTING.md), …) — execution guidance.
5. **HTML mirrors and the Mintlify site** — presentation only.

Markdown is the source of truth. If a mirror disagrees with its Markdown, fix the
Markdown first. ADRs link related RFCs/specs; specs link related ADRs/evals.

## `internal/`

| Directory | Contents |
|-----------|----------|
| `architecture/` | Canonical behavior doc + `c4/` architecture maps. |
| `decisions/` | ADRs (`NNNN-title.md`) — written before hard-to-reverse changes. |
| `specs/` | Per-rule behavior contracts (severity, detection, pass/fail examples, remediation). |
| `rfcs/` | Proposals for cross-cutting or risky changes, before implementation. |
| `catalog/` | MCP catalog governance — contribution process, false-positive validation gate, advisory curation. |
| `runbooks/` | Operational responses (release, CI failure, security incident). |
| `playbooks/` | Repeatable implementation flows (feature, bugfix, rule, CI, docs). |
| `designs/` | Visual identity — `DESIGN-TOKENS.md` is the canonical token reference (terminal + HTML report). |
| `audit/` | Manual rule-audit companions. |
| `demo/` | Demo tape/GIF, fixtures, and CLI-output mockups. |

## `product/`

The Mintlify site. `docs.json` defines navigation and **permanent URL contracts** —
rule pages live at `/rules/AE-*` because SARIF `helpUri` points there; moves go
through `docs.json` redirects.

| Area | Contents |
|------|----------|
| `introduction` · `installation` · `quickstart` · `design-philosophy` | Onboarding + the product commitments. |
| `concepts/` | How it works — readiness model, scoring & caps, fix engine, MCP safety, suppression governance. |
| `how-to/` | Task guides — adopt in a repo, CI, pre-commit, suppressing, using `fix` safely. |
| `config/` | `charter.yaml` and policy-profile reference. |
| `cli/` | Per-command reference. |
| `rules/` | One page per `AE-*` rule. Bootstrapped by `scripts/generate-rule-pages.ts`, then hand-maintained. |

Local preview: `npx mintlify dev` from `docs/product/`. Validate with
`moon run :docs` (mirrors, rule-page coverage, and `docs.json` resolution).
