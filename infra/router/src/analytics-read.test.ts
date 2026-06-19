import { describe, expect, it, vi } from "vitest";
import { handleAnalytics } from "./analytics-read";

// Minimal D1 stub: prepare/bind are inert; batch returns one result per query
// in the order build() issues them (the four scalar queries carry `n`).
function stubDb() {
	const batch = vi.fn(async () => [
		{ results: [] },
		{ results: [] },
		{ results: [] },
		{ results: [{ n: 0 }] },
		{ results: [{ n: 0 }] },
		{ results: [{ n: 0 }] },
		{ results: [] },
	]);
	const db = {
		prepare: () => ({ bind: () => ({}) }),
		batch,
	} as unknown as D1Database;
	return { db, batch };
}

const req = (headers: Record<string, string> = {}) =>
	new Request("https://use-charter.dev/dashboard/api/analytics", { headers });

describe("handleAnalytics", () => {
	it("rejects requests without a Cloudflare Access assertion (403) and never queries D1", async () => {
		const { db, batch } = stubDb();
		const res = await handleAnalytics(req(), { ANALYTICS_DB: db });
		expect(res.status).toBe(403);
		expect(batch).not.toHaveBeenCalled();
	});

	it("returns the aggregate shape for an authenticated request", async () => {
		const { db } = stubDb();
		const res = await handleAnalytics(
			req({ "Cf-Access-Authenticated-User-Email": "founder@use-charter.dev" }),
			{
				ANALYTICS_DB: db,
			},
		);
		expect(res.status).toBe(200);
		const body = (await res.json()) as { generatedAt: string };
		expect(body).toMatchObject({
			rangeDays: 30,
			pageviewsByDay: [],
			uniquesByDay: [],
			topPages: [],
			blogViews: 0,
			docsViews: 0,
			events: { install_copied: 0 },
			topCountries: [],
		});
		expect(typeof body.generatedAt).toBe("string");
	});
});

// A fresh module instance resets the in-isolate memo so each test starts cold.
async function freshHandle() {
	vi.resetModules();
	return (await import("./analytics-read")).handleAnalytics;
}

const authed = req({
	"Cf-Access-Authenticated-User-Email": "founder@use-charter.dev",
});

describe("handleAnalytics aggregation and memoization", () => {
	it("maps populated D1 rows into the dashboard JSON shape", async () => {
		const handle = await freshHandle();
		const batch = vi.fn(async () => [
			{ results: [{ day: "2026-06-01", count: 5 }] },
			{ results: [{ day: "2026-06-01", count: 3 }] },
			{ results: [{ path: "/", hits: 9 }] },
			{ results: [{ n: 4 }] },
			{ results: [{ n: 2 }] },
			{ results: [{ n: 1 }] },
			{ results: [{ country: "US", hits: 7 }] },
		]);
		const db = {
			prepare: () => ({ bind: () => ({}) }),
			batch,
		} as unknown as D1Database;
		const body = (await (
			await handle(authed, { ANALYTICS_DB: db })
		).json()) as Record<string, unknown>;
		expect(body).toMatchObject({
			pageviewsByDay: [{ day: "2026-06-01", count: 5 }],
			uniquesByDay: [{ day: "2026-06-01", count: 3 }],
			topPages: [{ path: "/", hits: 9 }],
			blogViews: 4,
			docsViews: 2,
			events: { install_copied: 1 },
			topCountries: [{ country: "US", hits: 7 }],
		});
	});

	it("defaults a scalar metric to 0 when its query returns no row", async () => {
		const handle = await freshHandle();
		const batch = vi.fn(async () => [
			{ results: [] },
			{ results: [] },
			{ results: [] },
			{ results: [] }, // blog scalar: no row → blogViews falls back to 0
			{ results: [{ n: 0 }] },
			{ results: [{ n: 0 }] },
			{ results: [] },
		]);
		const db = {
			prepare: () => ({ bind: () => ({}) }),
			batch,
		} as unknown as D1Database;
		const body = (await (
			await handle(authed, { ANALYTICS_DB: db })
		).json()) as { blogViews: number };
		expect(body.blogViews).toBe(0);
	});

	it("serves the second request from the memo without re-querying D1", async () => {
		const handle = await freshHandle();
		const { db, batch } = stubDb();
		await handle(authed, { ANALYTICS_DB: db });
		await handle(authed, { ANALYTICS_DB: db });
		expect(batch).toHaveBeenCalledTimes(1);
	});
});
