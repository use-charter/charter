package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The unit tests in this package drive newRootCommand().Execute() directly,
// which bypasses fang entirely. fang is only wired in main.go's execute(), so
// the only way to prove the fang-wrapped entrypoint preserves Charter's
// exit-code + silent-error contract is to build the real binary and run it as a
// subprocess. These tests are hermetic: they reuse the populated build/module
// cache (no network) and a throwaway git repo under t.TempDir().

// testVersion is injected into the built binary via -ldflags so the version
// assertions are deterministic. The binary's own build-info version is a VCS
// pseudo-version that differs from what version.Version() returns inside the
// test binary, so comparing against a known sentinel is the only stable check —
// and it doubles as proof that fang's WithVersion is fed from internal/version.
const testVersion = "v0.0.0-fangtest"

// buildCharterBinary compiles the cmd/charter main package (where fang is
// wired) into a temp dir and returns its path. `go test` runs with the package
// source directory as the working directory, so "." is cmd/charter.
func buildCharterBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "charter")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	ldflags := "-X go.use-charter.dev/charter/internal/version.injected=" + testVersion
	build := exec.Command("go", "build", "-ldflags", ldflags, "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build charter binary: %v\n%s", err, out)
	}
	return bin
}

// pipedEnv inherits the parent environment but strips every color-forcing (or
// color-suppressing) variable, so the ONLY color signal left for the binary is
// the captured pipe — which is never a TTY. That makes "piped output is plain"
// assertions deterministic across local shells and CI without leaning on
// NO_COLOR. Pass extra "KEY=VALUE" entries to layer on top (e.g. NO_COLOR=1).
func pipedEnv(extra ...string) []string {
	env := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		switch key {
		case "NO_COLOR", "CLICOLOR", "CLICOLOR_FORCE", "FORCE_COLOR", "COLORTERM":
			continue
		default:
			env = append(env, kv)
		}
	}
	return append(env, extra...)
}

type charterRun struct {
	stdout   string
	stderr   string
	exitCode int
}

// runCharter executes the built binary with stdin detached and stdout/stderr
// captured into buffers (i.e. genuinely piped, never a TTY).
func runCharter(t *testing.T, bin string, env []string, args ...string) charterRun {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("run charter %v: %v", args, err)
		}
	}
	return charterRun{stdout: stdout.String(), stderr: stderr.String(), exitCode: code}
}

// newScoredRepo builds a throwaway git repo: a blank Go module that scores
// below 100. That makes `charter doctor --threshold 0` pass (any score clears
// 0) and `--threshold 100` fail, giving the exit-0 and silent-exit-1 cases a
// deterministic fixture.
func newScoredRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.email", "charter@example.com")
	runGit(t, repo, "config", "user.name", "Charter Test")
	runGit(t, repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/demo\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	runGit(t, repo, "add", "-A")
	return repo
}

func containsANSI(s string) bool {
	return strings.ContainsRune(s, '\x1b')
}

