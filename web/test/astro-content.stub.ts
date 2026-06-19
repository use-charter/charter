// Test stub for the `astro:content` virtual module (unavailable outside an Astro
// build). Routes import `getCollection`; tests seed the data via `__setPosts`.
// Aliased in vitest.config.ts; both the route and the test resolve to this one
// module instance, so the seeded posts are shared.

// biome-ignore lint/suspicious/noExplicitAny: test fixture mirrors Astro's loose collection entry shape.
type Entry = any;

let posts: Entry[] = [];

export function __setPosts(next: Entry[]): void {
	posts = next;
}

export async function getCollection(
	_name: string,
	filter?: (entry: Entry) => boolean,
): Promise<Entry[]> {
	return filter ? posts.filter(filter) : posts;
}
