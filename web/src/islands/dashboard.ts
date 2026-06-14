// Mission Control dashboard island. Fetches /dashboard/api/stats (served by the
// Access-gated charter-router worker) and renders the bento: count-up KPIs, a
// gradient traffic hero chart, a community donut, delta pills. Pure DOM +
// inline SVG — no chart library, no framework.
import { initThemeSwitch } from './theme';

interface Series {
  day: string;
  count: number;
}
interface Stats {
  generatedAt?: string;
  repo?: string;
  growth?: { stars: number; forks: number; watchers: number; openIssues: number } | null;
  traffic?: { views14d: number; uniqueViews14d: number; clones14d: number; uniqueClones14d: number; viewsSeries: Series[]; clonesSeries: Series[] } | null;
  releases?: { latestTag: string | null; publishedAt: string | null; url: string | null; totalDownloads: number; assets: { name: string; downloads: number }[] } | null;
  adoption?: { actionRepos: number; schemaRefs: number; sampleAdopters: string[] };
  community?: { openIssues: number; closedIssues: number; openPRs: number };
  errors?: Record<string, string>;
  error?: string;
}

const NS = 'http://www.w3.org/2000/svg';
const $ = <T extends Element = HTMLElement>(s: string, r: ParentNode = document): T | null => r.querySelector<T>(s);
const reduceMotion = matchMedia('(prefers-reduced-motion: reduce)').matches;
const nf = new Intl.NumberFormat('en', { notation: 'compact', maximumFractionDigits: 1 });
const fmt = (n: number | undefined | null): string => (typeof n === 'number' ? nf.format(n) : '—');

function banner(kind: 'error' | 'info', html: string): void {
  const b = $('[data-banner]');
  if (!b) return;
  b.hidden = false;
  b.className = `mc__banner mc__banner--${kind}`;
  b.innerHTML = html;
}
function clearBanner(): void {
  const b = $('[data-banner]');
  if (b) b.hidden = true;
}

function countUp(el: HTMLElement | null, target: number | undefined | null): void {
  if (!el) return;
  if (typeof target !== 'number') {
    el.textContent = '—';
    return;
  }
  if (reduceMotion) {
    el.textContent = fmt(target);
    return;
  }
  const dur = 800;
  const start = performance.now();
  const step = (now: number): void => {
    const p = Math.min(1, (now - start) / dur);
    const e = 1 - Math.pow(1 - p, 3);
    el.textContent = fmt(Math.round(target * e));
    if (p < 1) requestAnimationFrame(step);
    else el.textContent = fmt(target);
  };
  requestAnimationFrame(step);
}

function setText(sel: string, text: string): void {
  const el = $(sel);
  if (el) el.textContent = text;
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.round(diff / 60000);
  if (m < 1) return 'updated just now';
  if (m < 60) return `updated ${m}m ago`;
  const h = Math.round(m / 60);
  if (h < 24) return `updated ${h}h ago`;
  return `updated ${Math.round(h / 24)}d ago`;
}

function deltaPill(sel: string, series: Series[]): void {
  const el = $(sel);
  if (!el) return;
  if (series.length < 4) {
    el.textContent = '';
    return;
  }
  const half = Math.floor(series.length / 2);
  const a = series.slice(0, half).reduce((s, p) => s + p.count, 0);
  const b = series.slice(half).reduce((s, p) => s + p.count, 0);
  const pct = a === 0 ? (b > 0 ? 100 : 0) : Math.round(((b - a) / a) * 100);
  const dir = pct > 0 ? 'up' : pct < 0 ? 'down' : 'flat';
  el.className = `delta delta--${dir}`;
  el.textContent = `${pct > 0 ? '▲' : pct < 0 ? '▼' : '→'} ${Math.abs(pct)}% vs prior 7d`;
}

function path(d: string, cls: string): SVGPathElement {
  const p = document.createElementNS(NS, 'path');
  p.setAttribute('d', d);
  p.setAttribute('class', cls);
  return p;
}

// Hero chart: views area (gradient) + clones dashed line + horizontal gridlines.
function drawTraffic(views: Series[], clones: Series[]): void {
  const svg = $<SVGSVGElement>('[data-chart="traffic"]');
  if (!svg) return;
  svg.innerHTML = '';
  const W = 600;
  const H = 200;
  const padX = 2;
  const padY = 14;
  if (views.length < 2) return;
  const max = Math.max(...views.map((p) => p.count), ...clones.map((p) => p.count), 1);
  const xy = (arr: Series[]): readonly [number, number][] =>
    arr.map((p, i) => [padX + (i * (W - padX * 2)) / (arr.length - 1), H - padY - (p.count / max) * (H - padY * 2)] as const);

  // gradient def
  const defs = document.createElementNS(NS, 'defs');
  defs.innerHTML = `<linearGradient id="mcViewGrad" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stop-color="var(--charter-blue)" stop-opacity="0.22"/><stop offset="100%" stop-color="var(--charter-blue)" stop-opacity="0"/></linearGradient>`;
  svg.append(defs);

  // gridlines (4)
  for (let g = 0; g <= 3; g++) {
    const y = padY + (g * (H - padY * 2)) / 3;
    const ln = document.createElementNS(NS, 'line');
    ln.setAttribute('x1', String(padX));
    ln.setAttribute('x2', String(W - padX));
    ln.setAttribute('y1', y.toFixed(1));
    ln.setAttribute('y2', y.toFixed(1));
    ln.setAttribute('class', 'chart__grid');
    svg.append(ln);
  }

  const v = xy(views);
  const c = xy(clones);
  const line = (pts: readonly [number, number][]): string => pts.map(([x, y], i) => `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`).join(' ');
  const vLine = line(v);
  const area = `${vLine} L${v[v.length - 1]![0].toFixed(1)},${H} L${v[0]![0].toFixed(1)},${H} Z`;
  svg.append(path(area, 'chart__area'));
  if (c.length >= 2) svg.append(path(line(c), 'chart__clones'));
  const vp = path(vLine, 'chart__views');
  svg.append(vp);

  if (!reduceMotion) {
    const len = vp.getTotalLength();
    vp.style.strokeDasharray = String(len);
    vp.style.strokeDashoffset = String(len);
    vp.getBoundingClientRect();
    vp.style.transition = 'stroke-dashoffset 900ms cubic-bezier(0.16,1,0.3,1)';
    vp.style.strokeDashoffset = '0';
  }

  // x-axis: 4 evenly-spaced day labels (M/D)
  const ax = $('[data-xaxis]');
  if (ax) {
    const idxs = [0, Math.floor(views.length / 3), Math.floor((2 * views.length) / 3), views.length - 1];
    ax.innerHTML = [...new Set(idxs)]
      .map((i) => {
        const d = views[i]?.day ?? '';
        const md = d ? `${Number(d.slice(5, 7))}/${Number(d.slice(8, 10))}` : '';
        return `<span>${md}</span>`;
      })
      .join('');
  }
}

