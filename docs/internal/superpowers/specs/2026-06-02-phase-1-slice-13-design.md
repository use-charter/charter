# Phase 1 Slice 13 Design — MCP Catalog v1

## Goal

Ship the static, founder-curated **MCP Catalog v1** (architecture M1.6) so `AE-MCP-001`/`AE-MCP-002` reference a versioned list of known servers and **re-fire on repos that previously passed** when a dependency is archived, receives a CVE, or releases a new stable version. Per ADR-0021 the catalog flags three things — **deprecated/archived packages** (HIGH), **advisory-affected pinned versions** (HIGH), and **behind-stable pins** (informational) — and supplies a **trusted vendor-host baseline** for `AE-MCP-002`. Staleness is made *safe*: exact-match comparison only, so a lagging catalog stays silent rather than wrong.

## Audience

- coding agents implementing this slice
- maintainers reviewing the severity model (why version-lag is informational, security is HIGH) and the zero-FP guarantee under catalog staleness
- the founder, who owns catalog curation, the CVE-advisory process (T1.6.2), and the pre-ship FP gate (T1.6.3)

## Scope

### In scope

- `internal/catalog`: embedded curated `data/catalog.yaml` (`//go:embed`); `Default()` (parse-once), `Parse([]byte)` (for tests); types (`Catalog`/`ServerEntry`/`Advisory`); lookups (`Lookup`, `AdvisoryFor`, `KnownBehind`, `TrustedHostSet`); a **catalog-validity test** (the curation contract)
- `internal/rules/mcp/mcp001.go`: `checkPinning(files, cat)` becomes catalog-aware with a one-finding-per-server precedence ladder (deprecated > unpinned > advisory > behind-stable > clean)
- `internal/rules/mcp/mcp.go`: merge `catalog.TrustedHostSet()` into the `AE-MCP-002` allowlist baseline
- `internal/rules/catalog`: broaden `AE-MCP-001`/`AE-MCP-002` metadata descriptions to mention catalog currency
- new testdata fixtures + facet unit tests (deprecated, advisory, behind-stable, catalog-trusted-host); zero-FP clean-fixture assertions preserved
- docs: ADR-0021 (done), this spec + plan, `AE-MCP-001`/`AE-MCP-002` spec updates, audit-checklist update, T1.6.2 contribution/CVE process doc, T1.6.3 FP-validation skeleton, architecture M1.6/T1.6.1 + rule-table reconciliation, carry-forward update

### Out of scope

- the `AE-MCP-001` *fixer* (auto-pin to stable / rewrite deprecated → successor) — CF-6, deferred to `charter fix` v1.1 (catalog now supplies the target)
- network/registry lookups, live CVE feeds, automated advisory monitoring (Phase 2); `AE-MCP-003` changes; YAML/Pkl MCP config scanning (ADR-0011 still JSON-only)
- the actual founder curation refresh + FP-validation run (a `⚑ FOUNDER` launch-gate task, Slice 17) — this slice ships the engine + seed + process, not the validated final dataset

## Why this slice

M1.6 is the architecture's named remedy for the "one-time setup trap" ("Some installs, zero stranger issues → Accelerate the static MCP catalog so new findings fire on repos that previously passed"). Grounding (2026-06) showed the catalog's durable value is **deprecation detection** (the official reference set shrank to 7; ~10 popular `@modelcontextprotocol/server-*` packages were archived to vendor/community successors) and a **vendor-host allowlist**, not fragile version-lag scoring.

## Grounding (verified against the code & ecosystem)

- `internal/rules/mcp`: `checkPinning(files)` resolves a runner package token (`packageTokenFromArgs`), `classifyPackageSpec` → `(name, version, pinned)`; unpinned → `AE-MCP-001` HIGH. `checkTrustedRemotes(files, allow)` flags non-local remotes absent from `allow` (distinct "no allowlist" message when `allow` empty). `Run` reads JSON MCP configs, `config.LoadTrustedRemotes`, runs the three checks. `findings.Finding` has `Informational bool` (excluded from scoring — `internal/scoring/score.go` skips it; precedent `AE-SUPPRESS-003`).
- Ecosystem facts (sources: `modelcontextprotocol/servers` README, npm registry, MCP ecosystem reference 2026):
  - **CalVer, not semver:** `@modelcontextprotocol/server-filesystem` → `2026.1.14` (prev `2025.12.18`, `2025.11.25`). A `<`/`>` comparator is unsafe across schemes → **exact-match only**.
  - **Active reference (7):** `filesystem` (npm), `memory` (npm), `sequentialthinking` (npm), `everything` (npm, test-only), `fetch`/`git`/`time` (pypi, `uvx`).
  - **Archived → successor:** `server-github`→`github/github-mcp-server`; `server-gdrive`→`taylorwilsdon/google_workspace_mcp`; `server-slack`→`zencoderai/slack-mcp-server`; `server-postgres`→`crystaldba/postgres-mcp`; `server-sqlite`→community fork; `server-puppeteer`→`microsoft/playwright-mcp`; `server-brave-search`→`brave/brave-search-mcp-server`; `server-redis`→`redis/mcp-redis`; `server-aws-kb-retrieval`→AWS Labs `bedrock-kb-retrieval-mcp-server`; (`server-gitlab` archived).
  - **Vendor hosts (OAuth 2.1):** `api.githubcopilot.com`, `mcp.sentry.dev`, `mcp.linear.app`, `mcp.atlassian.com`, `mcp.notion.com`, `mcp.stripe.com`, `mcp.vercel.com`, `mcp.asana.com`.

