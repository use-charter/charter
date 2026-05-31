package agentconfig

import "testing"

func TestParseHookConfigClaudeNested(t *testing.T) {
	raw := `{
  "hooks": {
    "PreToolUse": [
      { "matcher": "Bash", "hooks": [ { "type": "command", "command": "rm -rf ./build" } ] }
    ]
  }
}`
	cf, err := parseHookConfig(".claude/settings.json", []byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cf.Commands) != 1 || cf.Commands[0].Value != "rm -rf ./build" {
		t.Fatalf("got %+v", cf.Commands)
	}
	if cf.Commands[0].Line == 0 {
		t.Fatalf("expected a non-zero line, got %+v", cf.Commands[0])
	}
}

func TestParseHookConfigCursorFlatAndArgs(t *testing.T) {
	raw := `{ "version": 1, "hooks": { "beforeShellExecution": [ { "command": "git reset --hard" }, { "type": "command", "args": ["sudo", "rm", "-rf", "/"] } ] } }`
	cf, err := parseHookConfig(".cursor/hooks.json", []byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var values []string
	for _, c := range cf.Commands {
		values = append(values, c.Value)
	}
	if len(values) != 2 || values[0] != "git reset --hard" || values[1] != "sudo rm -rf /" {
		t.Fatalf("got %v", values)
	}
	// args entries are joined into one value; the joined string is not a
	// contiguous substring of the raw JSON, so its best-effort line is 0.
	for _, c := range cf.Commands {
		if c.Value == "sudo rm -rf /" && c.Line != 0 {
			t.Fatalf("expected joined-args command at file level (line 0), got %d", c.Line)
		}
	}
}

func TestParseHookConfigInvalidJSON(t *testing.T) {
	if _, err := parseHookConfig(".claude/settings.json", []byte("{ not json")); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestParseHookConfigNoHooks(t *testing.T) {
	cf, err := parseHookConfig(".claude/settings.json", []byte(`{ "model": "x" }`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cf.Commands) != 0 {
		t.Fatalf("expected no commands, got %+v", cf.Commands)
	}
}
