package cli

import (
	"github.com/gentleman-programming/gentle-ai/internal/model/seamtest"
)

// init flips the QE installer-flow seam OFF for the entire cli test package
// so the upstream flag-defaults / normalize tests see the unmodified upstream
// SDDMode default ("") without being edited. The QE build's single default is
// covered in production (seam ON) by the app parity tests, which assert the CLI
// install path matches the TUI (both single). Disable is the same centralized
// helper the tui and screens test packages use (see
// internal/model/seamtest/seamtest.go) so all three packages share one
// implementation of the seam-off pattern.
//
// Package `cli` also has an upstream TestMain (protocol_probe_test.go) that
// installs hermetic engram fakes; Go only allows one TestMain per package, so
// this file uses init() instead — init() runs before TestMain/m.Run() either
// way, so the seam is still off before any test in the package executes.
//
// This file is net-new fork overlay (registered in tools/qe-overlay/overlay.json).
func init() {
	seamtest.Disable()
}
