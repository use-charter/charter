/* Charter HTML report — client interactivity.
   Vanilla, zero network. Filter by severity, full-text search, expand/collapse,
   copy path:line, light/dark toggle. Keyboard-operable; aria-* kept in sync. */
(function () {
  "use strict";

  var root = document.documentElement;
  var activeSeverity = "all";

  function findingCards() {
    return Array.prototype.slice.call(document.querySelectorAll(".fc"));
  }
  function searchInput() {
    return document.getElementById("charter-search");
  }

  /* ── Expand / collapse a single card ── */
  function setOpen(card, open) {
    var hdr = card.querySelector(".fc-hdr");
    var body = card.querySelector(".fc-body");
    if (!hdr || !body) return;
    body.classList.toggle("open", open);
    hdr.setAttribute("aria-expanded", open ? "true" : "false");
  }
  function toggleCard(card) {
    var body = card.querySelector(".fc-body");
    setOpen(card, !body.classList.contains("open"));
    syncExpandAll();
  }

  /* ── Expand / collapse all currently visible cards ── */
  function visibleCards() {
    return findingCards().filter(function (c) {
      return c.style.display !== "none";
    });
  }
  function allVisibleOpen() {
    var vis = visibleCards();
    return vis.length > 0 && vis.every(function (c) {
      return c.querySelector(".fc-body").classList.contains("open");
    });
  }
  function syncExpandAll() {
    var btn = document.getElementById("expand-all");
    if (!btn) return;
    var open = allVisibleOpen();
    btn.setAttribute("aria-pressed", open ? "true" : "false");
    var label = btn.querySelector(".label");
    if (label) label.textContent = open ? "Collapse all" : "Expand all";
  }
  function toggleAll() {
    var open = !allVisibleOpen();
    visibleCards().forEach(function (c) {
      setOpen(c, open);
    });
    syncExpandAll();
  }

  /* ── Filter + search ── */
  function applyFilter() {
    var input = searchInput();
    var query = (input ? input.value : "").trim().toLowerCase();
    var anyVisible = false;
    findingCards().forEach(function (c) {
      var sevMatch = activeSeverity === "all" || c.getAttribute("data-sev") === activeSeverity;
      var haystack = (c.getAttribute("data-search") || c.textContent).toLowerCase();
      var queryMatch = query === "" || haystack.indexOf(query) !== -1;
      var show = sevMatch && queryMatch;
      c.style.display = show ? "" : "none";
      if (show) anyVisible = true;
    });
    var empty = document.getElementById("findings-empty");
    if (empty) empty.classList.toggle("show", !anyVisible && findingCards().length > 0);
    syncExpandAll();
  }
  function setSeverity(sev) {
    activeSeverity = sev;
    document.querySelectorAll(".fp").forEach(function (b) {
      b.setAttribute("aria-pressed", b.getAttribute("data-filter") === sev ? "true" : "false");
    });
    applyFilter();
  }

  /* ── Copy path:line ── */
  function copyPath(btn) {
    var text = btn.getAttribute("data-copy");
    if (!text) return;
    var done = function () {
      var label = btn.querySelector(".label");
      var prev = label ? label.textContent : "";
      btn.classList.add("copied");
      if (label) label.textContent = "Copied";
      window.setTimeout(function () {
        btn.classList.remove("copied");
        if (label) label.textContent = prev || "Copy";
      }, 1400);
    };
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(done, function () {});
    } else {
      done();
    }
  }

  /* ── Theme toggle (overrides prefers-color-scheme) ── */
  function prefersDark() {
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
  }
  function currentTheme() {
    return root.getAttribute("data-theme") || (prefersDark() ? "dark" : "light");
  }
  function toggleTheme(btn) {
    var next = currentTheme() === "dark" ? "light" : "dark";
    root.setAttribute("data-theme", next);
    if (btn) btn.setAttribute("aria-pressed", next === "dark" ? "true" : "false");
  }

  /* ── Wire events (delegation where practical) ── */
  function init() {
    document.addEventListener("click", function (e) {
      var copy = e.target.closest ? e.target.closest(".copy-btn") : null;
      if (copy) { e.stopPropagation(); copyPath(copy); return; }

      var fp = e.target.closest ? e.target.closest(".fp") : null;
      if (fp && !fp.disabled) { setSeverity(fp.getAttribute("data-filter")); return; }

      var reset = e.target.closest ? e.target.closest("[data-filter-reset]") : null;
      if (reset) {
        var input = searchInput();
        if (input) input.value = "";
        setSeverity("all");
        return;
      }

      var expandAll = e.target.closest ? e.target.closest("#expand-all") : null;
      if (expandAll) { toggleAll(); return; }

      var theme = e.target.closest ? e.target.closest(".theme-toggle") : null;
      if (theme) { toggleTheme(theme); return; }

      var hdr = e.target.closest ? e.target.closest(".fc-hdr") : null;
      if (hdr) { toggleCard(hdr.closest(".fc")); return; }
    });

    /* Keyboard activation for finding headers (Enter / Space) — WCAG 2.1.1 */
    document.addEventListener("keydown", function (e) {
      if (e.key !== "Enter" && e.key !== " " && e.key !== "Spacebar") return;
      var hdr = e.target.closest ? e.target.closest(".fc-hdr") : null;
      if (hdr) { e.preventDefault(); toggleCard(hdr.closest(".fc")); }
    });

    var input = searchInput();
    if (input) input.addEventListener("input", applyFilter);

    /* Reflect the live OS colour preference so screen readers announce the
       correct toggle state before any click (the static default is "false"). */
    var themeBtn = document.querySelector(".theme-toggle");
    if (themeBtn) themeBtn.setAttribute("aria-pressed", currentTheme() === "dark" ? "true" : "false");

    syncExpandAll();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
