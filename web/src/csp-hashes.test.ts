import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

// Guards the script-src hashes in public/_headers against the inline scripts they
// authorise. The CSP drops 'unsafe-inline' for scripts and instead allow-lists the
// SHA-256 of each `<script is:inline>` in Base.astro (the theme-resolve and
// scroll-reveal blocks). `is:inline` ships the body verbatim, so hashing the source
// matches what the browser hashes. If either script changes and the hash in
// _headers is not regenerated, the browser would block it — this test fails first.
//
// Scope: Base.astro is the only file with executable inline scripts (every other
// `is:inline` block is non-executing application/ld+json). Add a source here if
// that ever changes.

const read = (rel: string) => readFileSync(fileURLToPath(new URL(rel, import.meta.url)), 'utf8');

const sha256 = (body: string) => `sha256-${createHash('sha256').update(body, 'utf8').digest('base64')}`;

/** Bodies of executable inline scripts (`<script is:inline>` with no type attribute). */
function inlineScriptBodies(astro: string): string[] {
  return [...astro.matchAll(/<script is:inline>([\s\S]*?)<\/script>/g)].map((m) => m[1]);
}

/** sha256-* tokens allow-listed in the CSP's script-src directive. */
function scriptSrcHashes(headers: string): string[] {
  const csp = headers.match(/Content-Security-Policy:\s*(.+)/)?.[1] ?? '';
  const scriptSrc = csp.split(';').find((d) => d.trim().startsWith('script-src')) ?? '';
  return [...scriptSrc.matchAll(/'(sha256-[^']+)'/g)].map((m) => m[1]);
}

describe('CSP script-src hashes', () => {
  const headers = read('../public/_headers');
  const bodies = inlineScriptBodies(read('./layouts/Base.astro'));
  const allowed = scriptSrcHashes(headers);

  it('finds the inline scripts and their allow-listed hashes', () => {
    expect(bodies.length).toBeGreaterThan(0);
    expect(allowed.length).toBe(bodies.length);
  });

  it('allow-lists the hash of every inline script', () => {
    for (const body of bodies) expect(allowed).toContain(sha256(body));
  });

  it('has no stale hashes that match no inline script', () => {
    const live = new Set(bodies.map(sha256));
    for (const hash of allowed) expect(live).toContain(hash);
  });
});
