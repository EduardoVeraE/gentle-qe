package model

import "testing"

// TestQEFlowDefaultIsOnInProduction guards that the fork binary ships with the
// QE installer flow UNCONDITIONALLY enabled. qeFlowDefault is an immutable const
// (unaffected by any test package's TestMain flip of the QEInstallerFlow var),
// so this asserts the production default itself. If someone flips it, the QE
// simplified installer would silently stop applying for real users.
func TestQEFlowDefaultIsOnInProduction(t *testing.T) {
	if !qeFlowDefault {
		t.Fatal("qeFlowDefault must be true: the fork ships the QE installer flow unconditionally")
	}
}
