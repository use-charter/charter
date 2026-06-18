// Read side of the founder analytics (ADR-0027): aggregates the D1 counters
// into the JSON the dashboard renders. Gated by Cloudflare Access — the route
// lives under /dashboard*, and this handler additionally requires an Access
// assertion header as defense-in-depth. The Cache API is unavailable behind
// Access, so freshness comes from a short in-isolate memo over direct reads.

const RANGE_DAYS = 30;
const MEMO_MS = 30_000;

export interface AnalyticsReadEnv {
  ANALYTICS_DB: D1Database;
}

interface DayCount {
  day: string;
  count: number;
}
interface PathHits {
  path: string;
  hits: number;
}
interface CountryHits {
  country: string;
  hits: number;
}
interface Scalar {
  n: number;
}

let memo: { at: number; body: string } | null = null;

/** The UTC day `days` days before today, `YYYY-MM-DD` — the window's lower bound. */
function windowStart(days: number, now: Date = new Date()): string {
  const d = new Date(now);
  d.setUTCDate(d.getUTCDate() - days);
  return d.toISOString().slice(0, 10);
}

async function build(db: D1Database): Promise<string> {
  const since = windowStart(RANGE_DAYS - 1);
  const q = (sql: string) => db.prepare(sql).bind(since);
  const [pv, uq, tp, blog, docs, copies, geo] = await db.batch([
    q('SELECT day, SUM(hits) AS count FROM pageview WHERE day >= ? GROUP BY day ORDER BY day'),
    q('SELECT day, COUNT(*) AS count FROM visitor WHERE day >= ? GROUP BY day ORDER BY day'),
    q('SELECT path, SUM(hits) AS hits FROM pageview WHERE day >= ? GROUP BY path ORDER BY hits DESC LIMIT 10'),
    q("SELECT COALESCE(SUM(hits), 0) AS n FROM pageview WHERE day >= ? AND (path = '/blog' OR path LIKE '/blog/%')"),
    q("SELECT COALESCE(SUM(hits), 0) AS n FROM pageview WHERE day >= ? AND path = '/docs/*'"),
    q("SELECT COUNT(*) AS n FROM event WHERE day >= ? AND type = 'install_copied'"),
    q('SELECT country, SUM(hits) AS hits FROM geo WHERE day >= ? GROUP BY country ORDER BY hits DESC LIMIT 5'),
  ]);

  const scalar = (r: D1Result): number => ((r.results as Scalar[])[0]?.n ?? 0);
  return JSON.stringify({
    generatedAt: new Date().toISOString(),
    rangeDays: RANGE_DAYS,
    pageviewsByDay: (pv.results as DayCount[]).map((r) => ({ day: r.day, count: Number(r.count) })),
    uniquesByDay: (uq.results as DayCount[]).map((r) => ({ day: r.day, count: Number(r.count) })),
    topPages: (tp.results as PathHits[]).map((r) => ({ path: r.path, hits: Number(r.hits) })),
    blogViews: scalar(blog),
    docsViews: scalar(docs),
    events: { install_copied: scalar(copies) },
    topCountries: (geo.results as CountryHits[]).map((r) => ({ country: r.country, hits: Number(r.hits) })),
  });
}

function json(body: string, status: number): Response {
  return new Response(body, {
    status,
    headers: { 'Content-Type': 'application/json; charset=utf-8', 'Cache-Control': 'no-store' },
  });
}

export async function handleAnalytics(request: Request, env: AnalyticsReadEnv): Promise<Response> {
  // Defense-in-depth: only requests that passed Cloudflare Access carry these.
  const accessed =
    request.headers.has('Cf-Access-Jwt-Assertion') || request.headers.has('Cf-Access-Authenticated-User-Email');
  if (!accessed) return json(JSON.stringify({ error: 'forbidden' }), 403);

  const now = Date.now();
  if (memo && now - memo.at < MEMO_MS) return json(memo.body, 200);

  const body = await build(env.ANALYTICS_DB);
  memo = { at: now, body };
  return json(body, 200);
}
