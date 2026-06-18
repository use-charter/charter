# Founder Dashboard Analytics — Implementation Plan

**Document type:** Implementation plan (HOW, not WHAT)
**Based on:** `docs/internal/superpowers/specs/2026-06-19-dashboard-analytics.md`
**Decision record:** [ADR-0027](../../decisions/0027-founder-dashboard-analytics.md)
**Status:** Locked & Validated — grilled, doc-grounded (Wrangler v4 / Workers testing verified 2026-06-19)
**Branch:** `main`; `moon run :check` green before every push; `infra/**` pushes auto-deploy the router via `.github/workflows/deploy-workers.yml`

---

## Architecture recap

Collection lives in the existing **`charter-router`** Worker (fronts `use-charter.dev/*`). New resources: D1 db `charter_analytics` + KV namespace `ANALYTICS_SALT` (daily salt only). Site stays static (ADR-0026). Six metrics server-side; install-copy via one client beacon. Privacy = random daily salt in KV deleted ~48h; raw IP/UA never persisted.

---

## Cross-cutting invariants (apply to every phase)

1. **Collection must never affect serving.** All KV/D1 work runs inside `ctx.waitUntil(...)` wrapped in `try/catch`; a collection failure is swallowed (optionally `console.warn`). The response is computed and returned independent of analytics.
2. **Single counting chokepoint.** `record()` is invoked once, *after* the response is resolved (landing **or** Mintlify proxy branch), never per-branch — so `/docs` HTML is counted for docs-reach while Mintlify assets are not. It reads only `response.status` + `response.headers` (never the body) and returns the original response untouched.
3. **Kill-switch.** A `[vars] ANALYTICS_ENABLED = "true"` gates `record()` and `/api/event`; flipping it to `"false"` (dashboard, no code change) disables collection instantly.
4. **Pure, edge-independent core.** `qualifies`, `normalizePath`, `visitorHash` take primitives (path, status, content-type, UA, `Sec-Purpose`, IP, salt, country) as arguments — because `request.cf` and `CF-Connecting-IP` are **not populated under local `wrangler dev`**, this keeps the logic unit-testable without the edge.
5. **Identity sources:** IP = `request.headers.get('CF-Connecting-IP')`; UA = `request.headers.get('User-Agent')`; country = `request.cf?.country` (guard `undefined`; skip the geo write if absent or `T1`/`XX`). Hash via `crypto.subtle.digest('SHA-256', …)` (no compat flag needed).
6. **No personal data at rest, ever** — raw IP/UA/query strings appear in no D1 row, no KV value, no log.

---

## Sequencing & deploy order (strict)

```
0a. (manual) wrangler d1 create + kv namespace create  → paste IDs into wrangler.toml
0b. commit Phase 0 (bindings + migration)  → infra push deploys router (no writes yet, safe pre-migration)
0c. (manual) wrangler d1 migrations apply charter_analytics --remote   ← BEFORE Phase 1 deploy
1.  collection code  → deploy (now writes land in existing tables)
2,4 → 3,5 ; 6 anytime ; 7 last
```
Deploying collection code **before** the migration is applied would 500 every D1 write (harmless to serving per invariant 1, but loses data) — hence 0c precedes Phase 1.

---

## Manual founder steps (require Cloudflare auth)

```bash
cd infra/router
wrangler d1 create charter_analytics                       # v4 syntax; copy database_id → wrangler.toml
wrangler kv namespace create ANALYTICS_SALT                # v4: "kv namespace create" (space, no colon); copy id → wrangler.toml
wrangler d1 migrations apply charter_analytics --remote    # after 0001 migration committed (use the DB *name*)
# local dev/tests: wrangler d1 migrations apply charter_analytics --local
```
`database_id` / KV `id` are non-secret → committed in `wrangler.toml` (mirrors ADR-0026 origin vars). Tooling versions verified 2026-06-19: pinned `wrangler` 4.100.0 already has these v4 commands (latest is 4.102.0 — bump optional); `@cloudflare/workers-types` latest `4.20260617.1`.

