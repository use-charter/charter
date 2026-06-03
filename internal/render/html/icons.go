package html

import "html/template"

// icon returns an inline SVG for the given key as trusted HTML. Icons are inline
// (not an icon font) so the report stays fully self-contained and offline — no
// external stylesheet or font is ever requested. Unknown keys fall back to a dot.
func icon(key string) template.HTML {
	if key == "brandmark" {
		return template.HTML(brandMark) //nolint:gosec // static, author-controlled SVG
	}
	body, ok := iconPaths[key]
	if !ok {
		body = iconPaths["dot"]
	}
	return template.HTML(svgOpen + body + svgClose) //nolint:gosec // static, author-controlled SVG
}

const (
	svgOpen  = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">`
	svgClose = `</svg>`
)

// brandMark is the committed Charter [C] mark (docs/internal/designs/brand/mark.svg):
// a geometric bracket pair enclosing a 300° arc, drawn with currentColor.
const brandMark = `<svg viewBox="0 0 64 64" fill="currentColor" aria-hidden="true" focusable="false">` +
	`<rect x="3" y="12" width="4" height="40"/><rect x="3" y="12" width="11" height="4"/><rect x="3" y="48" width="11" height="4"/>` +
	`<rect x="57" y="12" width="4" height="40"/><rect x="50" y="12" width="11" height="4"/><rect x="50" y="48" width="11" height="4"/>` +
	`<path d="M41.53,26.5 A11,11 0 1 0 41.53,37.5" fill="none" stroke="currentColor" stroke-width="6" stroke-linecap="butt"/></svg>`

var iconPaths = map[string]string{
	"check":          `<path d="M5 12.5l4.5 4.5L19 7"/>`,
	"x":              `<path d="M6 6l12 12M18 6L6 18"/>`,
	"alert-triangle": `<path d="M12 3.5l9 15.5H3z"/><path d="M12 10v4"/><path d="M12 17h.01"/>`,
	"chevron-down":   `<path d="M6 9l6 6 6-6"/>`,
	"copy":           `<rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15V5a2 2 0 0 1 2-2h10"/>`,
	"search":         `<circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3"/>`,
	"external-link":  `<path d="M14 4h6v6"/><path d="M20 4l-9 9"/><path d="M19 14v4a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2h4"/>`,
	"filter":         `<path d="M4 5h16l-6 7v6l-4 2v-8z"/>`,
	"info":           `<circle cx="12" cy="12" r="9"/><path d="M12 11v5M12 7.5h.01"/>`,
	"sun":            `<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M2 12h2M20 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4"/>`,
	"moon":           `<path d="M21 12.5A8.5 8.5 0 1 1 11.5 3a6.5 6.5 0 0 0 9.5 9.5z"/>`,
	"grid":           `<rect x="4" y="4" width="7" height="7" rx="1"/><rect x="13" y="4" width="7" height="7" rx="1"/><rect x="4" y="13" width="7" height="7" rx="1"/><rect x="13" y="13" width="7" height="7" rx="1"/>`,
	"bug":            `<rect x="8" y="8" width="8" height="11" rx="4"/><path d="M9 8a3 3 0 0 1 6 0"/><path d="M4 12h4M16 12h4M5 7l3 2M19 7l-3 2M5 17l3-2M19 17l-3-2"/>`,
	"eye-off":        `<path d="M3 3l18 18"/><path d="M10.6 10.6a2 2 0 0 0 2.8 2.8"/><path d="M9.4 5.2A9 9 0 0 1 21 12a14 14 0 0 1-2.2 3M6.3 6.3A14 14 0 0 0 3 12a9 9 0 0 0 12 5"/>`,
	"chart":          `<path d="M3 21h18"/><rect x="5" y="11" width="3" height="7"/><rect x="11" y="6" width="3" height="12"/><rect x="17" y="9" width="3" height="9"/>`,
	"lock":           `<rect x="5" y="11" width="14" height="9" rx="2"/><path d="M8 11V8a4 4 0 0 1 8 0v3"/>`,
	"shield":         `<path d="M12 3l8 3v5c0 5-3.5 8-8 10-4.5-2-8-5-8-10V6z"/>`,
	"shield-check":   `<path d="M12 3l8 3v5c0 5-3.5 8-8 10-4.5-2-8-5-8-10V6z"/><path d="M9 12l2 2 4-4"/>`,
	"settings":       `<circle cx="12" cy="12" r="3"/><path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M19.1 4.9L17 7M7 17l-2.1 2.1"/>`,
	"package":        `<path d="M12 3l8 4v10l-8 4-8-4V7z"/><path d="M4 7l8 4 8-4"/><path d="M12 11v10"/>`,
	"git-branch":     `<circle cx="6" cy="6" r="2.2"/><circle cx="6" cy="18" r="2.2"/><circle cx="18" cy="8" r="2.2"/><path d="M6 8.2v7.6"/><path d="M18 10.2c0 4-6 2-6 6"/>`,
	"file-text":      `<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"/><path d="M14 3v5h5"/><path d="M8 13h8M8 17h6"/>`,
	"test-pipe":      `<path d="M9 3h6"/><path d="M10 3v7l-5 8a2 2 0 0 0 1.8 3h10.4a2 2 0 0 0 1.8-3l-5-8V3"/>`,
	"bolt":           `<path d="M13 2L4 14h7l-1 8 9-12h-7z"/>`,
	"dot":            `<circle cx="12" cy="12" r="3"/>`,
}
