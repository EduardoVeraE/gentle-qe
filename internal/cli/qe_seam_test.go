package cli

import (
	"os"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// TestMain flips the QE installer-flow seam OFF for the entire cli test package
// so the upstream flag-defaults / normalize tests see the unmodified upstream
// SDDMode default ("") without being edited. The QE build's single default is
// covered in production (seam ON) by the app parity tests, which assert the CLI
// install path matches the TUI (both single).
//
// This file is net-new fork overlay (registered in tools/qe-overlay/overlay.json).
func TestMain(m *testing.M) {
	model.QEInstallerFlow = false
	os.Exit(m.Run())
}
