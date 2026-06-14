import type { APIRoute } from 'astro';

// Dynamic robots.txt — reuses `site` from astro.config so the Sitemap URL can
// never drift. Charter is an open-source tool that WANTS to be surfaced in
// search and AI answers (GEO/LLMO), so every crawler — including AI training
// and AI-search agents — is explicitly allowed. No private paths exist.
const AI_AGENTS = [
  'GPTBot', // OpenAI crawl
  'OAI-SearchBot', // OpenAI search index
  'ChatGPT-User', // ChatGPT browsing on user request
  'ClaudeBot', // Anthropic crawl
  'Claude-User', // Claude browsing on user request
  'Claude-SearchBot', // Claude search index
  'anthropic-ai', // Anthropic (legacy token)
  'Google-Extended', // Gemini / Vertex AI training
  'Applebot-Extended', // Apple Intelligence training
  'PerplexityBot', // Perplexity crawl
  'Perplexity-User', // Perplexity browsing on user request
  'Amazonbot', // Amazon / Alexa
  'Meta-ExternalAgent', // Meta AI crawl
  'Bytespider', // ByteDance / TikTok
  'CCBot', // Common Crawl (feeds many models)
  'cohere-ai', // Cohere
  'Diffbot', // Diffbot
  'DuckAssistBot', // DuckDuckGo AI assist
  'MistralAI-User', // Mistral browsing on user request
  'YouBot', // You.com
];

const getRobotsTxt = (sitemapURL: URL, docsSitemapURL: URL) => `# https://www.robotstxt.org/robotstxt.html

# Default: every crawler may access the entire site.
User-agent: *
Allow: /

# AI crawlers and assistants are explicitly welcome.
${AI_AGENTS.map((ua) => `User-agent: ${ua}\nAllow: /`).join('\n\n')}

# Landing + legal pages (this Astro site) and the Mintlify product docs each
# publish their own sitemap; list both so crawlers discover every URL.
Sitemap: ${sitemapURL.href}
Sitemap: ${docsSitemapURL.href}
`;

export const GET: APIRoute = ({ site }) => {
  // `site` is guaranteed because astro.config sets it; fall back defensively.
  const base = site ?? new URL('https://use-charter.dev');
  const sitemapURL = new URL('sitemap-index.xml', base);
  // Served by charter-router: Mintlify's sitemap with the host rewritten to
  // this domain (see infra/router/src/index.ts).
  const docsSitemapURL = new URL('docs/sitemap.xml', base);
  return new Response(getRobotsTxt(sitemapURL, docsSitemapURL), {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
};
