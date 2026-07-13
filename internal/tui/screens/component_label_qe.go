package screens

import "github.com/gentleman-programming/gentle-ai/internal/model"

// component_label_qe.go — etiqueta visible de componente del overlay Gentle-QE.
//
// Renombra SOLO el label mostrado del componente de branding de OpenCode a
// "opencode-sdet-logo" en las pantallas del instalador (Components to install y
// Review), sin tocar el ComponentID upstream ("opencode-gentle-logo"). La
// identidad del componente NO cambia: planner, uninstall y upgrade lo siguen
// resolviendo por su constante original, y los goldens del preset custom no se
// alteran.
//
// Seam-aware: con el seam OFF (tests de parity upstream, TestMain lo apaga)
// devuelve el string crudo intacto, así los tests upstream pasan sin editar.
func qeComponentLabel(c model.ComponentID) string {
	if model.QEInstallerFlow && c == model.ComponentOpenCodeGentleLogo {
		return "opencode-sdet-logo"
	}
	return string(c)
}
