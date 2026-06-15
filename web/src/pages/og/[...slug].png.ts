import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { Resvg } from '@resvg/resvg-js';
import type { APIRoute } from 'astro';
import { getCollection } from 'astro:content';
import satori from 'satori';

// Per-post Open Graph images, rendered at build to /og/<slug>.png. satori lays
// the card out (HTML-ish tree) into SVG; resvg rasterizes to PNG. Branded to
// match the site: brand-dark backdrop, blue accent, the post title in Ruda.
// Static Ruda instances (satori can't parse the variable font). Read from
// source at build time — cwd is the web/ project dir under both the local
// `bun run build` and the CI `moon run web:build`. (import.meta.url is
// unreliable here: the endpoint is relocated into a build chunk.)
const fontDir = join(process.cwd(), 'src/assets/fonts');
const fontRegular = readFileSync(join(fontDir, 'Ruda-Regular.ttf'));
const fontBold = readFileSync(join(fontDir, 'Ruda-ExtraBold.ttf'));

export async function getStaticPaths() {
  const posts = await getCollection('blog', ({ data }) => !data.draft);
  return posts.map((post) => ({ params: { slug: post.id }, props: { post } }));
}

const BG = '#06080c';
const INK = '#eaedf3';
const BLUE = '#5aa2f7';
const MUTE = '#8b94a3';

interface Node {
  type: string;
  props: { style?: Record<string, unknown>; children?: Node | Node[] | string };
}
const el = (type: string, style: Record<string, unknown>, children?: Node | Node[] | string): Node => ({ type, props: { style, children } });

export const GET: APIRoute = async ({ props }) => {
  const { post } = props as { post: { data: { title: string; tags: string[] } } };
  const tagline = post.data.tags.length ? post.data.tags.join('  ·  ') : 'use-charter.dev';

  const tree = el(
    'div',
    {
      width: '100%',
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'space-between',
      padding: '72px 80px',
      backgroundColor: BG,
      backgroundImage: `radial-gradient(900px 520px at 82% -12%, rgba(90,162,247,0.20), transparent 62%)`,
      color: INK,
      fontFamily: 'Ruda',
    },
    [
      // brand row
      el('div', { display: 'flex', alignItems: 'center', gap: '16px' }, [
        el('div', { display: 'flex', fontSize: '30px', fontWeight: 800, color: BLUE, letterSpacing: '-0.02em' }, '[C]'),
        el('div', { display: 'flex', fontSize: '30px', fontWeight: 800, color: INK, letterSpacing: '-0.02em' }, 'charter'),
      ]),
      // title
      el('div', { display: 'flex', flexDirection: 'column', gap: '20px' }, [
        el('div', { display: 'flex', fontSize: '22px', fontWeight: 600, color: BLUE, letterSpacing: '0.14em', textTransform: 'uppercase' }, 'The Charter blog'),
        el(
          'div',
          { display: 'flex', fontSize: post.data.title.length > 52 ? '58px' : '72px', fontWeight: 800, lineHeight: 1.04, letterSpacing: '-0.035em', color: INK, maxWidth: '1040px' },
          post.data.title,
        ),
      ]),
      // footer row
      el('div', { display: 'flex', alignItems: 'center', justifyContent: 'space-between' }, [
        el('div', { display: 'flex', fontSize: '24px', color: MUTE, letterSpacing: '0.01em' }, tagline),
        el('div', { display: 'flex', fontSize: '24px', fontWeight: 600, color: INK }, 'use-charter.dev'),
      ]),
    ],
  );

  const svg = await satori(tree as unknown as Parameters<typeof satori>[0], {
    width: 1200,
    height: 630,
    fonts: [
      { name: 'Ruda', data: fontRegular, weight: 400, style: 'normal' },
      { name: 'Ruda', data: fontBold, weight: 800, style: 'normal' },
    ],
  });

  const png = new Resvg(svg, { fitTo: { mode: 'width', value: 1200 } }).render().asPng();
  return new Response(new Uint8Array(png), {
    headers: { 'Content-Type': 'image/png', 'Cache-Control': 'public, max-age=31536000, immutable' },
  });
};
