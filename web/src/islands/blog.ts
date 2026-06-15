// Blog post island: a top reading-progress bar and table-of-contents scroll-spy.
// Progress reflects how far through the article body the reader is; the TOC
// highlights the section currently in view.
import { initThemeSwitch } from './theme';

// Wire the shared three-state theme switcher in the nav bar.
initThemeSwitch();

const progress = document.querySelector<HTMLElement>('[data-progress]');
const post = document.querySelector<HTMLElement>('[data-post]');

let raf = 0;
function updateProgress(): void {
  raf = 0;
  if (!progress || !post) return;
  const total = post.offsetHeight - window.innerHeight;
  const scrolled = total <= 0 ? 1 : Math.min(1, Math.max(0, -post.getBoundingClientRect().top / total));
  progress.style.transform = `scaleX(${scrolled})`;
}
function onScroll(): void {
  if (!raf) raf = requestAnimationFrame(updateProgress);
}
window.addEventListener('scroll', onScroll, { passive: true });
window.addEventListener('resize', onScroll);
updateProgress();

// TOC scroll-spy via IntersectionObserver on the rendered headings.
const links = new Map<string, HTMLAnchorElement>();
document.querySelectorAll<HTMLAnchorElement>('[data-toc]').forEach((a) => {
  const id = a.getAttribute('data-toc');
  if (id) links.set(id, a);
});

if (links.size > 0) {
  const headings = Array.from(document.querySelectorAll<HTMLElement>('.bl-prose :is(h2, h3)')).filter((h) => h.id && links.has(h.id));
  let active = '';
  const setActive = (id: string): void => {
    if (id === active) return;
    if (active) links.get(active)?.classList.remove('is-active');
    links.get(id)?.classList.add('is-active');
    active = id;
  };
  const io = new IntersectionObserver(
    (entries) => {
      // Pick the topmost heading currently intersecting the upper band.
      const visible = entries.filter((e) => e.isIntersecting).map((e) => e.target as HTMLElement);
      if (visible.length > 0) {
        visible.sort((a, b) => a.getBoundingClientRect().top - b.getBoundingClientRect().top);
        if (visible[0]?.id) setActive(visible[0].id);
      }
    },
    { rootMargin: '-72px 0px -70% 0px', threshold: 0 },
  );
  headings.forEach((h) => io.observe(h));
}
