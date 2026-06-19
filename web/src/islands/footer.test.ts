import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { initFooterGlow } from "./footer";

function stubHoverCapable(hoverNone = false) {
	vi.stubGlobal(
		"matchMedia",
		vi.fn((q: string) => ({
			matches: q.includes("hover: none") ? hoverNone : false,
			media: q,
			addEventListener: () => {},
			removeEventListener: () => {},
		})),
	);
	window.matchMedia = globalThis.matchMedia;
}

// rAF runs synchronously so a single pointermove fully exercises the handler.
function stubRaf() {
	vi.stubGlobal("requestAnimationFrame", (cb: FrameRequestCallback) => {
		cb(0);
		return 0;
	});
}

function pointerMove(
	x: number,
	y: number,
	target: EventTarget = document.body,
) {
	const ev = new Event("pointermove", { bubbles: true });
	Object.assign(ev, { clientX: x, clientY: y });
	target.dispatchEvent(ev);
}

beforeEach(() => {
	document.body.innerHTML = '<div data-wordmark></div><a id="link">link</a>';
	const mark = document.querySelector("[data-wordmark]") as HTMLElement;
	// jsdom has no layout; pin the wordmark's box so band math is deterministic.
	mark.getBoundingClientRect = () =>
		({
			left: 0,
			right: 200,
			top: 100,
			bottom: 200,
			width: 200,
			height: 100,
			x: 0,
			y: 100,
			toJSON: () => ({}),
		}) as DOMRect;
});
afterEach(() => vi.unstubAllGlobals());

describe("initFooterGlow", () => {
	it("lights the wordmark when the pointer crosses its glyph band", () => {
		stubHoverCapable(false);
		stubRaf();
		initFooterGlow();
		const mark = document.querySelector("[data-wordmark]") as HTMLElement;
		pointerMove(100, 150); // inside band (top 112 .. 162)
		expect(mark.classList.contains("is-lit")).toBe(true);
		expect(mark.style.getPropertyValue("--mx")).toBe("100px");
		expect(mark.style.getPropertyValue("--my")).toBe("50px");
	});

	it("stays dark above the glyph band", () => {
		stubHoverCapable(false);
		stubRaf();
		initFooterGlow();
		const mark = document.querySelector("[data-wordmark]") as HTMLElement;
		pointerMove(100, 105); // above band start (112)
		expect(mark.classList.contains("is-lit")).toBe(false);
	});

	it("stays dark while hovering interactive footer text", () => {
		stubHoverCapable(false);
		stubRaf();
		initFooterGlow();
		const mark = document.querySelector("[data-wordmark]") as HTMLElement;
		pointerMove(100, 150, document.getElementById("link") as HTMLElement); // over a link
		expect(mark.classList.contains("is-lit")).toBe(false);
	});

	it("no-ops on touch (hover: none) devices", () => {
		stubHoverCapable(true);
		const add = vi.spyOn(document, "addEventListener");
		initFooterGlow();
		expect(add).not.toHaveBeenCalledWith("pointermove", expect.anything());
		add.mockRestore();
	});

	it("lights up when the pointer event targets a non-element (document)", () => {
		stubHoverCapable(false);
		stubRaf();
		initFooterGlow();
		const mark = document.querySelector("[data-wordmark]") as HTMLElement;
		pointerMove(100, 150, document); // target is the document, not an Element
		expect(mark.classList.contains("is-lit")).toBe(true);
	});

	it("no-ops when the wordmark is absent", () => {
		stubHoverCapable(false);
		document.body.innerHTML = "";
		expect(() => initFooterGlow()).not.toThrow();
	});
});
