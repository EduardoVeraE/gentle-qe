package tui

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

// ─── qeWelcomeCanonicalCursor ────────────────────────────────────────────────

// TestQEWelcomeCanonicalCursor_RemapTable drives all 7 collapsed cursor
// positions through both hasDetectedOpenCode variants and asserts the
// canonical index produced matches the two-gap remap table in design.md.
func TestQEWelcomeCanonicalCursor_RemapTable(t *testing.T) {
	enableQESeam(t)
	tests := []struct {
		name          string
		collapsed     int
		hasOpenCode   bool
		wantCanonical int
	}{
		{"Start installation, no OpenCode", 0, false, 0},
		{"Start installation, OpenCode detected", 0, true, 0},
		{"Upgrade tools, no OpenCode", 1, false, 1},
		{"Upgrade tools, OpenCode detected", 1, true, 1},
		{"Sync configs, no OpenCode", 2, false, 2},
		{"Sync configs, OpenCode detected", 2, true, 2},
		{"Upgrade + Sync, no OpenCode", 3, false, 3},
		{"Upgrade + Sync, OpenCode detected", 3, true, 3},
		{"Manage backups, no OpenCode", 4, false, 8},
		{"Manage backups, OpenCode detected", 4, true, 9},
		{"Managed uninstall, no OpenCode", 5, false, 9},
		{"Managed uninstall, OpenCode detected", 5, true, 10},
		{"Quit, no OpenCode", 6, false, 11},
		{"Quit, OpenCode detected", 6, true, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := qeModelWithOpenCodeDetection(tt.hasOpenCode)
			got := qeWelcomeCanonicalCursor(m, tt.collapsed)
			if got != tt.wantCanonical {
				t.Fatalf("qeWelcomeCanonicalCursor(collapsed=%d, hasOpenCode=%v) = %d, want %d",
					tt.collapsed, tt.hasOpenCode, got, tt.wantCanonical)
			}
		})
	}
}

// TestQEWelcomeCanonicalCursor_QuitBoundary is the explicit boundary case
// design.md calls out as the historical bug: collapsed Quit must land on the
// canonical Quit index, NOT on CommunityTools (the tail gap it must skip).
func TestQEWelcomeCanonicalCursor_QuitBoundary(t *testing.T) {
	enableQESeam(t)
	const collapsedQuit = 6

	t.Run("no OpenCode", func(t *testing.T) {
		m := qeModelWithOpenCodeDetection(false)
		got := qeWelcomeCanonicalCursor(m, collapsedQuit)
		const wantQuit = 11
		const communityToolsIdx = 10 // base+2, the tail gap qeWelcomeCanonicalCursor must skip
		if got == communityToolsIdx {
			t.Fatalf("qeWelcomeCanonicalCursor(Quit, no OpenCode) = %d routed to CommunityTools index, want %d (Quit)", got, wantQuit)
		}
		if got != wantQuit {
			t.Fatalf("qeWelcomeCanonicalCursor(Quit, no OpenCode) = %d, want %d", got, wantQuit)
		}
	})

	t.Run("OpenCode detected", func(t *testing.T) {
		m := qeModelWithOpenCodeDetection(true)
		got := qeWelcomeCanonicalCursor(m, collapsedQuit)
		const wantQuit = 12
		const communityToolsIdx = 11 // base+2, the tail gap qeWelcomeCanonicalCursor must skip
		if got == communityToolsIdx {
			t.Fatalf("qeWelcomeCanonicalCursor(Quit, OpenCode detected) = %d routed to CommunityTools index, want %d (Quit)", got, wantQuit)
		}
		if got != wantQuit {
			t.Fatalf("qeWelcomeCanonicalCursor(Quit, OpenCode detected) = %d, want %d", got, wantQuit)
		}
	})
}

// qeModelWithOpenCodeDetection builds a Model whose hasDetectedOpenCode()
// reflects the requested state, via Detection.Configs (the real upstream
// source hasDetectedOpenCode reads), not a synthetic field.
func qeModelWithOpenCodeDetection(detected bool) Model {
	m := NewModel(system.DetectionResult{}, "dev")
	if detected {
		m.Detection.Configs = []system.ConfigState{{Agent: string(model.AgentOpenCode), Exists: true}}
	}
	return m
}

