# Mintlify Deployment & Routing — Cloudflare

## Domain architecture

- **Primary domain**: `use-charter.dev` (bought from Cloudflare Registrar, DNS managed in Cloudflare)
- **Mintlify origin**: Mintlify-controlled subdomain, e.g. `charter.mintlify.dev`
- **Landing site**: will live on `use-charter.dev` (Slice 19, hosting TBD — Cloudflare Pages or Worker)
- **Root-domain routing required** (provided by a Cloudflare Worker on the zone):
  - `https://use-charter.dev/docs/*` -> proxy to Mintlify origin
  - `https://use-charter.dev/rules/*` -> proxy to Mintlify origin

## Why this design

- `helpUri` contract requires `https://use-charter.dev/rules/AE-*` to resolve
- Slice 19 owns the root domain for the marketing site
- Mintlify supports proxy/rewrite patterns documented at `/deploy/cloudflare.mdx`
- The domain is already on Cloudflare (Registrar + DNS), so Workers can route on the same zone
- No additional DNS changes needed beyond setting up a Worker route and an optional CNAME

## Cloudflare Worker — docs proxy

A single Worker on the `use-charter.dev` zone handles path-based routing. ES modules format.

```javascript
// docs-proxy - serves Charter public docs from Mintlify via Cloudflare Worker
// Deployed on the use-charter.dev zone. Routes /docs/* and /rules/* to the
// Mintlify docs origin. All other paths are placeholders until Slice 19.

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    // Pass through well-known paths (SSL verification, ACME challenges)
    if (path.startsWith('/.well-known/')) {
      return fetch(request);
    }

    // Route /docs/* and /rules/* to the Mintlify origin
    if (path.startsWith('/docs') || path.startsWith('/rules')) {
      const origin = env.MINTLIFY_ORIGIN || 'charter.mintlify.dev';
      const upstream = new URL(`https://${origin}${path}${url.search}`);
      const proxy = new Request(upstream, request);
      proxy.headers.set('Host', origin);
      proxy.headers.set('X-Forwarded-Host', url.hostname);
      proxy.headers.set('X-Forwarded-Proto', 'https');
      return fetch(proxy);
    }

    // Route to landing site when Slice 19 is live
    const landing = env.LANDING_ORIGIN;
    if (landing) {
      const dest = new URL(`https://${landing}${path}${url.search}`);
      return fetch(dest, {
        method: request.method,
        headers: request.headers,
        body: request.body,
        redirect: 'follow',
      });
    }

    // Before Slice 19: placeholder response
    return new Response(
      'Charter - AI-agent readiness scanner.\nDocs: /docs/  Rules: /rules/',
      { status: 200, headers: { 'Content-Type': 'text/plain' } },
    );
  },
};
```

### Worker configuration

```jsonc
// wrangler.jsonc - deploy via `npx wrangler deploy`
{
  "name": "docs-proxy",
  "main": "src/index.js",
  "compatibility_date": "2026-06-09",
  "compatibility_flags": ["nodejs_compat"]
}
```

Set `MINTLIFY_ORIGIN` as an environment variable (plain-text binding, not a secret).

### Worker routes

Add routes in the Cloudflare dashboard under **Workers & Pages > docs-proxy > Triggers > Routes**:

| Route | Description |
|---|---|
| `use-charter.dev/docs*` | All docs subpaths |
| `use-charter.dev/rules*` | All rules subpaths |

Alternatively, use a wildcard route `use-charter.dev/*` and let the Worker logic decide. The explicit route approach is preferred so non-matching paths are not unnecessarily routed through the Worker.

### Optional: docs subdomain

For direct access to the docs site during development:

- **DNS record**: `docs.use-charter.dev CNAME charter.mintlify.dev` (proxied — orange cloud)
- This is optional. The primary access path is `use-charter.dev/docs/*` via the Worker.

## Redirects (Mintlify-side)

Configure in `docs/product/docs.json`:
- Map any changed paths during restructuring
- All `/rules/AE-*` URLs are permanent and must not change

## Verification

At launch, verify from the public internet:

- `https://use-charter.dev/rules/AE-CTX-001` -> renders the rule page
- `https://use-charter.dev/rules/AE-SEC-001` -> renders the rule page
- `https://use-charter.dev/docs/quickstart` -> renders the quickstart page
- `https://use-charter.dev/docs/` -> renders the docs home or introduction
- Verify the `docs.json` redirects work for old -> new paths

## Handoff to Slice 19/20

This Worker must be updated when the landing site launches:
1. Set `LANDING_ORIGIN` env var to the landing site host
2. The catch-all else branch will then proxy to the landing site instead of returning the placeholder
3. If the landing site is on a different platform (not Cloudflare), the Worker can still proxy via `fetch`

## References

- Mintlify Cloudflare deployment guide: `https://docs.mintlify.com/deploy/cloudflare.mdx`
- Cloudflare Workers ES modules format: `https://developers.cloudflare.com/workers/reference/migrate-to-module-workers/`
- Cloudflare Workers Routes: `https://developers.cloudflare.com/workers/configuration/routing/routes/`
