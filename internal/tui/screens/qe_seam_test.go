package screens

import (
	"os"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model/seamtest"
)

// TestMain flips the QE installer-flow seam OFF for the entire screens test
// package, so upstream tests (e.g. persona_language_contract_test.go) see the
// unmodified upstream option lists without being edited. QE tests re-enable the
// seam locally via seamtest.Enable.
//
// This file is net-new fork overlay (registered in tools/qe-overlay/overlay.json).
func TestMain(m *testing.M) {
	seamtest.Disable()
	os.Exit(m.Run())
}

// enableQESeam turns the QE installer-flow seam ON for a single test and
// restores it afterwards. Tests using it MUST NOT call t.Parallel(): the seam
// is a package-global mutated here. Thin wrapper kept so existing call sites
// (enableQESeam(t)) don't need touching; the actual enable/cleanup pairing is
// centralized in seamtest.Enable (see internal/model/seamtest/seamtest.go).
func enableQESeam(t *testing.T) {
	t.Helper()
	seamtest.Enable(t)
}
