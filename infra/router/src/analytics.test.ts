import { describe, expect, it, vi } from "vitest";
import {
	type AnalyticsEnv,
	normalizePath,
	originAllowed,
	parseEventType,
	type QualifyParts,
	qualifies,
	utcDay,
	visitorHash,
} from "./analytics";

// A fresh module instance per stateful test so the module-scoped daily-salt memo
// starts empty and salt resolution is deterministic.
async function freshModule() {
	vi.resetModules();
	return import("./analytics");
}

const today = new Date().toISOString().slice(0, 10);

/** In-memory D1 + KV stubs recording what the collector writes. */
function stubEnv(overrides: Partial<AnalyticsEnv> = {}) {
	const batches: unknown[][] = [];
	const runs: { sql: string; args: unknown[] }[] = [];
	const kv = new Map<string, string>();
	const db = {
		prepare: (sql: string) => ({
			bind: (...args: unknown[]) => ({
				sql,
				args,
				run: async () => void runs.push({ sql, args }),
			}),
		}),
		batch: async (s: unknown[]) => {
			batches.push(s);
			return [];
		},
	} as unknown as D1Database;
	const salt = {
		get: async (k: string) => kv.get(k) ?? null,
		put: async (k: string, v: string) => void kv.set(k, v),
	} as unknown as KVNamespace;
	const env = {
		ANALYTICS_ENABLED: "true",
		ANALYTICS_DB: db,
		ANALYTICS_SALT: salt,
		...overrides,
	} as AnalyticsEnv;
	return { env, batches, runs, kv };
}

function page(headers: Record<string, string> = {}, country?: string) {
	const req = new Request("https://use-charter.dev/blog/introducing-charter", {
		headers,
	});
	if (country) Object.assign(req, { cf: { country } });
	return req;
}
const htmlOk = () =>
	new Response("<!doctype html>", {
		status: 200,
		headers: { "content-type": "text/html" },
	});

describe("normalizePath", () => {
	it("keeps root and known top-level pages", () => {
		expect(normalizePath("/")).toBe("/");
		expect(normalizePath("/blog")).toBe("/blog");
	});
	it("keeps bounded blog and legal slugs", () => {
		expect(normalizePath("/blog/introducing-charter")).toBe(
			"/blog/introducing-charter",
		);
		expect(normalizePath("/legal/privacy")).toBe("/legal/privacy");
	});
	it("collapses proxied doc families to one bucket each", () => {
		expect(normalizePath("/docs")).toBe("/docs/*");
		expect(normalizePath("/docs/quickstart")).toBe("/docs/*");
		expect(normalizePath("/cli/doctor")).toBe("/cli/*");
		expect(normalizePath("/rules/AE-SEC-001")).toBe("/rules/*");
		expect(normalizePath("/changelog/v1-0-0")).toBe("/changelog/*");
	});
	it("buckets unknown/scanner paths into /__other__", () => {
		expect(normalizePath("/wp-login.php")).toBe("/__other__");
		expect(normalizePath("/.env")).toBe("/__other__");
	});
	it("normalizes trailing slash and case", () => {
		expect(normalizePath("/Blog/")).toBe("/blog");
		expect(normalizePath("/legal/Terms/")).toBe("/legal/terms");
	});
});

const ok: QualifyParts = {
	method: "GET",
	path: "/blog/introducing-charter",
	status: 200,
	contentType: "text/html; charset=utf-8",
	ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
	secPurpose: "",
};

describe("qualifies", () => {
	it("counts a successful HTML GET from a real browser", () => {
		expect(qualifies(ok)).toBe(true);
	});
	it("rejects non-GET methods", () => {
		expect(qualifies({ ...ok, method: "POST" })).toBe(false);
	});
	it("rejects redirects and errors (not 2xx)", () => {
		expect(qualifies({ ...ok, status: 301 })).toBe(false);
		expect(qualifies({ ...ok, status: 404 })).toBe(false);
		expect(qualifies({ ...ok, status: 500 })).toBe(false);
	});
	it("rejects non-HTML responses (assets, xml)", () => {
		expect(qualifies({ ...ok, contentType: "application/xml" })).toBe(false);
		expect(qualifies({ ...ok, contentType: "text/css" })).toBe(false);
	});
	it("excludes api, dashboard, and mintlify asset paths", () => {
		expect(qualifies({ ...ok, path: "/api/event" })).toBe(false);
		expect(qualifies({ ...ok, path: "/dashboard" })).toBe(false);
		expect(qualifies({ ...ok, path: "/dashboard/api/analytics" })).toBe(false);
		expect(qualifies({ ...ok, path: "/mintlify-assets/x.css" })).toBe(false);
	});
	it("skips browser speculative prefetch", () => {
		expect(qualifies({ ...ok, secPurpose: "prefetch" })).toBe(false);
		expect(qualifies({ ...ok, secPurpose: "prefetch;prerender" })).toBe(false);
	});
	it("skips known bots", () => {
		expect(
			qualifies({
				...ok,
				ua: "Googlebot/2.1 (+http://www.google.com/bot.html)",
			}),
		).toBe(false);
	});
});

