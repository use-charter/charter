package main

import (
	"bytes"
	"context"
	"testing"
)

func TestDoctorCommandRuns(t *testing.T) {
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"doctor", "--path", "../..", "--threshold", "80", "--quiet"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected command to fail until doctor runner is implemented")
	}

	if err.Error() != "doctor runner not implemented" {
		t.Fatalf("expected placeholder doctor error, got %q", err.Error())
	}
}

func TestDoctorCommandHelpRuns(t *testing.T) {
	cmd := newRootCommand()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"doctor", "--help"})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected doctor help to run without error: %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("Scan a repository and compute a Charter score")) {
		t.Fatalf("expected help output to include doctor command description")
	}
}
