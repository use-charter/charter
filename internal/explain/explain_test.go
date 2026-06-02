package explain

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/rules/catalog"
	"go.use-charter.dev/charter/internal/terminal"
)

func disabledCaps() terminal.Capabilities { return terminal.Capabilities{Tier: terminal.Mono} }

func ansi16Caps() terminal.Capabilities {
	return terminal.Capabilities{Tier: terminal.ANSI16, IsTTY: true}
}

func trueColorCaps(hyperlinks bool) terminal.Capabilities {
	return terminal.Capabilities{Tier: terminal.TrueColor, IsTTY: true, Hyperlinks: hyperlinks}
}

func paletteFor(c terminal.Capabilities) terminal.Palette { return terminal.NewPalette(c, true) }

// TestJSONIsTheCatalogEntry asserts --format json emits the catalog.Entry shape
// verbatim (the catalog is the source of truth; explain invents nothing).
func TestJSONIsTheCatalogEntry(t *testing.T) {
	t.Parallel()

	entry, ok := catalog.Lookup("AE-SEC-001")
	if !ok {
		t.Fatal("AE-SEC-001 must exist in the catalog")
	}
	data, err := JSON(entry)
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}

	var got catalog.Entry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("explain JSON is not valid catalog.Entry JSON: %v", err)
	}
	if got != entry {
		t.Fatalf("explain JSON round-trip drifted\n got: %+v\nwant: %+v", got, entry)
	}
	for _, field := range []string{"ID", "Name", "Category", "ShortDescription", "HelpURI"} {
		if !bytes.Contains(data, []byte(field)) {
			t.Fatalf("explain JSON missing field %q:\n%s", field, data)
		}
	}
}

// TestTextPlainResolvesEveryRule walks every catalog ID and proves the plain
// projection shows ID, Name, Category, the short description, and the docs URL,
// with zero ANSI escape bytes.
func TestTextPlainResolvesEveryRule(t *testing.T) {
	t.Parallel()

	for _, id := range catalog.IDs() {
		entry, ok := catalog.Lookup(id)
		if !ok {
			t.Fatalf("catalog.IDs() returned %q but Lookup failed", id)
		}
		got := Text(entry, disabledCaps(), paletteFor(disabledCaps()))
		if bytes.IndexByte(got, 0x1b) != -1 {
			t.Fatalf("plain explain for %s must contain zero ANSI escape bytes, got: %q", id, got)
		}
		out := string(got)
		for _, want := range []string{entry.ID, entry.Name, entry.Category, entry.ShortDescription, entry.HelpURI} {
			if !strings.Contains(out, want) {
				t.Fatalf("plain explain for %s missing %q\nfull output:\n%s", id, want, out)
			}
		}
	}
}

// TestTextStyledHasContentAndAnsi covers the styled path: the same content plus
// ANSI escapes (TrueColor, no hyperlinks).
func TestTextStyledHasContentAndAnsi(t *testing.T) {
	t.Parallel()

	entry, _ := catalog.Lookup("AE-MCP-001")
	caps := trueColorCaps(false)
	got := Text(entry, caps, paletteFor(caps))

	if bytes.IndexByte(got, 0x1b) == -1 {
		t.Fatalf("styled explain must contain ANSI escape bytes, got: %q", got)
	}
	out := string(got)
	for _, want := range []string{entry.ID, entry.Name, entry.Category, entry.ShortDescription, entry.HelpURI} {
		if !strings.Contains(out, want) {
			t.Fatalf("styled explain missing %q\nfull output:\n%s", want, out)
		}
	}
	// No hyperlinks requested: the OSC 8 introducer must be absent.
	if strings.Contains(out, "\x1b]8;") {
		t.Fatalf("did not expect OSC 8 hyperlinks when caps.Hyperlinks is false, got: %q", out)
	}
}

// TestTextStyledHyperlinks covers the OSC 8 link branch on the rule ID and docs
// URL.
func TestTextStyledHyperlinks(t *testing.T) {
	t.Parallel()

	entry, _ := catalog.Lookup("AE-SEC-001")
	caps := trueColorCaps(true)
	out := string(Text(entry, caps, paletteFor(caps)))

	if !strings.Contains(out, "\x1b]8;") {
		t.Fatalf("expected OSC 8 hyperlinks when caps.Hyperlinks is true, got: %q", out)
	}
	if !strings.Contains(out, entry.HelpURI) {
		t.Fatalf("expected the docs URL in the hyperlinked output, got: %q", out)
	}
}

// TestTextStyledANSI16Faint exercises the faint-attribute branch of the styled
// path (neutral tokens have no faithful ANSI-16 color and degrade to faint).
func TestTextStyledANSI16Faint(t *testing.T) {
	t.Parallel()

	entry, _ := catalog.Lookup("AE-CTX-001")
	caps := ansi16Caps()
	got := Text(entry, caps, paletteFor(caps))
	if bytes.IndexByte(got, 0x1b) == -1 {
		t.Fatalf("styled ANSI16 explain must contain ANSI escape bytes, got: %q", got)
	}
	if !strings.Contains(string(got), entry.ID) {
		t.Fatalf("styled ANSI16 explain missing the rule ID, got: %q", got)
	}
}