// TestQECanonicalIndexForLabel_SurvivesUpstreamReorder is the direct proof
// that the reverse-label-lookup mechanism is immune to the class of bug
// qeWelcomeCanonicalCursor used to have: a stale hardcoded offset silently
// mis-routing after an upstream reorder/insert (the historical base 7->8
// sync bug this change hardens against). It feeds qeCanonicalIndexForLabel
// two DIFFERENT orderings of the same labels — simulating "before" and
// "after upstream reordered its menu" — and asserts every label resolves to
// its CURRENT position in whichever list is passed in, never a stale one.
func TestQECanonicalIndexForLabel_SurvivesUpstreamReorder(t *testing.T) {
	before := []string{"Start installation", "Sync configs", "Manage backups", "Managed uninstall", "Quit"}
	// "after": upstream inserted a new entry up front AND moved Quit from
	// last to first — the exact shape of drift that broke the old arithmetic.
	after := []string{"Quit", "A Brand New Upstream Entry", "Start installation", "Sync configs", "Manage backups", "Managed uninstall"}

	wantAfter := map[string]int{
		"Start installation": 2,
		"Sync configs":       3,
		"Manage backups":     4,
		"Managed uninstall":  5,
		"Quit":               0,
	}

	for _, label := range before {
		beforeIdx, ok := qeCanonicalIndexForLabel(before, label)
		if !ok {
			t.Fatalf("qeCanonicalIndexForLabel(before, %q) not found", label)
		}
		afterIdx, ok := qeCanonicalIndexForLabel(after, label)
		if !ok {
			t.Fatalf("qeCanonicalIndexForLabel(after, %q) not found", label)
		}
		if afterIdx != wantAfter[label] {
			t.Fatalf("qeCanonicalIndexForLabel(after, %q) = %d, want %d (its real position in the reordered list)", label, afterIdx, wantAfter[label])
		}
		// The point: the lookup tracks CONTENT, not a cached offset, so the
		// index legitimately differs across the two orderings for a label
		// whose position moved (e.g. "Quit": last in before, first in after).
		if label == "Quit" && beforeIdx == afterIdx {
			t.Fatalf("qeCanonicalIndexForLabel(%q) returned the same index (%d) in both orderings; test fixture did not actually move it", label, beforeIdx)
		}
	}
}

// TestQECanonicalIndexForLabel_NotFound documents the fail-loud contract at
// the pure-function level: a label absent from the upstream list must report
// ok=false, never a fabricated index.
func TestQECanonicalIndexForLabel_NotFound(t *testing.T) {
	_, ok := qeCanonicalIndexForLabel([]string{"A", "B"}, "C")
	if ok {
		t.Fatal("qeCanonicalIndexForLabel(missing label) ok = true, want false")
	}
}

// TestQEWelcomeCanonicalCursor_PanicsOnOutOfRangeCollapsedCursor asserts the
// fail-loud contract design.md always intended: an out-of-range collapsed
// cursor must panic, not silently pass through unchanged (the bug in the
// prior arithmetic's `default: return collapsed` branch).
func TestQEWelcomeCanonicalCursor_PanicsOnOutOfRangeCollapsedCursor(t *testing.T) {
	enableQESeam(t)
	m := qeModelWithOpenCodeDetection(false)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("qeWelcomeCanonicalCursor(collapsed=99) did not panic, want a fail-loud panic on an out-of-range collapsed cursor")
		}
	}()
	qeWelcomeCanonicalCursor(m, 99)
}

// TestQEWelcomeCanonicalCursor_PanicsOnNegativeCollapsedCursor triangulates
// with the above at the other boundary.
func TestQEWelcomeCanonicalCursor_PanicsOnNegativeCollapsedCursor(t *testing.T) {
	enableQESeam(t)
	m := qeModelWithOpenCodeDetection(false)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("qeWelcomeCanonicalCursor(collapsed=-1) did not panic, want a fail-loud panic on a negative collapsed cursor")
		}
	}()
	qeWelcomeCanonicalCursor(m, -1)
}

// ─── qeFilterPickerFlow ──────────────────────────────────────────────────────

