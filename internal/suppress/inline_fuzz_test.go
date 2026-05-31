package suppress

import "testing"

func FuzzParseInlineDirective(f *testing.F) {
	f.Add(`# charter:ignore AE-MCP-001 reason="x"`)
	f.Add(`// charter:ignore AE-CC-001`)
	f.Add(`<!-- charter:ignore AE-CC-002 expires=permanent -->`)
	f.Add(``)
	f.Add(`#`)
	f.Fuzz(func(t *testing.T, line string) {
		_, _ = parseInlineDirective(line)
	})
}
