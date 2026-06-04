package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadTrackedFileReadsTrackedInRootFile(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "AGENTS.md"), "# Context\n")

	content, ok := ReadTrackedFile(root, New([]string{"AGENTS.md"}), "AGENTS.md")
	if !ok {
		t.Fatalf("expected a tracked in-root file to be read")
	}
	if content != "# Context\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestReadTrackedFileRejectsUntrackedPath(t *testing.T) {
	root := t.TempDir()
	// File exists on disk but is absent from the inventory.
	writeTestFile(t, filepath.Join(root, "AGENTS.md"), "# Context\n")

	if _, ok := ReadTrackedFile(root, New(nil), "AGENTS.md"); ok {
		t.Fatalf("expected an untracked path to be rejected")
	}
}

func TestReadTrackedFileRejectsSymlinkEscapingRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "credentials")
	writeTestFile(t, secret, "AWS_SECRET_ACCESS_KEY=leak-me\n")

	// A malicious repo tracks AGENTS.md as a symlink that resolves outside root.
	link := filepath.Join(root, "AGENTS.md")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	content, ok := ReadTrackedFile(root, New([]string{"AGENTS.md"}), "AGENTS.md")
	if ok {
		t.Fatalf("expected an out-of-root symlink to be rejected, read %q", content)
	}
}

func TestReadTrackedFileRejectsOversizeFile(t *testing.T) {
	root := t.TempDir()
	big := make([]byte, maxTrackedFileBytes+1)
	writeTestFile(t, filepath.Join(root, "big.txt"), string(big))

	if _, ok := ReadTrackedFile(root, New([]string{"big.txt"}), "big.txt"); ok {
		t.Fatalf("expected a file over the size cap to be rejected")
	}
}

func TestReadTrackedFileAcceptsSymlinkedRoot(t *testing.T) {
	// The root itself is reached via a symlink (mirrors macOS /tmp -> /private/tmp
	// and t.TempDir() under /var -> /private/var). Resolving the root means this
	// legitimate file is NOT falsely rejected as an escape.
	realRoot := t.TempDir()
	writeTestFile(t, filepath.Join(realRoot, "go.mod"), "module example.com/x\n")

	linkRoot := filepath.Join(t.TempDir(), "link-root")
	if err := os.Symlink(realRoot, linkRoot); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	content, ok := ReadTrackedFile(linkRoot, New([]string{"go.mod"}), "go.mod")
	if !ok {
		t.Fatalf("expected a file under a symlinked root to be read")
	}
	if content != "module example.com/x\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestReadTrackedFileRejectsDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AGENTS.md"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if _, ok := ReadTrackedFile(root, New([]string{"AGENTS.md"}), "AGENTS.md"); ok {
		t.Fatalf("expected a directory path to be rejected")
	}
}
