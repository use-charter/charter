// Package explain formats a single Charter rule for `charter explain <RULE>`.
//
// It is a thin, pure projection over internal/rules/catalog: the catalog is the
// in-binary source of truth (ID, name, category, short description, help URI),
// and the full rationale/remediation prose lives at the help URL. This package
// invents no content beyond what the catalog carries.
//
// Like internal/render/text it has two text paths chosen solely by
// caps.ColorEnabled(): a plain, ANSI-free layout for non-TTY/NO_COLOR, and a
// styled layout (lipgloss + the internal/terminal palette) for a real TTY. It
// performs no capability detection and no I/O — the caller owns the TTY
// boundary and writes the returned bytes.
package explain

import (
	"bytes"
	"encoding/json"
	"fmt"

	"charm.land/lipgloss/v2"
	"go.use-charter.dev/charter/internal/rules/catalog"
	"go.use-charter.dev/charter/internal/terminal"
)

// JSON renders the catalog entry as indented JSON (the catalog.Entry shape).
func JSON(e catalog.Entry) ([]byte, error) {
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("render explain json: %w", err)
	}
	return b, nil
}

// Text renders a human-readable explanation of one rule. With color disabled it
// returns a plain, ANSI-free layout; otherwise it styles the rule ID, labels,
// and docs link via the palette (hyperlinking the docs URL when supported).
func Text(e catalog.Entry, caps terminal.Capabilities, pal terminal.Palette) []byte {
	if !caps.ColorEnabled() {
		return textPlain(e)
	}
	return textStyled(e, caps, pal)
}

// textPlain is the historical, ANSI-free layout: a rule heading followed by
// aligned Category / Summary / Docs rows.
func textPlain(e catalog.Entry) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s  %s\n", e.ID, e.Name)
	fmt.Fprintf(&b, "Category  %s\n", e.Category)
	fmt.Fprintf(&b, "Summary   %s\n", e.ShortDescription)
	fmt.Fprintf(&b, "Docs      %s\n", e.HelpURI)
	return b.Bytes()
}

// textStyled renders the same content with palette color and an optional OSC 8
// link on the rule ID and the docs URL.
func textStyled(e catalog.Entry, caps terminal.Capabilities, pal terminal.Palette) []byte {
	st := func(tok terminal.Token) lipgloss.Style {
		resolved := pal.Resolve(tok)
		s := lipgloss.NewStyle()
		if resolved.HasColor() {
			s = s.Foreground(resolved.Color)
		}
		if resolved.Bold {
			s = s.Bold(true)
		}
		if resolved.Faint {
			s = s.Faint(true)
		}
		if resolved.Reverse {
			s = s.Reverse(true)
		}
		return s
	}
	link := func(s lipgloss.Style) lipgloss.Style {
		if caps.Hyperlinks && e.HelpURI != "" {
			return s.Hyperlink(e.HelpURI)
		}
		return s
	}

	var b bytes.Buffer
	id := link(st(terminal.TextInfo).Bold(true)).Render(e.ID)
	name := st(terminal.TextPrimary).Bold(true).Render(e.Name)
	fmt.Fprintln(&b, id+"  "+name)
	fmt.Fprintln(&b, st(terminal.TextTertiary).Render("Category  ")+st(terminal.TextSecondary).Render(e.Category))
	fmt.Fprintln(&b, st(terminal.TextTertiary).Render("Summary   ")+st(terminal.TextPrimary).Render(e.ShortDescription))
	fmt.Fprintln(&b, st(terminal.TextTertiary).Render("Docs      ")+link(st(terminal.TextInfo)).Render(e.HelpURI))
	return b.Bytes()
}
