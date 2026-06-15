/**
 * Footer wordmark glow — shared by every page that renders <SiteFooter />.
 *
 * Lights the giant `charter` wordmark behind the colophon when the pointer
 * crosses its visible glyph band, but never while hovering footer text, links,
 * or controls. Pointer-driven, rAF-throttled, hover-capable devices only.
 */

// The glyphs occupy a centered band of the giant wordmark, not its full box.
const GLYPH_TOP = 0.12;
const GLYPH_BOTTOM = 0.62;
// Elements that should keep the wordmark dark while hovered.
const QUIET = 'a, button, input, label, [role="radiogroup"], h4, h5, p, li';

export function initFooterGlow(): void {
  const mark = document.querySelector<HTMLElement>('[data-wordmark]');
  if (!mark || window.matchMedia('(hover: none)').matches) return;

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
  });
}