// TestQEFilterPickerFlow_ExcludesDevOnlyScreens feeds qeFilterPickerFlow a
// synthetic raw slice shaped exactly like what upstream's pickerFlowSlice
// would build BEFORE the anchor's delegation (Claude/Kiro/Codex pickers +
// SDDMode present alongside Preset/StrictTDD/DependencyTree), and asserts it
// drops exactly those 4 slice-gated screens while preserving the rest of the
// chain. A synthetic slice is required here (rather than calling
// m.pickerFlowSlice()) because pickerFlowSlice's own anchor (model.go:3976)
// now already delegates to qeFilterPickerFlow — calling it would test the
// function against its own already-filtered output.
func TestQEFilterPickerFlow_ExcludesDevOnlyScreens(t *testing.T) {
	enableQESeam(t)
	raw := []Screen{
		ScreenPreset,
		ScreenClaudeModelPicker,
		ScreenKiroModelPicker,
		ScreenCodexModelPicker,
		ScreenSDDMode,
		ScreenStrictTDD,
		ScreenDependencyTree,
	}

	filtered := qeFilterPickerFlow(raw)

	for _, excluded := range []Screen{ScreenClaudeModelPicker, ScreenKiroModelPicker, ScreenCodexModelPicker, ScreenSDDMode} {
		if qeScreenInSlice(filtered, excluded) {
			t.Fatalf("qeFilterPickerFlow(%v) = %v, still contains excluded screen %v", raw, filtered, excluded)
		}
	}

	// ComponentSDD-driven StrictTDD must still be reachable — hiding a screen
	// must never remove preset functionality.
	want := []Screen{ScreenPreset, ScreenStrictTDD, ScreenDependencyTree}
	if len(filtered) != len(want) {
		t.Fatalf("qeFilterPickerFlow(%v) = %v, want %v", raw, filtered, want)
	}
	for i := range want {
		if filtered[i] != want[i] {
			t.Fatalf("qeFilterPickerFlow(%v) = %v, want %v", raw, filtered, want)
		}
	}
}

// TestQEFilterPickerFlow_KeepsNonGatedScreensUnchanged triangulates with a
// synthetic slice containing none of the 4 dev-only screens — qeFilterPickerFlow
// must be a no-op in that case.
func TestQEFilterPickerFlow_KeepsNonGatedScreensUnchanged(t *testing.T) {
	enableQESeam(t)
	raw := []Screen{ScreenPreset, ScreenStrictTDD, ScreenDependencyTree}

	filtered := qeFilterPickerFlow(raw)

	if len(filtered) != len(raw) {
		t.Fatalf("qeFilterPickerFlow(%v) = %v, want unchanged slice (no dev-only screens present)", raw, filtered)
	}
	for i := range raw {
		if raw[i] != filtered[i] {
			t.Fatalf("qeFilterPickerFlow(%v) = %v, want identical to input at index %d", raw, filtered, i)
		}
	}
}

// TestQEModel_PickerFlowSliceNeverExposesDevOnlyScreens is the integration-
// level companion: it drives the REAL m.pickerFlowSlice() (post-anchor) with
// a Selection that upstream would route through all 3 model pickers plus
// SDDMode, and asserts the anchor at model.go:3976 is actually wired —
// i.e. the live method's return value, not just the pure function in
// isolation, excludes the 4 dev-only screens.
func TestQEModel_PickerFlowSliceNeverExposesDevOnlyScreens(t *testing.T) {
	enableQESeam(t)
	m := NewModel(system.DetectionResult{}, "dev")
	m.Selection.Agents = []model.AgentID{
		model.AgentClaudeCode,
		model.AgentKiroIDE,
		model.AgentCodex,
		model.AgentOpenCode,
	}
	m.Selection.Components = []model.ComponentID{model.ComponentSDD}
	m.Selection.SDDMode = model.SDDModeSingle

	got := m.pickerFlowSlice()

	for _, excluded := range []Screen{ScreenClaudeModelPicker, ScreenKiroModelPicker, ScreenCodexModelPicker, ScreenSDDMode, ScreenStrictTDD} {
		if qeScreenInSlice(got, excluded) {
			t.Fatalf("m.pickerFlowSlice() = %v, still contains excluded screen %v (anchor not wired)", got, excluded)
		}
	}
	// Hiding StrictTDD (like SDDMode) must not drop the underlying SDD component.
	if !hasSelectedComponent(m.Selection.Components, model.ComponentSDD) {
		t.Fatalf("Selection.Components = %v, want ComponentSDD preserved despite hidden StrictTDD screen", m.Selection.Components)
	}
}

func qeScreenInSlice(s []Screen, target Screen) bool {
	for _, screen := range s {
		if screen == target {
			return true
		}
	}
	return false
}

// ─── qeSuppressCommunityTools / qeSuppressOpenCodePlugins ───────────────────

func TestQESuppressCommunityTools_AlwaysTrue(t *testing.T) {
	enableQESeam(t)
	if !qeSuppressCommunityTools() {
		t.Fatal("qeSuppressCommunityTools() = false, want true for the QE build")
	}
}

func TestQESuppressOpenCodePlugins_AlwaysTrue(t *testing.T) {
	enableQESeam(t)
	if !qeSuppressOpenCodePlugins() {
		t.Fatal("qeSuppressOpenCodePlugins() = false, want true for the QE build")
	}
}
