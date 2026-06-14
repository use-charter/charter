/**
 * Charter landing ck — client interactivity.
 *
 * The page server-renders its final ("scored") state, so everything works
 * with JavaScript disabled. This module layers motion on top:
 *   - boot hero: type the command, run a scan feed, count 0→82, sweep meters
 *   - re-run button replays the boot sequence
 *   - rule-catalog rail: click a category to swap the active panel
 *   - command tabs: click a command to swap the terminal panel (+ diff)
 *   - lifecycle stepper: light steps + animate fills on scroll-into-view
 *   - copy-install buttons, cursor-lit wordmark, theme switcher
 *
 * All motion respects prefers-reduced-motion.
 */
import { initWaitlistForm } from './WaitlistForm';
import { initThemeSwitch } from './theme';

const prefersReducedMotion = (): boolean =>
  typeof window.matchMedia === 'function' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const hasHover = (): boolean =>
  typeof window.matchMedia !== 'function' || !window.matchMedia('(hover: none)').matches;

const BOOT_CMD = 'charter doctor --strict';
const SCAN_FEED = ['AGENTS.md', 'charter.yaml', '.mcp.json', '.gitignore', 'package.json', '.github/workflows/', 'tests/', 'go.mod'];
const SCAN_TAGS = ['context', 'config', 'mcp', 'context', 'autonomy', 'ci', 'testing', 'env'];
const TARGET_SCORE = 82;

/** Cancellable timer registry so a re-run cleans up a half-finished sequence. */
class TimerBag {
  private timeouts: number[] = [];
  private intervals: number[] = [];
  private rafs: number[] = [];

  timeout(fn: () => void, ms: number): void {
    this.timeouts.push(window.setTimeout(fn, ms));
  }
  interval(fn: () => void, ms: number): number {
    const id = window.setInterval(fn, ms);
    this.intervals.push(id);
    return id;
  }
  raf(fn: FrameRequestCallback): number {
    const id = window.requestAnimationFrame(fn);
    this.rafs.push(id);
    return id;
  }
  clear(): void {
    this.timeouts.forEach((id) => window.clearTimeout(id));
    this.intervals.forEach((id) => window.clearInterval(id));
    this.rafs.forEach((id) => window.cancelAnimationFrame(id));
    this.timeouts = [];
    this.intervals = [];
    this.rafs = [];
  }
}

/** Copy-to-clipboard buttons: any element with [data-copy] holding a [data-copy-btn]. */
function initCopyButtons(): void {
  document.querySelectorAll<HTMLElement>('[data-copy]').forEach((host) => {
    const btn = host.querySelector<HTMLButtonElement>('[data-copy-btn]');
    const text = host.dataset.copy ?? '';
    if (!btn) return;
    btn.addEventListener('click', async () => {
      try {
        await navigator.clipboard.writeText(text);
        const prev = btn.textContent;
        btn.textContent = 'copied';
        window.setTimeout(() => { btn.textContent = prev; }, 1300);
      } catch {
        /* clipboard unavailable — leave the label unchanged */
      }
    });
  });
}

/** Scroll-active window highlighting in the tmux status bar. */
function initNavTracking(): void {
  const links = Array.from(document.querySelectorAll<HTMLAnchorElement>('[data-win]'));
  if (links.length === 0) return;
  let raf = 0;
  const onScroll = (): void => {
    if (raf) return;
    raf = window.requestAnimationFrame(() => {
      raf = 0;
      const vh = window.innerHeight;
      let current = '';
      links.forEach((a) => {
        const el = document.getElementById(a.dataset.win ?? '');
        if (el && el.getBoundingClientRect().top < vh * 0.45) current = a.dataset.win ?? '';
      });
      links.forEach((a) => a.classList.toggle('is-active', a.dataset.win === current && current !== ''));
    });
  };
  onScroll();
  window.addEventListener('scroll', onScroll, { passive: true });
}

