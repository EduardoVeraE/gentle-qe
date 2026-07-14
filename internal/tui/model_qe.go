package tui

import (
	"fmt"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/tui/screens"
)

// model_qe.go — Gentle-QE installer flow simplification overlay (package tui).
//
// Owns the QE-build-only logic that hides dev-only screens from the picker
// flow and remaps the collapsed Welcome menu cursor back to the untouched
// upstream confirmSelection switch. Zero upstream logic edits: model.go only
// carries 1-line delegating anchors to the functions below (see
// tools/qe-overlay/overlay.json inlineAnchors).

// The QE installer-flow testing seam is model.QEInstallerFlow (package model,
// shared by the tui, screens and cli packages). See internal/model/seam_qe.go.

// qeFilterPickerFlow drops the QE build's dev-only screens from an already
// built pickerFlowSlice(): the 3 model pickers (Claude/Kiro/Codex) and
// SDDMode. It reads only the slice elements passed in — never live m.Screen —
// preserving the invariant documented at pickerFlowSlice's call site.
func qeFilterPickerFlow(s []Screen) []Screen {
	if !model.QEInstallerFlow {
		return s
	}
	drop := map[Screen]bool{
		ScreenClaudeModelPicker: true,
		ScreenKiroModelPicker:   true,
		ScreenCodexModelPicker:  true,
		ScreenSDDMode:           true,
	}
	filtered := make([]Screen, 0, len(s))
	for _, screen := range s {
		if drop[screen] {
			continue
		}
		filtered = append(filtered, screen)
	}
	return filtered
}

// qeSuppressCommunityTools reports whether the QE build hides the standalone
// CommunityTools screen from the Welcome-menu install-flow auto-navigation
// guard (shouldShowCommunityToolsScreen). Always true for the QE build.
func qeSuppressCommunityTools() bool {
	return model.QEInstallerFlow
}

// qeSuppressOpenCodePlugins reports whether the QE build hides the
// OpenCodePlugins screen from its call-site guard
// (shouldShowOpenCodePluginsScreen). Always true for the QE build.
func qeSuppressOpenCodePlugins() bool {
	return model.QEInstallerFlow
}

// qeSuppressStrictTDD reports whether the QE build hides the Strict TDD Mode
// screen from its call-site guard (shouldShowStrictTDDScreen). Always true for
// the QE build, so the screen is skipped and Selection.StrictTDD stays at its
// zero value (false / OFF). An explicit CLI `--strict-tdd` is still respected
// upstream — only the QE TUI stops offering the choice.
func qeSuppressStrictTDD() bool {
	return model.QEInstallerFlow
}

// qeWelcomeCanonicalCursor remaps a collapsed Welcome menu cursor position
// (0..6, the 7 QE-essential entries returned by qeWelcomeOptions) to the
// canonical index the untouched upstream confirmSelection ScreenWelcome
// switch expects.
//
// Root-cause fix for the historical arithmetic bug (base 7->8 across the
// upstream v2.x sync; gentle-qe-cwd hardens this): instead of replaying a
// hand-derived offset table that silently drifts whenever upstream reorders,
// inserts, or removes a Welcome menu entry, this looks the collapsed
// cursor's LABEL up in the real, live, uncollapsed upstream menu
// (screens.QEWelcomeFullOptions — sourced from the exact opts slice
// WelcomeOptions() builds, via the capture anchor in welcome.go, NOT a
// second hardcoded copy of upstream's order). The canonical index is
// recomputed from upstream's actual current menu on every call, so it can
// never go stale the way a cached offset can.
//
// Both failure modes fail LOUDLY (panic) per design.md's original intent —
// "must fail loudly, not silently mis-route a menu action" — which the prior
// arithmetic's `default: return collapsed` branch contradicted (it silently
// passed an out-of-range cursor through unchanged):
//
//   - collapsed is outside qeWelcomeOptions' output for these inputs: the
//     collapse keep-set and the live Welcome menu have desynced.
//   - the collapsed entry's label has no match in the full upstream menu:
//     upstream renamed or removed an entry qeWelcomeOptions still keeps.
//
// A collapsed cursor is used, not stored, on the same synchronous
// confirmSelection() call it is produced for, so panicking here surfaces the
// drift at the exact keypress that would otherwise have misrouted — long
// before it could reach a release.
func qeWelcomeCanonicalCursor(m Model, collapsed int) int {
	if !model.QEInstallerFlow {
		return collapsed
	}

	showProfiles := m.hasDetectedOpenCode()
	profileCount := len(m.ProfileList)
	hasEngines := m.hasAgentBuilderEngines()

	collapsedOpts := screens.WelcomeOptions(m.UpdateResults, m.UpdateCheckDone, showProfiles, profileCount, hasEngines)
	if collapsed < 0 || collapsed >= len(collapsedOpts) {
		panic(fmt.Sprintf(
			"qeWelcomeCanonicalCursor: collapsed cursor %d out of range [0,%d) for the QE-collapsed Welcome menu %v — qeWelcomeOptions' keep-set and the live Welcome menu have diverged",
			collapsed, len(collapsedOpts), collapsedOpts))
	}
	label := collapsedOpts[collapsed]

	full := screens.QEWelcomeFullOptions(m.UpdateResults, m.UpdateCheckDone, showProfiles, profileCount, hasEngines)
	canonical, ok := qeCanonicalIndexForLabel(full, label)
	if !ok {
		panic(fmt.Sprintf(
			"qeWelcomeCanonicalCursor: collapsed entry %q (collapsed cursor %d) not found in the upstream Welcome menu %v — upstream likely renamed or removed a QE-kept entry; update qeWelcomeOptions' keep-set (welcome_qe.go)",
			label, collapsed, full))
	}
	return canonical
}

