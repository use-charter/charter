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

import { handleEvent, record } from "./analytics";
import { handleAnalytics } from "./analytics-read";
import { type DashboardEnv, handleDashboardStats } from "./dashboard";

export interface Env extends DashboardEnv {
	// Mintlify subdomain serving the product docs (e.g. tashfiq.mintlify.app).
	MINTLIFY_ORIGIN?: string;
	// Pages hostname for the landing site (e.g. charter-landing.pages.dev).
	// Until it is set, the worker returns a plain-text placeholder.
	LANDING_ORIGIN?: string;
	// First-party website analytics (ADR-0027). D1 holds aggregate UTC-day counts;
	// KV holds only the rotating daily salt. ANALYTICS_ENABLED is the kill-switch
	// ("true" collects; any other value disables collection + the event endpoint).
	ANALYTICS_DB: D1Database;
	ANALYTICS_SALT: KVNamespace;
	ANALYTICS_ENABLED?: string;
}

const DEFAULT_MINTLIFY_ORIGIN = "tashfiq.mintlify.app";

// Paths Mintlify owns when its site is proxied under this domain: the doc
// sections (content) and Mintlify's namespaced static assets (CSS/JS/favicons
// under /mintlify-assets/). Everything else is the landing site — which serves
// the curated /llms.txt index and /robots.txt. (/llms-full.txt is handled
// separately so its links can be host-rewritten to the public domain.)
function isMintlifyPath(path: string): boolean {
	return (
		path.startsWith("/docs") ||
		path.startsWith("/cli") ||
		path.startsWith("/rules") ||
		path.startsWith("/changelog") ||
		path.startsWith("/mintlify-assets")
	);
}

export default {
	async fetch(
		request: Request,
		env: Env,
		ctx: ExecutionContext,
	): Promise<Response> {
		const url = new URL(request.url);

		// Force HTTPS at the edge: redirect any plaintext request to its https
		// equivalent before doing any work. ACME http-01 challenges are excepted so
		// certificate validation can still answer over http. HSTS is configured at
		// the Cloudflare edge (SSL/TLS → Edge Certificates), which overrides any
		// header a worker could set, so it is not duplicated here.
		if (
			url.protocol === "http:" &&
			!url.pathname.startsWith("/.well-known/acme-challenge/")
		) {
			url.protocol = "https:";
			return Response.redirect(url.href, 301);
		}

		return route(request, env, ctx, url);
	},
} satisfies ExportedHandler<Env>;

async function route(
	request: Request,
	env: Env,
	ctx: ExecutionContext,
	url: URL,
): Promise<Response> {
	const path = url.pathname;

	// ACME / cert-validation challenges resolve at the edge, not at an origin.
	// Everything else under /.well-known/ (e.g. security.txt) falls through to
	// the landing-site proxy below.
	if (path.startsWith("/.well-known/acme-challenge/")) {
		return fetch(request);
	}

	// Founder dashboard stats API. Gated by Cloudflare Access on /dashboard*;
	// the handler also requires an Access assertion header (defense-in-depth).
	if (path === "/dashboard/api/stats") {
		return handleDashboardStats(request, env);
	}

	// Client analytics beacon (install-copy and similar). Same-origin only,
	// type allow-listed, always answers 204.
	if (path === "/api/event") {
		return handleEvent(request, env);
	}

	// Founder analytics read API. Access-gated like the stats endpoint.
	if (path === "/dashboard/api/analytics") {
		return handleAnalytics(request, env);
	}

	// Expose Mintlify's sitemap on this domain with origin→public host rewrite.
	// Mintlify generates it at the origin root (/sitemap.xml) listing the origin
	// host, so search engines never see the public docs URLs. Serve it at
	// /docs/sitemap.xml (declared in the landing robots.txt) with the host
	// swapped to the request host — every listed path is proxied unchanged, so
	// only the hostname needs rewriting.
	if (path === "/docs/sitemap.xml") {
		const origin = env.MINTLIFY_ORIGIN || DEFAULT_MINTLIFY_ORIGIN;
		const res = await fetch(`https://${origin}/sitemap.xml`, {
			headers: { Host: origin },
		});
		const body = (await res.text())
			.split(`https://${origin}`)
			.join(`https://${url.hostname}`);
		return new Response(body, {
			status: res.status,
			headers: {
				"Content-Type": "application/xml; charset=utf-8",
				"Cache-Control": "public, max-age=3600",
			},
		});
	}

	// /llms-full.txt is Mintlify's full concatenated docs corpus. Proxy it with
	// the origin→public host rewrite so the AI-readable URLs are canonical
	// (use-charter.dev), not the internal Mintlify origin. The curated /llms.txt
	// index is served by the landing site.
	if (path === "/llms-full.txt") {
		const origin = env.MINTLIFY_ORIGIN || DEFAULT_MINTLIFY_ORIGIN;
		const res = await fetch(`https://${origin}/llms-full.txt`, {
			headers: { Host: origin },
		});
		const body = (await res.text())
			.split(`https://${origin}`)
			.join(`https://${url.hostname}`);
		return new Response(body, {
			status: res.status,
			headers: {
				"Content-Type": "text/plain; charset=utf-8",
				"Cache-Control": "public, max-age=3600",
			},
		});
	}

	// Convenience 301: people and tools guess /sitemap.xml, but Astro's sitemap
	// integration emits an index at /sitemap-index.xml. Redirect the guess so it
	// resolves instead of 404ing. (The real sitemap is declared in robots.txt.)
	if (path === "/sitemap.xml") {
		return Response.redirect(new URL("/sitemap-index.xml", url).href, 301);
	}

	// Proxy Mintlify-owned paths to Mintlify, forwarding the public hostname so
	// Mintlify recognises use-charter.dev as its custom domain.
	if (isMintlifyPath(path)) {
		const origin = env.MINTLIFY_ORIGIN || DEFAULT_MINTLIFY_ORIGIN;
		const upstream = new URL(`https://${origin}${path}${url.search}`);
		const proxy = new Request(upstream, request);
		proxy.headers.set("Host", origin);
		proxy.headers.set("X-Forwarded-Host", url.hostname);
		proxy.headers.set("X-Forwarded-Proto", "https");
		proxy.headers.set(
			"CF-Connecting-IP",
			request.headers.get("CF-Connecting-IP") || "",
		);
		// Record analytics for proxied documentation HTML pages; non-HTML assets
		// are filtered out by `qualifies`.
		const response = await fetch(proxy);
		record(request, response, env, ctx);
		return response;
	}

	// Proxy everything else to the landing site (Cloudflare Pages). Re-basing
	// the original Request onto the Pages origin preserves method, headers and
	// body — so the /api/waitlist Function and every static asset are served.
	const landing = env.LANDING_ORIGIN;
	if (landing) {
		const dest = new URL(`https://${landing}${path}${url.search}`);
		// Record analytics for landing-site HTML pages.
		const response = await fetch(new Request(dest, request));
		record(request, response, env, ctx);
		return response;
	}

	// Before LANDING_ORIGIN is set: placeholder.
	return new Response(
		"Charter — AI-agent readiness scanner.\nDocs: /docs/  Rules: /rules/",
		{ status: 200, headers: { "Content-Type": "text/plain" } },
	);
}