describe("visitorHash", () => {
	it("is deterministic within a day (same salt/ip/ua)", async () => {
		const a = await visitorHash("salt-A", "203.0.113.5", "UA-1");
		const b = await visitorHash("salt-A", "203.0.113.5", "UA-1");
		expect(a).toBe(b);
	});
	it("differs across days (salt rotation) and across visitors", async () => {
		const day1 = await visitorHash("salt-A", "203.0.113.5", "UA-1");
		const day2 = await visitorHash("salt-B", "203.0.113.5", "UA-1");
		const other = await visitorHash("salt-A", "198.51.100.9", "UA-1");
		expect(day1).not.toBe(day2);
		expect(day1).not.toBe(other);
	});
	it("never leaks the raw IP or user-agent", async () => {
		const ip = "203.0.113.5";
		const ua = "Mozilla/5.0 SecretAgent";
		const h = await visitorHash("salt-A", ip, ua);
		expect(h).not.toContain(ip);
		expect(h).not.toContain("SecretAgent");
		expect(h).toMatch(/^[A-Za-z0-9_-]+$/); // base64url, no padding
	});
});

describe("utcDay", () => {
	it("formats a UTC calendar day", () => {
		expect(utcDay(new Date("2031-03-04T23:59:00Z"))).toBe("2031-03-04");
	});
});

const evt = (headers: Record<string, string>) =>
	new Request("https://example/api/event", { method: "POST", headers });

describe("originAllowed", () => {
	it("accepts requests originating from the site", () => {
		expect(originAllowed(evt({ Origin: "https://use-charter.dev" }))).toBe(
			true,
		);
		expect(
			originAllowed(evt({ Referer: "https://use-charter.dev/blog/x" })),
		).toBe(true);
	});
	it("rejects foreign or missing origins", () => {
		expect(originAllowed(evt({ Origin: "https://evil.example" }))).toBe(false);
		expect(originAllowed(evt({}))).toBe(false);
	});
});

describe("parseEventType", () => {
	it("returns an allow-listed type", () => {
		expect(parseEventType('{"type":"install_copied"}')).toBe("install_copied");
	});
	it("rejects unknown types, non-strings, malformed, and oversized bodies", () => {
		expect(parseEventType('{"type":"hack"}')).toBeNull();
		expect(parseEventType('{"type":123}')).toBeNull();
		expect(parseEventType("not json")).toBeNull();
		expect(
			parseEventType(`{"type":"install_copied","pad":"${"x".repeat(1100)}"}`),
		).toBeNull();
	});
});

