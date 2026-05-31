# Phase 1 Slice 5 Design

## Goal

Implement the MCP Static Config Scanner — `AE-MCP-001` (server packages must be pinned), `AE-MCP-002` (remote origins must be known or allowlisted), and `AE-MCP-003` (sensitive remote servers must declare auth metadata) — over JSON MCP config files, so `charter doctor` reports supply-chain, origin-trust, and authentication risks in a repo's MCP configuration with structured `path:line` locations. This completes milestone M1.3 (T1.3.1) and implements ADR-0011.

This slice does not build the M1.6 MCP catalog, `charter fix`, SARIF, or the AE-CC / AE-SUPPRESS rules.

## Audience

- coding agents implementing this vertical slice
- maintainers reviewing the MCP detection contract and the new `charter.yaml` config surface
- future contributors building the M1.6 catalog and the SARIF / fix consumers that reuse these findings

## Scope

### In scope

- a new `internal/rules/mcp` package: discovery, JSON parsing into a normalized server model, and the three checks
- a new `internal/config` package: a minimal `charter.yaml` loader for `mcp.trustedRemotes`
- adding `gopkg.in/yaml.v3` (latest, verified) as the YAML dependency for `charter.yaml`
- wiring the MCP rules into the `internal/doctor` pipeline with an error-propagating contract
- structured `path:line` locations on all MCP findings (per ADR-0009)
- fixtures under `testdata/repos/` and as-built alignment of the three `AE-MCP-*` specs, the audit checklist, and the architecture-doc examples

### Out of scope

- `mcp.yml` / Pkl MCP **server** scanning (v1 scans JSON config files; the `charter.yaml` allowlist is YAML)
- the M1.6 versioned MCP catalog / CVE-driven re-firing (v1 uses only the local `charter.yaml` allowlist)
- `charter fix` auto-remediation (M1.4) and `--format sarif` (M1.5) — both reuse the locations emitted here
- AE-CC-001/002 (M1.2) and AE-SUPPRESS-001/002/003 governance rules
- re-validating `AE-MCP-003` against the MCP `2026-07-28` release candidate once it ratifies (separate follow-up)
- changing scoring behavior (MCP findings are `High`, no cap)

## Why this slice

Secrets, env, and CI checks already ship; the MCP static config scanner is the one remaining task in M1.3, and finishing it closes "7 of 10 OWASP MCP Top 10 risks covered in v1". The MCP findings (`AE-MCP-001 .mcp.json:7`) are the centerpiece of the architecture doc's example outputs and the PR-comment trust scenario, so they carry high product value. The slice also leverages the just-shipped Slice 4 `Locations` contract directly and reuses the config-file scan pattern established by the secret rules.

## Grounding (verified live 2026-05-31, not from memory)

- **MCP config shape:** root key `mcpServers` (object: name → config). Stdio servers: `command` (string), `args` (`string[]`), `env` (object). Remote servers: `type` (`"http"` | `"sse"`; SSE deprecated early 2026 in favor of HTTP), `url` (string), `headers` (object). Env expansion uses `${VAR}` / `${VAR:-default}` / `$VAR`. All major clients share this structure; VS Code historically uses a top-level `servers` key — parsed as a fallback alias.
- **OWASP MCP Top 10** (official, v0.1 **beta**): `MCP04` Software Supply Chain → `AE-MCP-001`; `MCP09` Shadow MCP Servers → `AE-MCP-002`; `MCP07` Insufficient Authentication & Authorization → `AE-MCP-003`. Cited as "OWASP MCP Top 10 (beta)".
- **MCP spec revision:** `2025-11-25` is the latest stable revision; a `2026-07-28` release candidate hardens OAuth (RFC 9207) and makes the core stateless. Those are wire-protocol changes, not static-config changes, so `AE-MCP-003` checks for the presence of an auth declaration rather than specific OAuth field names.

## Precondition

