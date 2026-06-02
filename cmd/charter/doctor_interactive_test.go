package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

// exitCodeOf is shared with explain_test.go (same package): it extracts the
// exit code carried by a commandExitError.

// TestDoctorInteractiveRequiresTTY proves the containment gate: with a non-TTY
// stdout (a bytes.Buffer, exactly as under `go test` or a pipe) `doctor -i`
// must exit 2 with a clear message and must NOT launch the program. The test
// returning at all (rather than hanging on terminal input) is itself the
// no-hang assertion.
func TestDoctorInteractiveRequiresTTY(t *testing.T) {
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"doctor", "--path", t.TempDir(), "-i"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if got := exitCodeOf(t, err); got != 2 {
		t.Fatalf("exit code = %d, want 2", got)
	}
	if !strings.Contains(err.Error(), "requires a terminal") {
		t.Fatalf("error = %q, want it to mention a terminal requirement", err.Error())
	}
	if out.Len() != 0 {
		t.Fatalf("interactive gate must not write to stdout, got %q", out.String())
	}
}

// TestDoctorInteractiveThresholdComposes proves --threshold is compatible with
// -i (it is not in the rejected-flags set) and is carried into runInteractive:
// the combo clears the flag-conflict checks and stops only at the TTY gate, so
// a non-TTY run exits 2 with the terminal-requirement message (not a flag
// rejection). It guards against --threshold being accidentally rejected.
func TestDoctorInteractiveThresholdComposes(t *testing.T) {
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"doctor", "--path", t.TempDir(), "-i", "--threshold", "90"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if got := exitCodeOf(t, err); got != 2 {
		t.Fatalf("exit code = %d, want 2", got)
	}
	if !strings.Contains(err.Error(), "requires a terminal") {
		t.Fatalf("error = %q, want the TTY-gate message (not a flag rejection)", err.Error())
	}
	if out.Len() != 0 {
		t.Fatalf("interactive gate must not write to stdout, got %q", out.String())
	}
}

// TestDoctorInteractiveRejectsIncompatibleFlags proves -i is rejected (exit 2)
// when combined with the machine-readable / headless modes, before any scan or
// program launch.
func TestDoctorInteractiveRejectsIncompatibleFlags(t *testing.T) {
	repo := t.TempDir()
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"json", []string{"doctor", "--path", repo, "-i", "--format", "json"}, "--format"},
		{"sarif", []string{"doctor", "--path", repo, "-i", "--format", "sarif"}, "--format"},
		{"markdown", []string{"doctor", "--path", repo, "-i", "--format", "markdown"}, "--format"},
		{"quiet", []string{"doctor", "--path", repo, "-i", "--quiet"}, "--quiet"},
		{"out", []string{"doctor", "--path", repo, "-i", "--out", filepath.Join(t.TempDir(), "r.txt")}, "--out"},
		{"rule", []string{"doctor", "--path", repo, "-i", "--rule", "AE-SEC-001"}, "--rule"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newRootCommand()
			out := new(bytes.Buffer)
			cmd.SetOut(out)
			cmd.SetErr(new(bytes.Buffer))
			cmd.SetArgs(tc.args)
			cmd.SetContext(context.Background())

			err := cmd.Execute()
			if got := exitCodeOf(t, err); got != 2 {
				t.Fatalf("exit code = %d, want 2", got)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want it to mention %q", err.Error(), tc.want)
			}
			if out.Len() != 0 {
				t.Fatalf("rejected interactive combo must not write to stdout, got %q", out.String())
			}
		})
	}
}

// TestDoctorNonInteractiveUnchanged is a guard that the default (no -i) text
// path still runs and writes the historical output when stdout is not a TTY.
func TestDoctorNonInteractiveUnchanged(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"doctor", "--path", repo, "--threshold", "80"})
	cmd.SetContext(context.Background())

	_ = cmd.Execute() // pass/fail is fine; we only assert the plain contract holds.

	if !bytes.Contains(out.Bytes(), []byte("charter doctor: ")) {
		t.Fatalf("default text path changed; got:\n%s", out.String())
	}
	if i := bytes.IndexByte(out.Bytes(), 0x1b); i != -1 {
		t.Fatalf("non-TTY text output must contain zero ANSI escape bytes, found at %d", i)
	}
}