**Prerequisite to verify:** the Cloudflare Access application protects `/dashboard*`, so `/dashboard/api/analytics` is gated exactly like the existing `/dashboard/api/stats`.

---

## Testing approach

- **Unit (required, plain Vitest):** `normalizePath` (junk → `/__other__`, slash/case canonicalization), `qualifies` (rejects non-2xx, non-HTML, `/api`·`/dashboard`·Mintlify-asset paths, `Sec-Purpose: prefetch`, `isbot` UAs), `visitorHash` (stable within a day, differs across days, no IP/UA leakage). These are pure → no bindings needed.
- **Integration (best-effort):** event-guard + read-auth via `@cloudflare/vitest-pool-workers` (0.16.x — uses the `cloudflareTest()` plugin, **requires Vitest ^4.1**; per-test-file storage isolation; D1 schema loaded via `applyD1Migrations`/`readD1Migrations` — verify the `readD1Migrations` import path against the installed types, it moved in 0.16). If the harness setup is non-trivial, defer these to the Phase-7 live smoke rather than block.
- **Infra tooling:** add `vitest` (^4.1, matching web) + optionally `@cloudflare/vitest-pool-workers` to `infra/package.json`; add `infra/vitest.config.ts`; add a `test` task to `infra/moon.yml` and wire it into `:check`.
- TDD: failing test → implement, per repo norm.

---

## Phase 0 — Resources & schema

- **Task:** Add `[[d1_databases]]` (`binding="ANALYTICS_DB"`) + `[[kv_namespaces]]` (`binding="ANALYTICS_SALT"`) + `[vars] ANALYTICS_ENABLED="true"` to `infra/router/wrangler.toml`; extend the router `Env`. Author `infra/router/migrations/0001_init.sql`:
  ```sql
  CREATE TABLE IF NOT EXISTS pageview (day TEXT NOT NULL, path TEXT NOT NULL, hits INTEGER NOT NULL DEFAULT 0, PRIMARY KEY (day, path))    WITHOUT ROWID;
  CREATE TABLE IF NOT EXISTS visitor  (day TEXT NOT NULL, vhash TEXT NOT NULL,                                  PRIMARY KEY (day, vhash))   WITHOUT ROWID;
  CREATE TABLE IF NOT EXISTS geo      (day TEXT NOT NULL, country TEXT NOT NULL, hits INTEGER NOT NULL DEFAULT 0, PRIMARY KEY (day, country)) WITHOUT ROWID;
  CREATE TABLE IF NOT EXISTS event    (day TEXT NOT NULL, type TEXT NOT NULL, vhash TEXT NOT NULL,              PRIMARY KEY (day, type, vhash)) WITHOUT ROWID;
  ```
  - **Acceptance:** migration applies `--remote`; `bun run typecheck` green with new bindings on `Env`.
  - **Verify:** `wrangler d1 execute charter_analytics --remote --command "SELECT name FROM sqlite_master WHERE type='table'"` → `pageview, visitor, geo, event`.
  - **Files:** `infra/router/wrangler.toml`, `infra/router/migrations/0001_init.sql`, `infra/router/src/index.ts` (Env).
  - **Commit:** `feat(router): D1 + KV bindings and analytics schema`

## Phase 1 — Collection module

- **Task:** `infra/router/src/analytics.ts` exporting the pure core (`qualifies`, `normalizePath`, `visitorHash`) + `dailySalt(env)` (KV `salt:<UTC-date>` get-or-create 32B random, module-memo keyed by date, `expirationTtl` 172800) + `record(request, response, env, ctx)` which, when `ANALYTICS_ENABLED` and `qualifies`, computes day/path/vhash/country and `ctx.waitUntil(try{ env.ANALYTICS_DB.batch([pageview upsert, geo upsert, visitor insert-or-ignore]) }catch{})`. Upserts use `INSERT … VALUES(…,1) ON CONFLICT(pk) DO UPDATE SET hits = hits + 1`; visitor uses `INSERT OR IGNORE`. Bundle `isbot`. In `index.ts`, call `await record(...)` (non-blocking via waitUntil inside) at the **single return chokepoint** for both the landing and Mintlify-proxy responses.
  - **Acceptance:** unit tests green (invariant 4 list); `/docs` HTML counts, Mintlify assets and 3xx/4xx do not; serving is unaffected if `ANALYTICS_DB` write throws.
  - **Verify:** `bun run test` + `bun run typecheck` (infra).
  - **Files:** `infra/router/src/analytics.ts`, `infra/router/src/index.ts`, `infra/router/package.json`, `infra/router/vitest.config.ts`, `infra/router/src/analytics.test.ts`, `infra/moon.yml`.
  - **Commit:** `feat(router): first-party server-side pageview, visitor, and geo counting`