None blocking. This slice lives entirely in new packages (`internal/rules/mcp`, `internal/config`) plus `internal/doctor/run.go`; it does **not** touch the admin-blocked `internal/secrets/` or `internal/rules/secrets/` paths.

Cross-rule note: `AE-SEC-002` (secrets package) already scans MCP config files for raw secret values. This slice is complementary (pinning / origin / auth, not secret values). All fixtures must use `${VAR}` env-references for credential-shaped values so `AE-SEC-002` does not also fire on them.

## MCP config model

```go
package mcp

type Server struct {
    Name    string
    Command string            // stdio
    Args    []string          // stdio
    Env     map[string]string // stdio
    Type    string            // "http" | "sse" | "" (stdio)
    URL     string            // remote
    Headers map[string]string // remote
    Line    int               // 1-based best-effort source line of the server key; 0 if unknown
}

type ConfigFile struct {
    Path    string // repo-relative, forward-slash
    Servers []Server
}

func (s Server) IsRemote() bool // URL != "" || Type == "http" || Type == "sse"
```

Discovered files (tracked inventory only): `.mcp.json`, `mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json`. The discovery set must stay consistent with `AE-SEC-002`'s MCP target list (future drift-guard candidate, like `agentcontext`).

## charter.yaml allowlist contract

```yaml
mcp:
  trustedRemotes:
    - mcp.asana.com
    - api.example.com
```

- optional; absence yields an empty allowlist and no error
- `mcp.trustedRemotes` is a list of hostnames (no scheme, no path)
- read by `internal/config.LoadTrustedRemotes(root, inv) ([]string, error)`; a present-but-malformed file fails fast with a wrapped error

## Per-rule detection semantics

| Rule | Severity | OWASP | Detects | Location |
|---|---|---|---|---|
| AE-MCP-001 | High | MCP04 | runner package spec (`npx`/`bunx`/`uvx`/…) that is unpinned: `@latest`, missing version, dist-tag, semver range (`^ ~ >= > < * x`), or floating git ref (`github:` / `git+` / `#branch`) | config file + server-key line |
| AE-MCP-002 | High | MCP09 | remote server whose URL host is absent from `charter.yaml mcp.trustedRemotes`; when no allowlist exists, all non-local remotes are flagged as unverifiable | config file + server-key line |
| AE-MCP-003 | High | MCP07 | remote server (non-local) with no auth header (`Authorization`, `X-Api-Key`, `Api-Key`, `X-Auth-Token`) | config file + server-key line |

- Pinned = exact semver (`1.2.3`, optionally `v`/prerelease) or a digest (`sha256:…` / 40-hex). A dynamic version (`pkg@${VER}`) cannot be verified and is treated as unpinned.
- A package spec is the first non-flag arg of a recognized runner command; non-runner commands (`node`, `python3`, absolute paths) carry no package-pin assertion in v1.
- Localhost remotes (`localhost`, `127.0.0.1`, `::1`, `0.0.0.0`, `*.localhost`) are exempt from AE-MCP-002 and AE-MCP-003.
- A remote whose `url` is a bare env-reference (`${API_URL}`) yields no parseable host and is skipped (cannot be verified statically).

## Architecture / ownership

- `internal/config/` (new) — `charter.yaml` loader; reusable for future config/profiles; depends only on `repository` + `yaml.v3`.
- `internal/rules/mcp/` (new) — `config.go` (parse → model, pure), `mcp001.go` / `mcp002.go` / `mcp003.go` (pure checks over `[]ConfigFile`), `mcp.go` (`Run` discovers files, loads allowlist, runs checks). No rule package imports another rule package.
- `internal/doctor/run.go` — calls `gomcp.Run(root, inv)` after the CI rules and propagates its error (mirrors `gosecrets.RunSecretRules`).
- `internal/findings/` — unchanged; MCP findings reuse `Locations` and leave `Cap` zero.

Avoid: a generic config framework, a plugin/registry for rules, speculative interfaces, or `mcp.yml`/Pkl parsing in v1.

