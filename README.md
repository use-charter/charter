# Charter

Charter is an offline-first AI-agent-readiness scanner for software repositories. It audits agent context, MCP safety, reproducibility, CI posture, and governance so teams can safely adopt coding agents without guesswork.

This repository started as the AI-ready bootstrap baseline and now contains the first real `charter doctor` path with both text and JSON output. The repo itself remains the first dogfood target.

## Current State

- Phase: Phase 1 Slice 8 implemented; real `charter doctor` path with the full 15-rule v1 set, governance, suppression, SARIF output, and policy profiles
- Product authority: [`docs/internal/architecture/charter-architecture-2026.md`](./docs/internal/architecture/charter-architecture-2026.md)
- Module path: `go.use-charter.dev/charter`
- Repo contract: [`AGENTS.md`](./AGENTS.md)

Current implemented scope:

- repository resolver
- file inventory scanner (git-aware, respects .gitignore)
- finding model with structured locations (path:line, 1-based)
- score engine with hard caps (blocker ≤59, secret ≤49)
- `charter doctor` output with `--path`, `--threshold`, `--quiet`, and `--format text|json|markdown`
- 15 implemented rules: `AE-CTX-001`, `AE-CTX-002`, `AE-CTX-004`, `AE-ENV-001`, `AE-CI-002`, `AE-SEC-001`, `AE-SEC-002`, `AE-MCP-001`, `AE-MCP-002`, `AE-MCP-003`, `AE-CC-001`, `AE-CC-002`, `AE-SUPPRESS-001`, `AE-SUPPRESS-002`, `AE-SUPPRESS-003`
- agent context registry (`agentcontext` — shared source for context and secret scanning, drift guard)
- blocker-level secret detection with redacted evidence and score cap at `49`: `AE-SEC-001` (agent context), `AE-SEC-002` (MCP config)
- MCP config scanning of `.mcp.json` / `mcp.json` / `.cursor/mcp.json` / `.vscode/mcp.json`: server pinning (`AE-MCP-001`), trusted-remote allowlist via `charter.yaml` (`AE-MCP-002`), and remote auth declaration (`AE-MCP-003`)
- agent-config scanning of JSON hook configs (`.claude/settings.json`, `.cursor/hooks.json`): dangerous shell commands (`AE-CC-001`, Blocker) and explicit edit-scope declaration (`AE-CC-002`)
- suppression engine with governance: `.charter-suppress.yml` and inline `charter:ignore` directives, with `charter suppress <RULE>` to author entries; suppressed findings are excluded from the score and listed separately, and the governance rules `AE-SUPPRESS-001` (missing reason), `AE-SUPPRESS-002` (permanent waiver without approver), and `AE-SUPPRESS-003` (high suppression rate, informational) audit them
- text, JSON, Markdown (PR-comment), and SARIF 2.1.0 output with structured `path:line` finding locations; `--out <file>` writes any format to a file. SARIF is backed by a rule catalog (`internal/rules/catalog`) and carries GitHub Code Scanning metadata (`security-severity`, `tags`, `automationDetails`) so findings rank correctly in the Security tab
- policy profiles in `charter.yaml`: `policy.profile` (strict=90 / standard=80 / relaxed=60) or `policy.threshold`, with precedence `--threshold` flag > `policy.threshold` > `policy.profile` > default 80 (see `schemas/charter-config.schema.json`)

Documentation topology:

- root contract docs stay at repo root
- internal engineering docs live under `docs/internal/`
- future customer-facing docs live under `docs/product/`

Documentation authority ladder:

1. `docs/internal/architecture/charter-architecture-2026.md` for product behavior
2. `docs/internal/audit/charter-v1-audit-checklist.md` for manual rule-audit companion detail
3. ADRs in `docs/internal/decisions/` for irreversible constraints
4. root companion docs for execution guidance only
5. HTML artifacts as presentation mirrors only

## Golden Path

```bash
mise install
./scripts/install-hooks.sh
moon run :check
```

This repo commits `.miserc.toml` to ignore user-global `mise` config so project tool resolution stays hermetic.

All repo automation should route through the same task family:

- `moon run :check`
- `moon run :test`
- `moon run :vet`
- `moon run :lint`
- `moon run :build`
- `moon run :docs`
- `moon run :security`
- `moon run :eval`
- `moon run :actionlint`
- `moon run :zizmor`

## Core Conventions

- Latest docs first. Local manifests before memory.
- Project-local `mise` config only. User-global `mise` config is intentionally ignored in this repo.
- Bootstrap keeps tracked MCP configuration absent until a pinned, reviewed integration exists.
- Conventional Commits.
- SemVer.
- Single Go module. No `go.work`. No extra modules.
- Go command entrypoint lives in `cmd/charter/`. Public Go API is deferred until stable.
- Repo-owned helper scripts use TypeScript via Bun. No plain JavaScript helpers.
- ADRs in [`docs/internal/decisions/`](./docs/internal/decisions/).
- RFCs in [`docs/internal/rfcs/`](./docs/internal/rfcs/).
- Contract-first schemas in [`api/openapi/`](./api/openapi/) and [`schemas/`](./schemas/).
- Evals and verification artifacts in [`evals/`](./evals/).

GitHub-level repo health features stay split by capability:

- committed here: Renovate config and a disabled Scorecard workflow stub for future public/org enablement
- outside source control and should be enabled where supported: CodeQL default setup, required checks, branch protection, private vulnerability reporting

## Repo Map

- [`AGENTS.md`](./AGENTS.md): canonical agent instructions
- [`ARCHITECTURE.md`](./ARCHITECTURE.md): structure and architecture rules
- [`SECURITY.md`](./SECURITY.md): safety posture
- [`CONTRIBUTING.md`](./CONTRIBUTING.md): contribution workflow
- [`TESTING.md`](./TESTING.md): tests, fixtures, evals
- [`CONTEXT_MAP.md`](./CONTEXT_MAP.md): knowledge graph lite
- [`PERMISSIONS.md`](./PERMISSIONS.md): edit and escalation boundaries
- [`docs/internal/README.md`](./docs/internal/README.md): repo-internal engineering docs
- [`docs/product/README.md`](./docs/product/README.md): future customer-facing docs home

## License

Apache License 2.0 (`Apache-2.0`). See [`LICENSE`](./LICENSE).

## Contribution Model

- DCO first
- CLA deferred unless governance or commercial needs require it later

## Commercial Model

- Apache-2.0 OSS core
- paid hosted service, support, and enterprise features outside the OSS core