/** Boot hero: type → scan feed → count score → sweep meter + category bars. */
function initBootHero(): void {
  const hero = document.querySelector<HTMLElement>('[data-boot]');
  if (!hero) return;

  const cmdEl = hero.querySelector<HTMLElement>('[data-boot-cmd]');
  const caret = hero.querySelector<HTMLElement>('[data-boot-caret]');
  const scan = hero.querySelector<HTMLElement>('[data-boot-scan]');
  const numEl = hero.querySelector<HTMLElement>('[data-boot-num]');
  const statusNum = hero.querySelector<HTMLElement>('[data-boot-status-num]');
  const meter = hero.querySelector<HTMLElement>('[data-boot-meter]');
  const rerun = hero.querySelector<HTMLButtonElement>('[data-boot-rerun]');
  const bars = Array.from(hero.querySelectorAll<HTMLElement>('[data-rdx-fill]'));
  if (!cmdEl || !caret || !scan || !numEl || !meter) return;

  const targets = bars.map((b) => Number(b.dataset.rdxFill ?? '0'));
  const bag = new TimerBag();

  const settleFinal = (): void => {
    cmdEl.textContent = BOOT_CMD;
    caret.hidden = false; // a blinking cursor rests at the end of the line (CSS-driven)
    scan.innerHTML = '<span class="tk-faint">✓ 8 paths · 18 rules · 9 categories · 1.84s · 3 findings</span>';
    numEl.textContent = String(TARGET_SCORE);
    if (statusNum) statusNum.textContent = String(TARGET_SCORE);
    meter.style.width = TARGET_SCORE + '%';
    bars.forEach((b, i) => { b.style.width = targets[i] + '%'; });
  };

  if (prefersReducedMotion()) {
    settleFinal();
    return;
  }

  const run = (): void => {
    bag.clear();
    cmdEl.textContent = '';
    caret.hidden = false;
    scan.innerHTML = '&nbsp;';
    numEl.textContent = '0';
    if (statusNum) statusNum.textContent = '0';
    meter.style.width = '0%';
    bars.forEach((b) => { b.style.width = '0%'; });

    const countScore = (): void => {
      scan.innerHTML = '<span class="tk-faint">✓ 8 paths · 18 rules · 9 categories · 1.84s · 3 findings</span>';
      let start: number | undefined;
      const dur = 900;
      const tick = (ts: number): void => {
        if (start === undefined) start = ts;
        const p = Math.min(1, (ts - start) / dur);
        const eased = 1 - Math.pow(1 - p, 3);
        const v = Math.round(TARGET_SCORE * eased);
        numEl.textContent = String(v);
        if (statusNum) statusNum.textContent = String(v);
        meter.style.width = TARGET_SCORE * eased + '%';
        if (p < 1) bag.raf(tick);
      };
      bag.raf(tick);
      bag.timeout(() => bars.forEach((b, i) => { b.style.width = targets[i] + '%'; }), 140);
    };

    const scanning = (): void => {
      let n = 0;
      const fd = bag.interval(() => {
        const idx = n % SCAN_FEED.length;
        scan.innerHTML = `scanning <span class="tk-blue">${SCAN_FEED[idx]}</span> <span class="tk-faint">· ${SCAN_TAGS[idx]}</span>`;
        n += 1;
        if (n >= SCAN_FEED.length + 1) {
          window.clearInterval(fd);
          bag.timeout(countScore, 320);
        }
      }, 130);
    };

    // type the command
    bag.timeout(() => {
      let i = 0;
      const ty = bag.interval(() => {
        i += 1;
        cmdEl.textContent = BOOT_CMD.slice(0, i);
        if (i >= BOOT_CMD.length) {
          window.clearInterval(ty);
          bag.timeout(scanning, 240);
        }
      }, 42);
    }, 360);

    // safety net — settle if anything stalls
    bag.timeout(settleFinal, 6000);
  };

  run();
  rerun?.addEventListener('click', run);
}

/** Generic tab/panel switcher used by the rule catalog and the command surface. */
function initTabGroup(tabAttr: string, panelAttr: string, onSelect?: (index: string) => void): void {
  const tabs = Array.from(document.querySelectorAll<HTMLButtonElement>(`[${tabAttr}]`));
  const panels = Array.from(document.querySelectorAll<HTMLElement>(`[${panelAttr}]`));
  if (tabs.length === 0) return;
  const toCamel = (attr: string): string => attr.replace('data-', '').replace(/-([a-z])/g, (_, c: string) => c.toUpperCase());
  const tabKey = toCamel(tabAttr);
  const panelKey = toCamel(panelAttr);

  const select = (index: string): void => {
    tabs.forEach((t) => {
      const active = t.dataset[tabKey] === index;
      t.classList.toggle('is-active', active && tabAttr === 'data-cmd-tab');
      t.classList.toggle('on', active && tabAttr === 'data-rule-tab');
      t.setAttribute('aria-selected', active ? 'true' : 'false');
    });
    panels.forEach((p) => { p.hidden = p.dataset[panelKey] !== index; });
    onSelect?.(index);
  };

  tabs.forEach((tab) => {
    tab.addEventListener('click', () => {
      const index = tab.dataset[tabKey];
      if (index !== undefined) select(index);
    });
  });
}

/** Re-trigger the line fade-in animation when a command panel becomes active. */
function replayCommandFade(index: string): void {
  if (prefersReducedMotion()) return;
  const panel = document.querySelector<HTMLElement>(`[data-cmd-panel="${index}"]`);
  if (!panel) return;
  panel.querySelectorAll<HTMLElement>('.ck-term__body .ln').forEach((ln, i) => {
    ln.classList.remove('fade');
    // force reflow so the animation restarts
    void ln.offsetWidth;
    ln.style.animationDelay = i * 70 + 'ms';
    ln.classList.add('fade');
  });
}

