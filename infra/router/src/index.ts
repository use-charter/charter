// charter-router — edge router for use-charter.dev.
//
// Path logic lives here, not in DNS:
//   • /docs/*, /rules/*            → proxied to Mintlify (product docs)
//   • /.well-known/acme-challenge/ → resolved at the edge (cert validation)
//   • everything else              → proxied to the Cloudflare Pages landing
//                                     site (LANDING_ORIGIN), preserving method
//                                     and body so /api/waitlist works through it
//
// Mirrors the worker documented in docs/product/DEPLOY.md.

export interface Env {
  // Mintlify subdomain serving the product docs (e.g. tashfiq.mintlify.app).
  MINTLIFY_ORIGIN?: string;
  // Pages hostname for the landing site (e.g. charter-landing.pages.dev).
  // Until it is set, the worker returns a plain-text placeholder.
  LANDING_ORIGIN?: string;
}

const DEFAULT_MINTLIFY_ORIGIN = 'tashfiq.mintlify.app';

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const path = url.pathname;

    // ACME / cert-validation challenges resolve at the edge, not at an origin.
    // Everything else under /.well-known/ (e.g. security.txt) falls through to
    // the landing-site proxy below.
    if (path.startsWith('/.well-known/acme-challenge/')) {
      return fetch(request);
    }

    // Proxy the Mintlify-served sections (docs, CLI reference, rules,
    // changelog) to Mintlify, forwarding the public hostname so Mintlify
    // recognises use-charter.dev as its custom domain.
    if (
      path.startsWith('/docs') ||
      path.startsWith('/cli') ||
      path.startsWith('/rules') ||
      path.startsWith('/changelog')
    ) {
      const origin = env.MINTLIFY_ORIGIN || DEFAULT_MINTLIFY_ORIGIN;
      const upstream = new URL(`https://${origin}${path}${url.search}`);
      const proxy = new Request(upstream, request);
      proxy.headers.set('Host', origin);
      proxy.headers.set('X-Forwarded-Host', url.hostname);
      proxy.headers.set('X-Forwarded-Proto', 'https');
      proxy.headers.set('CF-Connecting-IP', request.headers.get('CF-Connecting-IP') || '');
      return fetch(proxy);
    }

    // Proxy everything else to the landing site (Cloudflare Pages). Re-basing
    // the original Request onto the Pages origin preserves method, headers and
    // body — so the /api/waitlist Function and every static asset are served.
    const landing = env.LANDING_ORIGIN;
    if (landing) {
      const dest = new URL(`https://${landing}${path}${url.search}`);
      return fetch(new Request(dest, request));
    }

    // Before LANDING_ORIGIN is set: placeholder.
    return new Response(
      'Charter — AI-agent readiness scanner.\nDocs: /docs/  Rules: /rules/',
      { status: 200, headers: { 'Content-Type': 'text/plain' } },
    );
  },
} satisfies ExportedHandler<Env>;
