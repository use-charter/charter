# Charter

Charter is an offline-first AI-agent-readiness scanner for software repositories. It audits agent context, MCP safety, reproducibility, CI posture, and governance so teams can safely adopt coding agents without guesswork.

This repository is intentionally bootstrapped as an AI-ready Go monorepo before product implementation begins. The repo itself is the first dogfood target.

## Current State

- Phase: Phase 0 foundation complete; product implementation not started
- Product authority: [`docs/architecture/charter-architecture-2026.md`](./docs/architecture/charter-architecture-2026.md)
- Module path: `go.charter.dev/charter`
- Repo contract: [`AGENTS.md`](./AGENTS.md)

Documentation authority ladder:

1. `docs/architecture/charter-architecture-2026.md` for product behavior
2. `docs/audit/charter-v1-audit-checklist.md` for manual rule-audit companion detail
3. ADRs in `decisions/` for irreversible constraints
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
- ADRs in [`decisions/`](./decisions/).
- RFCs in [`rfcs/`](./rfcs/).
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

## License

MIT. See [`LICENSE`](./LICENSE).
