package mcp

import (
	"net/url"
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/findings"
)

func remoteHost(rawURL string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" {
		return "", false
	}
	return strings.ToLower(u.Hostname()), true
}

func isLocalHost(host string) bool {
	switch host {
	case "localhost", "::1", "0.0.0.0":
		return true
	}
	return strings.HasPrefix(host, "127.") || strings.HasSuffix(host, ".localhost")
}

func checkTrustedRemotes(files []ConfigFile, allow []string) []findings.Finding {
	allowed := map[string]struct{}{}
	for _, h := range allow {
		allowed[strings.ToLower(strings.TrimSpace(h))] = struct{}{}
	}

	noAllowlist := len(allow) == 0

	var result []findings.Finding
	for _, cf := range files {
		for _, s := range cf.Servers {
			if !s.IsRemote() {
				continue
			}
			host, ok := remoteHost(s.URL)
			if !ok || isLocalHost(host) {
				continue
			}
			if _, trusted := allowed[host]; trusted {
				continue
			}
			summary := "Remote MCP server origin is not in the trusted-remote allowlist (OWASP MCP09 Shadow Servers)"
			if noAllowlist {
				summary = "Remote MCP server origin cannot be verified — no charter.yaml mcp.trustedRemotes allowlist is configured (OWASP MCP09)"
			}
			result = append(result, findings.Finding{
				RuleID:      "AE-MCP-002",
				Severity:    findings.SeverityHigh,
				Category:    "MCP Safety",
				Summary:     summary,
				Remediation: "Add the host to charter.yaml mcp.trustedRemotes after review, or replace the server with a trusted origin.",
				Evidence:    []string{cf.Path + ": server " + s.Name + " -> " + host},
				Locations:   []findings.Location{{Path: cf.Path, Line: s.Line}},
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		li, lj := result[i].Locations[0], result[j].Locations[0]
		if li.Path != lj.Path {
			return li.Path < lj.Path
		}
		return li.Line < lj.Line
	})
	return result
}
