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
}
