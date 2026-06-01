//go:build perf

// Package perf holds build-tagged performance validation for charter doctor.
// It is excluded from the default `:test` run and exercised only via the Moon
// `:perf` task (go test -tags=perf). The fixture is synthesized at test time so
// no large file set is ever committed.
package perf

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.use-charter.dev/charter/internal/doctor"
)

const (
	// dirCount * filesPerDir == 50,000 synthesized source files, spread across
	// nested directories to mirror a real repository layout.
	dirCount    = 500
	filesPerDir = 100
	sourceFiles = dirCount * filesPerDir

	// totalFiles also accounts for the root AGENTS.md and go.mod fixtures.
	totalFiles = sourceFiles + 2

	maxScanDuration = 2 * time.Second
	maxPeakRSSBytes = 256 * 1024 * 1024 // 256 MiB
)

// TestDoctorPerformance proves charter doctor scans a ~50,000-file repository
// within the wall-clock and (on Linux) peak-RSS budgets. Only doctor.Run is
// timed; fixture construction is deliberately excluded.
func TestDoctorPerformance(t *testing.T) {
	root := buildLargeRepo(t)

	start := time.Now()
	result, err := doctor.Run(root, 80, true)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("doctor.Run over %d-file repo: %v", totalFiles, err)
	}

	t.Logf(
		"scanned %d synthesized files in %s (score=%d, findings=%d)",
		totalFiles, elapsed, result.Score.Final, len(result.Findings),
	)

	if elapsed > maxScanDuration {
		t.Errorf("doctor.Run took %s, want <= %s", elapsed, maxScanDuration)
	}

	if runtime.GOOS == "linux" {
		peak, err := peakRSSBytes()
		if err != nil {
			t.Fatalf("read peak RSS: %v", err)
		}
		t.Logf("peak RSS: %d bytes (%.1f MiB)", peak, float64(peak)/(1024*1024))
		if peak > maxPeakRSSBytes {
			t.Errorf("peak RSS %d bytes exceeds budget %d bytes", peak, maxPeakRSSBytes)
		}
		return
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	t.Logf(
		"peak RSS assertion skipped on %s; runtime.MemStats proxy: Sys=%.1f MiB, HeapAlloc=%.1f MiB",
		runtime.GOOS, float64(mem.Sys)/(1024*1024), float64(mem.HeapAlloc)/(1024*1024),
	)
}

// buildLargeRepo synthesizes a git-backed repository of ~50,000 files in a temp
// directory. charter doctor resolves the root by walking to .git and builds its
// inventory via `git ls-files`, so the fixture is a real git repo with tracked
// files. Construction cost is intentionally outside the timed region.
func buildLargeRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	runGit(t, root, "init", "-q")

	writeFixtureFile(t, root, "AGENTS.md", agentsMarkdown)
	writeFixtureFile(t, root, "go.mod", "module example.com/perffixture\n\ngo 1.26\n")

	for dir := 0; dir < dirCount; dir++ {
		rel := filepath.Join("src", fmt.Sprintf("pkg%03d", dir))
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(abs, 0o755); err != nil {
			t.Fatalf("create fixture dir %s: %v", rel, err)
		}
		for file := 0; file < filesPerDir; file++ {
			name := fmt.Sprintf("file%03d.go", file)
			content := fmt.Sprintf("package pkg%03d\n\nvar Symbol%03d = %d\n", dir, file, file)
			if err := os.WriteFile(filepath.Join(abs, name), []byte(content), 0o644); err != nil {
				t.Fatalf("write fixture file %s/%s: %v", rel, name, err)
			}
		}
	}

	// Track the files so the inventory mirrors a real repository (git ls-files
	// --cached returns them, and AE-SEC-001's tracked-path query is exercised).
	runGit(t, root, "add", "-A")

	return root
}

func writeFixtureFile(t *testing.T, root, rel, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", rel, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
}

// peakRSSBytes reads VmHWM (peak resident set size, in kB) from
// /proc/self/status and returns it in bytes. Linux-only.
func peakRSSBytes() (uint64, error) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, fmt.Errorf("read /proc/self/status: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmHWM:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, fmt.Errorf("malformed VmHWM line: %q", line)
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse VmHWM value %q: %w", fields[1], err)
		}
		return kb * 1024, nil
	}

	return 0, fmt.Errorf("VmHWM not found in /proc/self/status")
}

// agentsMarkdown is a minimal but meaningful root context file so context rules
// have realistic content to evaluate during the scan.
const agentsMarkdown = `# AGENTS.md

## Project summary

This is a synthetic performance fixture. Charter is an offline-first Go CLI that
scores repositories for AI-agent readiness.

## Tech stack

- Go 1.26 module example.com/perffixture
- Deterministic, offline-only scanning

## Edit boundaries

- Off-limits: .git, generated state
- Safe for agent edits: src/

## Verification

- Verify with: charter doctor --path . --threshold 80
`
