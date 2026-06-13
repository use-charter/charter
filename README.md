<h1 align="center">Charter</h1>

<p align="center">
  <strong>AI-agent readiness, scored.</strong><br>
  Offline-first CLI that grades any repo 0–100 on how safely an AI coding agent can work in it — then tells you exactly what to fix.
</p>

<p align="center">
  <a href="https://github.com/use-charter/charter/releases"><img alt="Release" src="https://img.shields.io/github/v/release/use-charter/charter?sort=semver"></a>
  <a href="https://github.com/use-charter/charter/actions"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/use-charter/charter/ci.yml?branch=main"></a>
  <a href="./LICENSE"><img alt="License" src="https://img.shields.io/badge/license-Apache--2.0-blue"></a>
  <img alt="Go" src="https://img.shields.io/badge/go-1.26-00ADD8?logo=go&logoColor=white">
  <img alt="Offline" src="https://img.shields.io/badge/network-zero%20calls-success">
</p>

<p align="center">
  <img src="./docs/internal/demo/charter-demo.gif" alt="charter doctor scanning a repository" width="800">
</p>

## Why Charter

AI coding agents are only as safe and effective as the repo they operate in. Missing context files, unpinned MCP servers, leaked secrets, and absent CI gates quietly degrade every agent that touches your code. Charter makes that readiness **measurable** — one deterministic score, eighteen rules, zero network calls, zero LLM calls. Run it locally, gate it in CI, fix what it flags.

- **Deterministic** — same repo, same score, every time. No model in the loop.
- **Offline** — never phones home; safe for private and air-gapped codebases.
- **Actionable** — every finding has a rule ID, a reason, and a fix. `charter fix` repairs many of them diff-first.
- **Cross-vendor** — Claude Code, Cursor, Copilot, Gemini, Windsurf.

## Install

```bash
# Homebrew (macOS / Linux)
brew install use-charter/tap/charter

# Direct binary download
# Download the archive for your platform from GitHub Releases, extract, and put charter on your PATH.

# go install
go install go.use-charter.dev/charter/cmd/charter@latest

# Build from source
git clone https://github.com/use-charter/charter
cd charter && go build -o charter ./cmd/charter
```

## Quick start

```bash
charter doctor --path .
```

Charter scans the repo, evaluates 18 rules, and prints the score with finding details. Exit 0 = pass, exit 1 = below threshold, exit 2 = error.

```bash
charter init        # scaffold missing context files
charter fix         # diff-first auto-repair for supported rules
charter report      # self-contained offline HTML report
charter explain AE-CTX-001  # rule reference
```

## Score formula

```
score = max(0, 100 − B×20 − H×10 − M×4 − L×1)
final = min(base, applicable_cap)
```

Hard caps: raw secret → `≤ 49`, any blocker → `≤ 59`. Informational findings excluded.

## GitHub Action

```yaml
- uses: use-charter/charter-action@v1
  with:
    threshold: "80"
```

Downloads the signed `charter` binary, runs `charter doctor --format sarif`, and uploads to GitHub Code Scanning. See [`action/README.md`](./action/README.md).

## Performance

`charter doctor` scans a 50,000-file repository in ≤ 2 s / ≤ 256 MiB RSS, validated by `moon run :perf`.

## Design principles

Charter makes ten commitments: no network calls, no LLM calls, no file deletion, no silent mutation, every finding has a rule ID and fix guidance, every release is signed (SLSA L3), the score formula is public and stable within a major version, cross-vendor (Claude Code / Cursor / Copilot / Gemini / Windsurf), secrets never printed, CLI free forever.

## Documentation

- Customer-facing docs: [`docs/product/`](./docs/product/) — the Mintlify site at `https://use-charter.dev/docs`
- Rule reference: `https://use-charter.dev/rules/AE-*`
- Architecture: [`docs/internal/architecture/charter-architecture-2026.md`](./docs/internal/architecture/charter-architecture-2026.md)
- Repo contract: [`AGENTS.md`](./AGENTS.md)

## Developer setup

```bash
mise install
./scripts/install-hooks.sh
moon run :check
```

Task family:

```
moon run :check     # full quality gate
moon run :test      # go test -race ./...
moon run :lint      # gofumpt + golangci-lint + tsc
moon run :build     # go build
moon run :docs      # docs validation
moon run :security  # gitleaks + govulncheck + osv-scanner
moon run :perf      # 50k-file performance assertion (not in :check)
```

## Core conventions

- Single Go module `go.use-charter.dev/charter`. No `go.work`.
- Conventional Commits. SemVer. DCO sign-off on every commit.
- Repo-owned scripts in TypeScript via Bun. No plain JS helpers.
- Bootstrap keeps tracked MCP config absent until a pinned, reviewed integration exists.
- ADRs in [`docs/internal/decisions/`](./docs/internal/decisions/) for irreversible constraints.

## Repo map

- [`AGENTS.md`](./AGENTS.md) — agent instructions and hard constraints
- [`ARCHITECTURE.md`](./ARCHITECTURE.md) — module layout and seam rules
- [`CONTRIBUTING.md`](./CONTRIBUTING.md) — workflow, commits, PRs, ADR/RFC triggers
- [`SECURITY.md`](./SECURITY.md) — secrets, MCP, supply-chain posture
- [`TESTING.md`](./TESTING.md) — fixtures, evals, verification commands
- [`CONTEXT_MAP.md`](./CONTEXT_MAP.md) — context loading guide
- [`PERMISSIONS.md`](./PERMISSIONS.md) — off-limits paths and escalation policy

## License

Apache License 2.0. See [`LICENSE`](./LICENSE).

DCO-first contributions. CLA deferred unless governance needs require it.
