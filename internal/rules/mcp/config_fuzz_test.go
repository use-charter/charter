package mcp

import "testing"

func FuzzParseConfigFile(f *testing.F) {
	f.Add(`{"mcpServers":{"a":{"command":"npx","args":["-y","p@1.0.0"]}}}`)
	f.Add(`{"servers":{"x":{"type":"http","url":"https://h/mcp"}}}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`{not json}`)
	f.Fuzz(func(t *testing.T, data string) {
		// Must never panic; an error return is acceptable for junk input.
		_, _ = parseConfigFile(".mcp.json", []byte(data))
	})
}
