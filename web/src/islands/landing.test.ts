import { afterEach, describe, expect, it, vi } from "vitest";
import { sendCopyEvent } from "./landing";

describe("sendCopyEvent", () => {
	afterEach(() => vi.unstubAllGlobals());

	it("beacons the opted-in event type to the same-origin endpoint", async () => {
		const beacon = vi.fn(() => true);
		vi.stubGlobal("navigator", { sendBeacon: beacon });

		const host = document.createElement("span");
		host.dataset.copyEvent = "install_copied";
		sendCopyEvent(host);

		expect(beacon).toHaveBeenCalledOnce();
		const [url, body] = beacon.mock.calls[0] as unknown as [string, Blob];
		expect(url).toBe("/api/event");
		expect(body).toBeInstanceOf(Blob);
		expect(body.type).toBe("text/plain");
		expect(JSON.parse(await body.text())).toEqual({ type: "install_copied" });
	});

	it("does nothing for a host without a copy-event", () => {
		const beacon = vi.fn(() => true);
		vi.stubGlobal("navigator", { sendBeacon: beacon });
		sendCopyEvent(document.createElement("span"));
		expect(beacon).not.toHaveBeenCalled();
	});

	it("does nothing when sendBeacon is unavailable", () => {
		vi.stubGlobal("navigator", {});
		const host = document.createElement("span");
		host.dataset.copyEvent = "install_copied";
		expect(() => sendCopyEvent(host)).not.toThrow();
	});
});
