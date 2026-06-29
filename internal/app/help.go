package app

import (
	"fmt"
	"io"

	"github.com/gentleman-programming/gentle-ai/internal/branding"
)

func printHelp(w io.Writer, version string) {
	p := branding.Product
	fmt.Fprintf(w, `%s — %s: Unified AI Ecosystem for Testing and Reliability (%s)

USAGE
  %s                     Launch interactive TUI
  %s <command> [flags]

COMMANDS
  install      Configure AI coding agents on this machine
  uninstall    Remove %s managed files from this machine
  sync         Sync agent configs and skills to current version
  skill-registry refresh
               Refresh .atl/skill-registry.md with cache-hit fast path
  sdd-status [change]
               Print native SDD phase status for orchestrators
  sdd-continue [change]
               Print native SDD dispatcher routing output
  update       Check for available updates
  upgrade      Apply updates to managed tools
  restore      Restore a config backup
  doctor       Run ecosystem health diagnostics
  version      Print version

FLAGS
  --help, -h    Show this help

Run '%s help' for this message.
Documentation: https://github.com/%s/%s
`, p, branding.Display, version, p, p, branding.Display, p, branding.Owner, branding.Repo)
}
