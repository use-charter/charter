<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/product/images/logo-dark.svg?v=3">
  <img src="docs/product/images/logo-light.svg?v=3" alt="Charter" width="200">
</picture>

### AI-agent readiness, scored.

Charter grades any repository **0–100** on how safely an AI coding agent can work in it —
then hands you the exact fix for every gap. Offline, deterministic, no LLM in the loop.

[Documentation](https://use-charter.dev/docs) · [Rule catalog](https://use-charter.dev/rules) · [GitHub Action](https://use-charter.dev/docs/ci/github-action) · [Changelog](https://use-charter.dev/changelog)

<p>
  <a href="https://github.com/use-charter/charter/releases"><img alt="Release" src="https://img.shields.io/github/v/release/use-charter/charter?sort=semver&style=flat-square&color=5aa2f7&labelColor=0b0e14"></a>
  <a href="https://github.com/use-charter/charter/actions"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/use-charter/charter/ci.yml?branch=main&style=flat-square&labelColor=0b0e14"></a>
  <a href="./LICENSE"><img alt="License" src="https://img.shields.io/badge/license-Apache--2.0-5aa2f7?style=flat-square&labelColor=0b0e14"></a>
  <img alt="Go" src="https://img.shields.io/badge/go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white&labelColor=0b0e14">
  <img alt="SLSA Level 3" src="https://img.shields.io/badge/SLSA-level%203-45dd92?style=flat-square&labelColor=0b0e14">
  <img alt="Network" src="https://img.shields.io/badge/network-0%20calls-45dd92?style=flat-square&labelColor=0b0e14">
  <img alt="Output" src="https://img.shields.io/badge/SARIF-2.1.0-a78bfa?style=flat-square&labelColor=0b0e14">
</p>

<br>

<img src="docs/internal/demo/charter-demo.gif" alt="charter doctor scoring a repository: 85/100 PASS, one HIGH finding, per-category readiness bars" width="860">

</div>

---

## Why Charter

An AI coding agent is only as safe and effective as the repository it operates in. Missing
context files, unpinned MCP servers, leaked secrets, and absent CI gates quietly degrade
every agent that touches your code — and you find out in the diff, not before.

Charter makes that readiness **measurable**: one deterministic score, **18 rules** across
**9 categories**, **0 network calls**. Run it locally, gate it in CI, fix what it flags.

|  | |
|---|---|
| **Deterministic** | Same repo, same score, every time. No model in the loop, no flaky output. |
| **Offline** | Never phones home. Safe for private, regulated, and air-gapped codebases. |
| **Actionable** | Every finding carries a rule ID, a reason, and a fix. `charter fix` repairs many of them diff-first. |
| **Cross-vendor** | One score for Claude Code, Cursor, Copilot, Gemini, Windsurf, Codex, Zed, and Replit. |

> [!NOTE]
> Charter is a **static analyzer**, not an agent. It reads your files and prints a number.
> No code is sent anywhere, nothing is mutated without a diff you approve.

## Quickstart

```bash
# install (macOS / Linux)
brew install use-charter/tap/charter

# score the current repo
charter doctor
```

That's the whole loop. `charter doctor` scans the tree, evaluates 18 rules, and prints a
banded score with a per-category breakdown and every finding inline. Exit code is the gate:
`0` pass · `1` below threshold · `2` error.

> [!TIP]
> New repo? Run `charter init` first to scaffold the context files agents expect
> (`AGENTS.md`, `charter.yaml`, `.gitignore`), then `charter doctor` for an honest baseline.

## Install

Signed binaries ship for **macOS, Linux, and Windows** on both `amd64` and `arm64`. Pick your
platform below, or use Go on any of them.

### macOS

```bash
brew install use-charter/tap/charter
```

Works on Apple Silicon and Intel. Upgrade later with `brew upgrade charter`.

### Linux

```bash
# Homebrew (Linuxbrew)
brew install use-charter/tap/charter
```

No Homebrew? Grab the archive for your architecture from the [latest release](https://github.com/use-charter/charter/releases/latest)
— `charter_<version>_linux_amd64.tar.gz` (or `_arm64`) — then:

```bash
tar -xzf charter_*_linux_*.tar.gz
sudo install charter /usr/local/bin/charter
```

### Windows

Download `charter_<version>_windows_amd64.zip` (or `_arm64`) from the
[latest release](https://github.com/use-charter/charter/releases/latest), unzip it, and add
`charter.exe` to your `PATH`:

```powershell
# from the folder where you unzipped charter.exe
$dest = "$env:LOCALAPPDATA\Programs\charter"
New-Item -ItemType Directory -Force $dest | Out-Null
Move-Item charter.exe $dest -Force
[Environment]::SetEnvironmentVariable("Path", "$env:Path;$dest", "User")   # reopen the terminal
```

### Any platform — Go

```bash
go install go.use-charter.dev/charter/cmd/charter@latest   # requires Go 1.26+
```

### From source

```bash
git clone https://github.com/use-charter/charter && cd charter
go build -o charter ./cmd/charter
```

### Verify the download

Every release is **cosign-signed with SLSA Level 3 provenance** and ships `checksums.txt`, a
Sigstore bundle (`checksums.txt.sigstore.json`), and a per-archive SBOM. Confirm what you
installed at any time:

```bash
charter version --verify   # prints version, build provenance, and cosign verification status
```

## Commands

A small, sharp surface. The same seven commands behave identically in your shell and in CI.

| Command | What it does |
|---------|--------------|
| `charter doctor` | Scan + score the repo 0–100 with a per-category breakdown. |
| `charter init` | Scaffold missing context files. Never overwrites. |
| `charter fix` | Diff-first auto-repair for supported rules — nothing written until you approve. |
| `charter explain <RULE>` | Print a rule's category, summary, severity, and docs URL. |
| `charter suppress <RULE>` | Record a governed waiver with a reason, owner, and expiry. |
| `charter report` | Write a self-contained offline HTML report (fonts + data inlined). |
| `charter version` | Print version, build provenance, and supply-chain verification status. |

## What it checks

18 rules across 9 categories. Severity sets the score weight; every rule carries an `AE-*` ID
and a fix. Full reference at [use-charter.dev/rules](https://use-charter.dev/rules).

| Category | Rules | Checks for |
|----------|-------|-----------|
| **Context** | `AE-CTX-001/002/004/006` | A meaningful, accurate `AGENTS.md`; agent artifacts git-ignored; restrained emphasis. |
| **Secrets** | `AE-SEC-001/002` | No raw secrets in agent-visible files or MCP config. |
| **MCP Safety** | `AE-MCP-001/002/003` | MCP servers pinned, origins trusted, auth declared. |
| **Agent Config** | `AE-CC-001/002` | No dangerous hook commands; explicit agent edit scope. |
| **Environment** | `AE-ENV-001` | Reproducible toolchain — lockfile + pinned versions. |
| **CI** | `AE-CI-002` | Charter and workflow linters run in CI. |
| **Testing** | `AE-TEST-001` | Automated tests exist so an agent can self-verify. |
| **Autonomy** | `AE-AUTO-001` | The verification command is discoverable. |
| **Governance** | `AE-SUPPRESS-001/002/003` | Suppressions have a reason, an approver, and stay within a healthy rate. |

## Scoring

```
score = max(0, 100 − B×20 − H×10 − M×4 − L×1)
final = min(score, applicable_cap)
```

`B`/`H`/`M`/`L` are Blocker/High/Medium/Low finding counts. Informational findings are
excluded. Hard caps keep the dangerous cases honest:

- a raw secret in agent-visible content → **≤ 49**
- any active Blocker → **≤ 59**

The formula is public and stable within a major version. Same inputs, same score — always.

## Gate it in CI

```yaml
# .github/workflows/charter.yml
- uses: use-charter/charter-action@v1
  with:
    threshold: "80"   # fail PRs that score below this
    verify: true      # cosign + sha256 the binary before running
```

The action downloads the signed binary, runs `charter doctor --format sarif`, and uploads to
**GitHub Code Scanning** — findings land inline on the PR, no new dashboard to learn. See
[`action/README.md`](./action/README.md).

## The contract

Charter makes ten commitments, and shows its work on every one:

- ✅ No network calls — ever
- ✅ No LLM calls in the core
- ✅ No file deletion
- ✅ No silent mutation — diff-first fixes only
- ✅ Every finding has a rule ID and fix guidance
- ✅ Every release is signed (SLSA L3 + cosign)
- ✅ The score formula is public and stable within a major version
- ✅ Cross-vendor across every major coding agent
- ✅ Secrets are never printed
- ✅ The CLI is free, forever (Apache-2.0)

## Performance

`charter doctor` scans a **50,000-file repository in ≤ 2 s** using **≤ 256 MiB** RSS —
asserted in CI by `moon run :perf`.

## Tech stack

| Layer | Built with |
|-------|------------|
| **Core CLI** | ![Go](https://img.shields.io/badge/Go_1.26-00ADD8?style=flat-square&logo=go&logoColor=white) |
| **Build & tooling** | ![Moonrepo](https://img.shields.io/badge/Moonrepo-6F52F4?style=flat-square&logo=moonrepo&logoColor=white) ![mise](https://img.shields.io/badge/mise-A78BFA?style=flat-square) ![Bun](https://img.shields.io/badge/Bun-1B1817?style=flat-square&logo=bun&logoColor=white) ![hk](https://img.shields.io/badge/hk_hooks-2F6FDB?style=flat-square) |
| **Web & docs** | ![Astro](https://img.shields.io/badge/Astro-BC52EE?style=flat-square&logo=astro&logoColor=white) ![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=flat-square&logo=typescript&logoColor=white) ![Mintlify](https://img.shields.io/badge/Mintlify-18181B?style=flat-square&logo=mintlify&logoColor=white) |
| **Infra & CI/CD** | ![Cloudflare Workers](https://img.shields.io/badge/Cloudflare_Workers-F38020?style=flat-square&logo=cloudflareworkers&logoColor=white) ![Cloudflare Pages](https://img.shields.io/badge/Cloudflare_Pages-F38020?style=flat-square&logo=cloudflarepages&logoColor=white) ![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-2088FF?style=flat-square&logo=githubactions&logoColor=white) |
| **Supply chain** | ![SLSA](https://img.shields.io/badge/SLSA_L3-45DD92?style=flat-square) ![cosign](https://img.shields.io/badge/Sigstore_cosign-003A70?style=flat-square) ![SARIF](https://img.shields.io/badge/SARIF_2.1.0-A78BFA?style=flat-square) |

## Documentation

| | |
|---|---|
| **Product docs** | [use-charter.dev/docs](https://use-charter.dev/docs) (Mintlify) · source in [`docs/product/`](./docs/product/) |
| **Rule reference** | [use-charter.dev/rules](https://use-charter.dev/rules) |
| **Architecture** | [`docs/internal/architecture/charter-architecture-2026.md`](./docs/internal/architecture/charter-architecture-2026.md) |
| **Repo contract** | [`AGENTS.md`](./AGENTS.md) · [`ARCHITECTURE.md`](./ARCHITECTURE.md) · [`SECURITY.md`](./SECURITY.md) · [`CONTRIBUTING.md`](./CONTRIBUTING.md) · [`TESTING.md`](./TESTING.md) |

## Contributing

```bash
mise install               # toolchain (Go, Bun, Moon, hk)
./scripts/install-hooks.sh # pre-commit / pre-push hooks
moon run :check            # full quality gate
```

| Task | Runs |
|------|------|
| `moon run :check` | full quality gate |
| `moon run :test` | `go test -race ./...` |
| `moon run :lint` | gofumpt + golangci-lint + tsc |
| `moon run :build` | `go build` |
| `moon run :docs` | docs validation |
| `moon run :security` | gitleaks + govulncheck + osv-scanner |
| `moon run :perf` | 50k-file performance assertion |

Conventional Commits, SemVer, DCO sign-off on every commit. Irreversible decisions are
recorded as ADRs in [`docs/internal/decisions/`](./docs/internal/decisions/). See
[`CONTRIBUTING.md`](./CONTRIBUTING.md).

## Star history

<a href="https://star-history.com/#use-charter/charter&Date">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=use-charter/charter&type=Date&theme=dark">
    <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=use-charter/charter&type=Date" width="600">
  </picture>
</a>

## License

[Apache License 2.0](./LICENSE). DCO-first contributions.

<div align="center">
<br>
<sub>Built for the agent era · <a href="https://use-charter.dev">use-charter.dev</a></sub>
</div>
