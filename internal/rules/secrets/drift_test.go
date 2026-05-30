package secrets

import (
	"testing"

	"go.charter.dev/charter/internal/agentcontext"
)

// TestSEC001ScansEveryCanonicalContextFile guards the drift gap: every file the
// context rules recognize as agent context must also be scanned for secrets by
// AE-SEC-001. If a new context file is added to agentcontext.Files but not
// scanned here, a secret in it would silently slip through.
func TestSEC001ScansEveryCanonicalContextFile(t *testing.T) {
	scanned := make(map[string]bool, len(agentVisibleFileTargets))
	for _, target := range agentVisibleFileTargets {
		scanned[target] = true
	}

	for _, f := range agentcontext.Files {
		if !scanned[f] {
			t.Errorf("context file %q is recognized as agent context but not scanned by AE-SEC-001 (drift)", f)
		}
	}
}
