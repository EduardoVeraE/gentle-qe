package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/tui/screens"
)

// installer_flow_qe_test.go — QE installer-flow integration coverage (net-new
// fork overlay, registered in tools/qe-overlay/overlay.json).
//
// These tests drive the REAL Bubbletea Update() loop with the QE seam
// (model.QEInstallerFlow) turned ON via enableQESeam(t), replicating the key
// sequences used by the upstream dev-flow tests (preset_flow_test.go) so the
// oracle is the live Model.Screen/Model.Selection after Update(), not a
// hand-rolled model of the state machine. See
// openspec/changes/gentle-qe-dl9/specs/qe-installer-flow/spec.md for the
// requirements under test.
//
// Tests here MUST NOT call t.Parallel(): enableQESeam mutates the shared
// package-global model.QEInstallerFlow.

// qeHiddenScreens is the fixed set of 6 dev-only screens the QE build must
// never present in its navigation flow (spec: "Dev-Only Screens Hidden With
// Silent Defaults").
var qeHiddenScreens = map[Screen]string{
	ScreenClaudeModelPicker: "ScreenClaudeModelPicker",
	ScreenKiroModelPicker:   "ScreenKiroModelPicker",
	ScreenCodexModelPicker:  "ScreenCodexModelPicker",
	ScreenSDDMode:           "ScreenSDDMode",
	ScreenCommunityTools:    "ScreenCommunityTools",
	ScreenOpenCodePlugins:   "ScreenOpenCodePlugins",
}

// assertNotHiddenScreen fails the test if screen is one of the 6 QE-hidden
// screens. step names the point in the flow being checked, for diagnostics.
func assertNotHiddenScreen(t *testing.T, step string, screen Screen) {
	t.Helper()
	if name, hidden := qeHiddenScreens[screen]; hidden {
		t.Fatalf("%s: Model.Screen = %s, a QE-hidden screen must never become active", step, name)
	}
}

// TestQEInstallerFlow_HiddenScreensNeverActivated drives the full QE
// installation flow from Welcome through Complete with the default agent
// preselection — which, with an empty DetectionResult and no install state,
// includes every catalog agent (catalog.AllAgents()), i.e. Claude Code, Kiro
// IDE, Codex and OpenCode. That is exactly the selection that would upstream
// (seam OFF) route through all 3 model pickers, SDDMode, CommunityTools and
// OpenCodePlugins. The oracle is that none of the 6 QE-hidden screens is ever
// observed on Model.Screen, and that the underlying ComponentSDD selection
// and SDDMode default survive the flow untouched.
func TestQEInstallerFlow_HiddenScreensNeverActivated(t *testing.T) {
	enableQESeam(t)

	m := NewModel(system.DetectionResult{}, "dev")
	if !m.Selection.HasAgent(model.AgentClaudeCode) || !m.Selection.HasAgent(model.AgentOpenCode) {
		t.Fatalf("preconditions: default agent selection = %v, want it to include Claude Code and OpenCode", m.Selection.Agents)
	}

	assertNotHiddenScreen(t, "Welcome", m.Screen)

	steps := []flowAction{
		{key: tea.KeyMsg{Type: tea.KeyEnter}},                                                       // Welcome (Start installation) -> Detection
		{key: tea.KeyMsg{Type: tea.KeyEnter}},                                                       // Detection -> Agents
		{key: tea.KeyMsg{Type: tea.KeyEnter}, cursor: len(screens.AgentOptions()), setCursor: true}, // Agents Continue -> Persona
		{key: tea.KeyMsg{Type: tea.KeyEnter}},                                                       // Persona (SDET) -> Preset
		{key: tea.KeyMsg{Type: tea.KeyEnter}},                                                       // Preset (QE SDET) -> StrictTDD (pickers/SDDMode hidden)
		{key: tea.KeyMsg{Type: tea.KeyEnter}},                                                       // StrictTDD (Enable) -> DependencyTree (CommunityTools/OpenCodePlugins hidden)
		{key: tea.KeyMsg{Type: tea.KeyEnter}},                                                       // DependencyTree Continue -> Review
		{key: tea.KeyMsg{Type: tea.KeyEnter}},                                                       // Review Continue -> Installing
	}

	state := m
	for i, action := range steps {
		state = applyFlowAction(t, state, action)
		assertNotHiddenScreen(t, fmt.Sprintf("forward step %d", i+1), state.Screen)
	}

	if state.Screen != ScreenInstalling {
		t.Fatalf("after Review confirm, Screen = %v, want ScreenInstalling", state.Screen)
	}

	// Drive the manual installing fallback (no ExecuteFn) to Complete.
	for i := 0; i < 50 && state.Screen != ScreenComplete; i++ {
		state = applyFlowAction(t, state, flowAction{key: tea.KeyMsg{Type: tea.KeyEnter}})
		assertNotHiddenScreen(t, "installing loop", state.Screen)
	}
	if state.Screen != ScreenComplete {
		t.Fatalf("installing loop did not reach ScreenComplete within 50 iterations, last screen = %v", state.Screen)
	}

	// Spec: hiding a screen must never drop the underlying feature. ComponentSDD
	// must still be selected after the flow completed with the SDDMode/model
	// picker screens hidden.
	if !hasSelectedComponent(state.Selection.Components, model.ComponentSDD) {
		t.Fatalf("Selection.Components = %v, want ComponentSDD preserved despite hidden SDDMode screen", state.Selection.Components)
	}
	if state.Selection.SDDMode != model.SDDModeSingle {
		t.Fatalf("Selection.SDDMode = %v, want %v (documented QE default)", state.Selection.SDDMode, model.SDDModeSingle)
	}
}