## `internal/catalog` (new)

```go
type Catalog struct {
    Version      string         // catalog data version, e.g. "2026.06.02"
    Generated    string         // ISO date the seed/refresh was compiled
    Servers      []ServerEntry
    TrustedHosts []string       // lowercase hostnames
}
type ServerEntry struct {
    Package       string     // registry spec name, e.g. "@modelcontextprotocol/server-filesystem"
    Ecosystem     string     // "npm" | "pypi"
    Status        string     // "active" | "deprecated"
    StableVersion string     // exact; "" when deprecated
    KnownVersions []string   // ascending; last == StableVersion for active entries
    Successor     string     // migration target; required when deprecated
    Reference     string     // upstream URL
    Advisories    []Advisory
}
type Advisory struct {
    ID        string   // "CVE-…" / "GHSA-…"
    Affected  []string // exact versions
    FixedIn   string
    Severity  string   // "high" (v1)
    Summary   string
    Reference string
}
```

- `Default() *Catalog` — `sync.Once`-parses the embedded `data/catalog.yaml`; panics only on a malformed embed (caught by the validity test, so unreachable in shipped builds — never a `doctor` runtime error). Returns a read-only shared pointer.
- `Parse([]byte) (*Catalog, error)` — for tests to build inline catalogs (keeps facet tests independent of curated data).
- `(*Catalog) Lookup(pkg string) (ServerEntry, bool)` — exact package-name match.
- `(ServerEntry) AdvisoryFor(version string) (Advisory, bool)` — exact membership in any `Advisory.Affected`.
- `(ServerEntry) KnownBehind(version string) (stable string, behind bool)` — `behind` iff `version ∈ KnownVersions`, `version != StableVersion`, and no advisory covers it.
- `(*Catalog) TrustedHostSet() map[string]struct{}` — lowercased `TrustedHosts`.
- **Validity test** (`catalog_test.go`) asserts on `Default()`: non-empty; unique `Package`; each `Ecosystem ∈ {npm,pypi}`; `TrustedHosts` all lowercase + de-duplicated; for `active` — `StableVersion != ""`, `KnownVersions` non-empty + ascending-unique, `StableVersion == last(KnownVersions)`; for `deprecated` — `Successor != ""`, `StableVersion == ""`; every `Advisory` has `ID`, non-empty `Affected`, `FixedIn`, `Severity == "high"`; `Version`/`Generated` set.

## `AE-MCP-001` precedence ladder (`mcp001.go`)

`checkPinning(files []ConfigFile, cat *catalog.Catalog) []findings.Finding`. Per server with a resolvable registry token `name@version`:

1. `entry, ok := cat.Lookup(name)`.
2. **Deprecated** — `ok && entry.Status=="deprecated"` → **HIGH**: *"MCP server package `<name>` is archived/deprecated — migrate to `<successor>` (supply-chain maintenance risk)."* (fires even if unpinned; migration supersedes pinning.) → next server.
3. **Unpinned** — `!pinned` → existing **HIGH** (unchanged message). → next server.
4. **Advisory** — `ok`, pinned, `adv, hit := entry.AdvisoryFor(version)` → **HIGH**: *"MCP server `<name>@<version>` is affected by `<adv.ID>` — fixed in `<adv.FixedIn>` (`<adv.Summary>`)."* → next server.
5. **Behind-stable** — `ok`, pinned, `stable, behind := entry.KnownBehind(version)` → **informational** (`Informational:true`): *"MCP server `<name>` is pinned to `<version>`; catalog stable version is `<stable>` — upgrade available."* → next server.
6. else clean.

Severity/category unchanged (`MCP Safety`); `Evidence` carries `path: server <name> <token>` + the catalog detail; `Locations[0]` = server-key line. Deterministic sort retained. `cat` is injected (tests pass an inline catalog; `Run` passes `catalog.Default()`).

## `AE-MCP-002` baseline allowlist (`mcp.go`)

In `Run`, compute `allow = union(userTrustedRemotes, catalog.Default().TrustedHosts)` (case-folded, de-duplicated) and pass to `checkTrustedRemotes`. Effect: recognized vendor hosts pass without per-repo config (FP reduction); unknown remotes still flag; `charter.yaml` stays the override. `checkTrustedRemotes` signature/severity unchanged; its summary text generalized to "trusted-remote allowlist (MCP catalog or charter.yaml mcp.trustedRemotes)". The empty-`allow` "no allowlist configured" branch remains for the direct unit test but is unreachable in prod (catalog is non-empty).

## Architecture / ownership

