# Founder Dashboard — Website Analytics Specification

**Document type:** Specification (WHAT and WHY, not HOW)
**Status:** Locked & Validated — grilled, doc-grounded, zero open questions
**Date:** 2026-06-19
**Decision record:** [ADR-0027](../../decisions/0027-founder-dashboard-analytics.md)
**Last validated:** 2026-06-19 — every decision below grounded against live Cloudflare D1/Workers docs, MDN, and EDPB/Plausible privacy guidance (see [Grounding & sources](#grounding--sources)); router topology verified against `infra/router/`.

---

## Goal

Give the founder, on the existing Access-gated mission-control dashboard (`use-charter.dev/dashboard`), a truthful, real-time view of **website** behaviour alongside the GitHub signals already there — with **no third-party tracker, no cookies, no client-side storage, and no stored personal data**. Seven signals: daily pageviews, daily unique visitors, per-page views, blog views, install-command copies (the primary conversion), docs reach, and a top-5 visitor-country list.

---

## Problem Statement

The dashboard shows GitHub health but is blind to the site that drives adoption. The founder cannot see how many people visit, where they land, whether the blog is read, whether the **primary conversion** (`brew install …` copy) fires, how many continue to docs, or where visitors are. Off-the-shelf analytics (GA/Plausible/PostHog) contradicts Charter's "no telemetry" brand, adds a third-party script + dependency, and usually sets cookies. Cloudflare Web Analytics is cookieless but **cannot record custom events** and reports sampled "visits", not device-uniques. The enabler: the `charter-router` Worker already proxies **every** request to `use-charter.dev` and already serves the dashboard API — so six of seven metrics are measurable first-party, server-side, at zero marginal cost.

---

## Scope

### In scope
- **Collection:** server-side request counting inside `charter-router`; one client beacon for the install-copy event.
- **Storage:** one Cloudflare **D1** database (`charter_analytics`, aggregate rows, UTC-day keyed) + one **KV** namespace (`ANALYTICS_SALT`, holding the rotating daily salt only).
- **Read API:** new Access-gated `GET /dashboard/api/analytics` on the router.
- **Event API:** `POST /api/event` on the router (install-copy only).
- **Dashboard UI:** new tiles + a daily trend chart in `web/src/pages/dashboard.astro` (+ island), inheriting the terminal/dashboard design system (ADR-0024).
- **Client:** a single same-origin `navigator.sendBeacon('/api/event', …)` on copy success in `web/src/islands/landing.ts`.
- **Free tier only.**

### Out of scope
- Third-party analytics, any cookie, any client-side storage, any cross-site identifier, any consent banner.
- Sessionized "visits", funnels, cohorts, retention curves, per-visitor path trails.
- Pages→Workers migration (site stays static — ADR-0026).
- The separate Workers-Observability change (waitlist → router + `[observability]`) — tracked independently.
- Any change to the CLI/core, which stays zero-telemetry (ADR-0001).

---

## Metrics & definitions (contract)

Definitions are normative; UI labels must match them verbatim (no "unique visitors" where the value is an approximation, etc.).

| # | Metric (UI label) | Definition | Source |
|---|---|---|---|
| 1 | **Pageviews / day** | Count of qualifying page requests in the UTC day (raw, not sessionized) | server-side |
| 2 | **Unique visitors / day** | Count of distinct daily-salted visitor hashes for the day (privacy approximation — same person on 2 days = 2 uniques) | server-side |
| 3 | **Views per page** | Pageviews grouped by normalized `path` | server-side |
| 4 | **Blog views** | Subset of (3) where path is `/blog` or under `/blog/` | server-side |
| 5 | **Docs reach** | Hits where path is `/docs` or under `/docs/` (router-proxied to Mintlify) | server-side |
| 6 | **Install copies / day** | Distinct daily visitor hashes that emitted `install_copied` (deduped per visitor/day, so it is unspammable beyond 1/visitor/day) | client beacon |
| 7 | **Top-5 countries** | Top 5 by hits from `request.cf.country` (ISO-2; `T1` Tor and `XX` unknown excluded) | server-side |

---

## Collection rules (contract)

**Qualifying request** (counts toward 1–5, 7). ALL must hold:
- method `GET`; response status `< 400`; response `Content-Type` begins `text/html`;
- path is **not** under `/api/`, **not** `/dashboard` or under it (Access-gated, non-public), **not** a Mintlify static asset (`/mintlify-assets/…`), and not the `og`/`sitemap`/`robots`/`llms` machine endpoints (non-HTML anyway);
- request header `Sec-Purpose` does **not** start with `prefetch` (skip browser speculative prefetch/prerender hits — MDN-documented mechanism);
- `User-Agent` is **not** a known bot per the `isbot` denylist (bundled into the Worker; no free `request.cf` bot signal exists).

**Path normalization** (bounds row cardinality, defeats scanner-path bloat e.g. `/wp-login.php`):
1. drop query string and hash; 2. lowercase; 3. collapse to the canonical trailing-slash form the site uses; 4. map to one of the **known prefixes** `/`, `/blog`, `/blog/{slug}`, `/legal/{slug}`, `/docs`, `/docs/*` (collapsed to `/docs/*`), `/changelog*`; 5. anything else → the single bucket `/__other__`. The stored `path` set is therefore bounded and meaningful.

**Day boundary:** UTC. The dashboard labels the range "UTC" so the founder reads it correctly; no per-request timezone math.

---

## Visitor identity & privacy model (contract)

Unique visitors are counted **without cookies, without client storage, and without persisting any personal data**, following the Plausible/Fathom rotating-salt model (the design that lets the data be argued anonymous after 24h, not merely pseudonymous).

- **Daily salt:** a 32-byte random value generated once per UTC day, stored in KV under key `salt:<YYYY-MM-DD>` with a ~48h TTL, then **expires/deleted**. Each isolate reads it at most once per day (memoized in a module-level variable keyed by date) → ≤ a few hundred KV reads/day, 1 KV write/day (both well within KV free limits). A persistent/derived secret is explicitly **rejected**: per EDPB guidance a keyed hash with a long-lived key stays personal data; only salt **deletion** severs the link and supports the anonymized-after-24h posture.
- **Hash:** `vhash = base64url( SHA-256( dailySalt ‖ "use-charter.dev" ‖ client-IP ‖ user-agent ) )`, truncated to 16 bytes. Inputs follow Plausible (IP + UA + site scope under a rotating salt). The raw IP and user-agent are used **transiently in memory for hashing only** and are **never** written to D1, KV, logs, or disk.
- **Cross-day unlinkability:** because the salt rotates and the old salt is deleted, the same person yields a different hash each day and cannot be tracked across days — by design, and the basis for counting metric (2)/(6).
- **Midnight race:** at the UTC rollover, isolates briefly may generate competing salts (KV get-then-put, last-write-wins) → at most a sub-minute, single-digit uniques overcount once/day. Accepted.
- **Legal basis:** no cookie / no `localStorage` ⇒ ePrivacy Art. 5(3) consent trigger does not apply; GDPR processing rests on **legitimate interest (Art. 6(1)(f))** with a privacy-policy disclosure. This "no consent needed" stance is industry-standard (Plausible/Fathom) but legally **contested** — the privacy-policy disclosure is mandatory, not optional, and is part of this spec's deliverable.

---

## Storage & bindings (contract)

Cloudflare **D1** `charter_analytics`, bound to `charter-router`. Tables use `WITHOUT ROWID` so the primary key *is* the storage (no separate rowid table → halves rows-written per upsert):

```sql
CREATE TABLE pageview (day TEXT, path TEXT,    hits INTEGER, PRIMARY KEY (day, path))    WITHOUT ROWID;
CREATE TABLE visitor  (day TEXT, vhash TEXT,                 PRIMARY KEY (day, vhash))   WITHOUT ROWID;
CREATE TABLE geo      (day TEXT, country TEXT, hits INTEGER, PRIMARY KEY (day, country)) WITHOUT ROWID;
CREATE TABLE event    (day TEXT, type TEXT, vhash TEXT,      PRIMARY KEY (day, type, vhash)) WITHOUT ROWID;
```
Writes: `INSERT … ON CONFLICT(pk) DO UPDATE SET hits = hits + 1` for `pageview`/`geo`; `INSERT OR IGNORE` for `visitor`/`event` (dedup). D1 is single-threaded → upserts are atomic, no lost updates. The three per-pageview writes are issued as one `db.batch([...])` from `ctx.waitUntil()` (no added response latency).

**Bindings** (`infra/router/wrangler.toml`):
```toml
[[d1_databases]]
binding = "ANALYTICS_DB"
database_name = "charter_analytics"
database_id = "<dashboard-set>"

[[kv_namespaces]]
binding = "ANALYTICS_SALT"
id = "<dashboard-set>"
```
Schema ships as a checked-in Wrangler migration (`wrangler d1 migrations create/apply`).

**Derivations:** pageviews = `SUM(pageview.hits)`; uniques = `COUNT(visitor)` per day; per-page = `pageview` rows; blog/docs = path filters; install copies = `COUNT(event WHERE type='install_copied')` per day; countries = `geo … GROUP BY country ORDER BY hits DESC LIMIT 5`.

---

## Read & event APIs (contract)

**`GET /dashboard/api/analytics`** — Access-gated; rejects requests lacking the Cloudflare Access assertion header (`403`), mirroring `/dashboard/api/stats`. **Does not use the Cache API** (it is unavailable for Access-fronted Workers); freshness comes from a short module-level in-isolate memo (≤60 s) over direct D1 reads. Returns aggregates only:
```jsonc
{
  "generatedAt": "<ISO>", "rangeDays": 30,
  "pageviewsByDay": [{ "day": "2026-06-18", "count": 0 }],
  "uniquesByDay":   [{ "day": "2026-06-18", "count": 0 }],
  "topPages":       [{ "path": "/", "hits": 0 }],
  "blogViews": 0, "docsViews": 0,
  "events": { "install_copied": 0 },
  "topCountries":   [{ "country": "US", "hits": 0 }]
}
```

**`POST /api/event`** — body is `text/plain` carrying a tiny JSON string (keeps the `sendBeacon` request CORS-safelisted → no preflight). Server: reject unless the `Origin`/`Referer` host is `use-charter.dev`; `type` must be in the allow-list (`install_copied`); compute the same daily `vhash` and `INSERT OR IGNORE` into `event`; respond `204`. Forgery is bounded by Origin check + per-visitor/day dedup (a spammer with one IP/UA = at most 1 count/day).

**Client beacon** (`landing.ts`): on successful clipboard copy, `navigator.sendBeacon('/api/event', new Blob([JSON.stringify({type:'install_copied'})], {type:'text/plain'}))`. Fired on the copy action (not `unload`). No-JS / no-clipboard visitors are simply not counted (copy requires JS anyway).

---

## Dashboard UI

New section in `web/src/pages/dashboard.astro` (+ island), terminal/dashboard design system (ADR-0024): KPI tiles (pageviews, unique visitors, install copies — each with the UTC-30d total), a daily pageviews+uniques trend chart, a top-pages list, blog/docs reach tiles, and a top-5 countries list. Zero-state rendering when no data yet.

---

## Success Criteria

1. All seven metrics render on `/dashboard` with real values once traffic exists; zero-states before.
2. Collection is entirely Cloudflare **Free** tier; no paid feature enabled.
3. **No cookie, no client storage, no third-party request** — verifiable in the network/application tabs.
4. **No personal data at rest** — D1 holds only aggregates + daily-rotated hashes; raw IP/UA appear nowhere; the daily salt expires from KV within ~48h.
5. Install-copy beacon fires once per successful copy, is deduped per visitor/day, and is rejected for any non-allow-listed `type` or foreign `Origin`.
6. `GET /dashboard/api/analytics` returns `403` without a valid Access assertion.
7. Counting adds no measurable response latency (writes via `ctx.waitUntil()` `batch`); per-request CPU (one SHA-256 + isbot check) stays well within the 10 ms Free limit.
8. Bot UA traffic and `Sec-Purpose: prefetch` hits are excluded; UI labels match the definitions table verbatim and annotate "UTC" and "approximate uniques".
9. A privacy-policy disclosure for first-party cookieless analytics is published.

---

## Boundaries

- **Always:** count server-side where possible; write via `waitUntil`+`batch`; store only aggregates + irreversible daily hashes; rotate-and-delete the salt daily; gate the read API behind Access; stay on Free tier.
- **Ask first:** any new client script beyond the single copy beacon; a new D1 table or third binding; changing the visitor-hash scheme or salt lifecycle; adding sampling; layering in CF Web Analytics.
- **Never:** set a cookie or cross-site identifier; persist raw IP/user-agent/query strings/per-visitor paths; send data to a third party; add telemetry to the CLI/core; expose any analytics endpoint unauthenticated (read) or without an Origin+allow-list check (event).

---

## Resolved decisions (was Open Questions — all closed)

| Question | Decision | Why / source |
|---|---|---|
| "Visit" definition | **Pageviews/day** (raw) + **Unique visitors/day** (daily-salt hash). Not sessionized. | Matches the founder's ask and the Plausible counting model; sessionization adds a session-window with no founder value here. |
| Bot handling | Exclude via **`isbot`** UA denylist + skip `Sec-Purpose: prefetch`. | No free `request.cf` bot signal (Bot Management is Enterprise); `isbot` is the maintained 2026 standard; `Sec-Purpose` is MDN's documented prefetch-skip mechanism. |
| Unique-visitor identity | **Random daily salt in KV, deleted daily**, `SHA-256(salt‖domain‖IP‖UA)`; raw IP/UA never stored. | EDPB: persistent-key hash = personal data; only salt deletion yields the anonymized-after-24h posture (Plausible). |
| Read-endpoint caching | **No Cache API** (unavailable behind Access); short in-isolate memo over direct D1 reads. | Cloudflare docs: "For Workers fronted by Cloudflare Access, the Cache API is not currently available." |
| Write cost / cardinality | `WITHOUT ROWID` tables; `batch()` the 3 writes; allow-list+`/__other__` path bucketing. | Halves D1 rows-written (no rowid index); bounds `pageview` rows; ≈16k pageviews/day headroom on the 100k writes/day Free cap. |
| Event abuse | Origin host check + dedup per daily visitor hash. | Public POST is forgeable; dedup + Origin is the standard lightweight mitigation. |
| Retention | Keep daily **aggregate** rows indefinitely (negligible in 5 GB); salt auto-expires ~48h. | Aggregates carry no personal data; pruning unnecessary at this scale. |
| Default dashboard window | **30 days** (UTC). | Enough to show trend on a founder dashboard; range is a query param for later widening. |
| Workers Observability (waitlist→router) | **Deferred** to a separate change. | Independent concern; keeps this scope tight. |

---

## Failure modes & limits

- **Free-tier write ceiling:** ~2 D1 rows-written per upsert × 3 per pageview ⇒ ~6/pageview after `WITHOUT ROWID` reduces index amplification; 100k writes/day ⇒ **≈16k pageviews/day** headroom. Degradation if exceeded: drop the `visitor` table (aggregate-only), then head-sample writes. The Workers 100k req/day cap is the whole-site limit already and is unchanged by this work.
- **`waitUntil` durability:** D1 fire-and-forget after response is runtime-supported but not contractually guaranteed; under extreme edge eviction a write may be lost. Acceptable for vanity analytics (not billing).
- **Hot-row contention:** concurrent increments to one `(day,path)` row serialize on D1's single thread; fine at this scale.
- **Bot/uniques accuracy:** server-side counts include bots beyond the `isbot` list; "uniques" is a daily approximation. Both are labeled in the UI; optional CF Web Analytics overlay can later provide a bot-clean cross-check.

---

## Grounding & sources (verified 2026-06-19)

- D1 single-threaded execution / atomic auto-commit; Free limits 5M reads, **100k writes**, 5 GB, 500 MB/db, 10 dbs, 50 queries/invocation — https://developers.cloudflare.com/d1/platform/limits/ , https://developers.cloudflare.com/d1/platform/pricing/
- D1 `batch()` is a single transaction, sequential commit — https://developers.cloudflare.com/d1/worker-api/d1-database/
- D1 migrations + `[[d1_databases]]` binding — https://developers.cloudflare.com/d1/reference/migrations/ , https://developers.cloudflare.com/d1/get-started/
- `request.cf` geo fields free on all plans; `country` sentinels `T1` (Tor) / `XX` (unknown); `botManagement` is Bot-Management/Enterprise only — https://developers.cloudflare.com/workers/runtime-apis/request/ , https://developers.cloudflare.com/fundamentals/reference/http-request-headers/
- Cache API **unavailable behind Access**; does not cache POST — https://developers.cloudflare.com/workers/runtime-apis/cache/
- `Sec-Purpose: prefetch` forwarded; documented for skipping speculative page-visit counts — https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Sec-Purpose
- `sendBeacon` POST, `text/plain` is CORS-safelisted (no preflight), 64 KiB cap — https://developer.mozilla.org/en-US/docs/Web/API/Navigator/sendBeacon
- IP = personal data; salted hash = pseudonymisation unless salt deleted; rotating-daily-salt model + legitimate-interest/no-cookie basis (contested) — EDPB pseudonymisation guidelines (2025) https://www.edpb.europa.eu/system/files/2025-01/edpb_guidelines_202501_pseudonymisation_en.pdf , https://plausible.io/data-policy , https://plausible.io/blog/legal-assessment-gdpr-eprivacy
- `isbot` maintained UA bot-detection library (2026) — https://github.com/omrilotan/isbot
