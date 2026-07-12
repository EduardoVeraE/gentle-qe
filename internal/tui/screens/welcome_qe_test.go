package screens

import "testing"

// welcome_qe_test.go — unit tests for the Gentle-QE Welcome menu collapse
// (overlay Gentle-QE; ancla qe-overlay).

// qeWelcomeCollapsedCount mirrors the number of collapsed cases
// tui.qeWelcomeCanonicalCursor handles (7). Both sides are independently
// hardcoded to this value — see design.md's "structural fail-fast" invariant.
// Cross-package access from internal/tui/model_qe_test.go isn't possible
// (qeWelcomeOptions is unexported here), so the fail-fast lives in this file.
const qeWelcomeCollapsedCount = 7

// TestQEWelcomeOptions_CollapsesToSevenEssentials verifies the QE-essential
// keep-set, without profiles and with an available agent engine.
func TestQEWelcomeOptions_CollapsesToSevenEssentials(t *testing.T) {
	enableQESeam(t)
	full := WelcomeOptions(nil, true, false, 0, true)
	got := qeWelcomeOptions(full, false, true)

	want := []string{
		"Start installation",
		"Upgrade tools (up to date)",
		"Sync configs",
		"Upgrade + Sync",
		"Manage backups",
		"Managed uninstall",
		"Quit",
	}
	if len(got) != len(want) {
		t.Fatalf("qeWelcomeOptions() = %v (len %d), want %v (len %d)", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("qeWelcomeOptions()[%d] = %q, want %q; full got: %v", i, got[i], want[i], got)
		}
	}
}

// TestQEWelcomeOptions_ExcludesDevOnlyEntries triangulates with showProfiles
// true and hasEngines false (both dev-only levers flipped) and asserts the
// dev-only labels never survive the collapse regardless.
func TestQEWelcomeOptions_ExcludesDevOnlyEntries(t *testing.T) {
	enableQESeam(t)
	full := WelcomeOptions(nil, true, true, 3, false)
	got := qeWelcomeOptions(full, true, false)

	excluded := []string{
		"Configure models",
		"Create your own Agent",
		"Create your own Agent (no agents)",
		"OpenCode Community Plugins",
		"OpenCode SDD Profiles",
		"OpenCode SDD Profiles (3)",
		"Community Tools/Plugins",
	}
	for _, opt := range got {
		for _, bad := range excluded {
			if opt == bad {
				t.Fatalf("qeWelcomeOptions() = %v, must not contain dev-only entry %q", got, bad)
			}
		}
	}
}

// TestQEWelcomeOptions_StructuralFailFast is the fail-fast invariant from
// design.md: qeWelcomeOptions MUST always return exactly the number of
// collapsed cases qeWelcomeCanonicalCursor handles (7), regardless of
// showProfiles/hasEngines, so any future keep-set drift breaks loudly instead
// of silently mis-routing a Welcome menu action.
func TestQEWelcomeOptions_StructuralFailFast(t *testing.T) {
	enableQESeam(t)
	cases := []struct {
		showProfiles bool
		hasEngines   bool
	}{
		{false, false},
		{false, true},
		{true, false},
		{true, true},
	}
	for _, tc := range cases {
		full := WelcomeOptions(nil, true, tc.showProfiles, 1, tc.hasEngines)
		got := qeWelcomeOptions(full, tc.showProfiles, tc.hasEngines)
		if len(got) != qeWelcomeCollapsedCount {
			t.Fatalf("qeWelcomeOptions(showProfiles=%v, hasEngines=%v) len = %d, want %d; got: %v",
				tc.showProfiles, tc.hasEngines, len(got), qeWelcomeCollapsedCount, got)
		}
	}
}
