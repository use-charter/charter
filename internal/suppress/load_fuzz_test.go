package suppress

import "testing"

func FuzzParseSuppressFile(f *testing.F) {
	f.Add("suppressions:\n  - rule: AE-MCP-001\n    reason: x\n")
	f.Add("suppressions: []")
	f.Add("")
	f.Add("not: yaml: [")
	f.Fuzz(func(t *testing.T, data string) {
		_, _ = parseSuppressFile([]byte(data))
	})
}
