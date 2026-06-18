import { describe, expect, it, vi } from 'vitest';
import { handleAnalytics } from './analytics-read';

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
  const db = { prepare: () => ({ bind: () => ({}) }), batch } as unknown as D1Database;
  return { db, batch };
}

const req = (headers: Record<string, string> = {}) =>
  new Request('https://use-charter.dev/dashboard/api/analytics', { headers });

describe('handleAnalytics', () => {
  it('rejects requests without a Cloudflare Access assertion (403) and never queries D1', async () => {
    const { db, batch } = stubDb();
    const res = await handleAnalytics(req(), { ANALYTICS_DB: db });
    expect(res.status).toBe(403);
    expect(batch).not.toHaveBeenCalled();
  });

  it('returns the aggregate shape for an authenticated request', async () => {
    const { db } = stubDb();
    const res = await handleAnalytics(req({ 'Cf-Access-Authenticated-User-Email': 'founder@use-charter.dev' }), {
      ANALYTICS_DB: db,
    });
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
    expect(typeof body.generatedAt).toBe('string');
  });
});
