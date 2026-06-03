package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"go.use-charter.dev/charter/internal/doctor"
	renderjson "go.use-charter.dev/charter/internal/render/json"
	rendermarkdown "go.use-charter.dev/charter/internal/render/markdown"
)

// runReport executes `charter report` with args against a fresh root command,
// returning captured stdout/stderr plus the run error.
func runReport(t *testing.T, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	t.Helper()
	cmd := newRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(append([]string{"report"}, args...))
	cmd.SetContext(context.Background())
	return stdout, stderr, cmd.Execute()
}

// TestReportDefaultFormatWritesHTML proves the default --format is html and the
// default --out is charter-report.html in the working directory.
func TestReportDefaultFormatWritesHTML(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	// This test exercises the DEFAULT --out behavior: with no --out, the report
	// must land at charter-report.html relative to the cwd, so t.Chdir isolates a
	// temp cwd to assert that exact filename (and keep the artifact out of the
	// repo). makeTempDoctorRepo resolves its source via a cwd-relative path, so
	// chdir only AFTER the repo exists. t.Chdir also forbids t.Parallel().
	t.Chdir(t.TempDir())

	_, stderr, runErr := runReport(t, "--path", repo)
	if runErr != nil {
		t.Fatalf("expected exit 0, got %v", runErr)
	}

	data, rerr := os.ReadFile("charter-report.html")
	if rerr != nil {
		t.Fatalf("expected charter-report.html written: %v", rerr)
	}
	for _, marker := range []string{"data-charter-report", "<style>", "<script>"} {
		if !bytes.Contains(data, []byte(marker)) {
			t.Fatalf("expected HTML report to contain %q, got:\n%s", marker, data)
		}
	}
	if !bytes.Contains(stderr.Bytes(), []byte("charter: wrote charter-report.html")) {
		t.Fatalf("expected wrote-path note on stderr, got %q", stderr.String())
	}
}

// TestReportDefaultOutNames covers the per-format default filenames.
func TestReportDefaultOutNames(t *testing.T) {
	cases := []struct {
		format string
		file   string
	}{
		{format: "html", file: "charter-report.html"},
		{format: "markdown", file: "charter-report.md"},
		{format: "json", file: "charter-report.json"},
	}
	for _, tc := range cases {
		t.Run(tc.format, func(t *testing.T) {
			repo, err := makeTempDoctorRepo(t)
			if err != nil {
				t.Fatalf("fixture: %v", err)
			}
			// The default --out filename is resolved relative to the cwd, so
			// isolate a temp cwd to assert the per-format default lands there.
			// (t.Chdir genuinely tests the default-cwd behavior and forbids
			// t.Parallel(); explicit --out is covered by TestReportOutHonored.)
			t.Chdir(t.TempDir())

			_, _, runErr := runReport(t, "--path", repo, "--format", tc.format)
			if runErr != nil {
				t.Fatalf("expected exit 0, got %v", runErr)
			}
			if _, err := os.Stat(tc.file); err != nil {
				t.Fatalf("expected %s written: %v", tc.file, err)
			}
		})
	}
}

// TestReportOutHonored proves an explicit --out path overrides the default.
func TestReportOutHonored(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "custom-report.json")

	stdout, _, runErr := runReport(t, "--path", repo, "--format", "json", "--out", outPath)
	if runErr != nil {
		t.Fatalf("expected exit 0, got %v", runErr)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected nothing on stdout when --out is set, got %q", stdout.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected %s written: %v", outPath, err)
	}
}

// TestReportInvalidFormat proves an unknown --format is a usage error (exit 2).
func TestReportInvalidFormat(t *testing.T) {
	_, _, runErr := runReport(t, "--format", "yaml")
	if runErr == nil {
		t.Fatalf("expected invalid format error")
	}
	var signal interface{ ExitCode() int }
	if !errors.As(runErr, &signal) {
		t.Fatalf("expected command exit error, got %T", runErr)
	}
	if signal.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", signal.ExitCode())
	}
}

