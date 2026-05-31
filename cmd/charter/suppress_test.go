package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runSuppress(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs(append([]string{"suppress"}, args...))
	cmd.SetContext(context.Background())
	err := cmd.Execute()
	return out.String(), err
}

func TestSuppressWritesEntry(t *testing.T) {
	repo := initTempRepo(t)
	out, err := runSuppress(t, "AE-MCP-001", "--path", repo, "--reason", "vendored", "--expires", "90d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "written .charter-suppress.yml") {
		t.Fatalf("expected write confirmation, got:\n%s", out)
	}
	data, rerr := os.ReadFile(filepath.Join(repo, ".charter-suppress.yml"))
	if rerr != nil {
		t.Fatalf("expected file written: %v", rerr)
	}
	s := string(data)
	if !strings.Contains(s, "AE-MCP-001") || !strings.Contains(s, "vendored") {
		t.Fatalf("file missing entry:\n%s", s)
	}
	// 90d must be stored as an absolute date, not the literal "90d".
	if strings.Contains(s, "90d") {
		t.Fatalf("expected an absolute expiry date, got:\n%s", s)
	}
}

func TestSuppressDryRunWritesNothing(t *testing.T) {
	repo := initTempRepo(t)
	out, err := runSuppress(t, "AE-MCP-001", "--path", repo, "--reason", "x", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "dry run") {
		t.Fatalf("expected dry-run notice, got:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(repo, ".charter-suppress.yml")); !os.IsNotExist(err) {
		t.Fatalf("dry run must not write the file")
	}
}

func TestSuppressRejectsBadRuleAndMissingReason(t *testing.T) {
	repo := initTempRepo(t)
	if _, err := runSuppress(t, "not-a-rule", "--path", repo, "--reason", "x"); err == nil {
		t.Fatal("expected invalid rule id error")
	}
	if _, err := runSuppress(t, "AE-MCP-001", "--path", repo); err == nil {
		t.Fatal("expected missing-reason error")
	}
}

func TestSuppressPermanentWarnsWithoutApprover(t *testing.T) {
	repo := initTempRepo(t)
	out, err := runSuppress(t, "AE-MCP-001", "--path", repo, "--reason", "x", "--expires", "permanent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "warning:") || !strings.Contains(out, "AE-SUPPRESS-002") {
		t.Fatalf("expected permanent-without-approver warning, got:\n%s", out)
	}
}
