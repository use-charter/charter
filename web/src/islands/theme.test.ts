import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { initThemeSwitch } from "./theme";

// Controllable matchMedia: `dark` drives prefers-color-scheme, and the registered
// change listener is captured so a system theme flip can be simulated.
function stubMatchMedia(dark = false) {
	let cb: (() => void) | null = null;
	const mql = {
		matches: dark,
		media: "",
		addEventListener: (_e: string, fn: () => void) => {
			cb = fn;
		},
		removeEventListener: () => {},
	};
	vi.stubGlobal(
		"matchMedia",
		vi.fn(() => mql),
	);
	window.matchMedia = globalThis.matchMedia;
	return { mql, fire: () => cb?.() };
}

function mountPill() {
	document.body.innerHTML = `
    <div data-theme-switch>
      <button data-theme-seg="system"></button>
      <button data-theme-seg="light"></button>
      <button data-theme-seg="dark"></button>
    </div>`;
	return Array.from(
		document.querySelectorAll<HTMLButtonElement>("[data-theme-seg]"),
	);
}

const seg = (mode: string) =>
	document.querySelector<HTMLButtonElement>(
		`[data-theme-seg="${mode}"]`,
	) as HTMLButtonElement;

beforeEach(() => {
	localStorage.clear();
	document.documentElement.removeAttribute("data-theme");
	document.documentElement.removeAttribute("data-theme-mode");
});
afterEach(() => vi.unstubAllGlobals());

describe("initThemeSwitch", () => {
	it("defaults to system mode and resolves the theme from the OS preference", () => {
		stubMatchMedia(true); // OS prefers dark
		mountPill();
		initThemeSwitch();
		expect(document.documentElement.dataset.themeMode).toBe("system");
		expect(document.documentElement.dataset.theme).toBe("dark");
		expect(seg("system").getAttribute("aria-checked")).toBe("true");
	});

	it("restores a persisted light mode and reflects it on the pill", () => {
		localStorage.setItem("charter-theme", "light");
		stubMatchMedia(true);
		mountPill();
		initThemeSwitch();
		expect(document.documentElement.dataset.theme).toBe("light"); // forced, ignores OS
		expect(seg("light").getAttribute("aria-checked")).toBe("true");
	});

	it("switches mode on segment click and persists it", () => {
		stubMatchMedia(false);
		mountPill();
		initThemeSwitch();
		seg("dark").click();
		expect(document.documentElement.dataset.theme).toBe("dark");
		expect(localStorage.getItem("charter-theme")).toBe("dark");
		expect(seg("dark").getAttribute("aria-checked")).toBe("true");
		expect(seg("system").getAttribute("aria-checked")).toBe("false");
	});

	it("roves with the arrow keys across the segmented group", () => {
		stubMatchMedia(false);
		mountPill();
		initThemeSwitch();
		const pill = document.querySelector("[data-theme-switch]") as HTMLElement;
		pill.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight" })); // system → light
		expect(document.documentElement.dataset.themeMode).toBe("light");
		pill.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowLeft" })); // light → system
		expect(document.documentElement.dataset.themeMode).toBe("system");
		pill.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter" })); // ignored
		expect(document.documentElement.dataset.themeMode).toBe("system");
	});

	it("live-updates the resolved theme on an OS change while in system mode", () => {
		const mm = stubMatchMedia(false);
		mountPill();
		initThemeSwitch();
		expect(document.documentElement.dataset.theme).toBe("light");
		mm.mql.matches = true; // OS flips to dark
		mm.fire();
		expect(document.documentElement.dataset.theme).toBe("dark");
	});

	it("does not live-update once a fixed mode is chosen", () => {
		const mm = stubMatchMedia(false);
		mountPill();
		initThemeSwitch();
		seg("light").click();
		mm.mql.matches = true;
		mm.fire();
		expect(document.documentElement.dataset.theme).toBe("light"); // stays forced
	});

	it("ignores a click on a segment with no/invalid mode", () => {
		stubMatchMedia(false);
		document.body.innerHTML = `
      <div data-theme-switch>
        <button data-theme-seg="system"></button>
        <button data-theme-seg></button>
      </div>`;
		initThemeSwitch();
		const bad =
			document.querySelectorAll<HTMLButtonElement>("[data-theme-seg]")[1];
		bad.click(); // empty data-theme-seg → no mode change
		expect(document.documentElement.dataset.themeMode).toBe("system");
	});

	it("no-ops when there is no pill on the page", () => {
		stubMatchMedia(false);
		document.body.innerHTML = "";
		expect(() => initThemeSwitch()).not.toThrow();
	});

	it("falls back to system when localStorage is unreadable", () => {
		stubMatchMedia(false);
		const getItem = vi.spyOn(localStorage, "getItem").mockImplementation(() => {
			throw new Error("blocked");
		});
		mountPill();
		expect(() => initThemeSwitch()).not.toThrow();
		expect(document.documentElement.dataset.themeMode).toBe("system");
		getItem.mockRestore();
	});

	it("tolerates a write failure when persisting", () => {
		stubMatchMedia(false);
		mountPill();
		initThemeSwitch();
		const setItem = vi.spyOn(localStorage, "setItem").mockImplementation(() => {
			throw new Error("quota");
		});
		expect(() => seg("dark").click()).not.toThrow();
		expect(document.documentElement.dataset.theme).toBe("dark"); // selection still applied
		setItem.mockRestore();
	});
});
