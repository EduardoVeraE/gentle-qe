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

	for _, excluded := range []Screen{ScreenClaudeModelPicker, ScreenKiroModelPicker, ScreenCodexModelPicker, ScreenSDDMode} {
		if qeScreenInSlice(got, excluded) {
			t.Fatalf("m.pickerFlowSlice() = %v, still contains excluded screen %v (anchor not wired)", got, excluded)
		}
	}
	if !qeScreenInSlice(got, ScreenStrictTDD) {
		t.Fatalf("m.pickerFlowSlice() = %v, missing ScreenStrictTDD (ComponentSDD functionality must survive)", got)
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
