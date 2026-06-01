package fix

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.use-charter.dev/charter/internal/repository"
	"go.use-charter.dev/charter/internal/scaffold"
)

// ctx004Required are the ignore patterns AE-CTX-004 expects, in the order they
// are appended when missing.
var ctx004Required = []string{
	".charter/",
	"*.charter-session",
	".claude/local/",
	".cursor/cache/",
	".hk/",
	".env*",
}

// fixCTX004 plans a .gitignore repair covering the Charter / agent-session
// artifact patterns. When .gitignore is absent it plans a Create with the full
// scaffold block. When it exists it parses the present patterns, and either
// reports nothing to do (all required patterns present) or plans an Append of
// just the missing patterns (in required order) under a single section header,
// separated from existing content by exactly one blank line.
func fixCTX004(root string, inv repository.Inventory) (FilePlan, bool, error) {
	const path = ".gitignore"
	target := filepath.Join(root, path)

	onDisk := false
	if _, err := os.Stat(target); err == nil {
		onDisk = true
	}

	if !inv.Has(path) && !onDisk {
		contents := scaffold.Gitignore()
		return FilePlan{
			RuleID:   "AE-CTX-004",
			Path:     path,
			Action:   Create,
			Contents: contents,
			Diff:     buildCreateDiff(path, contents),
		}, true, nil
	}

	// #nosec G304 -- fixed .gitignore filename joined to the scan root, not user-controlled.
	existing, err := os.ReadFile(target)
	if err != nil {
		return FilePlan{}, false, fmt.Errorf("fix AE-CTX-004: read %s: %w", path, err)
	}

	present := gitignorePatterns(existing)
	var missing []string
	for _, pat := range ctx004Required {
		if _, ok := present[pat]; !ok {
			missing = append(missing, pat)
		}
	}
	if len(missing) == 0 {
		return FilePlan{}, false, nil
	}

	var body strings.Builder
	body.WriteString(appendSeparator(existing))
	body.WriteString("# Charter / agent session artifacts\n")
	for _, pat := range missing {
		body.WriteString(pat)
		body.WriteByte('\n')
	}
	added := []byte(body.String())

	return FilePlan{
		RuleID:   "AE-CTX-004",
		Path:     path,
		Action:   Append,
		Contents: added,
		Diff:     buildAppendDiff(path, existing, added),
	}, true, nil
}

// gitignorePatterns returns the set of non-comment, non-blank pattern lines in
// a .gitignore, trimmed of surrounding whitespace.
func gitignorePatterns(data []byte) map[string]struct{} {
	out := map[string]struct{}{}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		out[trimmed] = struct{}{}
	}
	return out
}

// appendSeparator returns the leading bytes needed so the appended block is
// separated from the existing content by exactly one blank line, regardless of
// how many trailing newlines the existing file has.
func appendSeparator(existing []byte) string {
	switch {
	case len(existing) == 0:
		return ""
	case bytes.HasSuffix(existing, []byte("\n\n")):
		return ""
	case bytes.HasSuffix(existing, []byte("\n")):
		return "\n"
	default:
		return "\n\n"
	}
}
