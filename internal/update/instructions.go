package update

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/branding"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

// updateHint returns a platform-specific instruction string for updating the given tool.
func updateHint(tool ToolInfo, profile system.PlatformProfile) string {
	switch tool.Name {
	case branding.Product: // overlay Gentle-QE (ancla qe-overlay)
		return gentleAIHint(profile)
	case "engram":
		return engramHint(profile)
	case "gga":
		return ggaHint(profile)
	case "opencode-subagent-statusline", "opencode-sdd-engram-manage":
		return branding.Product + " upgrade updates ~/.config/opencode npm deps, clears this plugin's @latest cache, then requires OpenCode restart/reload"
	default:
		return ""
	}
}

func updateHintForOwnership(tool ToolInfo, profile system.PlatformProfile, ownership HomebrewOwnership) string {
	if profile.PackageManager == "brew" && ownership != HomebrewNone {
		return fmt.Sprintf("brew upgrade --%s %s", ownership, tool.Name)
	}
	return updateHint(tool, profile)
}

func openCodeRegisteredNotMaterializedHint(tool ToolInfo) string {
	pkg := strings.TrimSpace(tool.NpmPackage)
	if pkg == "" {
		pkg = tool.Name
	}
	return fmt.Sprintf("registered in ~/.config/opencode/tui.json; pending npm dependency materialization for %s. Run %s upgrade to install/update ~/.config/opencode dependencies, then restart or reload OpenCode; if it stays pending, check OpenCode logs for package or peer dependency errors.", pkg, branding.Product)
}

func gentleAIHint(profile system.PlatformProfile) string { // overlay Gentle-QE (ancla qe-overlay)
	if profile.PackageManager == "brew" && homebrewPackageInstalled(branding.Product) {
		return "brew upgrade " + branding.Product
	}

	switch profile.OS {
	case "linux":
		return fmt.Sprintf("curl -fsSL https://raw.githubusercontent.com/%s/%s/main/scripts/install.sh | bash", branding.Owner, branding.Repo)
	case "darwin":
		return branding.Product + " upgrade (downloads pre-built binary)"
	case "windows":
		return fmt.Sprintf("irm https://raw.githubusercontent.com/%s/%s/main/scripts/install.ps1 | iex", branding.Owner, branding.Repo)
	default:
		return ""
	}
}

func engramHint(profile system.PlatformProfile) string {
	if profile.PackageManager == "brew" && homebrewPackageInstalled("engram") {
		return "brew upgrade engram"
	}
	return branding.Product + " upgrade (downloads pre-built binary)"
}

func ggaHint(profile system.PlatformProfile) string {
	if profile.PackageManager == "brew" && homebrewPackageInstalled("gga") {
		return "brew upgrade gga"
	}
	return "See https://github.com/Gentleman-Programming/gentleman-guardian-angel"
}
