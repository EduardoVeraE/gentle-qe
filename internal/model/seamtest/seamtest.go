// Package seamtest centralizes mutation of the Gentle-QE installer-flow
// testing seam (model.QEInstallerFlow) so the tui, screens and cli test
// packages share ONE implementation of the enable/disable/t.Cleanup pattern
// instead of each hand-rolling its own copy (previously duplicated verbatim
// in internal/tui/qe_seam_test.go and internal/tui/screens/qe_seam_test.go).
//
// It lives in its own package — not package model — specifically so package
// model's PRODUCTION code never imports "testing". This package is only ever
// imported from _test.go files across the repo, so importing "testing" here
// does not reach the production binary: go build only links packages that
// are part of the non-test dependency graph, and nothing in that graph
// imports seamtest.
//
// This file is net-new fork overlay (registered in tools/qe-overlay/overlay.json).
package seamtest

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// Disable flips the QE installer-flow seam OFF. Call it from a package's
// TestMain (no *testing.T is available there) so upstream dev-flow / parity
// tests see the unmodified upstream behavior by default; QE tests opt back
// in per-test via Enable.
func Disable() {
	model.QEInstallerFlow = false
}

// Enable flips the QE installer-flow seam ON for the duration of t and
// restores it via t.Cleanup once t finishes. This is the ONLY sanctioned way
// to turn the seam on in a test: pairing the mutation with its cleanup in one
// place means a test can never forget to restore it and leak the seam ON
// into the rest of the package's test run.
//
// Tests calling Enable MUST NOT call t.Parallel(): the seam is a
// package-global (model.QEInstallerFlow) mutated here.
func Enable(t *testing.T) {
	t.Helper()
	model.QEInstallerFlow = true
	t.Cleanup(func() { model.QEInstallerFlow = false })
}
