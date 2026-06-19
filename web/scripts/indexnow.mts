// Submit the site's canonical URLs to IndexNow so participating engines (Bing,
// Yandex, Seznam, Naver, Yep) recrawl promptly after a content change. IndexNow
// is free and needs no account — ownership is proven by hosting <KEY>.txt at the
// site root (public/c89d949c2d8cba4afbb3a7104a80e5f8.txt). Google does not
// consume IndexNow; Search Console + the sitemap cover Google.
//
//   bun scripts/indexnow.mts            # submit production URLs
//   bun scripts/indexnow.mts <baseUrl>  # override the host (e.g. a staging URL)

const KEY = 'c89d949c2d8cba4afbb3a7104a80e5f8';
const base = new URL(process.argv[2] ?? 'https://use-charter.dev');

// The URL set is the sitemap — the single source of truth for canonical pages.
const sitemap = await fetch(new URL('/sitemap-0.xml', base)).then((r) => r.text());
const urlList = [...sitemap.matchAll(/<loc>([^<]+)<\/loc>/g)].map((m) => m[1]);
if (urlList.length === 0) throw new Error(`No <loc> URLs found in ${base.origin}/sitemap-0.xml`);

const res = await fetch('https://api.indexnow.org/indexnow', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json; charset=utf-8' },
  body: JSON.stringify({ host: base.host, key: KEY, keyLocation: `${base.origin}/${KEY}.txt`, urlList }),
});

console.log(`IndexNow: submitted ${urlList.length} URL(s) -> ${res.status} ${res.statusText}`);
if (!res.ok) process.exit(1);
