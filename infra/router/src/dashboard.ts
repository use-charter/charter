// Dashboard stats aggregator for use-charter.dev/dashboard.
//
// Served at /dashboard/api/stats by the router. Aggregates the GitHub API into
// the four metric groups the dashboard renders (growth & traffic, releases &
// installs, adoption & signals, community). Cached briefly at the edge so a
// page refresh doesn't re-hit the API. Partial failures degrade gracefully —
// each group is independent and reports its own error rather than failing the
// whole payload.
//
// Auth: the route is gated by Cloudflare Access (only the founder's identity).
// As defense-in-depth this handler also requires an Access assertion header, so
// a request that somehow reaches the worker without passing Access is rejected.

const REPO = "use-charter/charter";
const GH = "https://api.github.com";
const CACHE_TTL = 300; // seconds

export interface DashboardEnv {
  // Fine-grained GitHub token (repo: Contents+Issues+Administration read) — a
  // Worker secret, never in wrangler.toml. Absent → handler returns a
  // "not configured" payload so the page can render a setup state.
  GITHUB_STATS_TOKEN?: string;
}

interface Series {
  day: string;
  count: number;
}

async function gh<T>(path: string, token: string, attempt = 0): Promise<T> {
  const res = await fetch(`${GH}${path}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: "application/vnd.github+json",
      "X-GitHub-Api-Version": "2022-11-28",
      "User-Agent": "charter-dashboard",
    },
  });
  // GitHub's search API enforces a strict secondary (burst) rate limit; a run of
  // back-to-back /search calls can transiently 403/429. Back off and retry.
  if ((res.status === 403 || res.status === 429) && attempt < 2) {
    const retryAfter = Number(res.headers.get("retry-after"));
    const waitMs = Math.min(Number.isFinite(retryAfter) && retryAfter > 0 ? retryAfter : 1.5, 5) * 1000;
    await new Promise((r) => setTimeout(r, waitMs));
    return gh<T>(path, token, attempt + 1);
  }
  if (!res.ok) throw new Error(`${path} → ${res.status}`);
  return (await res.json()) as T;
}

async function group<T>(fn: () => Promise<T>, errors: Record<string, string>, key: string): Promise<T | null> {
  try {
    return await fn();
  } catch (err) {
    errors[key] = err instanceof Error ? err.message : String(err);
    return null;
  }
}

interface RepoResp {
  stargazers_count: number;
  forks_count: number;
  subscribers_count: number;
  open_issues_count: number;
}
interface TrafficResp {
  count: number;
  uniques: number;
  views?: { timestamp: string; count: number; uniques: number }[];
  clones?: { timestamp: string; count: number; uniques: number }[];
}
interface ReleaseResp {
  tag_name: string;
  published_at: string;
  html_url: string;
  assets: { name: string; download_count: number }[];
}
interface SearchResp {
  total_count: number;
  items: { repository?: { full_name: string }; html_url: string }[];
}

async function buildStats(token: string): Promise<unknown> {
  const errors: Record<string, string> = {};

  const repo = await group(() => gh<RepoResp>(`/repos/${REPO}`, token), errors, "growth");
  const views = await group(() => gh<TrafficResp>(`/repos/${REPO}/traffic/views`, token), errors, "traffic_views");
  const clones = await group(() => gh<TrafficResp>(`/repos/${REPO}/traffic/clones`, token), errors, "traffic_clones");
  const releases = await group(() => gh<ReleaseResp[]>(`/repos/${REPO}/releases?per_page=10`, token), errors, "releases");
  const actionUse = await group(
    () => gh<SearchResp>(`/search/code?q=${encodeURIComponent('"use-charter/charter-action"')}&per_page=20`, token),
    errors,
    "adoption_action",
  );
  const schemaUse = await group(
    () => gh<SearchResp>(`/search/code?q=${encodeURIComponent('"use-charter.dev/schema/charter.schema.json"')}&per_page=20`, token),
    errors,
    "adoption_schema",
  );
  const openIssues = await group(
    () => gh<SearchResp>(`/search/issues?q=${encodeURIComponent(`repo:${REPO} type:issue state:open`)}&per_page=1`, token),
    errors,
    "community_open",
  );
  const closedIssues = await group(
    () => gh<SearchResp>(`/search/issues?q=${encodeURIComponent(`repo:${REPO} type:issue state:closed`)}&per_page=1`, token),
    errors,
    "community_closed",
  );
  const openPRs = await group(
    () => gh<SearchResp>(`/search/issues?q=${encodeURIComponent(`repo:${REPO} type:pr state:open`)}&per_page=1`, token),
    errors,
    "community_prs",
  );

  const series = (t: TrafficResp | null, k: "views" | "clones"): Series[] =>
    (t?.[k] ?? []).map((p) => ({ day: p.timestamp.slice(0, 10), count: p.count }));

  const adopters = [
    ...(actionUse?.items ?? []).map((i) => i.repository?.full_name).filter((n): n is string => !!n && !n.startsWith("use-charter/")),
  ];

  return {
    generatedAt: new Date().toISOString(),
    repo: REPO,
    growth: repo && {
      stars: repo.stargazers_count,
      forks: repo.forks_count,
      watchers: repo.subscribers_count,
      openIssues: repo.open_issues_count,
    },
    traffic: (views || clones) && {
      views14d: views?.count ?? 0,
      uniqueViews14d: views?.uniques ?? 0,
      clones14d: clones?.count ?? 0,
      uniqueClones14d: clones?.uniques ?? 0,
      viewsSeries: series(views, "views"),
      clonesSeries: series(clones, "clones"),
    },
    releases: releases &&
      releases.length > 0 && {
        latestTag: releases[0]?.tag_name ?? null,
        publishedAt: releases[0]?.published_at ?? null,
        url: releases[0]?.html_url ?? null,
        totalDownloads: releases.reduce((s, r) => s + r.assets.reduce((a, x) => a + x.download_count, 0), 0),
        assets: (releases[0]?.assets ?? []).map((a) => ({ name: a.name, downloads: a.download_count })),
      },
    adoption: {
      actionRepos: actionUse?.total_count ?? 0,
      schemaRefs: schemaUse?.total_count ?? 0,
      sampleAdopters: [...new Set(adopters)].slice(0, 8),
    },
    community: {
      openIssues: openIssues?.total_count ?? 0,
      closedIssues: closedIssues?.total_count ?? 0,
      openPRs: openPRs?.total_count ?? 0,
    },
    errors,
  };
}

export async function handleDashboardStats(request: Request, env: DashboardEnv): Promise<Response> {
  // Defense-in-depth: only requests that passed Cloudflare Access carry these.
  const accessed =
    request.headers.has("Cf-Access-Jwt-Assertion") || request.headers.has("Cf-Access-Authenticated-User-Email");
  if (!accessed) {
    return json({ error: "forbidden" }, 403);
  }

  if (!env.GITHUB_STATS_TOKEN) {
    return json({ error: "not_configured", message: "Set the GITHUB_STATS_TOKEN worker secret to enable the dashboard." }, 200);
  }

  // Short edge cache keyed on the request URL.
  const cache = caches.default;
  const cacheKey = new Request(new URL(request.url).toString(), { method: "GET" });
  const hit = await cache.match(cacheKey);
  if (hit) return hit;

  const stats = await buildStats(env.GITHUB_STATS_TOKEN);
  const hadErrors = Object.keys((stats as { errors?: Record<string, string> }).errors ?? {}).length > 0;
  // Don't cache partial results — otherwise a transient per-group failure would
  // stick (and outlive the Refresh button) for the whole TTL.
  const res = json(stats, 200, { "Cache-Control": hadErrors ? "no-store" : `public, max-age=${CACHE_TTL}` });
  if (!hadErrors) await cache.put(cacheKey, res.clone());
  return res;
}

function json(body: unknown, status: number, extra?: Record<string, string>): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json; charset=utf-8", ...(extra ?? {}) },
  });
}
