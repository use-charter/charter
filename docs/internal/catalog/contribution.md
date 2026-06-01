# MCP Catalog — Contribution & CVE-Advisory Process (T1.6.2)

The MCP catalog is **founder-curated** static data that powers the catalog-aware
facets of `AE-MCP-001` (deprecated / advisory / behind-stable) and the
`AE-MCP-002` trusted-host baseline. It is offline, embedded at build time, and
validated by a test — see ADR-0021 and `docs/internal/decisions/0021-mcp-catalog-v1.md`.

## Where the data lives

- Data: `internal/catalog/data/catalog.yaml` (single file, embedded via `//go:embed`).
- Loader + lookups: `internal/catalog/catalog.go`.
- Curation contract (the gate): `TestCatalogValid` in `internal/catalog/catalog_test.go`.

A malformed catalog is a **build/test failure**, never a `charter doctor` runtime
error. Always run `mise exec -- go test ./internal/catalog` after editing.

## Entry schema

```yaml
trustedHosts:            # lowercase hostnames; AE-MCP-002 baseline allowlist
  - api.githubcopilot.com

servers:
  - package: "@scope/name"   # exact registry spec (npm) or uvx/pip token (pypi)
    ecosystem: npm           # npm | pypi
    status: active           # active | deprecated
    stableVersion: "2026.1.14"          # active only; == last knownVersions
    knownVersions: ["2025.12.18", "2026.1.14"]  # ascending; recent recognized versions
    successor: "owner/repo"  # deprecated only; the migration target (required)
    reference: "https://…"   # upstream repo/docs
    advisories:              # optional; REAL CVE/GHSA IDs only
      - id: "CVE-2026-12345"
        affected: ["1.0.0", "1.0.1"]   # exact versions
        fixedIn: "1.0.2"
        severity: high
        summary: "short description"
        reference: "https://…"
```

**Exact-match invariant (ADR-0021):** Charter never orders versions across
schemes. A pinned version is flagged only if it *exactly* matches an advisory's
`affected` entry or is a non-last member of `knownVersions`. A version not listed
is silent. So `knownVersions` need only cover versions still plausibly pinned in
the wild — not full history.

## How to…

### Add an active server
Add an entry with `status: active`. Include `stableVersion` + `knownVersions`
only if you want the behind-stable nudge; otherwise omit both (the entry then
documents a known package and is a home for future advisories).

### Mark a package deprecated/archived
Set `status: deprecated`, add a `successor` (required), drop any version fields.
This is the catalog's highest-value, staleness-proof signal — archival is a fact.

### File a CVE/GHSA advisory
Add an `advisories[]` entry under the affected package with a **real** `id`
(`CVE-…` / `GHSA-…`), exact `affected` versions, `fixedIn`, `severity: high`, a
one-line `summary`, and a `reference`. **Never invent advisory IDs** — a security
tool must not ship fabricated CVEs.

### Add a trusted remote host
Append the lowercase hostname to `trustedHosts` (vendor-operated OAuth endpoints).

## SLA & roadmap

- **Phase 1 (now):** founder-owned. Target a **48h SLA** to land a catalog entry
  for a newly-disclosed CVE against a cataloged server.
- **Phase 2:** community-contributed catalog PRs + automated advisory monitoring
  (e.g. OSV/GHSA feeds) that open PRs against this file.

## Before a public release

Run the pre-ship FP gate — `docs/internal/catalog/fp-validation.md` (T1.6.3) —
and refresh version data to a verified ~20–30 server set. This is a launch gate
(carry-forward CF-12 / Slice 17).