## Go alignment (per golang-patterns / golang-testing)

- Pure functions take data and return concrete types (`[]findings.Finding`); the filesystem touch is isolated in `Run`, so checks unit-test without disk or git.
- Errors are values: `Run`, `config.LoadTrustedRemotes`, and `parseConfigFile` return wrapped errors (`%w`); `run.go` propagates; nothing is discarded with `_`. Fail fast on malformed config.
- No package-level mutable state (compiled regexes and runner/auth-name sets are read-only).
- Deterministic output: each check sorts its findings; discovery is sorted.
- Testing: table-driven `t.Run` subtests, `t.Helper()` helpers, `t.TempDir()` fixtures, `t.Parallel()` for independent units, a `FuzzParseConfigFile` target on the untrusted-JSON boundary, `-race` via `moon run :test`, ≥85% line coverage for `internal/rules/mcp`.

## Testing strategy

- unit: `classifyPackageSpec` (table-driven across pinned/latest/range/no-version/git-ref/dynamic), `packageTokenFromArgs`, `remoteHost`, `isLocalHost`, `hasAuthHeader`
- rule: `checkPinning` / `checkTrustedRemotes` / `checkRemoteAuth` assert finding `RuleID`, `Severity`, and `Locations[0].{Path,Line}`
- parser: stdio + remote + `servers` alias; `FuzzParseConfigFile` (no panic on junk input)
- integration: fixture repos under `testdata/repos/` exercised through `Run` (mirror the existing fixture/git strategy used by `pass-slice1`)
- dogfood: Charter keeps tracked MCP config absent, so `charter doctor` emits no MCP findings and the score stays 100
- the repo quality gate (`moon run :check`) stays green

## Risks

- **False positives on legitimate dist-tags / dynamic versions** — control: evidence names the exact offending spec; severity is High (not Blocker); users allowlist intent via pinning or `charter.yaml`.
- **`charter.yaml` parsing introduces the first YAML dependency** — control: minimal struct decode, latest pinned `yaml.v3`, recorded in ADR-0011.
- **MCP `2026-07-28` OAuth changes** — control: AE-MCP-003 is presence-based, not field-specific; a re-validation follow-up is tracked.
- **Discovery drift vs AE-SEC-002 MCP targets** — control: documented as a future drift-guard; discovery list kept explicit and small.
- **Noisy "no allowlist" output** — control: when no `charter.yaml` exists, AE-MCP-002 emits one finding per remote with a distinct "cannot verify origin" summary so the remediation (add an allowlist) is obvious.

## Success criteria

- `charter doctor` on a repo with an unpinned MCP server emits `AE-MCP-001` (High) with a `path:line` location; with an untrusted remote, `AE-MCP-002`; with an unauthenticated remote, `AE-MCP-003`.
- `charter.yaml mcp.trustedRemotes` suppresses AE-MCP-002 for listed hosts; localhost remotes never fire AE-MCP-002/003.
- a malformed `.mcp.json` or `charter.yaml` fails fast with a clear wrapped error (exit 2).
- the three `AE-MCP-*` specs are as-built; ADR-0011 is referenced; fixtures and the architecture-doc examples match real output.
- `moon run :check` is green and the dogfood `charter doctor` score is unchanged (100).

## References

- `docs/internal/decisions/0011-mcp-config-scanning.md`
- `docs/internal/decisions/0009-finding-location.md`, `0006-latest-docs-first.md`, `0004-contract-first-interfaces.md`, `0001-offline-first-core.md`
- `docs/internal/architecture/charter-architecture-2026.md` (§ rule catalog, M1.3 T1.3.1, output examples)
- `docs/internal/specs/AE-MCP-001.md`, `AE-MCP-002.md`, `AE-MCP-003.md`
- `docs/internal/superpowers/specs/2026-05-31-phase-1-slice-4-design.md`
- `docs/internal/superpowers/plans/2026-05-31-phase-1-slice-5.md` (derived implementation plan)
