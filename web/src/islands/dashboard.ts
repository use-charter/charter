// Mission Control dashboard island. Fetches /dashboard/api/stats (served by the
// charter-router worker, gated by Cloudflare Access) and renders the four metric
// groups. Pure DOM + inline-SVG sparklines — no chart library, no framework.
import { initThemeSwitch } from './theme';

interface Series {
  day: string;
  count: number;
}
interface Stats {
  generatedAt?: string;
  repo?: string;
  growth?: { stars: number; forks: number; watchers: number; openIssues: number } | null;
  traffic?: {
    views14d: number;
    uniqueViews14d: number;
    clones14d: number;
    uniqueClones14d: number;
    viewsSeries: Series[];
    clonesSeries: Series[];
  } | null;
  releases?: { latestTag: string | null; publishedAt: string | null; url: string | null; totalDownloads: number; assets: { name: string; downloads: number }[] } | null;
  adoption?: { actionRepos: number; schemaRefs: number; sampleAdopters: string[] };
  community?: { openIssues: number; closedIssues: number; openPRs: number };
  errors?: Record<string, string>;
  error?: string;
  message?: string;
}

const $ = <T extends Element = HTMLElement>(sel: string, root: ParentNode = document): T | null => root.querySelector<T>(sel);

const nf = new Intl.NumberFormat('en', { notation: 'compact', maximumFractionDigits: 1 });
const fmt = (n: number | undefined | null): string => (typeof n === 'number' ? nf.format(n) : '—');

function setText(sel: string, text: string): void {
  const el = $(sel);
  if (el) el.textContent = text;
}

function banner(kind: 'error' | 'info', html: string): void {
  const b = $('[data-banner]');
  if (!b) return;
  b.hidden = false;
  b.className = `dash__banner dash__banner--${kind}`;
  b.innerHTML = html;
}

function clearBanner(): void {
  const b = $('[data-banner]');
  if (b) b.hidden = true;
}

// Build a sparkline (line + soft area) from a series into the given <svg>.
function sparkline(svg: SVGSVGElement | null, series: Series[]): void {
  if (!svg) return;
  svg.innerHTML = '';
  if (series.length < 2) return;
  const W = 300;
  const H = 64;
  const pad = 4;
  const max = Math.max(...series.map((p) => p.count), 1);
  const step = (W - pad * 2) / (series.length - 1);
  const pts = series.map((p, i) => {
    const x = pad + i * step;
    const y = H - pad - (p.count / max) * (H - pad * 2);
    return [x, y] as const;
  });
  const line = pts.map(([x, y], i) => `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`).join(' ');
  const area = `${line} L${pts[pts.length - 1]![0].toFixed(1)},${H} L${pts[0]![0].toFixed(1)},${H} Z`;
  const ns = 'http://www.w3.org/2000/svg';
  const a = document.createElementNS(ns, 'path');
  a.setAttribute('d', area);
  a.setAttribute('class', 'spark__area');
  const l = document.createElementNS(ns, 'path');
  l.setAttribute('d', line);
  l.setAttribute('class', 'spark__line');
  svg.append(a, l);
}

function trend(series: Series[]): string {
  if (series.length < 2) return '';
  const half = Math.floor(series.length / 2);
  const a = series.slice(0, half).reduce((s, p) => s + p.count, 0);
  const b = series.slice(half).reduce((s, p) => s + p.count, 0);
  if (a === 0) return b > 0 ? '▲' : '';
  const pct = Math.round(((b - a) / a) * 100);
  return pct === 0 ? '→ 0%' : `${pct > 0 ? '▲' : '▼'} ${Math.abs(pct)}%`;
}

