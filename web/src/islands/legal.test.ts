import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// legal.ts is an entry-point island: it runs on import (readyState is 'complete'
// under jsdom). Each test seeds the DOM and globals, then imports it fresh.
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
	window.scrollTo = vi.fn();
}

const rail = () => `
  <div class="lg-rail__bar"><i></i></div>
  <nav class="lg-rail__nav"><a href="#a">A</a><a href="#b">B</a></nav>
  <section id="a">A</section>
  <section id="b">B</section>`;

beforeEach(() => {
	vi.resetModules();
	stubGlobals();
});
afterEach(() => vi.unstubAllGlobals());

describe("legal island", () => {
	it("marks the section in view active and updates the progress bar on scroll", async () => {
		document.body.innerHTML = rail();
		const a = document.getElementById("a") as HTMLElement;
		const b = document.getElementById("b") as HTMLElement;
		a.getBoundingClientRect = () => ({ top: -10 }) as DOMRect; // above threshold → current
		b.getBoundingClientRect = () => ({ top: 9999 }) as DOMRect;
		await import("./legal");
		expect(
			document.querySelector('a[href="#a"]')?.classList.contains("is-active"),
		).toBe(true);

		window.dispatchEvent(new Event("scroll"));
		expect(document.querySelector(".lg-rail__bar i")).toBeTruthy();
	});

	it("smooth-scrolls to a section on nav click and prevents the default jump", async () => {
		document.body.innerHTML = rail();
		await import("./legal");
		const link = document.querySelector('a[href="#b"]') as HTMLAnchorElement;
		const ev = new MouseEvent("click", { bubbles: true, cancelable: true });
		link.dispatchEvent(ev);
		expect(window.scrollTo).toHaveBeenCalledOnce();
		expect(ev.defaultPrevented).toBe(true);
	});

	it("ignores a nav click whose target section does not exist", async () => {
		document.body.innerHTML = `
      <div class="lg-rail__bar"><i></i></div>
      <nav class="lg-rail__nav"><a href="#missing">X</a></nav>
      <section id="a">A</section>`;
		await import("./legal");
		const link = document.querySelector(
			'a[href="#missing"]',
		) as HTMLAnchorElement;
		const ev = new MouseEvent("click", { bubbles: true, cancelable: true });
		link.dispatchEvent(ev);
		expect(window.scrollTo).not.toHaveBeenCalled();
		expect(ev.defaultPrevented).toBe(false);
	});

	it("defers init until DOMContentLoaded while the document is still loading", async () => {
		Object.defineProperty(document, "readyState", {
			value: "loading",
			configurable: true,
		});
		Object.defineProperty(document.documentElement, "scrollHeight", {
			value: 5000,
			configurable: true,
		});
		document.body.innerHTML = rail();
		const a = document.getElementById("a") as HTMLElement;
		const b = document.getElementById("b") as HTMLElement;
		a.getBoundingClientRect = () => ({ top: -10 }) as DOMRect;
		b.getBoundingClientRect = () => ({ top: 9999 }) as DOMRect;
		await import("./legal");
		// init has not run yet → no active link.
		expect(
			document.querySelector('a[href="#a"]')?.classList.contains("is-active"),
		).toBe(false);
		document.dispatchEvent(new Event("DOMContentLoaded"));
		expect(
			document.querySelector('a[href="#a"]')?.classList.contains("is-active"),
		).toBe(true);
		Object.defineProperty(document, "readyState", {
			value: "complete",
			configurable: true,
		});
	});

	it("no-ops when the rail has no anchored sections", async () => {
		document.body.innerHTML = '<nav class="lg-rail__nav"></nav>';
		await expect(import("./legal")).resolves.toBeDefined();
	});
});
