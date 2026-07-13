package screens

import (
	"strings"
	"sync"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/update"
)

// welcome_qe.go — Gentle-QE Welcome menu collapse overlay (package screens).
//
// Collapses the upstream Welcome menu to the 7 QE-essential entries. The
// keep-set is matched by label (not position) because "OpenCode SDD
// Profiles" is conditionally inserted upstream, which would otherwise shift
// every downstream index. Matching by label keeps the collapse correct
// regardless of showProfiles/hasEngines — those two parameters are accepted
// only for call-site symmetry with WelcomeOptions (welcome.go:55) and are not
// needed by the fixed collapsed keep-set itself.
func qeWelcomeOptions(opts []string, showProfiles, hasEngines bool) []string {
	if !model.QEInstallerFlow {
		return opts
	}
	keep := map[string]bool{
		"Start installation": true,
		"Sync configs":       true,
		"Upgrade + Sync":     true,
		"Manage backups":     true,
		"Managed uninstall":  true,
		"Quit":               true,
	}

	filtered := make([]string, 0, 7)
	for _, opt := range opts {
		if keep[opt] || strings.HasPrefix(opt, "Upgrade tools") {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}

// qeFullWelcomeOptionsMu guards qeFullWelcomeOptionsCapture below. Bubbletea's
// Update()/View() loop is single-goroutine in production, so this is never
// actually contended there — the mutex exists purely so that if a future test
// package ever grows a t.Parallel() (every existing QE seam test is
// documented NOT to use one; see qe_seam_test.go), that can never turn this
// into a silent `go test -race` failure instead of a loud one.
var (
	qeFullWelcomeOptionsMu      sync.Mutex
	qeFullWelcomeOptionsCapture []string
)

// qeCaptureFullWelcomeOptions is invoked via a single-line anchor in
// welcome.go, immediately before WelcomeOptions() hands its freshly built
// opts slice to qeWelcomeOptions' collapse filter above. It runs
// UNCONDITIONALLY — regardless of model.QEInstallerFlow — so both QE and
// upstream dev-flow test runs keep the capture correct, and stores an
// independent copy of the untouched, uncollapsed upstream menu: the single
// source of truth QEWelcomeFullOptions reads back below.
//
// This — not hardcoded index arithmetic — is what lets
// tui.qeWelcomeCanonicalCursor (internal/tui/model_qe.go) recover the real
// canonical index of a collapsed Welcome entry: it looks the entry's LABEL
// up in this real, live upstream list instead of replaying a hand-derived
// offset table that silently drifts whenever upstream reorders, inserts, or
// removes a Welcome menu entry (see design.md: this is the gentle-qe-cwd
// hardening of the historical base 7->8 sync bug).
func qeCaptureFullWelcomeOptions(opts []string) {
	qeFullWelcomeOptionsMu.Lock()
	defer qeFullWelcomeOptionsMu.Unlock()
	qeFullWelcomeOptionsCapture = append([]string(nil), opts...)
}

// QEWelcomeFullOptions returns the untouched, uncollapsed upstream Welcome
// menu — every entry, in upstream's real current order — for the given
// inputs (the same 5 parameters WelcomeOptions itself takes).
//
// It works by calling WelcomeOptions() itself, which (via the anchor in
// welcome.go) unconditionally re-runs qeCaptureFullWelcomeOptions
// synchronously as part of THIS call, then reads that capture back
// immediately. That makes it impossible for the result to be stale: it is
// never a leftover from an earlier View() render or a different Model
// state — it always reflects exactly what upstream's own construction
// built for these exact inputs, in this exact call.
func QEWelcomeFullOptions(updateResults []update.UpdateResult, updateCheckDone, showProfiles bool, profileCount int, hasEngines bool) []string {
	_ = WelcomeOptions(updateResults, updateCheckDone, showProfiles, profileCount, hasEngines)
	qeFullWelcomeOptionsMu.Lock()
	defer qeFullWelcomeOptionsMu.Unlock()
	return append([]string(nil), qeFullWelcomeOptionsCapture...)
}