describe("record (fire-and-forget pageview collection)", () => {
	it("writes a pageview + visitor batch for a qualifying request, minting a daily salt", async () => {
		const { record } = await freshModule();
		const { env, batches, kv } = stubEnv();
		const waits: Promise<unknown>[] = [];
		record(page(), htmlOk(), env, {
			waitUntil: (p: Promise<unknown>) => waits.push(p),
		} as unknown as ExecutionContext);
		await Promise.all(waits);
		expect(batches).toHaveLength(1);
		expect(batches[0]).toHaveLength(2); // pageview + visitor, no geo without a country
		expect(kv.get(`salt:${today}`)).toBeTruthy(); // salt minted and stored
	});

	it("adds a geo row when the request carries a usable country", async () => {
		const { record } = await freshModule();
		const { env, batches } = stubEnv();
		const waits: Promise<unknown>[] = [];
		record(page({}, "US"), htmlOk(), env, {
			waitUntil: (p: Promise<unknown>) => waits.push(p),
		} as unknown as ExecutionContext);
		await Promise.all(waits);
		expect(batches[0]).toHaveLength(3); // pageview + visitor + geo
	});

	it("skips the geo row for anonymized/unknown country codes", async () => {
		const { record } = await freshModule();
		const { env, batches } = stubEnv();
		const waits: Promise<unknown>[] = [];
		record(page({}, "T1"), htmlOk(), env, {
			waitUntil: (p: Promise<unknown>) => waits.push(p),
		} as unknown as ExecutionContext);
		await Promise.all(waits);
		expect(batches[0]).toHaveLength(2);
	});

	it("reuses an existing daily salt from KV instead of minting one", async () => {
		const { record } = await freshModule();
		const { env, kv } = stubEnv();
		kv.set(`salt:${today}`, "preexisting-salt");
		const waits: Promise<unknown>[] = [];
		record(page(), htmlOk(), env, {
			waitUntil: (p: Promise<unknown>) => waits.push(p),
		} as unknown as ExecutionContext);
		await Promise.all(waits);
		expect(kv.get(`salt:${today}`)).toBe("preexisting-salt"); // not overwritten
	});

	it("does nothing when analytics are disabled", async () => {
		const { record } = await freshModule();
		const { env, batches } = stubEnv({ ANALYTICS_ENABLED: "false" });
		const waits: Promise<unknown>[] = [];
		record(page(), htmlOk(), env, {
			waitUntil: (p: Promise<unknown>) => waits.push(p),
		} as unknown as ExecutionContext);
		await Promise.all(waits);
		expect(waits).toHaveLength(0);
		expect(batches).toHaveLength(0);
	});

	it("reuses the memoised salt across two collections in the same isolate", async () => {
		const { record } = await freshModule();
		const { env, batches } = stubEnv();
		const waits: Promise<unknown>[] = [];
		const fire = () =>
			record(page(), htmlOk(), env, {
				waitUntil: (p: Promise<unknown>) => waits.push(p),
			} as unknown as ExecutionContext);
		fire();
		fire(); // second collection hits the in-memory salt memo
		await Promise.all(waits);
		expect(batches).toHaveLength(2);
	});

	it("treats a response with no content-type as non-qualifying", async () => {
		const { record } = await freshModule();
		const { env, batches } = stubEnv();
		const waits: Promise<unknown>[] = [];
		record(page(), new Response("x", { status: 200 }), env, {
			waitUntil: (p: Promise<unknown>) => waits.push(p),
		} as unknown as ExecutionContext);
		await Promise.all(waits);
		expect(batches).toHaveLength(0);
	});

	it("does nothing for a non-qualifying response (e.g. a redirect)", async () => {
		const { record } = await freshModule();
		const { env, batches } = stubEnv();
		const waits: Promise<unknown>[] = [];
		const redirect = new Response(null, {
			status: 301,
			headers: { "content-type": "text/html" },
		});
		record(page(), redirect, env, {
			waitUntil: (p: Promise<unknown>) => waits.push(p),
		} as unknown as ExecutionContext);
		await Promise.all(waits);
		expect(batches).toHaveLength(0);
	});
});

describe("handleEvent (client beacon ingestion)", () => {
	const beacon = (body: string, headers: Record<string, string> = {}) =>
		new Request("https://use-charter.dev/api/event", {
			method: "POST",
			headers: { Origin: "https://use-charter.dev", ...headers },
			body,
		});

	it("records one event row for a valid same-origin beacon and answers 204", async () => {
		const { handleEvent } = await freshModule();
		const { env, runs } = stubEnv();
		const res = await handleEvent(beacon('{"type":"install_copied"}'), env);
		expect(res.status).toBe(204);
		expect(runs).toHaveLength(1);
		expect(runs[0].args).toEqual([today, "install_copied", expect.any(String)]);
	});

	it("answers 204 without writing when analytics are disabled", async () => {
		const { handleEvent } = await freshModule();
		const { env, runs } = stubEnv({ ANALYTICS_ENABLED: "false" });
		const res = await handleEvent(beacon('{"type":"install_copied"}'), env);
		expect(res.status).toBe(204);
		expect(runs).toHaveLength(0);
	});

	it("answers 204 without writing for a non-POST method", async () => {
		const { handleEvent } = await freshModule();
		const { env, runs } = stubEnv();
		const req = new Request("https://use-charter.dev/api/event", {
			headers: { Origin: "https://use-charter.dev" },
		});
		expect((await handleEvent(req, env)).status).toBe(204);
		expect(runs).toHaveLength(0);
	});

	it("answers 204 without writing for a foreign origin", async () => {
		const { handleEvent } = await freshModule();
		const { env, runs } = stubEnv();
		await handleEvent(
			beacon('{"type":"install_copied"}', { Origin: "https://evil.example" }),
			env,
		);
		expect(runs).toHaveLength(0);
	});

	it("answers 204 without writing for an invalid event type", async () => {
		const { handleEvent } = await freshModule();
		const { env, runs } = stubEnv();
		await handleEvent(beacon('{"type":"nope"}'), env);
		expect(runs).toHaveLength(0);
	});
});
