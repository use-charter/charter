import { describe, expect, it } from "vitest";
import { __setPosts } from "./astro-content.stub";
import { GET } from "../src/pages/blog/rss.xml";

// Two published posts (out of date order) and one draft.
__setPosts([
	{
		id: "older",
		data: {
			title: "Older Post",
			description: "old",
			date: new Date("2026-01-01T00:00:00Z"),
			tags: [],
			draft: false,
		},
	},
	{
		id: "newer",
		data: {
			title: "New & Shiny",
			description: "a <tagged> note",
			date: new Date("2026-02-01T00:00:00Z"),
			tags: ["ai", "ci"],
			draft: false,
		},
	},
	{
		id: "wip",
		data: {
			title: "Draft Post",
			description: "wip",
			date: new Date("2026-03-01T00:00:00Z"),
			tags: [],
			draft: true,
		},
	},
]);

describe("blog RSS feed", () => {
	it("lists published posts newest-first, escapes content, and excludes drafts", async () => {
		const res = await GET({} as Parameters<typeof GET>[0]);
		expect(res.headers.get("Content-Type")).toContain("application/xml");
		const xml = await res.text();

		expect(xml).toContain("<title>Charter Blog</title>");
		expect(xml.indexOf("New &amp; Shiny")).toBeLessThan(
			xml.indexOf("Older Post"),
		); // newest-first
		expect(xml).toContain("<title>New &amp; Shiny</title>"); // escaped title
		expect(xml).toContain("a &lt;tagged&gt; note"); // escaped description
		expect(xml).toContain("<category>ai</category>");
		expect(xml).toContain("<category>ci</category>");
		expect(xml).not.toContain("Draft Post"); // drafts excluded
		expect(xml).toContain("https://use-charter.dev/blog/newer");
		expect(xml).toContain(new Date("2026-02-01T00:00:00Z").toUTCString());
	});
});
