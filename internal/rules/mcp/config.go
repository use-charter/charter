package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Server is a normalized MCP server entry from any supported client config.
type Server struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
	Type    string // "http" | "sse" | "" (stdio)
	URL     string
	Headers map[string]string
	Line    int // 1-based best-effort line of the server key; 0 if unknown
}

// ConfigFile is one parsed MCP config file.
type ConfigFile struct {
	Path    string
	Servers []Server
}

// IsRemote reports whether the server is a remote (http/sse) endpoint.
func (s Server) IsRemote() bool {
	return s.URL != "" || s.Type == "http" || s.Type == "sse"
}

type rawServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	HTTPURL string            `json:"httpUrl"` // Gemini CLI streamable-HTTP transport
	Headers map[string]string `json:"headers"`
}

// remoteURL returns the server's remote endpoint, preferring "url" (SSE / VS
// Code) and falling back to Gemini CLI's "httpUrl" (streamable HTTP). Empty for
// stdio servers.
func (r rawServer) remoteURL() string {
	if r.URL != "" {
		return r.URL
	}
	return r.HTTPURL
}

type rawConfig struct {
	MCPServers map[string]rawServer `json:"mcpServers"`
	Servers    map[string]rawServer `json:"servers"`
}

func parseConfigFile(path string, data []byte) (ConfigFile, error) {
	var rc rawConfig
	if err := json.Unmarshal(data, &rc); err != nil {
		return ConfigFile{}, fmt.Errorf("parse %s: %w", path, err)
	}

	// mcpServers takes precedence; servers is a client alias used by VS Code.
	entries := rc.MCPServers
	if len(entries) == 0 {
		entries = rc.Servers
	}

	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)

	cf := ConfigFile{Path: path, Servers: make([]Server, 0, len(names))}
	for _, name := range names {
		r := entries[name]
		cf.Servers = append(cf.Servers, Server{
			Name:    name,
			Command: r.Command,
			Args:    r.Args,
			Env:     r.Env,
			Type:    r.Type,
			URL:     r.remoteURL(),
			Headers: r.Headers,
			Line:    bestEffortLineOfKey(string(data), name),
		})
	}
	return cf, nil
}

// bestEffortLineOfKey returns the 1-based line of the first occurrence of "name" as a JSON
// key. Deterministic best-effort; returns 0 if not found.
// NOTE: if name is a substring of an earlier key (e.g. "s" inside "servers"),
// the returned line may point to that earlier key. Sufficient for diagnostic display.
func bestEffortLineOfKey(text, name string) int {
	needle := `"` + name + `"`
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, needle) {
			return i + 1
		}
	}
	return 0
}
