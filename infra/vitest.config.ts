import { defineConfig } from 'vitest/config';

// Unit tests for the edge workers' pure, edge-independent logic (path
// normalization, request qualification, visitor hashing). Runs in Node — no
// Cloudflare bindings — so request.cf / D1 / KV are deliberately out of scope
// here; binding-level behaviour is verified by the live smoke test.
export default defineConfig({
  test: {
    include: ['router/src/**/*.test.ts', 'go-vanity/src/**/*.test.ts'],
  },
});
