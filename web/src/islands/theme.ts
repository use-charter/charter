/**
 * Three-state theme switcher — shared by the landing and legal navs.
 *
 * A compact segmented pill with three segments: system · light · dark.
 *   - `system` follows the OS preference and live-updates while selected.
 *   - `light` / `dark` force the resolved theme.
 *
 * The chosen MODE ('system' | 'light' | 'dark') persists to localStorage.
 * The EFFECTIVE theme ('light' | 'dark') is applied to
 * <html data-theme>; the chosen mode is mirrored on <html data-theme-mode>
 * so the pill can restore its active segment on load. A single sliding
 * highlight indicator animates behind the active segment.
 */

const STORAGE_KEY = 'charter-theme';

type ThemeMode = 'system' | 'light' | 'dark';
type ResolvedTheme = 'light' | 'dark';

const MODES: readonly ThemeMode[] = ['system', 'light', 'dark'];

const prefersDark = (): boolean =>
  typeof window.matchMedia === 'function' &&
  window.matchMedia('(prefers-color-scheme: dark)').matches;

const readMode = (): ThemeMode => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === 'system' || stored === 'light' || stored === 'dark') return stored;
  } catch {
    /* storage unavailable — fall through to default */
  }
  return 'system';
};

const resolve = (mode: ThemeMode): ResolvedTheme =>
  mode === 'system' ? (prefersDark() ? 'dark' : 'light') : mode;

const apply = (mode: ThemeMode): void => {
  const root = document.documentElement;
  root.dataset.theme = resolve(mode);
  root.dataset.themeMode = mode;
};

const persist = (mode: ThemeMode): void => {
  try {
    localStorage.setItem(STORAGE_KEY, mode);
  } catch {
    /* storage unavailable — selection still holds for this session */
  }
};

/**
 * Wire every [data-theme-switch] pill on the page to the shared theme state.
 * Idempotent and safe to call from multiple island entry points.
 */
export function initThemeSwitch(): void {
  const pills = Array.from(
    document.querySelectorAll<HTMLElement>('[data-theme-switch]'),
  );
  if (pills.length === 0) return;

  let mode = readMode();

  const syncPill = (pill: HTMLElement): void => {
    const segs = Array.from(
      pill.querySelectorAll<HTMLButtonElement>('[data-theme-seg]'),
    );
    const activeIndex = MODES.indexOf(mode);
    segs.forEach((seg, i) => {
      const on = seg.dataset.themeSeg === mode;
      seg.setAttribute('aria-checked', on ? 'true' : 'false');
      seg.tabIndex = on ? 0 : -1;
      seg.classList.toggle('is-active', on);
    });
    // Slide the highlight to the active segment via a 0-based column index.
    pill.style.setProperty('--seg-index', String(Math.max(0, activeIndex)));
  };

  const syncAll = (): void => pills.forEach(syncPill);

  const setMode = (next: ThemeMode): void => {
    mode = next;
    apply(mode);
    persist(mode);
    syncAll();
  };

  for (const pill of pills) {
    const segs = Array.from(
      pill.querySelectorAll<HTMLButtonElement>('[data-theme-seg]'),
    );
    segs.forEach((seg) => {
      seg.addEventListener('click', () => {
        const next = seg.dataset.themeSeg as ThemeMode | undefined;
        if (next && MODES.includes(next)) setMode(next);
      });
    });
    // Roving arrow-key navigation across the radiogroup.
    pill.addEventListener('keydown', (e: KeyboardEvent) => {
      if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return;
      e.preventDefault();
      const dir = e.key === 'ArrowRight' ? 1 : -1;
      const at = MODES.indexOf(mode);
      const nextIndex = (at + dir + MODES.length) % MODES.length;
      setMode(MODES[nextIndex]);
      segs[nextIndex]?.focus();
    });
  }

  // Re-resolve on OS change while in system mode (live update, no reload).
  if (typeof window.matchMedia === 'function') {
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const onChange = (): void => {
      if (mode === 'system') apply(mode);
    };
    if (typeof mq.addEventListener === 'function') {
      mq.addEventListener('change', onChange);
    }
  }

  // Reconcile pre-paint state set by Base.astro, then reflect on the pills.
  apply(mode);
  syncAll();
}
