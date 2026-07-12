package cli

import (
	"os"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model/seamtest"
)

// TestMain flips the QE installer-flow seam OFF for the entire cli test package
// so the upstream flag-defaults / normalize tests see the unmodified upstream
// SDDMode default ("") without being edited. The QE build's single default is
// covered in production (seam ON) by the app parity tests, which assert the CLI
// install path matches the TUI (both single). Disable is the same centralized
// helper the tui and screens test packages use (see
// internal/model/seamtest/seamtest.go) so all three packages share one
// implementation of the seam-off pattern.
//
// This file is net-new fork overlay (registered in tools/qe-overlay/overlay.json).
func TestMain(m *testing.M) {
	seamtest.Disable()
	os.Exit(m.Run())
}
