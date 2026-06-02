package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/rules/catalog"
)

// runExplain executes `charter explain <args...>` against a buffered root and
// returns stdout plus the command error (nil on success).
func runExplain(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(append([]string{"explain"}, args...))
	cmd.SetContext(context.Background())
	err := cmd.Execute()
	return out.String(), err
}

func exitCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error carrying an exit code, got nil")
	}
	var signal interface{ ExitCode() int }
	if !errors.As(err, &signal) {
		t.Fatalf("expected a commandExitError, got %T: %v", err, err)
	}
	return signal.ExitCode()
}

// TestExplainResolvesEveryRule proves the command resolves and renders every
// catalog rule in the default (plain, piped) text mode.
func TestExplainResolvesEveryRule(t *testing.T) {
	for _, id := range catalog.IDs() {
		entry, _ := catalog.Lookup(id)
		out, err := runExplain(t, id)
		if err != nil {
			t.Fatalf("explain %s: unexpected error: %v", id, err)
		}
		for _, want := range []string{entry.ID, entry.Name, entry.Category, entry.ShortDescription, entry.HelpURI} {
			if !strings.Contains(out, want) {
				t.Fatalf("explain %s missing %q\nfull output:\n%s", id, want, out)
			}
		}
		if strings.IndexByte(out, 0x1b) != -1 {
			t.Fatalf("explain %s (piped) must be ANSI-free, got: %q", id, out)
		}
	}
}

// TestExplainJSON asserts --format json emits the catalog.Entry verbatim.
func TestExplainJSON(t *testing.T) {
	out, err := runExplain(t, "AE-SEC-001", "--format", "json")
	if err != nil {
		t.Fatalf("explain --format json: %v", err)
	}
	var got catalog.Entry
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("explain --format json is not valid catalog.Entry JSON: %v\n%s", err, out)
	}
	want, _ := catalog.Lookup("AE-SEC-001")
	if got != want {
		t.Fatalf("explain json = %+v, want %+v", got, want)
	}
}

// TestExplainUnknownRule covers the unknown-rule guidance path (exit 2, lists
// valid IDs).
func TestExplainUnknownRule(t *testing.T) {
	_, err := runExplain(t, "AE-NOPE-999")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("unknown rule exit code = %d, want 2", code)
	}
	if !strings.Contains(err.Error(), "unknown rule") || !strings.Contains(err.Error(), "AE-SEC-001") {
		t.Fatalf("expected unknown-rule guidance listing valid IDs, got: %v", err)
	}
}

// TestExplainArgCount covers the no-arg and >1-arg usage errors (exit 2).
func TestExplainArgCount(t *testing.T) {
	for _, args := range [][]string{{}, {"AE-SEC-001", "AE-MCP-001"}} {
		_, err := runExplain(t, args...)
		if code := exitCodeOf(t, err); code != 2 {
			t.Fatalf("explain %v exit code = %d, want 2", args, code)
		}
	}
}

// TestExplainInvalidFormat and TestExplainInvalidColor cover the flag-validation
// usage errors.
func TestExplainInvalidFormat(t *testing.T) {
	_, err := runExplain(t, "AE-SEC-001", "--format", "yaml")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("invalid format exit code = %d, want 2", code)
	}
}

func TestExplainInvalidColor(t *testing.T) {
	_, err := runExplain(t, "AE-SEC-001", "--color", "rainbow")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("invalid color exit code = %d, want 2", code)
	}
}

// TestExplainInvalidColorJSON proves --color is validated even on the JSON
// path: an invalid --color must exit 2 regardless of --format, matching doctor.
func TestExplainInvalidColorJSON(t *testing.T) {
	_, err := runExplain(t, "AE-SEC-001", "--color", "bogus", "--format", "json")
	if code := exitCodeOf(t, err); code != 2 {
		t.Fatalf("explain --color bogus --format json exit code = %d, want 2", code)
	}
}

// TestExplainColorPrecedence covers --color always (forces styled even when
// piped), --no-color (plain), and --no-color winning over --color.
func TestExplainColorPrecedence(t *testing.T) {
	always, err := runExplain(t, "AE-SEC-001", "--color", "always")
	if err != nil {
		t.Fatalf("explain --color always: %v", err)
	}
	if strings.IndexByte(always, 0x1b) == -1 {
		t.Fatalf("--color always must force styled output (ANSI) even when piped, got: %q", always)
	}

	off, err := runExplain(t, "AE-SEC-001", "--no-color")
	if err != nil {
		t.Fatalf("explain --no-color: %v", err)
	}
	if strings.IndexByte(off, 0x1b) != -1 {
		t.Fatalf("--no-color must produce zero ANSI, got: %q", off)
	}

	conflict, err := runExplain(t, "AE-SEC-001", "--color", "always", "--no-color")
	if err != nil {
		t.Fatalf("explain --color always --no-color: %v", err)
	}
	if strings.IndexByte(conflict, 0x1b) != -1 {
		t.Fatalf("--no-color must win over --color=always (zero ANSI), got: %q", conflict)
	}
}

// TestExplainNoColorEnvHonored confirms NO_COLOR keeps the default (auto) path
// plain.
func TestExplainNoColorEnvHonored(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out, err := runExplain(t, "AE-SEC-001")
	if err != nil {
		t.Fatalf("explain with NO_COLOR: %v", err)
	}
	if strings.IndexByte(out, 0x1b) != -1 {
		t.Fatalf("NO_COLOR must keep output plain, got: %q", out)
	}
}
