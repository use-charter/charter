package repository

import (
	"os"
	"path/filepath"
	"strings"
)

// maxTrackedFileBytes caps how many bytes ReadTrackedFile will read from a
// single tracked file. Charter only ever reads manifests, configs, agent
// context, workflows, and source files for deterministic offline scanning, all
// of which are far below this bound. The cap is a denial-of-service guard: a
// malicious repository could otherwise track a multi-gigabyte file and force
// Charter to load it entirely into memory while scanning. 10 MiB is generous
// for any legitimately scanned file while keeping a hard per-file ceiling.
const maxTrackedFileBytes = 10 << 20 // 10 MiB

// ReadTrackedFile reads one repository file for scanning and returns its
// contents only when every safety gate passes. It is the single safe entry
// point for reading attacker-controlled repository content:
//
//   - Inventory-gated: path must be in the tracked inventory (git
//     ls-files-known). Untracked or ignored paths are never read.
//   - Symlink-contained: both the root and the resolved file are passed through
//     filepath.EvalSymlinks, and the resolved file must stay within the resolved
//     root. A symlink that escapes the repository root (e.g. AGENTS.md ->
//     ~/.aws/credentials) is rejected and never read, so its bytes cannot leak
//     into report evidence. Resolving the root too means a legitimately
//     symlinked root (macOS /tmp -> /private/tmp, t.TempDir() under /var ->
//     /private/var) is not falsely rejected.
//   - Size-capped: files larger than maxTrackedFileBytes are rejected (DoS
//     guard) so a single tracked file cannot exhaust memory.
//
// On any failed gate (not tracked, EvalSymlinks error for a missing or raced
// path, containment failure, oversize, or read error) it returns ("", false)
// and does not read the file. It performs no network or other side effects and
// is pure, deterministic, and offline.
func ReadTrackedFile(root string, inv Inventory, path string) (string, bool) {
	if !inv.Has(path) {
		return "", false
	}

	// Resolve both sides so a legitimately symlinked root is not mistaken for an
	// escape, while a file symlink that actually leaves the root is rejected.
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", false
	}
	realPath, err := filepath.EvalSymlinks(filepath.Join(realRoot, filepath.FromSlash(path)))
	if err != nil {
		return "", false
	}
	if realPath != realRoot && !strings.HasPrefix(realPath, realRoot+string(os.PathSeparator)) {
		return "", false
	}

	info, err := os.Stat(realPath)
	if err != nil || info.Size() > maxTrackedFileBytes {
		return "", false
	}

	// #nosec G304 -- realPath is an inventory-gated, root-contained, size-capped
	// tracked repository path resolved via EvalSymlinks (see the gates above).
	data, err := os.ReadFile(realPath)
	if err != nil {
		return "", false
	}
	return string(data), true
}
