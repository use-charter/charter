import { afterEach, describe, expect, it, vi } from "vitest";
import { type DashboardEnv, handleDashboardStats } from "./dashboard";

const ACCESS = {
	"Cf-Access-Authenticated-User-Email": "founder@use-charter.dev",
};
const req = (headers: Record<string, string> = {}) =>
	new Request("https://use-charter.dev/dashboard/api/stats", { headers });
const env = (overrides: Partial<DashboardEnv> = {}): DashboardEnv => ({
	GITHUB_STATS_TOKEN: "gh-token",
	...overrides,
});

const jsonRes = (
	body: unknown,
	status = 200,
	headers: Record<string, string> = {},
) => new Response(JSON.stringify(body), { status, headers });

// Default GitHub API responses keyed by the path fragment buildStats requests.
function ghResponder(over: Record<string, () => Response> = {}) {
	const routes: Record<string, () => Response> = {
		"/traffic/views": () =>
			jsonRes({
				count: 120,
				uniques: 80,
				views: [{ timestamp: "2026-06-01T00:00:00Z", count: 10, uniques: 7 }],
			}),
		"/traffic/clones": () =>
			jsonRes({
				count: 9,
				uniques: 6,
				clones: [{ timestamp: "2026-06-01T00:00:00Z", count: 2, uniques: 2 }],
			}),
		"/releases": () =>
			jsonRes([
				{
					tag_name: "v1.0.0",
					published_at: "2026-06-10T00:00:00Z",
					html_url: "https://github.com/use-charter/charter/releases/v1.0.0",
					assets: [{ name: "charter_linux", download_count: 42 }],
				},
			]),
		"/search/code": () =>
			jsonRes({
				total_count: 3,
				items: [
					{ repository: { full_name: "acme/widgets" } },
					{ repository: { full_name: "use-charter/charter" } },
				],
			}),
		"/search/issues": () => jsonRes({ total_count: 5, items: [] }),
		"/repos/use-charter/charter": () =>
			jsonRes({
				stargazers_count: 100,
				forks_count: 12,
				subscribers_count: 9,
				open_issues_count: 4,
			}),
		...over,
	};
	return vi.fn(async (input: string) => {
		const url = String(input);
		// Order matters: the repo root must not shadow its /traffic and /releases subpaths.
		const key = Object.keys(routes).find((k) => url.includes(k));
		return key ? routes[key]() : jsonRes({}, 404);
	});
}

function stubCaches() {
	const store = {
		match: vi.fn(async (): Promise<Response | undefined> => undefined),
		put: vi.fn(async () => {}),
	};
	vi.stubGlobal("caches", { default: store });
	return store;
}

afterEach(() => vi.unstubAllGlobals());

