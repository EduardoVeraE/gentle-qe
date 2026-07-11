package tui

import (
	"os"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// TestMain flips the QE installer-flow seam OFF for the entire tui test
// package, so the upstream dev-flow tests exercise the unmodified upstream
// behavior (pickerFlowSlice, Welcome navigation, model pickers, SDDMode, …)
// without being edited. The seam lives in the screens package (shared) — tui
// consumes screens' option lists, so flipping the single shared var here also
// disables the collapse/filter that screens applies. QE tests re-enable it via
// enableQESeam.
//
// This file is net-new fork overlay (registered in tools/qe-overlay/overlay.json).
func TestMain(m *testing.M) {
	model.QEInstallerFlow = false
	os.Exit(m.Run())
}

// enableQESeam turns the QE installer-flow seam ON for a single test and
// restores it afterwards. Tests using it MUST NOT call t.Parallel(): the seam
// is a package-global mutated here.
func enableQESeam(t *testing.T) {
	t.Helper()
	model.QEInstallerFlow = true
	t.Cleanup(func() { model.QEInstallerFlow = false })
}
