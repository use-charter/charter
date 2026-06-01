# MCP Catalog — False-Positive Validation (T1.6.3)

> **⚑ FOUNDER / LAUNCH GATE.** This validation **blocks the public tag (Slice 17)**.
> Until it is completed with a recorded ≤ 10% false-positive rate, the catalog
> ships as an engine + seed only (carry-forward CF-12).

## Why

The catalog makes `AE-MCP-001`/`AE-MCP-002` re-fire on previously-passing repos.
If it produces noise on legitimate servers, early adopters mute Charter — the
exact "rules too noisy" failure the architecture's validation analysis warns
about. This gate proves the catalog is trustworthy before it reaches users.

The design already bounds the risk (ADR-0021): behind-stable findings are
informational (non-deducting), and comparison is exact-match (a stale catalog
under-reports). This gate validates the **HIGH** signals — deprecated packages
and CVE advisories — plus the AE-MCP-002 host baseline.

## Method

1. Select **5+ real public repos** with committed MCP configs (`.mcp.json`,
   `.cursor/mcp.json`, etc.), spanning vendors and the archived-package set.
2. Build Charter at the candidate catalog version and scan each:
   `charter doctor --path <repo> --format json`.
3. For **every** `AE-MCP-001` and `AE-MCP-002` finding, classify as
   **true positive** or **false positive** and record the reasoning below.
4. Compute `FP rate = false positives / total MCP findings`. Target **≤ 10%**.
5. Fix the catalog (or rule) for any FP, re-run, and re-record until green.

## Results — run 1 (2026-06-02)

- Catalog version under test: `2026.06.02`
- Reviewer: founder (agent-assisted); method: real `.mcp.json` files fetched from public GitHub repos via `gh`, each scanned in an isolated git repo with no `charter.yaml` (catalog baseline only — the new-user case).
- Repos scanned: **11** (getsentry/spotlight, openfort-xyz/agent-skills, HaiBang1010/Portfolio, trezero/archon-trinity, Sternrassler/evolution, Enriquegonzalezz/app-control-prenatal, radesh20/ai-exception-system, frozo-ai/frozo-tradingview-mcp, chrisranderson/ai-coding-demo, chrisstoy/mockingbird, Faishal24/learn-vue-laravel).

Scope of the gate is **AE-MCP-001 + AE-MCP-002** (the catalog rules). AE-MCP-003 is recorded separately below.

| # | Repo | Finding | Class | Note |
|---|------|---------|-------|------|
| 2 | openfort-xyz/agent-skills | AE-MCP-002 `www.openfort.io` | TP | Niche, non-cataloged origin → correct "verify/allowlist" prompt (Charter can't bless every company). |
| 3 | HaiBang1010/Portfolio | AE-MCP-001 `shadcn@latest` | TP | Unpinned `@latest`. |
| 4 | trezero/archon-trinity | AE-MCP-002 `172.16.1.230` | **FP → fixed** | Private RFC1918 LAN address is internal, not a public shadow origin. Fixed: `isLocalHost` now exempts private/link-local/internal addresses. |
| 5 | Sternrassler/evolution | AE-MCP-001 `serena` `git+https://…` | TP | Floating git ref. |
| 6 | Enriquegonzalezz/app-control-prenatal | AE-MCP-001 `@supabase/mcp-server-supabase@latest` | TP | Unpinned `@latest`. |
| 7 | radesh20/ai-exception-system | AE-MCP-001 `@modelcontextprotocol/server-slack` (archived) | TP | **Catalog deprecation flag firing on a real archived package in the wild** — the engagement loop working. |
| 7 | radesh20/ai-exception-system | AE-MCP-001 `@gongrzhe/server-gmail-autoauth-mcp` | TP | Unpinned (no version). |
| 9 | chrisranderson/ai-coding-demo | AE-MCP-001 `mcp-server-git`, `mcp-server-time` | TP | Unpinned `uvx` servers. |
| 10 | chrisstoy/mockingbird | AE-MCP-001 `playwright` | TP | Unpinned. |

**Catalog wins (true negatives worth noting):** `mcp.sentry.dev`, `mcp.context7.com`, and `mcp.atlassian.com` all passed AE-MCP-002 on the catalog baseline with no `charter.yaml` — the trusted-host expansion working in the wild. Repos 1, 8, 11 had no AE-MCP-001/002 findings.

- Total AE-MCP-001/002 findings: **10**  ·  False positives: **1** (pre-fix)  ·  **FP rate: 10% → 0% after the fix.**
- **Gate: PASS** (≤ 10%; 0% after the private-address fix).

### Fix landed this run
- **AE-MCP-002/003 — private/internal address exemption** (`internal/rules/mcp/mcp002.go`, `isLocalHost`): loopback, RFC1918 private, link-local, the unspecified address, and reserved internal TLDs (`.localhost`, `.local`, `.internal`) are now treated as local — a LAN/internal server is not a public shadow origin. Eliminated the one FP.

### Refresh landed this run (T1.6.2)
- Added **real, verified CVEs** for `mcp-server-git` (CVE-2026-27735 / CVE-2025-68145 / CVE-2025-68143, all CWE-22 path traversal; fixes `2026.1.14` / `2025.12.18` / `2025.9.25`), using the new `affectedBelow` range matcher. Verified: a repo pinning `mcp-server-git@2025.8.0` now fires AE-MCP-001 HIGH → CVE-2026-27735.

## AE-MCP-003 observation — RESOLVED (CF-13)
AE-MCP-003 ("remote server declares no auth metadata") fired on **every** OAuth-based vendor remote server configured without a static `Authorization` header (sentry, context7, atlassian, openfort). For modern OAuth 2.1 remote servers, auth is declared via the OAuth flow, not a config header — a systematic FP. **Fixed:** `checkRemoteAuth` now exempts catalog `trustedHosts` (known OAuth vendor servers), so these no longer flag. Non-catalog remotes without auth still flag.

## Sign-off

- [x] FP rate ≤ 10% recorded above (10% → 0% after fix).
- [x] The FP has a rule fix landed and re-verified.
- [x] Real CVE advisories added for a cataloged package (T1.6.2).
- [x] CF-13 (AE-MCP-003 OAuth FP) resolved — catalog OAuth hosts exempt.
- [ ] Broaden the run to more repos over time; continue version-data curation (CF-12) before the public tag (Slice 17).
- [ ] Final founder sign-off at the Slice 17 release gate.
