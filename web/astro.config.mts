import { readFileSync, readdirSync } from 'node:fs';
import sitemap from '@astrojs/sitemap';
import { defineConfig } from 'astro/config';
import icon from 'astro-icon';

// Per-page sitemap lastmod from real content dates, not a uniform build
// timestamp: Google only trusts <lastmod> when it looks accurate and ignores it
// when every URL shares one date that bumps on each deploy. Blog posts carry
// `updated ?? date` in frontmatter; other pages omit lastmod (better none than a
// fabricated value).
const blogDir = new URL('./src/content/blog/', import.meta.url);
const blogLastmod = new Map<string, string>();
for (const file of readdirSync(blogDir)) {
  if (!file.endsWith('.md') || file.startsWith('_')) continue;
  const frontmatter = readFileSync(new URL(file, blogDir), 'utf8').split('---')[1] ?? '';
  const pick = (key: string): string | undefined =>
    frontmatter.match(new RegExp(`^${key}:\\s*(.+)$`, 'm'))?.[1]?.trim().replace(/['"]/g, '');
  const stamp = pick('updated') || pick('date');
  if (stamp) blogLastmod.set(file.replace(/\.md$/, ''), new Date(stamp).toISOString());
}

export default defineConfig({
  output: 'static',
  site: 'https://use-charter.dev',
  integrations: [
    icon(),
    sitemap({
      // The founder dashboard is Cloudflare-Access-gated and unlisted — keep it
      // out of the public sitemap.
      filter: (page) => !page.includes('/dashboard'),
      // Drop priority/changefreq (Google ignores both). Set lastmod only where
      // it can be substantiated (blog frontmatter); omit it everywhere else.
      serialize(item) {
        delete item.changefreq;
        delete item.priority;
        const slug = item.url.match(/\/blog\/([^/]+)\/?$/)?.[1];
        const lastmod = slug ? blogLastmod.get(slug) : undefined;
        if (lastmod) item.lastmod = lastmod;
        else delete item.lastmod;
        return item;
      },
    }),
  ],
  // Astro 7 changed the compressHTML default to 'jsx', which strips whitespace
  // between inline elements (`<span>a</span> <em>b</em>` would render as `ab`).
  // Pin the v6 lossless behaviour so the whitespace policy stays in this one
  // config seam instead of leaking `{" "}` markers into every component.
  compressHTML: true,
  markdown: {
    // No Shiki: it injects an inline dark theme on code blocks that ignores our
    // theme tokens (dark block on a light page, unreadable). Fenced code renders
    // as plain <pre><code>, styled theme-aware by .lg-doc__body pre.
    // Astro 7 renders Markdown with Sätteri by default (GFM + SmartyPants, as
    // before); we use no remark/rehype plugins, so no processor override needed.
    syntaxHighlight: false,
  },
  build: {
    // Inline all CSS into the HTML <style> so there is no render-blocking
    // stylesheet request — critical on high-latency mobile (Slow 4G). The
    // ~50KB compresses to ~10KB and ships in the single HTML download.
    inlineStylesheets: 'always',
  },
  // The Content Security Policy is served as a real header from public/_headers
  // rather than generated here: a static host cannot mint per-request nonces,
  // and a hash-based policy cannot cover the inline styles this site legitimately
  // relies on (syntax-highlight token colours and CSSOM-driven animations).
});
