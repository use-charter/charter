package mcp

import (
	"net"
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

// isLocalHost reports whether a host is local/internal rather than a public
// remote origin. AE-MCP-002 (shadow servers) and AE-MCP-003 (remote auth) target
// public origins, so loopback, RFC1918 private and link-local addresses, the
// unspecified address, and the reserved internal-only TLDs (.localhost, .local
// mDNS, .internal) are exempt — a server on your own machine or LAN is not a
// public shadow server. (FP fix from the M1.6 catalog FP-validation: a config
// pointing at a 172.16.x.x LAN address was wrongly flagged as an untrusted
// public remote.)
func isLocalHost(host string) bool {
	switch host {
	case "localhost", "::1", "0.0.0.0":
		return true
	}
	if strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()
	}
	return strings.HasPrefix(host, "127.")
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
			summary := "Remote MCP server origin is not in the trusted-remote allowlist (MCP catalog or charter.yaml mcp.trustedRemotes) (OWASP MCP09 Shadow Servers)"
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
