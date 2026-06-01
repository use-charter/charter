// Package fix is the diff-first repair engine behind `charter fix`. It plans
// file changes that resolve a subset of doctor findings, renders unified-diff
// previews, and applies them under three non-negotiable safety properties:
//
//   - It backs up any existing target before mutating it.
//   - It never deletes, truncates, or otherwise touches an unrelated file, and
//     never writes outside the resolved repository root.
//   - It only ever acts on rules with a registered fixer; secret and dangerous
//     rules (AE-SEC-001/002, AE-CC-001, ...) have no fixer and can never be
//     targeted, even if a caller hands Apply a hand-built plan for one.
//
// Plan is pure (no disk writes); Apply owns all I/O.
package fix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.use-charter.dev/charter/internal/doctor"
	"go.use-charter.dev/charter/internal/repository"
)

// Action is the kind of file change a plan performs.
type Action int

const (
	// Create writes a full file to a path that must be absent at apply time.
	Create Action = iota
	// Append adds bytes to the end of an existing file (after backing it up).
	Append
)

// FilePlan is a single planned, previewable file change.
type FilePlan struct {
	RuleID   string
	Path     string // repo-relative, forward-slash
	Action   Action
	Contents []byte // Create: full file; Append: the bytes to append
	Diff     string // unified diff preview
}

// Options narrows planning to a single rule when Rule is set; empty means all
// fixable rules.
type Options struct{ Rule string }

// fixer computes the plan for one rule. The bool reports whether a fix is
// actually warranted (false means the rule is already satisfied / nothing to
// do). It performs reads but never writes.
type fixer func(root string, inv repository.Inventory) (FilePlan, bool, error)

var registry = map[string]fixer{
	"AE-CTX-001": fixCTX001,
	"AE-CTX-004": fixCTX004,
	"AE-CI-002":  fixCI002,
}

// Fixable reports whether a rule has a registered fixer. Rules without one
// (notably the secret and dangerous-config rules) are never fixable.
func Fixable(ruleID string) bool {
	_, ok := registry[ruleID]
	return ok
}

// Plan builds the fix plans for the active findings whose rule has a registered
// fixer (filtered by opts.Rule when set). It is pure: no disk writes. At most
// one plan is produced per rule, in first-seen finding order, and a rule whose
// fixer reports "nothing to do" yields no plan.
func Plan(result doctor.Result, root string, inv repository.Inventory, opts Options) ([]FilePlan, error) {
	var plans []FilePlan
	seen := map[string]struct{}{}

	for _, f := range result.Findings {
		if opts.Rule != "" && f.RuleID != opts.Rule {
			continue
		}
		fn, ok := registry[f.RuleID]
		if !ok {
			continue
		}
		if _, done := seen[f.RuleID]; done {
			continue
		}
		seen[f.RuleID] = struct{}{}

		plan, want, err := fn(root, inv)
		if err != nil {
			return nil, err
		}
		if !want {
			continue
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

// Apply writes the plans, backing up any existing target to
// .charter/backups/<ts>/<path> before mutating it. A single timestamp is used
// for the whole call. Create writes only to a path that is still absent at
// apply time (a re-stat guards against overwriting). Append reads the existing
// file, backs it up, then rewrites it as existing+Contents. Apply never deletes
// or truncates an unrelated file, never writes outside root, and refuses any
// plan whose rule has no registered fixer. It returns the written repo-relative
// paths and the relative backup directory used (empty when nothing was backed
// up).
func Apply(root string, plans []FilePlan) (written []string, backupDir string, err error) {
	ts := time.Now().UTC().Format("20060102T150405Z")
	relBackupRoot := filepath.ToSlash(filepath.Join(".charter", "backups", ts))

	for _, plan := range plans {
		if !Fixable(plan.RuleID) {
			return written, backupDir, fmt.Errorf("fix: refusing to apply unregistered rule %q", plan.RuleID)
		}

		rel := filepath.FromSlash(plan.Path)
		target := filepath.Join(root, rel)
		if !withinRoot(root, target) {
			return written, backupDir, fmt.Errorf("fix: refusing to write outside root: %s", plan.Path)
		}

		switch plan.Action {
		case Create:
			if _, statErr := os.Stat(target); statErr == nil {
				// Target appeared since planning; never overwrite.
				continue
			}
			if writeErr := writeFile(target, plan.Contents); writeErr != nil {
				return written, backupDir, writeErr
			}
			written = append(written, plan.Path)

		case Append:
			// #nosec G304 -- target is a registered fixer's repo-relative plan path joined to root and root-bounded above.
			existing, readErr := os.ReadFile(target)
			switch {
			case readErr == nil:
				if backupErr := backupFile(root, ts, plan.Path, existing); backupErr != nil {
					return written, backupDir, backupErr
				}
				backupDir = relBackupRoot
			case os.IsNotExist(readErr):
				existing = nil // append to an absent file degrades to a create
			default:
				return written, backupDir, fmt.Errorf("fix: read %s: %w", plan.Path, readErr)
			}

			merged := make([]byte, 0, len(existing)+len(plan.Contents))
			merged = append(merged, existing...)
			merged = append(merged, plan.Contents...)
			if writeErr := writeFile(target, merged); writeErr != nil {
				return written, backupDir, writeErr
			}
			written = append(written, plan.Path)

		default:
			return written, backupDir, fmt.Errorf("fix: unknown action %d for %s", plan.Action, plan.Path)
		}
	}

	return written, backupDir, nil
}

// backupFile copies data to .charter/backups/<ts>/<path> under root, preserving
// the original's repo-relative layout so a restore is a straight copy back.
func backupFile(root, ts, relPath string, data []byte) error {
	dest := filepath.Join(root, ".charter", "backups", ts, filepath.FromSlash(relPath))
	return writeFile(dest, data)
}

// writeFile creates parents and writes data, mirroring the rest of the repo's
// world-readable file/dir modes (these are generated project files, not
// secrets).
func writeFile(target string, data []byte) error {
	// #nosec G301 -- 0o755 matches the surrounding repo tree; created dirs hold generated project files, not secrets.
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("fix: mkdir %s: %w", filepath.Dir(target), err)
	}
	// #nosec G306 G703 -- target is root-bounded by withinRoot before any call; 0o644 matches the repo tree (generated project files, not secrets).
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("fix: write %s: %w", target, err)
	}
	return nil
}

// withinRoot reports whether target resolves to a path at or below root,
// rejecting any `..` traversal smuggled through a plan path.
func withinRoot(root, target string) bool {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
