package mcp

import (
	"strings"
	"testing"
)

func remoteFixture() []ConfigFile {
	return []ConfigFile{{Path: ".mcp.json", Servers: []Server{
		{
			Name: "asana", Type: "http", URL: "https://mcp.asana.com/mcp",
			Headers: map[string]string{"Authorization": "Bearer ${ASANA_TOKEN}"}, Line: 3,
		},
		{Name: "shadow", Type: "http", URL: "https://unknown.example.net/mcp", Line: 6},
		{Name: "local", Type: "http", URL: "http://127.0.0.1:8080/mcp", Line: 9},
	}}}
}

func TestCheckTrustedRemotes(t *testing.T) {
	fs := checkTrustedRemotes(remoteFixture(), []string{"mcp.asana.com"})
	// asana allowlisted, local exempt -> only "shadow" flagged.
	if len(fs) != 1 || fs[0].RuleID != "AE-MCP-002" {
		t.Fatalf("expected 1 AE-MCP-002, got %+v", fs)
	}
	if fs[0].Locations[0].Line != 6 {
		t.Fatalf("wrong location: %+v", fs[0].Locations)
	}
	if !strings.Contains(fs[0].Summary, "not in the trusted-remote allowlist") {
		t.Fatalf("expected allowlist summary, got %q", fs[0].Summary)
	}
}

func TestCheckTrustedRemotesNoAllowlistFlagsAll(t *testing.T) {
	fs := checkTrustedRemotes(remoteFixture(), nil)
	// no allowlist -> both non-local remotes flagged (asana + shadow).
	if len(fs) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(fs))
	}
	for _, f := range fs {
		if !strings.Contains(f.Summary, "no charter.yaml mcp.trustedRemotes allowlist is configured") {
			t.Fatalf("expected no-allowlist summary, got %q", f.Summary)
		}
	}
}

func TestCheckRemoteAuth(t *testing.T) {
	fs := checkRemoteAuth(remoteFixture())
	// asana has Authorization header, local exempt -> only "shadow" flagged.
	if len(fs) != 1 || fs[0].RuleID != "AE-MCP-003" {
		t.Fatalf("expected 1 AE-MCP-003, got %+v", fs)
	}
	if fs[0].Locations[0].Line != 6 {
		t.Fatalf("wrong location: %+v", fs[0].Locations)
	}
}

func TestRemoteHost(t *testing.T) {
	cases := []struct {
		in   string
		host string
		ok   bool
	}{
		{"https://mcp.asana.com/mcp", "mcp.asana.com", true},
		{"https://Mcp.Asana.COM/mcp", "mcp.asana.com", true}, // lowercased
		{"https://host:8443/mcp", "host", true},              // host:port
		{"api.example.com/mcp", "", false},                   // scheme-less -> skipped
		{"${API_URL}", "", false},                            // env-ref -> skipped
		{"", "", false},
	}
	for _, c := range cases {
		h, ok := remoteHost(c.in)
		if h != c.host || ok != c.ok {
			t.Errorf("remoteHost(%q) = (%q,%v), want (%q,%v)", c.in, h, ok, c.host, c.ok)
		}
	}
}

func TestIsLocalHost(t *testing.T) {
	local := []string{
		"localhost", "127.0.0.1", "127.0.0.2", "::1", "0.0.0.0", "foo.localhost",
		// RFC1918 private + link-local + internal-only TLDs (FP-validation fix).
		"10.0.0.5", "172.16.1.230", "192.168.1.10", "169.254.1.1", "fe80::1",
		"archon.local", "db.internal",
	}
	for _, h := range local {
		if !isLocalHost(h) {
			t.Errorf("expected %q to be local/internal", h)
		}
	}
	for _, h := range []string{"mcp.asana.com", "example.com", "8.8.8.8", "172.32.0.1"} {
		if isLocalHost(h) {
			t.Errorf("expected %q to be a public remote", h)
		}
	}
}

func TestHasAuthHeader(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
		want    bool
	}{
		{"authorization", map[string]string{"Authorization": "Bearer ${T}"}, true},
		{"x-api-key", map[string]string{"X-Api-Key": "${K}"}, true},
		{"api-key", map[string]string{"Api-Key": "${K}"}, true},
		{"x-auth-token", map[string]string{"X-Auth-Token": "${K}"}, true},
		{"blank value", map[string]string{"Authorization": ""}, false},
		{"non-auth header", map[string]string{"X-Client-Id": "abc"}, false},
		{"nil headers", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasAuthHeader(c.headers); got != c.want {
				t.Errorf("hasAuthHeader(%v) = %v, want %v", c.headers, got, c.want)
			}
		})
	}
}