## Phase 2 — Event endpoint

- **Task:** `POST /api/event` matched **early** in `index.ts` (before proxy branches, beside `/dashboard/api/stats`): gated by `ANALYTICS_ENABLED`; require `Origin`/`Referer` host == `use-charter.dev`; `await request.text()` (reject > 1 KB); `JSON.parse` in try/catch; `type` ∈ `["install_copied"]`; compute `visitorHash`; `INSERT OR IGNORE` into `event`; **always respond `204`** (including rejections — no probing signal).
  - **Acceptance:** foreign Origin / unknown type / malformed body → nothing recorded; repeat from same visitor/day → 0 increment (dedup).
  - **Verify:** guard + dedup test (pool-workers or unit on the parse/guard helper); `bun run typecheck`.
  - **Files:** `infra/router/src/index.ts`, `infra/router/src/analytics.ts`, test.
  - **Commit:** `feat(router): install-copy event ingest with origin check and per-visitor dedup`

## Phase 3 — Client beacon

- **Task:** In `web/src/islands/landing.ts`, on successful clipboard copy: `navigator.sendBeacon?.('/api/event', new Blob([JSON.stringify({type:'install_copied'})], {type:'text/plain'}))` (text/plain → CORS-safelisted, no preflight; same-origin). No-op when absent; never blocks the copy UX.
  - **Acceptance:** one beacon per successful copy, `text/plain`, copy feedback unchanged.
  - **Verify:** Vitest spy on `navigator.sendBeacon`; `bun run test` + `bun run build` (web).
  - **Files:** `web/src/islands/landing.ts`, `web/src/islands/landing.test.ts`.
  - **Commit:** `feat(web): emit install-copy analytics beacon on copy`

## Phase 4 — Read API

- **Task:** `infra/router/src/analytics-read.ts`, wired early in `index.ts` at `GET /dashboard/api/analytics`: reuse the stats handler's Access-assertion check (`Cf-Access-Jwt-Assertion` / `Cf-Access-Authenticated-User-Email` present → else `403`); query D1 for the trailing 30 UTC days (pageviews/day, uniques/day = `COUNT(visitor)`, top pages, blog/docs sums, `install_copied` count, top-5 countries excl. `T1`/`XX`); return the spec JSON shape. Freshness via a ~30 s module-level memo — **no Cache API** (unavailable behind Access).
  - **Acceptance:** `403` without assertion; correct shape with assertion; reads ≪ 5M/day budget.
  - **Verify:** auth-gate + shape test; `bun run typecheck`.
  - **Files:** `infra/router/src/analytics-read.ts`, `infra/router/src/index.ts`, test.
  - **Commit:** `feat(router): access-gated analytics read API`

## Phase 5 — Dashboard UI

- **Task:** Extend `web/src/pages/dashboard.astro` + `web/src/islands/dashboard.ts` + `web/src/styles/dashboard.css`: fetch `/dashboard/api/analytics` in parallel with the existing `/dashboard/api/stats`; render KPI tiles (pageviews, unique visitors, install copies — UTC-30d), a daily pageviews+uniques trend **reusing the island's existing hand-rolled SVG chart** (no chart library — JS budget), a top-pages list, blog/docs reach, and top-5 countries; zero-states; labels annotate "UTC" and "approx. uniques" verbatim per the spec definitions table; terminal/dashboard design system (ADR-0024).
  - **Acceptance:** renders with empty + populated data; labels match the spec; both themes intentional.
  - **Verify:** `bunx astro check` + `bun run build`; chrome-devtools screenshots at 1440 + mobile, light + dark.
  - **Files:** `web/src/pages/dashboard.astro`, `web/src/islands/dashboard.ts`, `web/src/styles/dashboard.css`.
  - **Commit:** `feat(web): website analytics on the founder dashboard`