// TestFangBinaryExitContract asserts that the fang-wrapped binary keeps the four
// load-bearing exit-code cases intact, that --version works without breaking the
// version subcommand, and that styled help/error output is plain (zero ANSI
// escapes) when piped.
func TestFangBinaryExitContract(t *testing.T) {
	bin := buildCharterBinary(t)
	repo := newScoredRepo(t)

	t.Run("doctor passing exits 0 with no error box", func(t *testing.T) {
		got := runCharter(t, bin, pipedEnv(), "doctor", "--path", repo, "--threshold", "0")
		if got.exitCode != 0 {
			t.Fatalf("exit = %d, want 0; stderr:\n%s", got.exitCode, got.stderr)
		}
		if got.stderr != "" {
			t.Fatalf("expected empty stderr on pass, got:\n%s", got.stderr)
		}
		if strings.TrimSpace(got.stdout) == "" {
			t.Fatal("expected a report on stdout for a passing doctor run")
		}
	})

	t.Run("doctor below threshold exits 1 with NO error box", func(t *testing.T) {
		got := runCharter(t, bin, pipedEnv(), "doctor", "--path", repo, "--threshold", "100")
		if got.exitCode != 1 {
			t.Fatalf("exit = %d, want 1; stderr:\n%s", got.exitCode, got.stderr)
		}
		// The silent FAIL contract: the report already shows FAIL on stdout, so
		// fang must print NOTHING to stderr (no styled error box).
		if got.stderr != "" {
			t.Fatalf("silent FAIL must produce empty stderr, got:\n%q", got.stderr)
		}
	})

	t.Run("invalid usage exits 2 with a message", func(t *testing.T) {
		got := runCharter(t, bin, pipedEnv(), "doctor", "--format", "bogus")
		if got.exitCode != 2 {
			t.Fatalf("exit = %d, want 2; stderr:\n%s", got.exitCode, got.stderr)
		}
		// fang's styled handler capitalizes the first letter and appends a
		// period, so match case-insensitively.
		if !strings.Contains(strings.ToLower(got.stderr), "invalid format") {
			t.Fatalf("expected the invalid-format message on stderr, got:\n%q", got.stderr)
		}
		if containsANSI(got.stderr) {
			t.Fatalf("piped error output must contain no ANSI escapes, got:\n%q", got.stderr)
		}
	})

	t.Run("unknown command exits 2 via the non-exitSignal path", func(t *testing.T) {
		got := runCharter(t, bin, pipedEnv(), "frobnicate")
		if got.exitCode != 2 {
			t.Fatalf("exit = %d, want 2; stderr:\n%s", got.exitCode, got.stderr)
		}
		if !strings.Contains(strings.ToLower(got.stderr), "unknown command") {
			t.Fatalf("expected an unknown-command error on stderr, got:\n%q", got.stderr)
		}
		if containsANSI(got.stderr) {
			t.Fatalf("piped error output must contain no ANSI escapes, got:\n%q", got.stderr)
		}
	})

	t.Run("--version exits 0 and reports the version", func(t *testing.T) {
		got := runCharter(t, bin, pipedEnv(), "--version")
		if got.exitCode != 0 {
			t.Fatalf("exit = %d, want 0; stderr:\n%s", got.exitCode, got.stderr)
		}
		if !strings.Contains(got.stdout, testVersion) {
			t.Fatalf("expected injected version %q in --version output, got:\n%q", testVersion, got.stdout)
		}
	})

	t.Run("version subcommand still works", func(t *testing.T) {
		got := runCharter(t, bin, pipedEnv(), "version")
		if got.exitCode != 0 {
			t.Fatalf("exit = %d, want 0; stderr:\n%s", got.exitCode, got.stderr)
		}
		for _, label := range []string{"charter", "commit", "built", "go", "platform"} {
			if !strings.Contains(got.stdout, label) {
				t.Fatalf("version subcommand output missing %q:\n%s", label, got.stdout)
			}
		}
		if !strings.Contains(got.stdout, testVersion) {
			t.Fatalf("version subcommand output missing injected version %q:\n%s", testVersion, got.stdout)
		}
	})

	t.Run("piped help is plain (no ANSI)", func(t *testing.T) {
		// No NO_COLOR here on purpose: the only color signal is the non-TTY
		// pipe, so this proves fang downgrades color on a piped fd.
		got := runCharter(t, bin, pipedEnv(), "--help")
		if got.exitCode != 0 {
			t.Fatalf("exit = %d, want 0; stderr:\n%s", got.exitCode, got.stderr)
		}
		if containsANSI(got.stdout) {
			t.Fatalf("piped help must contain no ANSI escapes, got:\n%q", got.stdout)
		}
	})

	t.Run("NO_COLOR keeps help plain", func(t *testing.T) {
		got := runCharter(t, bin, pipedEnv("NO_COLOR=1"), "--help")
		if got.exitCode != 0 {
			t.Fatalf("exit = %d, want 0; stderr:\n%s", got.exitCode, got.stderr)
		}
		if containsANSI(got.stdout) {
			t.Fatalf("NO_COLOR help must contain no ANSI escapes, got:\n%q", got.stdout)
		}
	})
}
