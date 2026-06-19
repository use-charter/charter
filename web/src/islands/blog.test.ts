import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// blog.ts is an entry-point island: it runs on import. Each test seeds the DOM
// and globals, then imports a fresh module instance.
let ioCallback:
	| ((entries: { isIntersecting: boolean; target: Element }[]) => void)
	| null = null;

function stubGlobals() {
	vi.stubGlobal("requestAnimationFrame", (cb: FrameRequestCallback) => {
		cb(0);
		return 1;
	});
	vi.stubGlobal(
		"matchMedia",
		vi.fn((q: string) => ({
			matches: false,
			media: q,
			addEventListener: () => {},
			removeEventListener: () => {},
		})),
	);
	window.matchMedia = globalThis.matchMedia;
	class IO {
		constructor(
			cb: (e: { isIntersecting: boolean; target: Element }[]) => void,
		) {
			ioCallback = cb;
		}
		observe() {}
		unobserve() {}
		disconnect() {}
		takeRecords() {
			return [];
		}
	}
	vi.stubGlobal("IntersectionObserver", IO);
}

const article = () => `
  <div data-progress></div>
  <article data-post>
    <nav><a data-toc="s1"></a><a data-toc="s2"></a></nav>
    <div class="bl-prose"><h2 id="s1">A</h2><h3 id="s2">B</h3></div>
  </article>`;

beforeEach(() => {
	ioCallback = null;
	vi.resetModules();
	stubGlobals();
});
afterEach(() => vi.unstubAllGlobals());

describe("blog island", () => {
	it("sets the reading-progress transform from the article scroll position", async () => {
		document.body.innerHTML = article();
		const post = document.querySelector("[data-post]") as HTMLElement;
		Object.defineProperty(post, "offsetHeight", {
			value: 5000,
			configurable: true,
		});
		post.getBoundingClientRect = () => ({ top: -1000 }) as DOMRect;
		await import("./blog");
		const bar = document.querySelector("[data-progress]") as HTMLElement;
		expect(bar.style.transform).toMatch(/^scaleX\(/);

		window.dispatchEvent(new Event("scroll"));
		expect(bar.style.transform).not.toBe("scaleX(0)");
	});

	it("caps progress at 1 when the article is shorter than the viewport", async () => {
		document.body.innerHTML = article();
		const post = document.querySelector("[data-post]") as HTMLElement;
		Object.defineProperty(post, "offsetHeight", {
			value: 10,
			configurable: true,
		}); // < innerHeight
		await import("./blog");
		const bar = document.querySelector("[data-progress]") as HTMLElement;
		expect(bar.style.transform).toBe("scaleX(1)");
	});

	it("activates the TOC link for the heading in view and switches as it changes", async () => {
		document.body.innerHTML = article();
		await import("./blog");
		expect(ioCallback).toBeTypeOf("function");
		const h1 = document.getElementById("s1") as HTMLElement;
		const h2 = document.getElementById("s2") as HTMLElement;
		h1.getBoundingClientRect = () => ({ top: 10 }) as DOMRect;
		h2.getBoundingClientRect = () => ({ top: 400 }) as DOMRect;

		ioCallback?.([{ isIntersecting: true, target: h1 }]);
		expect(
			document
				.querySelector('[data-toc="s1"]')
				?.classList.contains("is-active"),
		).toBe(true);

		ioCallback?.([{ isIntersecting: true, target: h2 }]);
		expect(
			document
				.querySelector('[data-toc="s2"]')
				?.classList.contains("is-active"),
		).toBe(true);
		expect(
			document
				.querySelector('[data-toc="s1"]')
				?.classList.contains("is-active"),
		).toBe(false);
	});

	it("ignores valueless TOC links, unlinked headings, and empty observer batches", async () => {
		document.body.innerHTML = `
      <div data-progress></div>
      <article data-post>
        <nav><a data-toc="s1"></a><a data-toc></a></nav>
        <div class="bl-prose"><h2 id="s1">A</h2><h3>no id</h3></div>
      </article>`;
		await import("./blog");
		// An observer callback with nothing intersecting must not change the active link.
		ioCallback?.([
			{
				isIntersecting: false,
				target: document.getElementById("s1") as Element,
			},
		]);
		expect(
			document
				.querySelector('[data-toc="s1"]')
				?.classList.contains("is-active"),
		).toBe(false);
	});

	it("no-ops the progress bar when the article markup is absent", async () => {
		document.body.innerHTML = "<div data-progress></div>"; // no [data-post]
		await import("./blog");
		window.dispatchEvent(new Event("scroll"));
		const bar = document.querySelector("[data-progress]") as HTMLElement;
		expect(bar.style.transform).toBe("");
	});
});
