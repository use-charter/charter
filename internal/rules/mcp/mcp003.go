package mcp

import (
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/findings"
)

// authHeaderNames are case-insensitive header keys that declare authentication.
var authHeaderNames = []string{"authorization", "x-api-key", "api-key", "x-auth-token"}

func hasAuthHeader(headers map[string]string) bool {
	for k, v := range headers {
		if strings.TrimSpace(v) == "" {
			continue
		}
		lk := strings.ToLower(strings.TrimSpace(k))
		for _, name := range authHeaderNames {
			if lk == name {
				return true
			}
		}
	}
	return false
}

// checkRemoteAuth emits AE-MCP-003 for remote servers that declare no auth
// metadata. oauthHosts are catalog-known vendor hosts that authenticate via the
// OAuth 2.1 flow rather than a static config header, so they are exempt — flagging
// them is a false positive (a config for e.g. mcp.sentry.dev legitimately carries
// no Authorization header). FP fix from the M1.6 catalog FP-validation (CF-13).
func checkRemoteAuth(files []ConfigFile, oauthHosts map[string]struct{}) []findings.Finding {
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
			if _, trusted := oauthHosts[host]; trusted {
				continue
			}
			if hasAuthHeader(s.Headers) {
				continue
			}
			result = append(result, findings.Finding{
				RuleID:      "AE-MCP-003",
				Severity:    findings.SeverityHigh,
				Category:    "MCP Safety",
				Summary:     "Remote MCP server declares no authentication metadata (OWASP MCP07)",
				Remediation: "Declare an auth header (e.g. Authorization referencing an environment variable) for the remote MCP server, or switch to a local/trusted integration mode.",
				Evidence:    []string{cf.Path + ": server " + s.Name + " (" + host + ") has no auth header"},
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