## Phase 6 — Privacy disclosure (parallel)

- **Task:** Add a first-party-analytics section to the privacy page (`web/src/lib/legal-content.ts`): cookieless, no client storage, aggregate counts + a daily-rotated irreversible visitor hash deleted within 48 h, no IP/UA stored, legal basis legitimate interest, founder-only access.
  - **Acceptance:** privacy page states the above accurately and matches the spec posture.
  - **Verify:** `bun run build`; section renders.
  - **Files:** `web/src/lib/legal-content.ts`.
  - **Commit:** `docs(web): privacy disclosure for first-party cookieless analytics`

## Phase 7 — Deploy & verify

- **Task:** Confirm Phase-0 manual steps done (resources + migration); push so CI deploys the router and Pages builds the web changes. Live smoke.
  - **Acceptance:** real visits populate `pageview`/`visitor`/`geo`; a copy increments `event`; the dashboard renders live numbers.
  - **Verify:**
    - visit `/` and `/docs/quickstart` in a **real browser** (not curl — `isbot`/non-HTML would skew), then `wrangler d1 execute charter_analytics --remote --command "SELECT * FROM pageview ORDER BY day DESC LIMIT 5"` shows rows incl. a `/docs/*` row;
    - `curl -s -o /dev/null -w '%{http_code}' https://use-charter.dev/dashboard/api/analytics` → `403`;
    - DevTools on `/` → **no cookie, no storage, no third-party request**;
    - `SELECT * FROM visitor LIMIT 3` and `SELECT * FROM event LIMIT 3` → only hashes, no IP/UA;
    - **empirically confirm** whether no-op `INSERT OR IGNORE` counts as a billable write (check `meta.rows_written`) — if it inflates writes materially, gate the visitor insert behind an in-request "first-seen" check.
  - **Files:** optional `docs/product/DEPLOY.md` / `infra/README.md` update recording the D1/KV resources.
  - **Commit:** `docs(infra): record analytics D1/KV resources in deploy runbook`

---

## Risks & mitigations

- **Deploy-before-binding / before-migration:** sequenced (0a→0b→0c→1); collection failures are swallowed (invariant 1) so even a mis-sequence never breaks serving.
- **`request.cf`/IP absent in local dev:** pure-function core takes these as args; integration relies on `--remote` or Phase-7 live smoke.
- **`waitUntil` D1 write loss / local-dev parity undocumented:** accepted for vanity analytics; the kill-switch and try/catch bound the blast radius.
- **No-op `INSERT OR IGNORE` write cost:** undocumented → Phase-7 empirical check + fallback (first-seen guard).
- **Free-tier write cap:** ≈16k pageviews/day headroom (`WITHOUT ROWID`); degradation = drop `visitor` table (aggregate-only) → sample.
- **Midnight salt race:** sub-minute, single-digit uniques overcount once/day; accepted.
- **`vitest-pool-workers` 0.16 API churn:** plugin-based config + version-sensitive `readD1Migrations` import; unit tests (pure funcs) are the floor, integration is best-effort.
- **isbot bundle:** small pure-JS regex; verify router bundle type-checks + deploys.

## Verification checkpoint (between every phase)
- [ ] Prior phase committed; `moon run :check` green; infra deployed / Pages built without regression.
- [ ] No cookie / client storage / third-party request introduced.
- [ ] No raw IP/UA/query string written anywhere.
- [ ] Serving path unaffected by any analytics failure (kill-switch + try/catch intact).

## Out of scope (per spec)
Sessionization, retention pruning, CF Web Analytics overlay, the waitlist→router observability move, any CLI/core change.
