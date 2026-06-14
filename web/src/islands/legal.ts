/* Legal pages: reading-progress bar, scroll-spy active section, smooth
   anchor jumps, and the cursor-lit colophon wordmark. Progressive
   enhancement — the page is fully readable and navigable without it. */

import { initThemeSwitch } from './theme';

const SPY_RATIO = 0.36;
const JUMP_OFFSET = 58;
const GLYPH_TOP = 0.12;
const GLYPH_BOTTOM = 0.62;

function initRail(): void {
  const bar = document.querySelector<HTMLElement>('.lg-rail__bar i');
  const links = Array.from(
    document.querySelectorAll<HTMLAnchorElement>('.lg-rail__nav a'),
  );
  const ids = links
    .map((a) => a.getAttribute('href')?.slice(1) ?? '')
    .filter(Boolean);
  if (ids.length === 0) return;

  const setActive = (id: string): void => {
    for (const a of links) {
      a.classList.toggle('is-active', a.getAttribute('href') === `#${id}`);
    }
  };

  let raf = 0;
  const onScroll = (): void => {
    if (raf) return;
    raf = requestAnimationFrame(() => {
      raf = 0;
      const doc = document.documentElement;
      const max = doc.scrollHeight - window.innerHeight;
      const pct = max > 0 ? Math.min(100, Math.max(0, (window.scrollY / max) * 100)) : 0;
      if (bar) bar.style.width = `${pct}%`;

      let current = ids[0];
      const threshold = window.innerHeight * SPY_RATIO;
      for (const id of ids) {
        const el = document.getElementById(id);
        if (el && el.getBoundingClientRect().top < threshold) current = id;
      }
      setActive(current);
    });
  };

  for (const a of links) {
    a.addEventListener('click', (e) => {
      const id = a.getAttribute('href')?.slice(1);
      const el = id ? document.getElementById(id) : null;
      if (!el) return;
      e.preventDefault();
      const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
      const y = el.getBoundingClientRect().top + window.scrollY - JUMP_OFFSET;
      window.scrollTo({ top: y, behavior: reduce ? 'auto' : 'smooth' });
    });
  }

  onScroll();
  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('resize', onScroll);
}

function initColophonGlow(): void {
  const mark = document.querySelector<HTMLElement>('.ck-man__mark');
  if (!mark || window.matchMedia('(hover: none)').matches) return;

  // Only light the wordmark when the pointer is over its visible glyph band
  // and NOT over the footer content (links, subscribe pane, social icons) or
  // any interactive element.
  const QUIET = 'a, button, input, label, [role="radiogroup"], h5, p, li';

  let queued = false;
  let lx = 0;
  let ly = 0;
  let target: EventTarget | null = null;
  const onMove = (e: PointerEvent): void => {
    lx = e.clientX;
    ly = e.clientY;
    target = e.target;
    if (queued) return;
    queued = true;
    requestAnimationFrame(() => {
      queued = false;
      const r = mark.getBoundingClientRect();
      const inBand =
        lx >= r.left &&
        lx <= r.right &&
        ly >= r.top + r.height * GLYPH_TOP &&
        ly <= r.top + r.height * GLYPH_BOTTOM;
      const node = target instanceof Element ? target : null;
      const overQuiet = node != null && node.closest(QUIET) != null;
      const inside = inBand && !overQuiet;
      mark.classList.toggle('is-lit', inside);
      if (inside) {
        mark.style.setProperty('--mx', `${lx - r.left}px`);
        mark.style.setProperty('--my', `${ly - r.top}px`);
      }
    });
  };
  document.addEventListener('pointermove', onMove);
}

function initSubscribe(): void {
  const form = document.querySelector<HTMLFormElement>('.ck-sub form');
  if (!form) return;
  const input = form.querySelector<HTMLInputElement>('.ck-sub__input');
  const ok = form.querySelector<HTMLElement>('.ck-sub__ok');
  form.addEventListener('submit', (e) => {
    e.preventDefault();
    const value = input?.value.trim() ?? '';
    const valid = /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(value);
    if (ok) {
      ok.textContent = valid
        ? '✓ queued — we’ll ping you when new rules and releases land'
        : '✗ that doesn’t look like an email';
    }
    if (valid && input) input.value = '';
  });
}

function init(): void {
  initRail();
  initColophonGlow();
  initSubscribe();
  initThemeSwitch();
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init, { once: true });
} else {
  init();
}
