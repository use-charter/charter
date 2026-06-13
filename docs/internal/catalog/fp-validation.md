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

## Results — run 2 (2026-06-14, broadened pool)

- Catalog version under test: `2026.06.02` (unchanged); Charter built from `main` (`d8314126`).
- Method identical to run 1: real committed MCP configs fetched from public GitHub via `gh api`, each scanned in an isolated git repo (committed, no `charter.yaml` — catalog baseline / new-user case).
- Repos scanned: **12**, none overlapping run 1, spanning all four config filenames (`.mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json`, `.gemini/settings.json`) and a vendor-diverse mix (Automattic, EPAM, NG-ZORRO, Vercel-host, Atlassian-host, Figma-host, Svelte-host).

| # | Repo | Config | Finding | Class | Note |
|---|------|--------|---------|-------|------|
| 1 | Automattic/wp-calypso | `.mcp.json` | AE-MCP-001 `playwright` | TP | Unpinned npx package. |
| 2 | sevos/omablog | `.mcp.json` | AE-MCP-001 `chrome-devtools-mcp@latest` | TP | `@latest`. `tidewave` SSE `localhost:3000` correctly **not** flagged (TN). |
| 3 | PaulRBerg/next-template | `.mcp.json` | AE-MCP-001 `@upstash/context7-mcp` (no ver) + `next-devtools-mcp@latest` | TP×2 | Unpinned. |
| 4 | epam/ai-dial-ui-kit | `.mcp.json` | — | TN | `mcp.atlassian.com` + `mcp.figma.com` trusted hosts; `node` local cmd → correctly silent. |
| 5 | inspec/inspec | `.vscode/mcp.json` | — | TN | `servers` (VS Code alias) parsed; `mcp.atlassian.com` trusted → silent. |
| 6 | un-pany/v3-admin-vite | `.cursor/mcp.json` | — | TN | `localhost:3333` → internal, correctly silent. |
| 7 | c15t/c15t | `.cursor/mcp.json` | — | TN | `mcp.vercel.com` trusted → silent. |
| 8 | jasonjgardner/blockbench-mcp-plugin | `.vscode/mcp.json` | — | TN | `localhost:3000` → internal, silent. |
| 9 | NG-ZORRO/ng-zorro-antd | `.gemini/settings.json` | — | **MISS** | `@angular/cli` (unpinned) + `@eslint/mcp@latest` went undetected — **`.gemini/settings.json` is not a recognized config path** (see coverage gap below). False **negative**, not FP. |
| 10 | itswadesh/svelte-commerce | `.gemini/settings.json` | — | (not evaluated) | Same Gemini gap; would otherwise prompt verify on non-cataloged `mcp.svelte.dev`. |
| 11 | syscoin/sys-claude | `.mcp.json` | AE-MCP-001 ×4 (`@upstash/context7-mcp@latest`, `@playwright/mcp@latest`, `context-mode@latest`, `memsearch-mcp@latest`) | TP×4 | All `@latest`. |
| 12 | JCouce/poe-assistant | `.mcp.json` | — | TN | `uv --directory mcp-poe-tools run` → local server, no registry pin → correctly silent. |

- Total AE-MCP-001/002 findings emitted: **8**  ·  False positives: **0**  ·  **FP rate: 0%.**
- **Gate: PASS** (≤ 10%). All 8 emitted findings are legitimate unpinned packages; every trusted-host and local/loopback target was correctly silent (5 true negatives confirming the trustedHosts + internal-address exemptions in the wild).

### Coverage gap found this run (detection, NOT a false positive)
- **`.gemini/settings.json` is not scanned.** `internal/rules/mcp/mcp.go` recognizes `.mcp.json`, `mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json` — but not Gemini CLI's `.gemini/settings.json` (`mcpServers` key). Repo 9 had two unpinned packages that went undetected solely because of the path. This **under-reports** (fail-safe per ADR-0021 — never misreports), but it contradicts the cross-vendor claim ("Claude Code / Cursor / Copilot / Gemini / Windsurf"). **Recommended fix:** add `.gemini/settings.json` to the recognized MCP config paths (and consider `.vscode/mcp.json` is already covered). Low-risk — the parser already reads `mcpServers`/`servers`. Tracked as a launch-eve recommendation; not a gate blocker.
- **Candidate trusted-host:** `mcp.svelte.dev` (official Svelte docs MCP) appeared as a legit vendor remote — consider adding to `trustedHosts` after the Gemini path lands so it can be validated end-to-end.
- No new deprecated/archived packages or CVEs observed; no advisory additions required this run.

## AE-MCP-003 observation — RESOLVED (CF-13)
AE-MCP-003 ("remote server declares no auth metadata") fired on **every** OAuth-based vendor remote server configured without a static `Authorization` header (sentry, context7, atlassian, openfort). For modern OAuth 2.1 remote servers, auth is declared via the OAuth flow, not a config header — a systematic FP. **Fixed:** `checkRemoteAuth` now exempts catalog `trustedHosts` (known OAuth vendor servers), so these no longer flag. Non-catalog remotes without auth still flag.

## Sign-off

- [x] FP rate ≤ 10% recorded above (10% → 0% after fix).
- [x] The FP has a rule fix landed and re-verified.
- [x] Real CVE advisories added for a cataloged package (T1.6.2).
- [x] CF-13 (AE-MCP-003 OAuth FP) resolved — catalog OAuth hosts exempt.
- [x] Broaden the run to more repos (run 2, 2026-06-14): **12 more repos, 0% FP, gate PASS**. Surfaced one detection-coverage gap (`.gemini/settings.json` unsupported) — recorded above as a launch-eve recommendation, not a blocker (fail-safe under-report per ADR-0021).
- [ ] Apply the `.gemini/settings.json` config-path fix (cross-vendor coverage) + consider `mcp.svelte.dev` trustedHost — TDD, then re-run run-2 repo 9 to confirm the two unpinned packages fire.
- [ ] Final founder sign-off at the Slice 17 release gate.
