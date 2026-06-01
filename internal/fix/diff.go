package fix

import (
	"fmt"
	"strings"
)

// buildCreateDiff renders the unified diff for creating path from nothing. The
// hunk header counts every line of contents as an addition against /dev/null.
func buildCreateDiff(path string, contents []byte) string {
	lines := splitLines(contents)

	var b strings.Builder
	b.WriteString("--- /dev/null\n")
	fmt.Fprintf(&b, "+++ b/%s\n", path)
	fmt.Fprintf(&b, "@@ -0,0 +1,%d @@\n", len(lines))
	for _, line := range lines {
		b.WriteByte('+')
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

// buildAppendDiff renders the unified diff for appending added to the end of an
// existing file. It shows up to three trailing context lines of existing
// (space-prefixed) followed by every added line (`+`-prefixed) and a hunk
// header whose old count is the context line count and whose new count is the
// context plus the additions. It is deterministic and tolerates a missing
// trailing newline in either input.
func buildAppendDiff(path string, existing, added []byte) string {
	existingLines := splitLines(existing)
	addedLines := splitLines(added)

	ctx := len(existingLines)
	if ctx > 3 {
		ctx = 3
	}
	ctxLines := existingLines[len(existingLines)-ctx:]

	// For a non-empty file the hunk starts at the first context line on both
	// sides; an empty file appends as a fresh hunk at line 1 of the new side.
	oldStart := len(existingLines) - ctx + 1
	newStart := oldStart
	if len(existingLines) == 0 {
		oldStart = 0
		newStart = 1
	}

	var b strings.Builder
	fmt.Fprintf(&b, "--- a/%s\n", path)
	fmt.Fprintf(&b, "+++ b/%s\n", path)
	fmt.Fprintf(&b, "@@ -%d,%d +%d,%d @@\n", oldStart, ctx, newStart, ctx+len(addedLines))
	for _, line := range ctxLines {
		b.WriteByte(' ')
		b.WriteString(line)
		b.WriteByte('\n')
	}
	for _, line := range addedLines {
		b.WriteByte('+')
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

// splitLines splits b into logical lines, discarding exactly one trailing
// newline so a file ending in "\n" does not yield a phantom empty final line.
// Empty input yields no lines.
func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	s := strings.TrimSuffix(string(b), "\n")
	return strings.Split(s, "\n")
}
