package version

import "testing"

func TestVersionNonEmpty(t *testing.T) {
	if Version() == "" {
		t.Fatal("Version() must never be empty")
	}
}

func TestVersionDefaultsInTestBuild(t *testing.T) {
	// A plain `go test` build injects no ldflags and reports module version
	// "(devel)", so Version() falls through to the development placeholder.
	if got := Version(); got != "0.0.0-dev" {
		t.Fatalf("expected the development default in a test build, got %q", got)
	}
}

func TestCommitAndDateNeverEmpty(t *testing.T) {
	if Commit() == "" {
		t.Fatal("Commit() must never be empty")
	}
	if Date() == "" {
		t.Fatal("Date() must never be empty")
	}
}
