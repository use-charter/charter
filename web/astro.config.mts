import sitemap from '@astrojs/sitemap';
import { defineConfig } from 'astro/config';
import icon from 'astro-icon';

export default defineConfig({
  output: 'static',
  site: 'https://use-charter.dev',
  integrations: [
    icon(),
    sitemap({
      changefreq: 'weekly',
      priority: 1.0,
      lastmod: new Date(),
    }),
  ],
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