- `internal/catalog/` (new): data + parse-once loader + pure lookups + validity test. No network, no LLM, no new runtime dependency (reuses `gopkg.in/yaml.v3`).
- `internal/rules/mcp/`: `mcp001.go` catalog-aware; `mcp.go` merges trusted hosts. `mcp002.go`/`mcp003.go` logic unchanged.
- `internal/rules/catalog/`: metadata description text only.
- Avoid: cross-scheme version ordering; runtime catalog errors; double-reporting a server; making the catalog a network/Phase-2 dependency; touching `AE-MCP-003`.

## Go alignment

- Pure lookups; `Parse` wraps errors `%w`; `Default()` parse-once via `sync.Once`. Deterministic ordering preserved. Embedded data via `//go:embed data/catalog.yaml`. Tests: catalog-validity (curation contract), `Parse`/lookup/advisory/known-behind units (inline catalogs), `mcp001` facet matrix, `mcp.go` integration (trusted-host pass), `-race`; ≥85% for `internal/catalog`.

## Testing & verification strategy

- **catalog-validity:** the contract test above on `Default()` (fails the build if curation is malformed).
- **catalog units (inline `Parse`):** `Lookup` hit/miss; `AdvisoryFor` exact-match (affected hits, unaffected/newer misses); `KnownBehind` (in-list non-stable → behind; stable → not; absent → not; advisory-covered → not behind, advisory wins); `TrustedHostSet` lowercasing.
- **`AE-MCP-001` facets (inline catalog):** deprecated package → HIGH + successor in message (even when `@latest`); advisory version → HIGH + `id`/`fixed_in`; behind-stable → informational, `Informational==true`, **does not** deduct (scoring test); stable pin → clean; version absent from `known_versions` → clean (staleness-safe); precedence (deprecated beats unpinned; advisory beats behind).
- **`AE-MCP-002` integration:** a remote whose host ∈ catalog `TrustedHosts` passes with **no** `charter.yaml`; an unknown remote still flags. Existing `mcp_test.go`/`remote_test.go` stay green (asana flows unaffected; `checkTrustedRemotes(_, nil)` direct test unchanged).
- **Zero-FP regression:** `pass-mcp-clean` (pins `…server-filesystem@1.0.4`, absent from `known_versions`) stays clean; `TestRunCleanNoFindings`/`TestCheckPinningAllPinned` pass (update their `checkPinning` call sites to pass `catalog.Default()`).
- **Dogfood:** Charter's own `charter doctor` stays **100** (Charter has no `.mcp.json`); a synthetic repo pinning `@modelcontextprotocol/server-github@<v>` now scores lower via the new deprecated-HIGH (proves the recurring-engagement loop).
- `moon run :check` green.

## Risks

- **Catalog staleness emits wrong findings** — control: exact-match only; behind-stable is informational (no score impact); a version absent from `known_versions` is silent; deprecation/host data ages well. Net: a stale catalog under-reports (safe), never over-reports.
- **Version-lag HIGH would tank scores across repos on stale data (Commitment #9)** — control: ADR-0021 makes version-lag informational; only deprecation/CVE (durable, high-value) deduct.
- **Double-reporting a server** — control: one-finding-per-server precedence ladder; facet test asserts a single finding.
- **Wrong/seed version data before launch** — control: package names + successors + hosts are accurate now; version fields flagged for the founder refresh + FP-validation (T1.6.3) gate before the public tag (carry-forward); the engine ships independent of final curation.
- **Curation burden** — control: documented T1.6.2 process; `known_versions` need only cover plausibly-still-pinned recent versions (absent = silent), not full history.

## Success criteria

- `internal/catalog` ships with an embedded, validity-tested seed (active reference set + archived-package advisories with successors + vendor trusted-host baseline).
- `AE-MCP-001` flags deprecated (HIGH) and advisory-affected (HIGH) pins and nudges behind-stable (informational, non-deducting), one finding per server; `AE-MCP-002` passes catalog vendor hosts by default.
- Clean fixtures stay zero-FP; Charter's own scan stays 100; all existing MCP tests green.
- ADR-0021 + this spec + the plan committed; `AE-MCP-001`/`AE-MCP-002` specs, audit checklist, architecture M1.6/T1.6.1 + rule-table, and the T1.6.2/T1.6.3 docs updated; carry-forward annotated; `moon run :check` green.

## References

- `docs/internal/decisions/0021-mcp-catalog-v1.md` (ADR-0021); `0011` (MCP config scanning), `0008` (score), `0013` (`Informational`/`AE-SUPPRESS-003`), `0005`/`0020` (`fix` deferral, CF-6)
- `docs/internal/architecture/charter-architecture-2026.md` (§M1.6 T1.6.1–T1.6.3; `AE-MCP-001`/`AE-MCP-002` rule table)
- `internal/rules/mcp/{mcp001.go,mcp002.go,mcp.go,config.go}`; `internal/findings/finding.go`; `internal/scoring/score.go`
- `docs/internal/superpowers/plans/2026-06-02-phase-1-slice-13.md` (derived implementation plan)
