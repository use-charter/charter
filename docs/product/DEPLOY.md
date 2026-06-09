# Mintlify Deployment — Setup Guide

Deploys Charter's customer-facing docs at `https://use-charter.dev/docs` and `https://use-charter.dev/rules`.

**Two phases:**
- **Phase A** — Mintlify preview (5 min, no Cloudflare needed, sandbox for content review)
- **Phase B** — Cloudflare custom domain (wire the real domain once content is approved)

---

## Phase A — Mintlify preview

### Step 1: Create a Mintlify account

Go to [mintlify.com](https://mintlify.com) → **Sign up with GitHub**. Use the same GitHub account that owns `use-charter/charter`.

---

### Step 2: Create the project

In the Mintlify dashboard:

1. **New project** → **Connect GitHub**
2. Select the `use-charter/charter` repository
3. **Docs directory**: `docs/product` ← set this exactly; Mintlify reads `docs.json` from this path
4. **Branch**: `main`
5. Click **Deploy**

Mintlify deploys automatically. The first build takes ~1–2 minutes.

---

### Step 3: Get your preview URL

After the deploy completes, Mintlify assigns a subdomain:

```
https://charter.mintlify.dev
```

(The exact subdomain is shown in the dashboard under **Project → Deployments**.)

Every push to `main` triggers a new deploy automatically.

---

### Step 4: Verify the deploy

Check these URLs on your preview domain:

| URL | Expected |
|---|---|
| `/` | Introduction page with four Cards |
| `/quickstart` | Clean install instructions, no "launch-gated" language |
| `/installation` | Four install paths (brew, binary, go install, source) |
| `/rules/AE-CTX-001` | New anatomy: Why this rule, What triggers it, Examples |
| `/rules/AE-SEC-001` | Score impact section, Related rules cross-links |
| `/concepts/fix-engine` | Fix engine page renders |
| `/design-philosophy` | Ten commitments page renders |
| `/changelog` | v1.0 entry |

If images (logo, favicon) don't appear, check **Project Settings → Custom Domain** — the `images/` path is served relative to `docs/product/`.

---

### Step 5: Iterate

- All content changes: edit `.mdx` files and push to `main` — Mintlify redeploys in ~30s
- Navigation changes: edit `docs.json` and push
- Rule pages: edit `docs/internal/specs/AE-*.md` and run `bun scripts/generate-rule-pages.ts`, then push

---

## Phase B — Cloudflare custom domain

Do this after content is approved on the preview URL.

### Architecture

```
use-charter.dev  (Cloudflare Registrar + DNS)
     │
     ├── /docs/*  ─── Cloudflare Worker ──► charter.mintlify.dev
     ├── /rules/* ─── Cloudflare Worker ──► charter.mintlify.dev
     └── /*       ─── Cloudflare Worker ──► LANDING_ORIGIN (Slice 19) or placeholder
```

The Worker proxies `/docs/*` and `/rules/*` to Mintlify, forwarding the correct `X-Forwarded-Host` header so Mintlify recognises `use-charter.dev` as its public hostname.

---

### Step 1: Add custom domain in Mintlify

In the Mintlify dashboard → **Project Settings → Custom Domain**:

- Enter `use-charter.dev`
- Mintlify shows:
  - CNAME target: `cname.mintlify.builders`
  - Two TXT verification values — copy both, you need them in the next step

---

### Step 2: Add DNS records in Cloudflare

In the [Cloudflare dashboard](https://dash.cloudflare.com) → DNS for `use-charter.dev`, add all five records:

**Verification TXT records** (add these first — Mintlify needs them before it can issue the SSL cert):

| Type | Name | Content | Proxy status |
|---|---|---|---|
| TXT | `_acme-challenge.use-charter.dev` | `<value from Mintlify dashboard>` | ⬜ DNS only |
| TXT | `_cf-custom-hostname.use-charter.dev` | `<value from Mintlify dashboard>` | ⬜ DNS only |

**Routing records** (add after TXT records are verified in Mintlify):

| Type | Name | Content | Proxy status |
|---|---|---|---|
| A | `use-charter.dev` | `192.0.2.1` | ✅ Proxied (orange cloud) |
| CNAME | `docs` | `cname.mintlify.builders` | ⬜ DNS only (grey cloud) |

**Why the TXT records:** Mintlify uses Let's Encrypt for TLS (`_acme-challenge`) and Cloudflare for Hostname validation (`_cf-custom-hostname`). Without both TXT records present and verified, Mintlify cannot provision the SSL certificate for your domain. Add these before the CNAME.

**Why the A record:** Cloudflare Workers only intercept requests when traffic routes through Cloudflare's network. A proxied A record on the root domain enables that. `192.0.2.1` is a non-routable RFC 5737 address — the Worker intercepts before the IP is ever used.

**Why grey cloud on CNAME:** The `docs.use-charter.dev` CNAME points directly to Mintlify. Orange cloud (proxied) makes Cloudflare terminate TLS, breaking Mintlify's certificate provisioning. DNS-only (grey cloud) lets Mintlify handle SSL end-to-end. This CNAME is optional — the Worker route is the primary path.

---

### Step 2a: Confirm Mintlify verified the domain

Back in the Mintlify dashboard → **Project Settings → Custom Domain**: wait for the status to show **Verified** or **Active** before proceeding. This usually takes 1–5 minutes after the TXT records propagate.

---

### Step 3: Create the Cloudflare Worker

In the [Cloudflare dashboard](https://dash.cloudflare.com) → **Workers & Pages → Create Worker**:

Name: `docs-proxy`

Paste this script:

```javascript
// docs-proxy — routes /docs/* and /rules/* to Mintlify.
// Set MINTLIFY_ORIGIN env var to your Mintlify subdomain (e.g. charter.mintlify.dev).
// Set LANDING_ORIGIN env var when the Slice 19 landing site is deployed.

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    // Pass through SSL verification and ACME challenge paths
    if (path.startsWith('/.well-known/')) {
      return fetch(request);
    }

    // Proxy /docs/* and /rules/* to Mintlify
    if (path.startsWith('/docs') || path.startsWith('/rules')) {
      const origin = env.MINTLIFY_ORIGIN || 'charter.mintlify.dev';
      const upstream = new URL(`https://${origin}${path}${url.search}`);
      const proxy = new Request(upstream, request);
      proxy.headers.set('Host', origin);
      proxy.headers.set('X-Forwarded-Host', url.hostname);
      proxy.headers.set('X-Forwarded-Proto', 'https');
      proxy.headers.set('CF-Connecting-IP', request.headers.get('CF-Connecting-IP') || '');
      return fetch(proxy);
    }

    // Proxy everything else to the landing site (Slice 19)
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

    // Before Slice 19: placeholder
    return new Response(
      'Charter — AI-agent readiness scanner.\nDocs: /docs/  Rules: /rules/',
      { status: 200, headers: { 'Content-Type': 'text/plain' } },
    );
  },
};
```

---

### Step 4: Set environment variable

In the Worker → **Settings → Variables and Secrets**:

| Variable | Value | Type |
|---|---|---|
| `MINTLIFY_ORIGIN` | `charter.mintlify.dev` | Plain text |

Do **not** add `LANDING_ORIGIN` until the Slice 19 landing site is deployed.

---

### Step 5: Add Worker routes

In the Worker → **Triggers → Routes → Add route**:

| Route pattern | Zone |
|---|---|
| `use-charter.dev/docs*` | use-charter.dev |
| `use-charter.dev/rules*` | use-charter.dev |

---

### Step 6: Verify end-to-end

From a browser (incognito, no cache):

```
https://use-charter.dev/docs/quickstart         → renders quickstart page
https://use-charter.dev/docs/installation       → renders install page
https://use-charter.dev/rules/AE-CTX-001        → renders rule page
https://use-charter.dev/rules/AE-SEC-001        → renders rule page
https://use-charter.dev/                        → placeholder text (until Slice 19)
```

---

## Redirects

If any page paths change during content updates, add redirects in `docs/product/docs.json`:

```json
"redirects": [
  { "source": "/old-path", "destination": "/new-path" }
]
```

The `/rules/AE-*` paths are **permanent** — SARIF `helpUri`s point at them. Never change them without a redirect.

---

## Handoff to Slice 19

When the landing site deploys:

1. Set `LANDING_ORIGIN` in the Worker env vars (value = the landing site hostname)
2. The Worker's catch-all branch then proxies to it instead of returning the placeholder
3. Test: `https://use-charter.dev/` should serve the landing site; `/docs/*` and `/rules/*` still route to Mintlify

---

## References

- Mintlify custom domain docs: `https://mintlify.com/docs/customize/custom-domain`
- Mintlify Cloudflare deployment: `https://mintlify.com/docs/deploy/cloudflare`
- Cloudflare Workers: `https://developers.cloudflare.com/workers/`
- Cloudflare Workers routes: `https://developers.cloudflare.com/workers/configuration/routing/routes/`
