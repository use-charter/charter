package mcp

import "testing"

func TestParseConfigFileStdioAndRemote(t *testing.T) {
	raw := `{
  "mcpServers": {
    "fs": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem@1.0.4"] },
    "asana": { "type": "http", "url": "https://mcp.asana.com/mcp", "headers": { "Authorization": "Bearer ${ASANA_TOKEN}" } }
  }
}`
	cf, err := parseConfigFile(".mcp.json", []byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.Path != ".mcp.json" {
		t.Fatalf("expected path .mcp.json, got %q", cf.Path)
	}
	if len(cf.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(cf.Servers))
	}
	byName := map[string]Server{}
	for _, s := range cf.Servers {
		byName[s.Name] = s
	}
	if byName["fs"].Command != "npx" || len(byName["fs"].Args) != 2 {
		t.Fatalf("stdio parse wrong: %+v", byName["fs"])
	}
	if byName["asana"].Type != "http" || byName["asana"].URL != "https://mcp.asana.com/mcp" {
		t.Fatalf("remote parse wrong: %+v", byName["asana"])
	}
	if byName["asana"].Line == 0 {
		t.Fatalf("expected a non-zero best-effort line for asana")
	}
}

func TestParseConfigFileServersAlias(t *testing.T) {
	raw := `{ "servers": { "x": { "command": "node", "args": ["s.js"] } } }`
	cf, err := parseConfigFile(".vscode/mcp.json", []byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cf.Servers) != 1 || cf.Servers[0].Name != "x" {
		t.Fatalf("alias parse wrong: %+v", cf.Servers)
	}
}

func TestParseConfigFileEmpty(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"empty object", "{}"},
		{"empty mcpServers", `{"mcpServers":{}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cf, err := parseConfigFile(".mcp.json", []byte(tc.raw))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cf.Path != ".mcp.json" {
				t.Fatalf("expected path .mcp.json, got %q", cf.Path)
			}
			if len(cf.Servers) != 0 {
				t.Fatalf("expected 0 servers, got %d", len(cf.Servers))
			}
		})
	}
}

func TestServerIsRemote(t *testing.T) {
	tests := []struct {
		name string
		s    Server
		want bool
	}{
		{"url only", Server{URL: "https://h/mcp"}, true},
		{"type http", Server{Type: "http"}, true},
		{"type sse", Server{Type: "sse"}, true},
		{"url and type", Server{URL: "https://h/mcp", Type: "http"}, true},
		{"stdio empty", Server{Command: "npx"}, false},
		{"zero value", Server{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsRemote(); got != tt.want {
				t.Errorf("IsRemote() = %v, want %v", got, tt.want)
			}
		})
	}
}
