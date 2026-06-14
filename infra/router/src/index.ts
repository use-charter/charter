// charter-router — edge router for use-charter.dev.
//
// Path logic lives here, not in DNS:
//   • Mintlify-owned paths (see isMintlifyPath) → proxied to Mintlify
//   • /.well-known/acme-challenge/               → resolved at the edge
//   • everything else                            → proxied to the Cloudflare
//                                                   Pages landing site
//                                                   (LANDING_ORIGIN), preserving
//                                                   method and body so
//                                                   /api/waitlist works through it
//
// Mirrors the worker documented in docs/product/DEPLOY.md, following Mintlify's
// official subpath reverse-proxy guidance
// (mintlify.com/docs/deploy/docs-subpath, /deploy/reverse-proxy): the doc
// sections AND Mintlify's namespaced static assets (/mintlify-assets/*) plus the
// root LLM index files must all reach Mintlify — otherwise the asset requests
// fall through to the landing site and the docs render unstyled.

export interface Env {
  // Mintlify subdomain serving the product docs (e.g. tashfiq.mintlify.app).
  MINTLIFY_ORIGIN?: string;
  // Pages hostname for the landing site (e.g. charter-landing.pages.dev).
  // Until it is set, the worker returns a plain-text placeholder.
  LANDING_ORIGIN?: string;
}

const DEFAULT_MINTLIFY_ORIGIN = 'tashfiq.mintlify.app';

// Paths Mintlify owns when its site is proxied under this domain: the doc
// sections (content), Mintlify's namespaced static assets (CSS/JS/favicons under
// /mintlify-assets/), and the root-level LLM index files. Everything else is the
// landing site.
function isMintlifyPath(path: string): boolean {
  return (
    path.startsWith('/docs') ||
    path.startsWith('/cli') ||
    path.startsWith('/rules') ||
    path.startsWith('/changelog') ||
    path.startsWith('/mintlify-assets') ||
    path === '/llms.txt' ||
    path === '/llms-full.txt'
  );
}

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

    // Expose Mintlify's sitemap on this domain with origin→public host rewrite.
    // Mintlify generates it at the origin root (/sitemap.xml) listing the origin
    // host, so search engines never see the public docs URLs. Serve it at
    // /docs/sitemap.xml (declared in the landing robots.txt) with the host
    // swapped to the request host — every listed path is proxied unchanged, so
    // only the hostname needs rewriting.
    if (path === '/docs/sitemap.xml') {
      const origin = env.MINTLIFY_ORIGIN || DEFAULT_MINTLIFY_ORIGIN;
      const res = await fetch(`https://${origin}/sitemap.xml`, {
        headers: { Host: origin },
      });
      const body = (await res.text()).split(`https://${origin}`).join(`https://${url.hostname}`);
      return new Response(body, {
        status: res.status,
        headers: {
          'Content-Type': 'application/xml; charset=utf-8',
          'Cache-Control': 'public, max-age=3600',
        },
      });
    }

    // Proxy Mintlify-owned paths to Mintlify, forwarding the public hostname so
    // Mintlify recognises use-charter.dev as its custom domain.
    if (isMintlifyPath(path)) {
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
