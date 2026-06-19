import { afterEach, describe, expect, it, vi } from "vitest";
import type { Env } from "./index";
import worker from "./index";

const ctx = {
	waitUntil: () => {},
	passThroughOnException: () => {},
} as unknown as ExecutionContext;

// Analytics disabled so proxy paths never touch D1/KV; origins are explicit.
function env(overrides: Partial<Env> = {}): Env {
	return {
		MINTLIFY_ORIGIN: "origin.mintlify.app",
		LANDING_ORIGIN: "landing.pages.dev",
		ANALYTICS_ENABLED: "false",
		ANALYTICS_DB: {} as D1Database,
		ANALYTICS_SALT: {} as KVNamespace,
		...overrides,
	} as Env;
}

const call = (url: string, init?: RequestInit, e: Env = env()) =>
	worker.fetch(new Request(url, init), e, ctx);

afterEach(() => vi.unstubAllGlobals());

describe("charter-router edge behaviour", () => {
	it("redirects plaintext HTTP to HTTPS with a 301", async () => {
		const res = await call("http://use-charter.dev/foo?q=1");
		expect(res.status).toBe(301);
		expect(res.headers.get("Location")).toBe("https://use-charter.dev/foo?q=1");
	});

	it("does not redirect ACME http-01 challenges (answers them over http)", async () => {
		const fetchMock = vi.fn(async () => new Response("challenge-token"));
		vi.stubGlobal("fetch", fetchMock);
		const res = await call(
			"http://use-charter.dev/.well-known/acme-challenge/abc",
		);
		expect(res.status).toBe(200);
		expect(await res.text()).toBe("challenge-token");
		expect(fetchMock).toHaveBeenCalledOnce();
	});
});

describe("charter-router routing", () => {
	it("delegates /dashboard/api/stats and rejects it without Access (403)", async () => {
		const res = await call("https://use-charter.dev/dashboard/api/stats");
		expect(res.status).toBe(403);
	});

	it("delegates /api/event to the beacon handler (204 when disabled)", async () => {
		const res = await call("https://use-charter.dev/api/event", {
			method: "POST",
		});
		expect(res.status).toBe(204);
	});

	it("delegates /dashboard/api/analytics and rejects it without Access (403)", async () => {
		const res = await call("https://use-charter.dev/dashboard/api/analytics");
		expect(res.status).toBe(403);
	});

	it("serves /docs/sitemap.xml from Mintlify with the host rewritten to the public domain", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				async () =>
					new Response(
						"<loc>https://origin.mintlify.app/docs/quickstart</loc>",
					),
			),
		);
		const res = await call("https://use-charter.dev/docs/sitemap.xml");
		const body = await res.text();
		expect(res.headers.get("Content-Type")).toContain("application/xml");
		expect(body).toContain("https://use-charter.dev/docs/quickstart");
		expect(body).not.toContain("origin.mintlify.app");
	});

	it("serves /llms-full.txt from Mintlify with the host rewritten", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				async () =>
					new Response("see https://origin.mintlify.app/llms-full.txt"),
			),
		);
		const body = await (
			await call("https://use-charter.dev/llms-full.txt")
		).text();
		expect(body).toBe("see https://use-charter.dev/llms-full.txt");
	});

	it("301-redirects the guessed /sitemap.xml to the real sitemap index", async () => {
		const res = await call("https://use-charter.dev/sitemap.xml");
		expect(res.status).toBe(301);
		expect(res.headers.get("Location")).toBe(
			"https://use-charter.dev/sitemap-index.xml",
		);
	});

	it("proxies Mintlify-owned paths to the docs origin", async () => {
		const fetchMock = vi.fn(
			async (_input: Request | string) =>
				new Response("<html>docs</html>", {
					headers: { "content-type": "text/html" },
				}),
		);
		vi.stubGlobal("fetch", fetchMock);
		const res = await call("https://use-charter.dev/docs/quickstart?x=1");
		expect(res.status).toBe(200);
		const proxied = fetchMock.mock.calls[0][0] as Request;
		expect(proxied.url).toBe("https://origin.mintlify.app/docs/quickstart?x=1");
	});

	it("falls back to the default Mintlify origin when MINTLIFY_ORIGIN is unset", async () => {
		const fetchMock = vi.fn(
			async (_input: Request | string) =>
				new Response("ok", { headers: { "content-type": "text/html" } }),
		);
		vi.stubGlobal("fetch", fetchMock);
		await call(
			"https://use-charter.dev/rules/AE-SEC-001",
			undefined,
			env({ MINTLIFY_ORIGIN: undefined }),
		);
		expect((fetchMock.mock.calls[0][0] as Request).url).toContain(
			"tashfiq.mintlify.app",
		);
	});

	it("proxies everything else to the landing site, preserving the path", async () => {
		const fetchMock = vi.fn(
			async (_input: Request | string) =>
				new Response("landing", { headers: { "content-type": "text/html" } }),
		);
		vi.stubGlobal("fetch", fetchMock);
		const res = await call("https://use-charter.dev/blog/introducing-charter");
		expect(await res.text()).toBe("landing");
		expect((fetchMock.mock.calls[0][0] as Request).url).toBe(
			"https://landing.pages.dev/blog/introducing-charter",
		);
	});

	it("returns the placeholder before LANDING_ORIGIN is configured", async () => {
		const res = await call(
			"https://use-charter.dev/",
			undefined,
			env({ LANDING_ORIGIN: undefined }),
		);
		expect(res.status).toBe(200);
		expect(await res.text()).toContain("Charter — AI-agent readiness scanner.");
	});

	it("falls back to the default origin for /docs/sitemap.xml when MINTLIFY_ORIGIN is unset", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				async () =>
					new Response("<loc>https://tashfiq.mintlify.app/docs/x</loc>"),
			),
		);
		const body = await (
			await call(
				"https://use-charter.dev/docs/sitemap.xml",
				undefined,
				env({ MINTLIFY_ORIGIN: undefined }),
			)
		).text();
		expect(body).toContain("https://use-charter.dev/docs/x");
	});

	it("falls back to the default origin for /llms-full.txt when MINTLIFY_ORIGIN is unset", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				async () =>
					new Response("at https://tashfiq.mintlify.app/llms-full.txt"),
			),
		);
		const body = await (
			await call(
				"https://use-charter.dev/llms-full.txt",
				undefined,
				env({ MINTLIFY_ORIGIN: undefined }),
			)
		).text();
		expect(body).toBe("at https://use-charter.dev/llms-full.txt");
	});
});
