import { defineConfig } from "vitest/config";

// Unit tests for the edge workers' pure, edge-independent logic (routing, path
// normalization, request qualification, visitor hashing, GitHub aggregation).
// Runs in Node — Cloudflare bindings (request.cf / D1 / KV / caches) are stubbed
// per-test, so binding-level behaviour is also verified by the live smoke test.
export default defineConfig({
	test: {
		include: ["router/src/**/*.test.ts", "go-vanity/src/**/*.test.ts"],
		coverage: {
			provider: "v8",
			// `include` instruments every matching worker source, tested or not
			// (Vitest 4 folded the old `all` option into this), so an untested file
			// counts as 0 rather than vanishing from the report.
			include: ["router/src/**/*.ts", "go-vanity/src/**/*.ts"],
			exclude: ["**/*.test.ts"],
			thresholds: { statements: 90, branches: 90, functions: 90, lines: 90 },
		},
	},
});
