package tui

import "github.com/gentleman-programming/gentle-ai/internal/model"

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

// qeWelcomeCanonicalCursor remaps a collapsed Welcome menu cursor position
// (0..6, the 7 QE-essential entries returned by qeWelcomeOptions) to the
// canonical index the untouched upstream confirmSelection ScreenWelcome
// switch expects. It replays the SAME conditional counter that switch uses,
// stepping both non-uniform gaps documented in design.md:
//
//	collapsed 0..3           -> canonical == collapsed (static leaders)
//	base := 7 (8 if hasDetectedOpenCode(), +1 for the OpenCode SDD Profiles
//	           leader-gap entry)
//	collapsed 4 (Backups)    -> base
//	collapsed 5 (Uninstall)  -> base + 1
//	collapsed 6 (Quit)       -> base + 3 (tail gap: skips CommunityTools at
//	                            base+2)
//
// MUST stay in sync with len(qeWelcomeOptions(...)) == 7 — see the
// structural fail-fast test in welcome_qe_test.go. Any future change to the
// collapsed keep-set that is not reflected here must fail loudly, not
// silently mis-route a menu action.
func qeWelcomeCanonicalCursor(m Model, collapsed int) int {
	if !model.QEInstallerFlow {
		return collapsed
	}
	if collapsed < 4 {
		return collapsed
	}

	base := 7
	if m.hasDetectedOpenCode() {
		base = 8
	}

	switch collapsed {
	case 4: // Manage backups
		return base
	case 5: // Managed uninstall
		return base + 1
	case 6: // Quit — steps over the tail gap (CommunityTools at base+2)
		return base + 3
	default:
		return collapsed
	}
}
