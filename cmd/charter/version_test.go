package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/version"
)

func runVersion(t *testing.T, args ...string) string {
	t.Helper()
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(append([]string{"version"}, args...))
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version %v: %v", args, err)
	}
	return out.String()
}

func TestVersionCommand(t *testing.T) {
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"version"})
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version: %v", err)
	}
	got := out.String()
	for _, label := range []string{"charter", "commit", "built", "go", "platform"} {
		if !strings.Contains(got, label) {
			t.Fatalf("missing label %q in output:\n%s", label, got)
		}
	}
	if !strings.Contains(got, version.Version()) {
		t.Fatalf("expected version %q in output:\n%s", version.Version(), got)
	}
}

func TestVersionShort(t *testing.T) {
	got := strings.TrimSpace(runVersion(t, "--short"))
	if got != version.Version() {
		t.Fatalf("--short = %q, want %q", got, version.Version())
	}
	if strings.Contains(got, "commit") {
		t.Fatalf("--short should print only the version, got %q", got)
	}
}

func TestVersionJSON(t *testing.T) {
	var v struct {
		Version, Commit, Date, Go, Platform string
	}
	if err := json.Unmarshal([]byte(runVersion(t, "--format", "json")), &v); err != nil {
		t.Fatalf("version --format json is not valid JSON: %v", err)
	}
	if v.Version != version.Version() || v.Go == "" || v.Platform == "" {
		t.Fatalf("unexpected json version payload: %+v", v)
	}
}

func TestVersionBadFormat(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"version", "--format", "yaml"})
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error for an unknown --format")
	}
}