// Community donut: open vs closed issues, two arcs on a 120-box ring.
function drawDonut(open: number, closed: number): void {
  const svg = $<SVGSVGElement>('[data-donut]');
  if (!svg) return;
  svg.innerHTML = '';
  const total = open + closed;
  const cx = 60;
  const cy = 60;
  const rad = 48;
  const circ = 2 * Math.PI * rad;
  const track = document.createElementNS(NS, 'circle');
  track.setAttribute('cx', String(cx));
  track.setAttribute('cy', String(cy));
  track.setAttribute('r', String(rad));
  track.setAttribute('fill', 'none');
  track.setAttribute('stroke', 'var(--color-border-tertiary)');
  track.setAttribute('stroke-width', '12');
  svg.append(track);
  if (total === 0) return;
  const seg = (frac: number, offset: number, cls: string): void => {
    const a = document.createElementNS(NS, 'circle');
    a.setAttribute('cx', String(cx));
    a.setAttribute('cy', String(cy));
    a.setAttribute('r', String(rad));
    a.setAttribute('fill', 'none');
    a.setAttribute('class', cls);
    a.setAttribute('stroke-width', '12');
    a.setAttribute('stroke-linecap', 'round');
    a.setAttribute('stroke-dasharray', `${(frac * circ).toFixed(1)} ${circ.toFixed(1)}`);
    a.setAttribute('stroke-dashoffset', String(-offset * circ));
    svg.append(a);
  };
  seg(closed / total, 0, 'donut__closed');
  seg(open / total, closed / total, 'donut__open');
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
  setText('[data-stamp]', s.generatedAt ? relativeTime(s.generatedAt) : 'live');

  // Tier 1 KPIs
  countUp($('[data-kpi="stars"] [data-v]'), s.growth?.stars);
  countUp($('[data-kpi="forks"] [data-v]'), s.growth?.forks);
  countUp($('[data-kpi="watchers"] [data-v]'), s.growth?.watchers);
  countUp($('[data-kpi="openIssues"] [data-v]'), s.growth?.openIssues);

  // Hero traffic
  countUp($('[data-hero="views"]'), s.traffic?.views14d);
  setText('[data-hero="uniqueViews"]', s.traffic ? `${fmt(s.traffic.uniqueViews14d)} unique` : 'no traffic access');
  deltaPill('[data-delta="views"]', s.traffic?.viewsSeries ?? []);
  drawTraffic(s.traffic?.viewsSeries ?? [], s.traffic?.clonesSeries ?? []);

  // Releases
  const r = s.releases;
  const tag = $<HTMLAnchorElement>('[data-rel-tag]');
  if (tag) {
    tag.textContent = r?.latestTag ?? 'no releases yet';
    if (r?.url) tag.href = r.url;
  }
  setText('[data-rel-date]', r?.publishedAt ? new Date(r.publishedAt).toLocaleDateString() : '');
  countUp($('[data-hero="downloads"]'), r?.totalDownloads ?? 0);
  const assets = $('[data-rel-assets]');
  if (assets) {
    assets.innerHTML = (r?.assets ?? []).map((a) => `<li><span>${a.name}</span><b>${fmt(a.downloads)}</b></li>`).join('') || '<li class="muted">no assets</li>';
  }

  // Adoption
  countUp($('[data-hero="action"]'), s.adoption?.actionRepos);
  countUp($('[data-hero="schema"]'), s.adoption?.schemaRefs);
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
  const open = s.community?.openIssues ?? 0;
  const closed = s.community?.closedIssues ?? 0;
  setText('[data-c="openIssues"]', fmt(open));
  setText('[data-c="closedIssues"]', fmt(closed));
  setText('[data-c="openPRs"]', fmt(s.community?.openPRs));
  countUp($('[data-donut-total]'), open + closed);
  drawDonut(open, closed);
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
      banner('info', `Stats API returned <code>${res.status}</code> (not JSON). It's served by the <code>charter-router</code> worker — not by local <code>astro preview</code>. Deploy the router and set <code>GITHUB_STATS_TOKEN</code> to see live metrics.`);
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
