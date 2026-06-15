/* Legal pages: reading-progress bar, scroll-spy active section, smooth
   anchor jumps, and the cursor-lit colophon wordmark. Progressive
   enhancement — the page is fully readable and navigable without it. */

import { initThemeSwitch } from './theme';
import { initFooterGlow } from './footer';

const SPY_RATIO = 0.36;
const JUMP_OFFSET = 58;

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

function init(): void {
  initRail();
  initFooterGlow();
  initThemeSwitch();
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init, { once: true });
} else {
  init();
}
