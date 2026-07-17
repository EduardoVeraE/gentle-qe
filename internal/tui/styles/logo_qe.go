package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// logo_qe.go — Gentle-QE overlay: replaces the upstream neon-rose logo with
// the Gentle-QE brand. The upstream file logo.go declares logoLines and
// gradientColors as package-level vars; this init() reassigns both after the
// var initializers run, so the original file is never edited. RenderLogo()'s
// banding loop applies unchanged; with a single band the whole logo renders
// in the solid Gentle-QE brand color.
//
// The wordmark alone stays as the default logoLines because RenderLogo()
// callers render into frames that wrap long lines and have tight height
// budgets (see welcomePrimaryContentHeight in screens/welcome.go). The full
// brand scene is opt-in via RenderLogoFit, which falls back to the wordmark
// whenever the terminal cannot fit the scene unwrapped.
func init() {
	logoLines = qeLogoLines
	gradientColors = []lipgloss.Color{qeLogoColor}
}

// qeLogoColor is the Gentle-QE brand color for the terminal logo. lipgloss
// degrades it automatically to the nearest ANSI color on terminals without
// truecolor support.
const qeLogoColor = lipgloss.Color("#0A91B2")

// qeLogoLines is the Gentle-QE wordmark (figlet ANSI Shadow) plus a tagline.
var qeLogoLines = []string{
	` ██████╗ ███████╗███╗   ██╗████████╗██╗     ███████╗     ██████╗ ███████╗`,
	`██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██║     ██╔════╝    ██╔═══██╗██╔════╝`,
	`██║  ███╗█████╗  ██╔██╗ ██║   ██║   ██║     █████╗█████╗██║   ██║█████╗`,
	`██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██║     ██╔══╝╚════╝██║▄▄ ██║██╔══╝`,
	`╚██████╔╝███████╗██║ ╚████║   ██║   ███████╗███████╗    ╚██████╔╝███████╗`,
	` ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚══════╝╚══════╝     ╚══▀▀═╝ ╚══════╝`,
	`             Q U A L I T Y   E N G I N E E R`,
}

// qeBrandSceneLines is a braille render of the Gentle-QE brand rocket
// ascending, centered over the 73-char wordmark. Max line width is 52 chars.
var qeBrandSceneLines = []string{
	`                                     ⢠⣄`,
	`                                    ⣰⠏⠹⣆`,
	`                                   ⣴⠃  ⠘⣧`,
	`                                 ⢀⡼⠃    ⠈⢷⡀        ✦`,
	`                                 ⣼⠁      ⠈⣷`,
	`                                ⢀⡟        ⢹⡄`,
	`                                ⢸⡇⢀⡤⠖⠒⠒⠒⢤⡀⢸⡇`,
	`                                ⢸⡇⠘⠧⣄⣀⣀⣀⡴⠃⢸⡇`,
	`                                ⢸⡇        ⢸⡇`,
	`                               ⣠⢿⡇        ⢸⡿⣆`,
	`                          ✦  ⢀⡾⠁⢸⡇        ⢸⡇⠈⢳⡄`,
	`                           ⢀⡴⠋  ⣸⢧⣤⣤⣤⣤⣤⣤⣤⣤⡼⣇  ⠙⢦⡀`,
	`                           ⠻⢅⣠⠴⠋⠁  ⠘⣦⣀⣀⣴⠇   ⠙⠢⣄⣠⠟`,
	`                                   ⢠⣿⣿⣿⣿⡄`,
	`                                   ⠸⣿⣿⣿⣿⠏          ✦`,
	`                                    ⠙⣿⣿⠏`,
	`                                     ⠘⠏`,
}

// Minimum terminal size for the full brand scene. Width must fit the widest
// line (73) plus the Welcome frame chrome (10, see welcomeContentWidth);
// height must absorb the extra scene rows on top of the wordmark-only
// Welcome view without pushing the menu off-screen.
const (
	qeBrandSceneMinWidth  = 84
	qeBrandSceneMinHeight = 42
)

// RenderLogoFit returns the full brand scene above the wordmark when the
// terminal is large enough to show it unwrapped, and falls back to the plain
// wordmark logo otherwise. Zero width/height (unknown size) uses the
// fallback. The gradient banding matches RenderLogo().
func RenderLogoFit(width, height int) string {
	if width < qeBrandSceneMinWidth || height < qeBrandSceneMinHeight {
		return RenderLogo()
	}

	lines := make([]string, 0, len(qeBrandSceneLines)+len(qeLogoLines))
	lines = append(lines, qeBrandSceneLines...)
	lines = append(lines, qeLogoLines...)

	bands := len(gradientColors)
	total := len(lines)
	var b strings.Builder
	for i, line := range lines {
		bandIdx := (i * bands) / total
		if bandIdx >= bands {
			bandIdx = bands - 1
		}
		style := lipgloss.NewStyle().Foreground(gradientColors[bandIdx])
		b.WriteString(style.Render(line))
		if i < total-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