describe("handleDashboardStats", () => {
	it("rejects requests without a Cloudflare Access assertion (403)", async () => {
		const res = await handleDashboardStats(req(), env());
		expect(res.status).toBe(403);
	});

	it("reports not_configured when the GitHub token is absent (200)", async () => {
		const res = await handleDashboardStats(
			req(ACCESS),
			env({ GITHUB_STATS_TOKEN: undefined }),
		);
		expect(res.status).toBe(200);
		expect((await res.json()) as { error: string }).toMatchObject({
			error: "not_configured",
		});
	});

	it("returns a fresh cached response on a cache hit without calling GitHub", async () => {
		const caches = stubCaches();
		caches.match.mockResolvedValueOnce(jsonRes({ cached: true }));
		const fetchMock = ghResponder();
		vi.stubGlobal("fetch", fetchMock);
		const res = await handleDashboardStats(req(ACCESS), env());
		expect((await res.json()) as { cached: boolean }).toEqual({ cached: true });
		expect(fetchMock).not.toHaveBeenCalled();
	});

	it("aggregates the GitHub API into the four metric groups and caches the result", async () => {
		const caches = stubCaches();
		vi.stubGlobal("fetch", ghResponder());
		const res = await handleDashboardStats(req(ACCESS), env());
		expect(res.status).toBe(200);
		expect(res.headers.get("Cache-Control")).toContain("max-age=300");
		const body = (await res.json()) as Record<string, any>;
		expect(body.growth).toMatchObject({
			stars: 100,
			forks: 12,
			watchers: 9,
			openIssues: 4,
		});
		expect(body.traffic).toMatchObject({ views14d: 120, clones14d: 9 });
		expect(body.traffic.viewsSeries).toEqual([
			{ day: "2026-06-01", count: 10 },
		]);
		expect(body.releases).toMatchObject({
			latestTag: "v1.0.0",
			totalDownloads: 42,
		});
		// Adoption counts distinct external repos: the search returns acme/widgets
		// and use-charter/charter, so only the external repo counts → 1.
		expect(body.adoption).toMatchObject({
			actionRepos: 1,
			schemaRefs: 1,
			sampleAdopters: ["acme/widgets"],
		});
		expect(body.community).toMatchObject({
			openIssues: 5,
			closedIssues: 5,
			openPRs: 5,
		});
		expect(body.errors).toEqual({});
		expect(caches.put).toHaveBeenCalledOnce();
	});

	it("degrades gracefully and does not cache when a metric group fails", async () => {
		stubCaches();
		const caches = stubCaches();
		vi.stubGlobal(
			"fetch",
			ghResponder({ "/releases": () => jsonRes({ message: "boom" }, 500) }),
		);
		const res = await handleDashboardStats(req(ACCESS), env());
		expect(res.status).toBe(200);
		expect(res.headers.get("Cache-Control")).toBe("no-store");
		const body = (await res.json()) as {
			errors: Record<string, string>;
			releases: unknown;
		};
		expect(body.releases).toBeFalsy();
		expect(body.errors.releases).toContain("500");
		expect(caches.put).not.toHaveBeenCalled();
	});

	it("emits null/zero groups (not crashes) when every GitHub call fails", async () => {
		stubCaches();
		vi.stubGlobal(
			"fetch",
			vi.fn(async () => jsonRes({ message: "down" }, 500)),
		);
		const body = (await (
			await handleDashboardStats(req(ACCESS), env())
		).json()) as Record<string, any>;
		expect(body.growth).toBeFalsy();
		expect(body.traffic).toBeFalsy();
		expect(body.releases).toBeFalsy();
		expect(body.adoption).toEqual({
			actionRepos: 0,
			schemaRefs: 0,
			sampleAdopters: [],
		});
		expect(body.community).toEqual({
			openIssues: 0,
			closedIssues: 0,
			openPRs: 0,
		});
		expect(Object.keys(body.errors).length).toBeGreaterThan(4);
	});

	it("treats an empty release list as no release and tolerates partial traffic", async () => {
		stubCaches();
		vi.stubGlobal(
			"fetch",
			ghResponder({
				"/releases": () => jsonRes([]),
				"/traffic/views": () => jsonRes({ message: "down" }, 500), // views fail, clones succeed
				"/search/code": () =>
					jsonRes({ total_count: 1, items: [{ html_url: "x" }] }), // item without repository
			}),
		);
		const body = (await (
			await handleDashboardStats(req(ACCESS), env())
		).json()) as Record<string, any>;
		expect(body.releases).toBeFalsy(); // truthy array but length 0 → no release block
		expect(body.traffic).toMatchObject({ views14d: 0, clones14d: 9 }); // views fall back to 0
		expect(body.adoption.sampleAdopters).toEqual([]); // item lacked a repository
	});

	it("retries a rate-limited (403) GitHub call and succeeds", async () => {
		stubCaches();
		vi.stubGlobal("setTimeout", ((fn: () => void) => {
			fn();
			return 0;
		}) as unknown as typeof setTimeout);
		let repoHits = 0;
		const fetchMock = ghResponder({
			"/repos/use-charter/charter": () => {
				repoHits += 1;
				return repoHits === 1
					? jsonRes({ message: "rate limited" }, 403, { "retry-after": "1" })
					: jsonRes({
							stargazers_count: 100,
							forks_count: 12,
							subscribers_count: 9,
							open_issues_count: 4,
						});
			},
		});
		vi.stubGlobal("fetch", fetchMock);
		const res = await handleDashboardStats(req(ACCESS), env());
		const body = (await res.json()) as {
			growth: { stars: number };
			errors: Record<string, string>;
		};
		expect(repoHits).toBe(2); // retried once, then succeeded
		expect(body.growth.stars).toBe(100);
		expect(body.errors.growth).toBeUndefined();
	});

	it("retries with the default backoff when no Retry-After header is present", async () => {
		stubCaches();
		vi.stubGlobal("setTimeout", ((fn: () => void) => {
			fn();
			return 0;
		}) as unknown as typeof setTimeout);
		let hits = 0;
		vi.stubGlobal(
			"fetch",
			ghResponder({
				"/repos/use-charter/charter": () => {
					hits += 1;
					return hits === 1
						? jsonRes({ message: "rate limited" }, 429) // no retry-after → default 1.5s backoff
						: jsonRes({
								stargazers_count: 7,
								forks_count: 0,
								subscribers_count: 0,
								open_issues_count: 0,
							});
				},
			}),
		);
		const body = (await (
			await handleDashboardStats(req(ACCESS), env())
		).json()) as Record<string, any>;
		expect(hits).toBe(2);
		expect(body.growth.stars).toBe(7);
	});

	it("records a non-Error rejection by its string form", async () => {
		stubCaches();
		vi.stubGlobal(
			"fetch",
			vi.fn(async () => {
				throw "string failure"; // a non-Error rejection
			}),
		);
		const body = (await (
			await handleDashboardStats(req(ACCESS), env())
		).json()) as Record<string, any>;
		expect(Object.values(body.errors)).toContain("string failure");
	});

	it("zeroes the failing half of traffic while keeping the other", async () => {
		stubCaches();
		vi.stubGlobal(
			"fetch",
			ghResponder({
				"/traffic/clones": () => jsonRes({ message: "down" }, 500),
			}),
		);
		const body = (await (
			await handleDashboardStats(req(ACCESS), env())
		).json()) as Record<string, any>;
		expect(body.traffic).toMatchObject({ views14d: 120, clones14d: 0 });
		expect(body.traffic.clonesSeries).toEqual([]); // failed half → empty series
	});

	it("tolerates a release missing its optional fields", async () => {
		stubCaches();
		vi.stubGlobal(
			"fetch",
			ghResponder({ "/releases": () => jsonRes([{ assets: [] }]) }),
		);
		const body = (await (
			await handleDashboardStats(req(ACCESS), env())
		).json()) as Record<string, any>;
		expect(body.releases).toMatchObject({
			latestTag: null,
			publishedAt: null,
			url: null,
			totalDownloads: 0,
			assets: [],
		});
	});
});
