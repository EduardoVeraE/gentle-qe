package screens

import (
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/model"
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
