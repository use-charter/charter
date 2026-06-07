package mcp

import (
	"regexp"
	"sort"
	"strings"

	"go.use-charter.dev/charter/internal/catalog"
	"go.use-charter.dev/charter/internal/findings"
)

// exactVersionPattern matches a pinned exact semver; digestPattern matches a
// content digest. A version matching either is considered pinned.
var (
	exactVersionPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$`)
	digestPattern       = regexp.MustCompile(`^(?:sha256:[0-9a-f]{64}|[0-9a-f]{40})$`)
)

// classifyPackageSpec splits a package token into name + version and reports
// whether the version is exactly pinned. Unpinned: missing version, "latest"
// and dist-tags, semver ranges, and floating git refs.
func classifyPackageSpec(token string) (name, version string, pinned bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", "", false
	}
	if strings.HasPrefix(token, "github:") || strings.HasPrefix(token, "git+") || strings.Contains(token, "#") {
		return token, "", false // floating git ref
	}
	at := strings.LastIndex(token, "@")
	// at == 0 means the only '@' is the scope prefix -> no version present.
	if at <= 0 {
		return token, "", false
	}
	name = token[:at]
	version = token[at+1:]
	return name, version, isPinnedVersion(version)
}

func isPinnedVersion(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	return exactVersionPattern.MatchString(v) || digestPattern.MatchString(v)
}

// directRunners install and run a remote package directly: <runner> [flags] <pkg>.
var directRunners = map[string]struct{}{"npx": {}, "bunx": {}, "uvx": {}}

// dlxRunners fetch a remote package only via the "dlx" subcommand:
// <runner> dlx [flags] <pkg>. Their "exec"/"run"/script forms launch local
// binaries, which are not registry package specs.
var dlxRunners = map[string]struct{}{"pnpm": {}, "yarn": {}}

// packageTokenFromArgs returns the remote package spec token for a runner-based
// stdio server. Returns ok=false when the command is not a recognized runner,
// when no package arg applies (local binary, script, or "exec"/"run" form), or
// when the candidate is a local path rather than a registry spec.
func packageTokenFromArgs(command string, args []string) (string, bool) {
	base := command
	if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
		base = base[i+1:]
	}

	rest := args
	if _, ok := dlxRunners[base]; ok {
		// Only "dlx" launches a remote package; exec/run/scripts are local.
		i := firstNonFlag(args)
		if i < 0 || args[i] != "dlx" {
			return "", false
		}
		rest = args[i+1:]
	} else if _, ok := directRunners[base]; !ok {
		return "", false
	}

	i := firstNonFlag(rest)
	if i < 0 {
		return "", false
	}
	tok := rest[i]
	// Local paths are never registry package specs.
	if strings.HasPrefix(tok, "/") || strings.HasPrefix(tok, "./") || strings.HasPrefix(tok, "../") {
		return "", false
	}
	return tok, true
}

// firstNonFlag returns the index of the first arg that is not a "-"-prefixed
// flag, or -1 if none.
func firstNonFlag(args []string) int {
	for i, a := range args {
		if !strings.HasPrefix(a, "-") {
			return i
		}
	}
	return -1
}

// checkPinning evaluates AE-MCP-001 across all servers, emitting at most one
// finding per server via a precedence ladder (ADR-0021):
// deprecated > unpinned > advisory > behind-stable > clean. Catalog comparisons
// are exact-match only, so a version absent from the catalog's known set is
// silent (a stale catalog under-reports rather than misreporting).
func checkPinning(files []ConfigFile, cat *catalog.Catalog) []findings.Finding {
	var result []findings.Finding
	for _, cf := range files {
		for _, s := range cf.Servers {
			token, ok := packageTokenFromArgs(s.Command, s.Args)
			if !ok {
				continue
			}
			if f, ok := catalogFinding(cf, s, token, cat); ok {
				result = append(result, f)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		li, lj := findings.FirstLocation(result[i]), findings.FirstLocation(result[j])
		if li.Path != lj.Path {
			return li.Path < lj.Path
		}
		return li.Line < lj.Line
	})
	return result
}

func catalogFinding(cf ConfigFile, s Server, token string, cat *catalog.Catalog) (findings.Finding, bool) {
	name, version, pinned := classifyPackageSpec(token)
	entry, known := cat.Lookup(name)
	loc := []findings.Location{{Path: cf.Path, Line: s.Line}}
	base := cf.Path + ": server " + s.Name

	// 1. Deprecated/archived package — migration supersedes pinning.
	if known && entry.Status == "deprecated" {
		return findings.Finding{
			RuleID:      "AE-MCP-001",
			Severity:    findings.SeverityHigh,
			Category:    "MCP Safety",
			Summary:     "MCP server package " + name + " is archived/deprecated — migrate to " + entry.Successor + " (supply-chain maintenance risk)",
			Remediation: "Replace " + name + " with its maintained successor: " + entry.Successor + ".",
			Evidence:    []string{base + " uses " + token + " (archived; see " + entry.Reference + ")"},
			Locations:   loc,
		}, true
	}

	// 2. Unpinned — the original AE-MCP-001 supply-chain check.
	if !pinned {
		return findings.Finding{
			RuleID:      "AE-MCP-001",
			Severity:    findings.SeverityHigh,
			Category:    "MCP Safety",
			Summary:     "MCP server package is not pinned to an exact version (supply-chain risk, OWASP MCP04)",
			Remediation: "Pin the MCP server package to an exact version or digest instead of @latest, a semver range, or a floating git ref.",
			Evidence:    []string{base + " uses " + token},
			Locations:   loc,
		}, true
	}

	if !known {
		return findings.Finding{}, false
	}

	// 3. Known advisory affecting the pinned version.
	if adv, hit := entry.AdvisoryFor(version); hit {
		summary := "MCP server " + name + "@" + version + " is affected by " + adv.ID + " — fixed in " + adv.FixedIn
		if adv.Summary != "" {
			summary += " (" + adv.Summary + ")"
		}
		ev := base + " uses " + token + "; advisory " + adv.ID
		if adv.Reference != "" {
			ev += " (" + adv.Reference + ")"
		}
		return findings.Finding{
			RuleID:      "AE-MCP-001",
			Severity:    findings.SeverityHigh,
			Category:    "MCP Safety",
			Summary:     summary,
			Remediation: "Upgrade " + name + " to " + adv.FixedIn + " or later.",
			Evidence:    []string{ev},
			Locations:   loc,
		}, true
	}

	// 4. Behind the catalog's stable version — informational nudge (non-deducting,
	// re-surfaces). Mirrors AE-SUPPRESS-003's informational shape.
	if stable, behind := entry.KnownBehind(version); behind {
		return findings.Finding{
			RuleID:        "AE-MCP-001",
			Severity:      findings.SeverityLow,
			Category:      "MCP Safety",
			Summary:       "MCP server " + name + " is pinned to " + version + "; catalog stable version is " + stable + " — upgrade available",
			Remediation:   "Bump " + name + " to " + stable + " when convenient.",
			Evidence:      []string{base + " uses " + token + " (catalog stable: " + stable + ")"},
			Locations:     loc,
			Informational: true,
		}, true
	}

	// 5. Clean.
	return findings.Finding{}, false
}
