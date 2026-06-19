import type { APIContext } from "astro";
import { describe, expect, it } from "vitest";
import { GET as llms } from "./pages/llms.txt";
import { GET as robots } from "./pages/robots.txt";

// The route handlers only read `site`; a partial context is enough.
const ctx = (site?: string) =>
	({ site: site ? new URL(site) : undefined }) as unknown as APIContext;
const body = (r: Response | Promise<Response>) =>
	Promise.resolve(r).then((res) => res.text());

describe("robots.txt", () => {
	it("allows all crawlers and lists both sitemaps off the configured site", async () => {
		const res = await robots(ctx("https://use-charter.dev"));
		expect(res.headers.get("Content-Type")).toContain("text/plain");
		const txt = await res.text();
		expect(txt).toContain("User-agent: *\nAllow: /");
		expect(txt).toContain("Sitemap: https://use-charter.dev/sitemap-index.xml");
		expect(txt).toContain("Sitemap: https://use-charter.dev/docs/sitemap.xml");
	});

	it("explicitly welcomes the AI crawlers (GPTBot, ClaudeBot, …)", async () => {
		const txt = await body(robots(ctx("https://use-charter.dev")));
		for (const ua of [
			"GPTBot",
			"ClaudeBot",
			"PerplexityBot",
			"Google-Extended",
		]) {
			expect(txt).toContain(`User-agent: ${ua}`);
		}
	});

	it("falls back to the canonical origin when site is absent", async () => {
		const txt = await body(robots(ctx()));
		expect(txt).toContain("https://use-charter.dev/sitemap-index.xml");
	});
});

describe("llms.txt", () => {
	it("emits the curated brief with absolute links and the full-corpus pointer", async () => {
		const res = await llms(ctx("https://use-charter.dev"));
		expect(res.headers.get("Content-Type")).toContain("text/plain");
		const txt = await res.text();
		expect(txt).toContain("# Charter");
		expect(txt).toContain("https://use-charter.dev/docs/cli/doctor");
		expect(txt).toContain("https://use-charter.dev/llms-full.txt");
	});

	it("falls back to the canonical origin when site is absent", async () => {
		expect(await body(llms(ctx()))).toContain(
			"https://use-charter.dev/docs/quickstart",
		);
	});
});
