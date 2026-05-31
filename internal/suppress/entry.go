// Package suppress loads and applies Charter suppressions from .charter-suppress.yml
// (external) and inline charter:ignore comment directives (in-source), partitioning
// findings into active and suppressed for the doctor pipeline.
package suppress

import (
	"strings"

	"go.charter.dev/charter/internal/findings"
)

// File is the repo-relative suppression file path.
const File = ".charter-suppress.yml"

// Source identifies where a suppression came from. Values mirror SARIF
// suppression.kind for the future SARIF renderer.
const (
	SourceExternal = "external" // .charter-suppress.yml
	SourceInSource = "inSource" // inline charter:ignore directive
)

// FileEntry is one .charter-suppress.yml record (used for both read and write).
type FileEntry struct {
	Rule     string `yaml:"rule"`
	Reason   string `yaml:"reason,omitempty"`
	Expires  string `yaml:"expires,omitempty"`
	Approver string `yaml:"approver,omitempty"`
	Path     string `yaml:"path,omitempty"`
}

type suppressDoc struct {
	Suppressions []FileEntry `yaml:"suppressions"`
}

// Entry is a normalized suppression from either source.
type Entry struct {
	Rule     string
	Reason   string
	Approver string
	Expires  string // ISO YYYY-MM-DD, the literal "permanent", or "" (=> permanent)
	Path     string // file source: optional finding-path scope; inSource: the directive's file
	Source   string // SourceExternal | SourceInSource
	Line     int    // inSource directive line; 0 for file entries
}

// Suppressed pairs a suppressed finding with the suppression that muted it.
type Suppressed struct {
	Finding  findings.Finding
	Source   string
	Reason   string
	Approver string
	Expires  string
}

// IsPermanent reports whether an entry has no finite expiry (blank or "permanent").
func IsPermanent(e Entry) bool {
	x := strings.TrimSpace(e.Expires)
	return x == "" || strings.EqualFold(x, "permanent")
}
