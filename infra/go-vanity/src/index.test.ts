import { describe, expect, it } from "vitest";
import worker from "./index";

const get = (url: string) => worker.fetch(new Request(url));

describe("go-vanity worker", () => {
	it("answers the go tool (?go-get=1) with the go-import meta tags only", async () => {
		const res = await get("https://go.use-charter.dev/charter?go-get=1");
		const body = await res.text();
		expect(res.headers.get("Content-Type")).toBe("text/html; charset=utf-8");
		expect(body).toContain(
			'<meta name="go-import" content="go.use-charter.dev/charter git https://github.com/use-charter/charter">',
		);
		expect(body).toContain('<meta name="go-source"');
		// The go-get response is meta-only: no human redirect.
		expect(body).not.toContain('http-equiv="refresh"');
	});

	it("redirects a human at the module root to pkg.go.dev for the module", async () => {
		const body = await (await get("https://go.use-charter.dev/")).text();
		expect(body).toContain("https://pkg.go.dev/go.use-charter.dev/charter");
		expect(body).toContain('http-equiv="refresh"');
	});

	it("preserves the subpath when redirecting a human", async () => {
		const body = await (
			await get("https://go.use-charter.dev/charter/cmd/charter")
		).text();
		expect(body).toContain(
			"https://pkg.go.dev/go.use-charter.dev/charter/cmd/charter",
		);
	});

	it("strips trailing slashes from the human redirect target", async () => {
		const body = await (
			await get("https://go.use-charter.dev/charter/")
		).text();
		expect(body).toContain("https://pkg.go.dev/go.use-charter.dev/charter");
		expect(body).not.toContain("charter//");
	});

	it("serves the RFC 9116 security.txt as text/plain", async () => {
		const res = await get("https://go.use-charter.dev/.well-known/security.txt");
		const body = await res.text();
		expect(res.status).toBe(200);
		expect(res.headers.get("Content-Type")).toBe("text/plain; charset=utf-8");
		expect(body).toContain("Contact: https://github.com/use-charter/charter/security/advisories/new");
		expect(body).toContain("Canonical: https://use-charter.dev/.well-known/security.txt");
	});
});
