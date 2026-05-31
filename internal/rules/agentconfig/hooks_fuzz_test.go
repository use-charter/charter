package agentconfig

import "testing"

func FuzzParseHookConfig(f *testing.F) {
	f.Add(`{"hooks":{"PreToolUse":[{"hooks":[{"command":"rm -rf /"}]}]}}`)
	f.Add(`{"version":1,"hooks":{"beforeShellExecution":[{"command":"x"}]}}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`{not json}`)
	f.Fuzz(func(t *testing.T, data string) {
		_, _ = parseHookConfig(".claude/settings.json", []byte(data))
	})
}
