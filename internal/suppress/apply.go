package suppress

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.use-charter.dev/charter/internal/findings"
)

// Apply partitions findings into active and suppressed using .charter-suppress.yml
// entries and inline charter:ignore directives, honoring expiry/approval against
// now. used reports the suppressions governance should audit: every non-expired
// file entry plus every inline directive discovered at a finding location. It
// fails fast (wrapped error) when a finding's source file cannot be read.
func Apply(root string, all []findings.Finding, fileEntries []Entry, now time.Time) (active []findings.Finding, suppressed []Suppressed, used []Entry, err error) {
	for _, e := range fileEntries {
		if auditable(e, now) {
			used = append(used, e)
		}
	}

	cache := map[string][]string{}
	for _, f := range all {
		if e, ok := matchFileEntry(f, fileEntries, now); ok {
			suppressed = append(suppressed, toSuppressed(f, e))
			continue
		}

		inline, ok, derr := matchInline(root, f, cache)
		if derr != nil {
			return nil, nil, nil, derr
		}
		if ok && auditable(inline, now) {
			used = append(used, inline)
		}
		if ok && honored(inline, now) {
			suppressed = append(suppressed, toSuppressed(f, inline))
			continue
		}

		active = append(active, f)
	}
	return active, suppressed, used, nil
}

func toSuppressed(f findings.Finding, e Entry) Suppressed {
	return Suppressed{Finding: f, Source: e.Source, Reason: e.Reason, Approver: e.Approver, Expires: e.Expires}
}

// matchFileEntry returns the first honored file entry whose rule matches the
// finding and whose path scope (if any) matches a finding location.
func matchFileEntry(f findings.Finding, entries []Entry, now time.Time) (Entry, bool) {
	for _, e := range entries {
		if e.Rule != f.RuleID {
			continue
		}
		if e.Path != "" && !locationMatches(f, e.Path) {
			continue
		}
		if !honored(e, now) {
			continue
		}
		return e, true
	}
	return Entry{}, false
}

func locationMatches(f findings.Finding, path string) bool {
	for _, loc := range f.Locations {
		if loc.Path == path {
			return true
		}
	}
	return false
}

// matchInline reads each finding location's source line and returns a
// charter:ignore directive for the finding's rule if present on that line.
func matchInline(root string, f findings.Finding, cache map[string][]string) (Entry, bool, error) {
	for _, loc := range f.Locations {
		if loc.Path == "" || loc.Line <= 0 {
			continue
		}
		lines, err := fileLines(root, loc.Path, cache)
		if err != nil {
			return Entry{}, false, err
		}
		if loc.Line > len(lines) {
			continue
		}
		e, ok := parseInlineDirective(lines[loc.Line-1])
		if !ok || e.Rule != f.RuleID {
			continue
		}
		e.Path = loc.Path
		e.Line = loc.Line
		return e, true, nil
	}
	return Entry{}, false, nil
}

func fileLines(root, path string, cache map[string][]string) ([]string, error) {
	if lines, ok := cache[path]; ok {
		return lines, nil
	}
	// #nosec G304 -- path is a finding location drawn from the tracked inventory scan.
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return nil, fmt.Errorf("read %s for inline suppression: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")
	cache[path] = lines
	return lines, nil
}

// honored reports whether a suppression actually mutes its finding: a blank
// expires (default TTL), a valid non-expired explicit date, or an explicit
// permanent waiver that names an approver. A permanent waiver without an approver,
// an expired date, or a malformed date is not honored (fail closed).
func honored(e Entry, now time.Time) bool {
	x := strings.TrimSpace(e.Expires)
	if x == "" {
		return true // default TTL: honored
	}
	if strings.EqualFold(x, "permanent") {
		return strings.TrimSpace(e.Approver) != ""
	}
	end, ok := expiryEnd(x)
	if !ok {
		return false // malformed date: fail closed
	}
	return now.Before(end)
}

// auditable reports whether governance should inspect a suppression: blank-expiry
// (default) and explicit-permanent entries (so AE-SUPPRESS-001/002 can flag them)
// plus non-expired dated entries. Expired or malformed-date entries are inert.
func auditable(e Entry, now time.Time) bool {
	x := strings.TrimSpace(e.Expires)
	if x == "" || strings.EqualFold(x, "permanent") {
		return true
	}
	end, ok := expiryEnd(x)
	if !ok {
		return false
	}
	return now.Before(end)
}

// expiryEnd parses an ISO date and returns the instant the suppression stops
// applying (start of the day after the expiry date). ok=false on a malformed date.
func expiryEnd(expires string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(expires))
	if err != nil {
		return time.Time{}, false
	}
	return t.Add(24 * time.Hour), true
}