// qePersonaAutoPreset maps each QE persona to the preset it auto-selects. Both
// QE personas skip the preset picker entirely: SDET installs the full SDET
// stack (PresetQESDET), Dev FullStack the upstream foundation skills
// (PresetDevFullStack). The QE-simplified TUI never shows the preset screen —
// choosing the persona is the single install decision. The other QE presets
// (Front/API/Perf) remain reachable via the CLI --preset flag.
var qePersonaAutoPreset = map[model.PersonaID]model.PresetID{
	model.PersonaSDET:         model.PresetQESDET,
	model.PersonaDevFullStack: model.PresetDevFullStack,
}

// qeAutoSelectPersonaPreset couples persona→preset in the QE build ONLY: when
// the user picks a QE persona, it fixes that persona's preset and jumps
// straight to the install-plan screen, skipping the preset picker and the
// intermediate pickers the QE build already suppresses (qeFilterPickerFlow /
// qeSuppress*). It mirrors the piOnly shortcut in confirmSelection's
// ScreenAgents case (model.go).
//
// Returns (m, true) when the shortcut applied; (m, false) when it does not
// apply (a non-QE persona, or seam OFF), leaving the normal ScreenPreset flow
// untouched.
//
// This helper depends on the QE-build invariant "suppressed pickers → the
// preset step falls straight through to DependencyTree"; if the QE build ever
// re-enables an intermediate picker, revisit this jump.
func (m Model) qeAutoSelectPersonaPreset() (Model, bool) {
	if !model.QEInstallerFlow {
		return m, false
	}
	preset, ok := qePersonaAutoPreset[m.Selection.Persona]
	if !ok {
		return m, false
	}
	m.Selection.Preset = preset
	m.Selection.Components = componentsForPreset(preset, m.Selection.Persona)
	m.buildDependencyPlan()
	m.setScreen(ScreenDependencyTree)
	return m, true
}

// qeCanonicalIndexForLabel finds label's position in full — the real,
// current upstream Welcome menu — by CONTENT, never by position. This is the
// entire mechanism that makes qeWelcomeCanonicalCursor immune to upstream
// reordering or insertion: reorder full and the same label still resolves to
// its new position, instead of a stale offset pointing at whatever now sits
// there. Pulled out as its own pure function so that property is directly
// testable without driving the full Bubbletea Update() loop — see
// TestQECanonicalIndexForLabel_SurvivesUpstreamReorder in model_qe_test.go.
func qeCanonicalIndexForLabel(full []string, label string) (int, bool) {
	for idx, opt := range full {
		if opt == label {
			return idx, true
		}
	}
	return -1, false
}
