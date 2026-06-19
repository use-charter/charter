// jsdom under this Node build does not expose a working localStorage, so provide
// a minimal Map-backed Storage for the islands that persist preferences.
if (typeof globalThis.localStorage === "undefined") {
	const store = new Map<string, string>();
	const localStoragePolyfill: Storage = {
		getItem: (k) => (store.has(k) ? (store.get(k) as string) : null),
		setItem: (k, v) => void store.set(k, String(v)),
		removeItem: (k) => void store.delete(k),
		clear: () => store.clear(),
		key: (i) => [...store.keys()][i] ?? null,
		get length() {
			return store.size;
		},
	};
	Object.defineProperty(globalThis, "localStorage", {
		value: localStoragePolyfill,
		configurable: true,
	});
	if (typeof window !== "undefined") {
		Object.defineProperty(window, "localStorage", {
			value: localStoragePolyfill,
			configurable: true,
		});
	}
}
