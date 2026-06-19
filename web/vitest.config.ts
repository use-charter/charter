import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

export default defineConfig({
	resolve: {
		// `astro:content` is an Astro build-time virtual module; route handlers that
		// import it resolve to a controllable stub under test.
		alias: {
			"astro:content": fileURLToPath(
				new URL("./test/astro-content.stub.ts", import.meta.url),
			),
		},
	},
	test: {
		environment: "jsdom",
		// A concrete origin so jsdom exposes a working localStorage (about:blank is opaque).
		environmentOptions: { jsdom: { url: "https://use-charter.dev" } },
		globals: true,
		setupFiles: ["./test/setup.ts"],
		coverage: {
			provider: "v8",
			// `include` instruments every matching module, tested or not (Vitest 4
			// folded the old `all` option into this), so untested files count as 0.
			include: ["src/**/*.ts"],
			exclude: [
				"**/*.test.ts",
				"src/content.config.ts", // declarative Astro content schema (needs the astro:content runtime)
				"src/pages/og/**", // build-time OG image rendering (satori + resvg native deps, no business logic)
				// Animation/visual islands: a hero boot sequence and a fetch-driven chart
				// dashboard, both rAF/canvas-timing heavy. Per web/testing.md these are
				// visual-regression territory (verified by manual QA + the 100 Lighthouse
				// run), where unit assertions are brittle and low-signal. sendCopyEvent —
				// the one piece of pure logic in landing.ts — is still unit-tested.
				"src/islands/landing.ts",
				"src/islands/dashboard.ts",
			],
			thresholds: { statements: 90, branches: 90, functions: 90, lines: 90 },
		},
	},
});
