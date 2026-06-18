# infra/ — Cloudflare edge

Source-of-truth config for the two Cloudflare Workers that front Charter's
public domains. The landing site itself is a **Cloudflare Pages** project
(`web/`, config in `web/wrangler.toml`); these Workers sit in front of it.

| Worker | Host | Job |
|---|---|---|
| `charter-router` (`router/`) | `use-charter.dev/*` | `/docs`,`/rules` → Mintlify; everything else → the Pages landing site (`LANDING_ORIGIN`). Also counts page views into first-party analytics (D1) and serves the Access-gated founder dashboard APIs |
| `charter-go-vanity` (`go-vanity/`) | `go.use-charter.dev` | Serves the `go-import` meta tag so `go install go.use-charter.dev/charter/...` resolves to `github.com/use-charter/charter` |

```
                 ┌──────────────────────────── use-charter.dev ────────────────────────────┐
   request ──▶  charter-router  ──┬─ /docs/*, /rules/*    ──▶  tashfiq.mintlify.app
                                  ├─ /dashboard/api/*     ──▶  GitHub + analytics  (Access-gated)
                                  ├─ /api/event           ──▶  install-copy beacon → D1
                                  └─ /*                   ──▶  LANDING_ORIGIN (charter-landing.pages.dev)
                                                                 └─ /api/waitlist → Resend
                                  every HTML page view ──▶ counted into D1 (daily salt in KV)

   go install ──▶  go.use-charter.dev  ──▶  charter-go-vanity  ──▶  go-import → github.com/use-charter/charter
```

## Prerequisites

- `wrangler` is pinned in `package.json`. Install once: `bun install` (run in `infra/`).
- Authenticate: `wrangler login`, or set `CLOUDFLARE_API_TOKEN` (+ `CLOUDFLARE_ACCOUNT_ID`)
  for CI. The account ID and the `use-charter.dev` zone ID live in the Cloudflare
  dashboard and in CI secrets/vars — they are not committed here.

## Commands

```bash
bun run typecheck        # tsc --noEmit over both workers
bun run dev:router       # local: wrangler dev router/
bun run deploy:router    # deploy charter-router
bun run deploy:go-vanity # deploy charter-go-vanity
bun run deploy           # deploy both
```

## Deploy order & one-time setup

1. **Pages landing site** (`web/`) — create via dashboard (Git integration:
   root `web`, build `bun run build`, output `dist`) or `wrangler pages deploy`.
   Set `RESEND_API_KEY` (secret) and `WAITLIST_TO` (var) on the Pages project.
   Note its `*.pages.dev` hostname.
2. **`charter-router`** — `bun run deploy:router`, then in the dashboard set
   `LANDING_ORIGIN` to the `pages.dev` hostname from step 1, and set
   `MINTLIFY_ORIGIN` (defaults to `tashfiq.mintlify.app`). The `use-charter.dev/*`
   route needs a proxied (orange-cloud) DNS record at the apex to bind.
3. **Analytics store** (bound to `charter-router`) — first-party, cookieless
   website analytics. One-time:
   ```bash
   wrangler d1 create charter_analytics             # → database_id in router/wrangler.toml
   wrangler kv namespace create ANALYTICS_SALT      # → id in router/wrangler.toml
   wrangler d1 migrations apply charter_analytics --remote
   ```
   Bindings (`router/wrangler.toml`): `ANALYTICS_DB` (D1 — aggregate UTC-day
   counts), `ANALYTICS_SALT` (KV — the rotating daily salt only), and the
   `ANALYTICS_ENABLED` kill-switch var (`"true"` collects; any other value
   disables collection and the event endpoint). The Access-gated dashboard read
   APIs also need the `GITHUB_STATS_TOKEN` worker secret. Rationale:
   `docs/internal/decisions/0027-founder-dashboard-analytics.md`.
4. **`charter-go-vanity`** — `bun run deploy:go-vanity`. The `custom_domain`
   route auto-provisions `go.use-charter.dev` + its TLS cert. Verify:
   ```bash
   curl -s "https://go.use-charter.dev/charter?go-get=1" | grep go-import
   go install go.use-charter.dev/charter/cmd/charter@latest
   ```

Full runbook (Mintlify DNS, Pages settings, verification curls): `docs/product/DEPLOY.md`.
Rationale for this topology: `docs/internal/decisions/0026-go-public-deploy-pages-and-vanity-import.md`.