// TestQEInstallerFlow_SDDModeDefaultsToSingle asserts SDDMode is
// model.SDDModeSingle immediately after NewModel with the QE seam ON, and
// that it survives the picker flow untouched — there is no way for the user
// to flip it since the SDDMode screen that would let them choose Multi is
// hidden.
func TestQEInstallerFlow_SDDModeDefaultsToSingle(t *testing.T) {
	enableQESeam(t)

	m := NewModel(system.DetectionResult{}, "dev")
	if m.Selection.SDDMode != model.SDDModeSingle {
		t.Fatalf("NewModel with QE seam ON: Selection.SDDMode = %v, want %v", m.Selection.SDDMode, model.SDDModeSingle)
	}

	m.Selection.Agents = []model.AgentID{
		model.AgentClaudeCode,
		model.AgentKiroIDE,
		model.AgentCodex,
		model.AgentOpenCode,
	}
	m.Screen = ScreenPreset
	m.Cursor = presetCursor(t, model.PresetQESDET)

	state := applyFlowAction(t, m, flowAction{key: tea.KeyMsg{Type: tea.KeyEnter}})
	if state.Screen != ScreenStrictTDD {
		t.Fatalf("Screen = %v, want ScreenStrictTDD (SDDMode/model pickers hidden)", state.Screen)
	}
	if state.Selection.SDDMode != model.SDDModeSingle {
		t.Fatalf("Selection.SDDMode = %v, want %v to remain the documented default after navigating past the hidden SDDMode screen", state.Selection.SDDMode, model.SDDModeSingle)
	}
}

// TestQEInstallerFlow_StrictTDDScreenIsReachable asserts ScreenStrictTDD is
// NOT among the screens the QE build hides — driving Update() from
// ScreenPreset with ComponentSDD selected must land on it.
func TestQEInstallerFlow_StrictTDDScreenIsReachable(t *testing.T) {
	enableQESeam(t)

	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPreset
	m.Cursor = presetCursor(t, model.PresetQESDET)

	state := applyFlowAction(t, m, flowAction{key: tea.KeyMsg{Type: tea.KeyEnter}})

	if state.Screen != ScreenStrictTDD {
		t.Fatalf("Screen = %v, want ScreenStrictTDD to remain visible in the QE build", state.Screen)
	}
	if !qeScreenInSlice(state.pickerFlowSlice(), ScreenStrictTDD) {
		t.Fatalf("pickerFlowSlice() = %v, missing ScreenStrictTDD", state.pickerFlowSlice())
	}
}

