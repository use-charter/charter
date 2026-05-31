package agentconfig

import (
	"regexp"
	"sort"
	"strings"

	"go.charter.dev/charter/internal/findings"
)

// multiTokenPatterns are specific multi-token phrases; substring matching is
// accurate for them. Operator-chaining / command-substitution injection
// (&&, ||, ;, $(...), backticks) is intentionally deferred (false-positive-prone).
var multiTokenPatterns = []string{
	"rm -rf", "git reset --hard", "git clean -fd", "sudo ", "chmod 777", "chown -r ",
}

// singleWordRE matches dangerous single-word executables at a word boundary, so
// "dd" does not fire inside "add"/"git add", and "truncate" does not fire inside
// "untruncated".
var singleWordRE = regexp.MustCompile(`\bdd\b|\btruncate\b|\bmkfs`)

func isDangerousCommand(cmd string) bool {
	lower := strings.ToLower(strings.Join(strings.Fields(cmd), " "))
	for _, p := range multiTokenPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return singleWordRE.MatchString(lower)
}

func checkDangerousCommands(files []ConfigFile) []findings.Finding {
	var result []findings.Finding
	for _, cf := range files {
		for _, c := range cf.Commands {
			if !isDangerousCommand(c.Value) {
				continue
			}
			result = append(result, findings.Finding{
				RuleID:      "AE-CC-001",
				Severity:    findings.SeverityBlocker,
				Category:    "Agent Config",
				Summary:     "Agent hook config contains a dangerous shell command (OWASP MCP05 Command Injection & Execution)",
				Remediation: "Replace the destructive or privilege-escalating command with an explicit, scoped command; prefer array-form (args) execution.",
				Evidence:    []string{cf.Path + ": hook command uses " + strings.TrimSpace(c.Value)},
				Locations:   []findings.Location{{Path: cf.Path, Line: c.Line}},
			})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		li, lj := result[i].Locations[0], result[j].Locations[0]
		if li.Path != lj.Path {
			return li.Path < lj.Path
		}
		return li.Line < lj.Line
	})
	return result
}
