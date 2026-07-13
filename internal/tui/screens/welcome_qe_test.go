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

// ─── QEWelcomeFullOptions ────────────────────────────────────────────────────
//
// QEWelcomeFullOptions is the single source of truth
// tui.qeWelcomeCanonicalCursor's reverse-label-lookup relies on (see
// model_qe.go). These tests guard the two properties that lookup depends on:
// (1) it always returns the FULL, uncollapsed upstream menu — even with the
// QE seam ON, where WelcomeOptions() itself returns the collapsed list —
// and (2) its order/content matches upstream's real construction exactly,
// so a future upstream sync that reorders/inserts/renames a Welcome entry
// changes what THIS test observes too (structural fail-fast: it fails loudly
// here, at the seam, rather than silently mis-routing a keypress downstream).

// TestQEWelcomeFullOptions_MatchesUpstreamUncollapsedOrder pins the exact
// upstream Welcome menu (labels + order) for a known input set. If upstream
// reorders, renames, inserts, or removes a menu entry on the next sync, this
// test fails immediately and names exactly what changed — the early,
// noisy-by-design detection the fork's fragility mitigation calls for.
func TestQEWelcomeFullOptions_MatchesUpstreamUncollapsedOrder(t *testing.T) {
	got := QEWelcomeFullOptions(nil, true, false, 0, true)

	want := []string{
		"Start installation",
		"Upgrade tools (up to date)",
		"Sync configs",
		"Upgrade + Sync",
		"Configure models",
		"Create your own Agent",
		"OpenCode Community Plugins",
		"Uninstall OpenCode Plugin",
		"Manage backups",
		"Managed uninstall",
		"Community Tools/Plugins",
		"Quit",
	}
	if len(got) != len(want) {
		t.Fatalf("QEWelcomeFullOptions() = %v (len %d), want %v (len %d) — upstream Welcome menu structure changed; update qeWelcomeOptions' keep-set if needed", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("QEWelcomeFullOptions()[%d] = %q, want %q; full got: %v", i, got[i], want[i], got)
		}
	}
}

// TestQEWelcomeFullOptions_ReturnsFullEvenWhenSeamCollapsesWelcomeOptions is
// the property the reverse-lookup depends on most directly: with the QE seam
// ON, WelcomeOptions() itself returns only the 7 collapsed entries, but
// QEWelcomeFullOptions must still return every upstream entry.
func TestQEWelcomeFullOptions_ReturnsFullEvenWhenSeamCollapsesWelcomeOptions(t *testing.T) {
	enableQESeam(t)

	collapsed := WelcomeOptions(nil, true, false, 0, true)
	if len(collapsed) != qeWelcomeCollapsedCount {
		t.Fatalf("preconditions: WelcomeOptions() with seam ON = %v (len %d), want the collapsed 7-entry menu", collapsed, len(collapsed))
	}

	full := QEWelcomeFullOptions(nil, true, false, 0, true)
	if len(full) <= len(collapsed) {
		t.Fatalf("QEWelcomeFullOptions() len = %d, want > collapsed WelcomeOptions() len %d", len(full), len(collapsed))
	}

	const devOnlyEntry = "Configure models"
	if qeStringInSlice(collapsed, devOnlyEntry) {
		t.Fatalf("preconditions: collapsed WelcomeOptions() = %v, must not contain dev-only entry %q", collapsed, devOnlyEntry)
	}
	if !qeStringInSlice(full, devOnlyEntry) {
		t.Fatalf("QEWelcomeFullOptions() = %v, missing upstream dev-only entry %q that qeWelcomeCanonicalCursor's reverse lookup needs to see", full, devOnlyEntry)
	}
}

// TestQEWelcomeFullOptions_InsertsProfilesEntryWhenShown covers the
// non-uniform-gap case (the original arithmetic bug's root cause): the
// "OpenCode SDD Profiles" entry only exists in the upstream list when
// showProfiles is true, shifting every subsequent index by 1.
func TestQEWelcomeFullOptions_InsertsProfilesEntryWhenShown(t *testing.T) {
	without := QEWelcomeFullOptions(nil, true, false, 0, true)
	with := QEWelcomeFullOptions(nil, true, true, 2, true)

	if len(with) != len(without)+1 {
		t.Fatalf("QEWelcomeFullOptions(showProfiles=true) len = %d, want %d (without) + 1", len(with), len(without))
	}
	const profilesLabel = "OpenCode SDD Profiles (2)"
	if !qeStringInSlice(with, profilesLabel) {
		t.Fatalf("QEWelcomeFullOptions(showProfiles=true) = %v, missing %q", with, profilesLabel)
	}

	backupsWithout, _ := qeIndexOf(without, "Manage backups")
	backupsWith, _ := qeIndexOf(with, "Manage backups")
	if backupsWith != backupsWithout+1 {
		t.Fatalf("\"Manage backups\" index = %d with profiles shown, want %d (shifted by the inserted Profiles entry)", backupsWith, backupsWithout+1)
	}
}

func qeStringInSlice(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

func qeIndexOf(s []string, target string) (int, bool) {
	for i, v := range s {
		if v == target {
			return i, true
		}
	}
	return -1, false
}
