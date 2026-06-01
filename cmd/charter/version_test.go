package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/version"
)

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