/** Lifecycle stepper: light steps sequentially + animate fills + count percentages. */
function initLifecycle(): void {
  const step = document.querySelector<HTMLElement>('[data-step]');
  if (!step) return;
  const items = Array.from(step.querySelectorAll<HTMLElement>('.ck-step__item'));
  const fills = items.map((it) => it.querySelector<HTMLElement>('[data-step-fill]'));
  const pcts = items.map((it) => it.querySelector<HTMLElement>('[data-step-pct]'));
  const targets = items.map((it) => Number(it.dataset.pct ?? '0'));

  const light = (count: number): void => {
    step.classList.add('on');
    items.forEach((it, i) => {
      const lit = i < count;
      it.classList.toggle('on', lit);
      const fill = fills[i];
      if (fill) fill.style.width = (lit ? targets[i] : 0) + '%';
      const pct = pcts[i];
      if (pct) pct.textContent = (lit ? targets[i] : 0) + '%';
    });
  };

  const countTo = (el: HTMLElement, to: number, delay: number): void => {
    let start: number | undefined;
    const dur = 800;
    const begin = (): void => {
      const tick = (ts: number): void => {
        if (start === undefined) start = ts;
        const p = Math.min(1, (ts - start) / dur);
        el.textContent = Math.round(to * (1 - Math.pow(1 - p, 3))) + '%';
        if (p < 1) window.requestAnimationFrame(tick);
      };
      window.requestAnimationFrame(tick);
    };
    window.setTimeout(begin, delay);
  };

  if (prefersReducedMotion()) {
    light(items.length);
    return;
  }

  // Reset to the pre-animation state so the full entrance (nodes light, rails
  // and bars fill, percentages count up) plays when the section scrolls in. The
  // page server-renders the finished state for no-JS / reduced-motion.
  step.classList.remove('on');
  items.forEach((it, i) => {
    it.classList.remove('on');
    fills[i]?.style.setProperty('width', '0%');
    if (pcts[i]) pcts[i]!.textContent = '0%';
  });

  let done = false;
  const reveal = (): void => {
    if (done) return;
    const r = step.getBoundingClientRect();
    if (r.top < window.innerHeight * 0.8 && r.bottom > 0) {
      done = true;
      window.removeEventListener('scroll', onScroll);
      step.classList.add('on');
      // Stagger every step — including the first — so each one visibly eases in
      // from the reset baseline (node lights, rail + bar fill, percentage counts).
      items.forEach((it, idx) => {
        window.setTimeout(() => {
          it.classList.add('on');
          const fill = fills[idx];
          if (fill) fill.style.width = targets[idx] + '%';
          const pct = pcts[idx];
          if (pct) countTo(pct, targets[idx], 120);
        }, 200 + idx * 360);
      });
    }
  };
  let raf = 0;
  const onScroll = (): void => {
    if (raf) return;
    raf = window.requestAnimationFrame(() => { raf = 0; reveal(); });
  };
  reveal();
  window.addEventListener('scroll', onScroll, { passive: true });
  // safety net
  window.setTimeout(() => { if (!done) { done = true; light(items.length); } }, 4500);
}

/**
 * Cursor-lit wordmark in the colophon — a glow layer masked to a soft circle.
 * The glow only lights when the pointer is actually over the wordmark's
 * visible glyph band AND not hovering any footer content (links, columns,
 * subscribe pane, social icons) or other interactive elements.
 */
function initWordmark(): void {
  const mark = document.querySelector<HTMLElement>('[data-wordmark]');
  if (!mark || !hasHover()) return;
  // The wordmark sits behind the footer content. Light it whenever the pointer
  // is within its visible glyph band — including over the content layer's empty
  // padding — but never while hovering actual footer text, links, or controls.
  const QUIET = 'a, button, input, label, [role="radiogroup"], h5, p, li';
  // The glyphs occupy a centered band of the giant wordmark, not its full box.
  const GLYPH_TOP = 0.12;
  const GLYPH_BOTTOM = 0.62;
  let queued = false;
  let lx = 0;
  let ly = 0;
  let target: EventTarget | null = null;
  document.addEventListener('pointermove', (e: PointerEvent) => {
    lx = e.clientX;
    ly = e.clientY;
    target = e.target;
    if (queued) return;
    queued = true;
    window.requestAnimationFrame(() => {
      queued = false;
      const r = mark.getBoundingClientRect();
      const inBand =
        lx >= r.left && lx <= r.right &&
        ly >= r.top + r.height * GLYPH_TOP && ly <= r.top + r.height * GLYPH_BOTTOM;
      const node = target instanceof Element ? target : null;
      const overQuiet = node != null && node.closest(QUIET) != null;
      const inside = inBand && !overQuiet;
      mark.classList.toggle('is-lit', inside);
      if (inside) {
        mark.style.setProperty('--mx', lx - r.left + 'px');
        mark.style.setProperty('--my', ly - r.top + 'px');
      }
    });
  });
}

function init(): void {
  initCopyButtons();
  initThemeSwitch();
  initNavTracking();
  initBootHero();
  initTabGroup('data-rule-tab', 'data-rule-panel');
  initTabGroup('data-cmd-tab', 'data-cmd-panel', replayCommandFade);
  initLifecycle();
  initWordmark();
  initWaitlistForm();
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}

export {};
