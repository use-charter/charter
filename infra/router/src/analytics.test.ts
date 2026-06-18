import { describe, expect, it } from 'vitest';
import { normalizePath, originAllowed, parseEventType, qualifies, utcDay, visitorHash, type QualifyParts } from './analytics';

describe('normalizePath', () => {
  it('keeps root and known top-level pages', () => {
    expect(normalizePath('/')).toBe('/');
    expect(normalizePath('/blog')).toBe('/blog');
  });
  it('keeps bounded blog and legal slugs', () => {
    expect(normalizePath('/blog/introducing-charter')).toBe('/blog/introducing-charter');
    expect(normalizePath('/legal/privacy')).toBe('/legal/privacy');
  });
  it('collapses proxied doc families to one bucket each', () => {
    expect(normalizePath('/docs')).toBe('/docs/*');
    expect(normalizePath('/docs/quickstart')).toBe('/docs/*');
    expect(normalizePath('/cli/doctor')).toBe('/cli/*');
    expect(normalizePath('/rules/AE-SEC-001')).toBe('/rules/*');
    expect(normalizePath('/changelog/v1-0-0')).toBe('/changelog/*');
  });
  it('buckets unknown/scanner paths into /__other__', () => {
    expect(normalizePath('/wp-login.php')).toBe('/__other__');
    expect(normalizePath('/.env')).toBe('/__other__');
  });
  it('normalizes trailing slash and case', () => {
    expect(normalizePath('/Blog/')).toBe('/blog');
    expect(normalizePath('/legal/Terms/')).toBe('/legal/terms');
  });
});

const ok: QualifyParts = {
  method: 'GET',
  path: '/blog/introducing-charter',
  status: 200,
  contentType: 'text/html; charset=utf-8',
  ua: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
  secPurpose: '',
};

describe('qualifies', () => {
  it('counts a successful HTML GET from a real browser', () => {
    expect(qualifies(ok)).toBe(true);
  });
  it('rejects non-GET methods', () => {
    expect(qualifies({ ...ok, method: 'POST' })).toBe(false);
  });
  it('rejects redirects and errors (not 2xx)', () => {
    expect(qualifies({ ...ok, status: 301 })).toBe(false);
    expect(qualifies({ ...ok, status: 404 })).toBe(false);
    expect(qualifies({ ...ok, status: 500 })).toBe(false);
  });
  it('rejects non-HTML responses (assets, xml)', () => {
    expect(qualifies({ ...ok, contentType: 'application/xml' })).toBe(false);
    expect(qualifies({ ...ok, contentType: 'text/css' })).toBe(false);
  });
  it('excludes api, dashboard, and mintlify asset paths', () => {
    expect(qualifies({ ...ok, path: '/api/event' })).toBe(false);
    expect(qualifies({ ...ok, path: '/dashboard' })).toBe(false);
    expect(qualifies({ ...ok, path: '/dashboard/api/analytics' })).toBe(false);
    expect(qualifies({ ...ok, path: '/mintlify-assets/x.css' })).toBe(false);
  });
  it('skips browser speculative prefetch', () => {
    expect(qualifies({ ...ok, secPurpose: 'prefetch' })).toBe(false);
    expect(qualifies({ ...ok, secPurpose: 'prefetch;prerender' })).toBe(false);
  });
  it('skips known bots', () => {
    expect(qualifies({ ...ok, ua: 'Googlebot/2.1 (+http://www.google.com/bot.html)' })).toBe(false);
  });
});

describe('visitorHash', () => {
  it('is deterministic within a day (same salt/ip/ua)', async () => {
    const a = await visitorHash('salt-A', '203.0.113.5', 'UA-1');
    const b = await visitorHash('salt-A', '203.0.113.5', 'UA-1');
    expect(a).toBe(b);
  });
  it('differs across days (salt rotation) and across visitors', async () => {
    const day1 = await visitorHash('salt-A', '203.0.113.5', 'UA-1');
    const day2 = await visitorHash('salt-B', '203.0.113.5', 'UA-1');
    const other = await visitorHash('salt-A', '198.51.100.9', 'UA-1');
    expect(day1).not.toBe(day2);
    expect(day1).not.toBe(other);
  });
  it('never leaks the raw IP or user-agent', async () => {
    const ip = '203.0.113.5';
    const ua = 'Mozilla/5.0 SecretAgent';
    const h = await visitorHash('salt-A', ip, ua);
    expect(h).not.toContain(ip);
    expect(h).not.toContain('SecretAgent');
    expect(h).toMatch(/^[A-Za-z0-9_-]+$/); // base64url, no padding
  });
});

describe('utcDay', () => {
  it('formats a UTC calendar day', () => {
    expect(utcDay(new Date('2031-03-04T23:59:00Z'))).toBe('2031-03-04');
  });
});

const evt = (headers: Record<string, string>) => new Request('https://example/api/event', { method: 'POST', headers });

describe('originAllowed', () => {
  it('accepts requests originating from the site', () => {
    expect(originAllowed(evt({ Origin: 'https://use-charter.dev' }))).toBe(true);
    expect(originAllowed(evt({ Referer: 'https://use-charter.dev/blog/x' }))).toBe(true);
  });
  it('rejects foreign or missing origins', () => {
    expect(originAllowed(evt({ Origin: 'https://evil.example' }))).toBe(false);
    expect(originAllowed(evt({}))).toBe(false);
  });
});

describe('parseEventType', () => {
  it('returns an allow-listed type', () => {
    expect(parseEventType('{"type":"install_copied"}')).toBe('install_copied');
  });
  it('rejects unknown types, non-strings, malformed, and oversized bodies', () => {
    expect(parseEventType('{"type":"hack"}')).toBeNull();
    expect(parseEventType('{"type":123}')).toBeNull();
    expect(parseEventType('not json')).toBeNull();
    expect(parseEventType(`{"type":"install_copied","pad":"${'x'.repeat(1100)}"}`)).toBeNull();
  });
});
