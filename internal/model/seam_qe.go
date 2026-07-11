package model

// seam_qe.go — Gentle-QE testing seam (net-new fork overlay).
//
// QEInstallerFlow is the single shared seam that gates every QE installer
// behavior overlay (screen hiding, Welcome collapse, option-list filtering,
// and the SDDMode=single default) across the tui, screens and cli packages.
//
// It is true in the production fork binary (which never compiles _test.go), so
// the QE flow is UNCONDITIONAL for real users. Each affected test package flips
// it OFF in a net-new TestMain so the upstream dev-flow / parity tests pass
// unedited; QE tests opt back in locally. It lives here (package model) because
// every consumer already imports model, avoiding a tui→cli or cli→screens
// dependency. Any test that mutates it MUST NOT call t.Parallel().
const qeFlowDefault = true

// QEInstallerFlow gates the Gentle-QE unconditional installer flow. See above.
var QEInstallerFlow = qeFlowDefault

// QEDefaultSDDMode returns the SDD orchestrator mode for the QE build. With the
// seam ON (production) it DEFAULTS an unset mode to SDDModeSingle — hiding the
// installer's SDDMode screen and closing the ""→SDDModeMulti auto-promotion
// path (app.go:664-666, sync.go:681-682) — for BOTH the TUI (NewModel, always
// unset) and the CLI install path, keeping them in parity. An explicit mode
// (e.g. CLI `--sdd-mode multi`) is respected, not overridden. With the seam OFF
// (dev/parity tests) it returns the upstream value unchanged.
func QEDefaultSDDMode(cur SDDModeID) SDDModeID {
	if QEInstallerFlow && cur == "" {
		return SDDModeSingle
	}
	return cur
}