// TestReportMarkdownByteMatch proves the markdown report is the existing
// render/markdown output for the same scan (with emit's trailing-newline
// normalization), i.e. the renderer is reused, not reimplemented.
func TestReportMarkdownByteMatch(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "report.md")

	_, _, runErr := runReport(t, "--path", repo, "--format", "markdown", "--out", outPath)
	if runErr != nil {
		t.Fatalf("expected exit 0, got %v", runErr)
	}

	result, err := doctor.Run(repo, 80, false)
	if err != nil {
		t.Fatalf("doctor.Run: %v", err)
	}
	expected, err := rendermarkdown.Render(result)
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}

	got, rerr := os.ReadFile(outPath)
	if rerr != nil {
		t.Fatalf("read report: %v", rerr)
	}
	if !bytes.Equal(got, ensureTrailingNewline(expected)) {
		t.Fatalf("markdown report does not match render/markdown output\n--- got ---\n%s\n--- want ---\n%s", got, expected)
	}
}

// TestReportJSONByteMatch is the JSON analogue of the markdown byte-match.
func TestReportJSONByteMatch(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "report.json")

	_, _, runErr := runReport(t, "--path", repo, "--format", "json", "--out", outPath)
	if runErr != nil {
		t.Fatalf("expected exit 0, got %v", runErr)
	}

	result, err := doctor.Run(repo, 80, false)
	if err != nil {
		t.Fatalf("doctor.Run: %v", err)
	}
	expected, err := renderjson.Render(result)
	if err != nil {
		t.Fatalf("render json: %v", err)
	}

	got, rerr := os.ReadFile(outPath)
	if rerr != nil {
		t.Fatalf("read report: %v", rerr)
	}
	if !bytes.Equal(got, ensureTrailingNewline(expected)) {
		t.Fatalf("json report does not match render/json output\n--- got ---\n%s\n--- want ---\n%s", got, expected)
	}
}

// TestReportExitsZeroBelowThreshold proves report is a generator, not a gate:
// a below-threshold repo still exits 0 and writes the artifact.
func TestReportExitsZeroBelowThreshold(t *testing.T) {
	repo := initTempRepo(t) // README-only repo scores below the default threshold
	outPath := filepath.Join(t.TempDir(), "report.html")

	_, _, runErr := runReport(t, "--path", repo, "--threshold", "80", "--out", outPath)
	if runErr != nil {
		t.Fatalf("expected exit 0 even below threshold, got %v", runErr)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected %s written: %v", outPath, err)
	}
}

// TestReportOpenIsBestEffort stubs the opener so the suite never spawns a
// browser, and proves --open neither changes the exit code nor the artifact.
func TestReportOpenIsBestEffort(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "report.html")

	var opened string
	original := openInBrowser
	openInBrowser = func(path string) error {
		opened = path
		return nil
	}
	t.Cleanup(func() { openInBrowser = original })

	_, _, runErr := runReport(t, "--path", repo, "--out", outPath, "--open")
	if runErr != nil {
		t.Fatalf("expected exit 0 with --open, got %v", runErr)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected %s written: %v", outPath, err)
	}
	if opened != outPath {
		t.Fatalf("expected opener invoked with %q, got %q", outPath, opened)
	}
}

// TestReportOpenFailureStillExitsZero proves a failing opener is downgraded to a
// note: the artifact was already written, so the command still succeeds.
func TestReportOpenFailureStillExitsZero(t *testing.T) {
	repo, err := makeTempDoctorRepo(t)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "report.html")

	original := openInBrowser
	openInBrowser = func(string) error { return errors.New("no opener available") }
	t.Cleanup(func() { openInBrowser = original })

	_, stderr, runErr := runReport(t, "--path", repo, "--out", outPath, "--open")
	if runErr != nil {
		t.Fatalf("expected exit 0 even when opener fails, got %v", runErr)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("could not open")) {
		t.Fatalf("expected opener-failure note on stderr, got %q", stderr.String())
	}
}

// TestReportHelpDocumentsNonGate proves the help text states report never gates.
func TestReportHelpDocumentsNonGate(t *testing.T) {
	stdout, _, runErr := runReport(t, "--help")
	if runErr != nil {
		t.Fatalf("expected report help to run, got %v", runErr)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("not a CI gate")) &&
		!bytes.Contains(stdout.Bytes(), []byte("not a gate")) {
		t.Fatalf("expected help to document the non-gate exit-0 contract, got:\n%s", stdout.String())
	}
}

// ensureTrailingNewline mirrors emit's normalization: exactly one trailing
// newline, matching what report writes to disk.
func ensureTrailingNewline(b []byte) []byte {
	if bytes.HasSuffix(b, []byte{'\n'}) {
		return b
	}
	return append(b, '\n')
}