// TestQEInstallerFlow_WelcomeCollapsedQuitDispatchesQuit drives the REAL
// Update() loop (not qeWelcomeCanonicalCursor directly) with the cursor on
// the collapsed Welcome menu's last entry ("Quit", collapsed index 6) and
// asserts the returned command is tea.Quit, for both hasDetectedOpenCode
// states. This is the integration-level companion to
// TestQEWelcomeCanonicalCursor_QuitBoundary (model_qe_test.go): that test
// checks the pure remap function; this one checks the remap is actually
// wired into confirmSelection's ScreenWelcome switch via a full key press.
func TestQEInstallerFlow_WelcomeCollapsedQuitDispatchesQuit(t *testing.T) {
	enableQESeam(t)

	tests := []struct {
		name        string
		hasOpenCode bool
	}{
		{"no OpenCode detected", false},
		{"OpenCode detected", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detection := system.DetectionResult{}
			if tt.hasOpenCode {
				detection.Configs = []system.ConfigState{{Agent: string(model.AgentOpenCode), Exists: true}}
			}
			m := NewModel(detection, "dev")
			m.Screen = ScreenWelcome
			// Collapsed Welcome cursor 6 = "Quit", the last of the 7
			// QE-essential entries (see qeWelcomeOptions / model_qe.go).
			m.Cursor = 6

			updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			state := updated.(Model)

			if cmd == nil {
				t.Fatal("Update(Enter on collapsed Quit) cmd = nil, want tea.Quit")
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatalf("Update(Enter on collapsed Quit) command = %T, want tea.QuitMsg (cursor remap must not misroute to another action)", cmd())
			}
			if state.Screen == ScreenCommunityTools {
				t.Fatal("Screen = ScreenCommunityTools, want the Quit action (cursor remap bug: landed on the hidden CommunityTools tail-gap index instead of Quit)")
			}
		})
	}
}

// TestQEInstallerFlow_ReverseNavigationMirrorsForwardAndAvoidsHiddenScreens
// drives the picker chain forward from ScreenPreset with every model-picker
// agent selected (Claude Code, Kiro IDE, Codex, OpenCode) plus ComponentSDD —
// the exact selection upstream's TestInstallNavigationRoundTrips
// ("all picker agents SDD single round-trips through every picker") drives
// through Claude -> Kiro -> Codex -> SDDMode -> StrictTDD. With the QE seam
// ON, the picker chain collapses to Preset -> StrictTDD -> DependencyTree.
// Esc must walk back through the exact reverse of the forward sequence,
// never landing on a hidden screen.
func TestQEInstallerFlow_ReverseNavigationMirrorsForwardAndAvoidsHiddenScreens(t *testing.T) {
	enableQESeam(t)

	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPreset
	m.Selection.Agents = []model.AgentID{
		model.AgentClaudeCode,
		model.AgentKiroIDE,
		model.AgentCodex,
		model.AgentOpenCode,
	}
	m.Cursor = presetCursor(t, model.PresetQESDET)

	forwardActions := []flowAction{
		{key: tea.KeyMsg{Type: tea.KeyEnter}}, // Preset -> StrictTDD (Claude/Kiro/Codex/SDDMode hidden)
		{key: tea.KeyMsg{Type: tea.KeyEnter}}, // StrictTDD (Enable) -> DependencyTree (CommunityTools/OpenCodePlugins hidden)
	}
	forwardScreens := []Screen{ScreenStrictTDD, ScreenDependencyTree}
	reverseScreens := []Screen{ScreenStrictTDD, ScreenPreset}

	state := m
	for idx, action := range forwardActions {
		state = applyFlowAction(t, state, action)
		if state.Screen != forwardScreens[idx] {
			t.Fatalf("forward step %d: Screen = %v, want %v", idx+1, state.Screen, forwardScreens[idx])
		}
		assertNotHiddenScreen(t, "forward", state.Screen)
	}

	for idx, want := range reverseScreens {
		state = applyFlowAction(t, state, flowAction{key: tea.KeyMsg{Type: tea.KeyEsc}})
		if state.Screen != want {
			t.Fatalf("reverse step %d: Screen = %v, want %v", idx+1, state.Screen, want)
		}
		assertNotHiddenScreen(t, "reverse", state.Screen)
	}
}