function render(s: Stats): void {
  if (s.error === 'not_configured') {
    banner('info', `<b>Dashboard not configured.</b> Set the <code>GITHUB_STATS_TOKEN</code> worker secret on <code>charter-router</code> to enable live metrics.`);
    return;
  }
  if (s.error === 'forbidden') {
    banner('error', 'Access denied.');
    return;
  }
  clearBanner();

  if (s.repo) setText('[data-repo]', s.repo);
  if (s.generatedAt) setText('[data-stamp]', `updated ${new Date(s.generatedAt).toLocaleString()}`);

  // KPIs
  const kv = (key: string, v: string, sub = ''): void => {
    const card = $(`[data-kpi="${key}"]`);
    if (!card) return;
    const vel = $('[data-v]', card);
    const sel = $('[data-sub]', card);
    if (vel) vel.textContent = v;
    if (sel) sel.textContent = sub;
  };
  kv('stars', fmt(s.growth?.stars));
  kv('forks', fmt(s.growth?.forks));
  kv('views', fmt(s.traffic?.views14d), s.traffic ? `${fmt(s.traffic.uniqueViews14d)} unique` : 'no traffic access');
  kv('clones', fmt(s.traffic?.clones14d), s.traffic ? `${fmt(s.traffic.uniqueClones14d)} unique` : '');

  // Growth & traffic
  sparkline($<SVGSVGElement>('[data-spark="views"]'), s.traffic?.viewsSeries ?? []);
  sparkline($<SVGSVGElement>('[data-spark="clones"]'), s.traffic?.clonesSeries ?? []);
  setText('[data-trend="views"]', s.traffic ? trend(s.traffic.viewsSeries) : '');
  setText('[data-trend="clones"]', s.traffic ? trend(s.traffic.clonesSeries) : '');
  setText('[data-g="watchers"]', fmt(s.growth?.watchers));
  setText('[data-g="uniqueViews"]', fmt(s.traffic?.uniqueViews14d));
  setText('[data-g="uniqueClones"]', fmt(s.traffic?.uniqueClones14d));
  setText('[data-g="openIssues"]', fmt(s.growth?.openIssues));

  // Releases
  const r = s.releases;
  const tag = $<HTMLAnchorElement>('[data-rel-tag]');
  if (tag) {
    tag.textContent = r?.latestTag ?? 'no releases yet';
    if (r?.url) tag.href = r.url;
  }
  setText('[data-rel-date]', r?.publishedAt ? new Date(r.publishedAt).toLocaleDateString() : '');
  setText('[data-rel-downloads]', fmt(r?.totalDownloads));
  const assets = $('[data-rel-assets]');
  if (assets) {
    assets.innerHTML = '';
    for (const a of r?.assets ?? []) {
      const li = document.createElement('li');
      li.innerHTML = `<span>${a.name}</span><b>${fmt(a.downloads)}</b>`;
      assets.append(li);
    }
  }

  // Adoption
  setText('[data-a="action"]', fmt(s.adoption?.actionRepos));
  setText('[data-a="schema"]', fmt(s.adoption?.schemaRefs));
  const list = $('[data-adopters]');
  if (list) {
    const repos = s.adoption?.sampleAdopters ?? [];
    list.innerHTML = repos.length
      ? repos.map((n) => `<li><a href="https://github.com/${n}" target="_blank" rel="noopener">${n}</a></li>`).join('')
      : '<li class="muted">none yet</li>';
  }
  const sig = $<HTMLAnchorElement>('[data-signals]');
  if (sig && s.repo) sig.href = `https://github.com/${s.repo}/issues?q=is%3Aissue+label%3Alaunch-signals`;

  // Community
  setText('[data-c="openIssues"]', fmt(s.community?.openIssues));
  setText('[data-c="closedIssues"]', fmt(s.community?.closedIssues));
  setText('[data-c="openPRs"]', fmt(s.community?.openPRs));
}

async function load(): Promise<void> {
  const root = $('[data-dash]');
  root?.classList.add('is-loading');
  try {
    const res = await fetch('/dashboard/api/stats', { headers: { Accept: 'application/json' } });
    if (res.status === 403) {
      render({ error: 'forbidden' });
      return;
    }
    const ct = res.headers.get('content-type') ?? '';
    if (!res.ok || !ct.includes('json')) {
      banner(
        'info',
        `Stats API returned <code>${res.status}</code> (not JSON). It's served by the <code>charter-router</code> worker — not by local <code>astro preview</code>. Deploy the router and set <code>GITHUB_STATS_TOKEN</code> to see live metrics.`,
      );
      return;
    }
    const data = (await res.json()) as Stats;
    render(data);
    if (data.errors && Object.keys(data.errors).length > 0) {
      banner('info', `Some metrics are unavailable: <code>${Object.keys(data.errors).join(', ')}</code>. Check the token scopes.`);
    }
  } catch (err) {
    banner('error', `Failed to load metrics: ${err instanceof Error ? err.message : String(err)}`);
  } finally {
    root?.classList.remove('is-loading');
  }
}

function init(): void {
  initThemeSwitch();
  const btn = $<HTMLButtonElement>('[data-refresh]');
  btn?.addEventListener('click', () => {
    btn.classList.add('is-spin');
    void load().finally(() => setTimeout(() => btn.classList.remove('is-spin'), 600));
  });
  void load();
}

if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
else init();
