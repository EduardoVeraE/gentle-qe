package screens

import "github.com/gentleman-programming/gentle-ai/internal/model"

// persona_qe.go — opción de persona SDET del overlay Gentle-QE para el picker.
//
// La descripción se inyecta al map upstream vía init() (cero edición inline).
// PersonaOptions delega a qeFilterPersonaOptions(opts) con una sola línea; el
// parámetro opts (la lista dev-only upstream) se ignora a propósito — el
// build QE siempre devuelve únicamente las personas QE, nunca un append.

func qeFilterPersonaOptions(_ []model.PersonaID) []model.PersonaID {
	return []model.PersonaID{model.PersonaSDET}
}

// qePersonaOptionsForBuild is the seam-aware entry the PersonaOptions anchor
// calls. In the production QE build (seam ON) it returns the QE-only list; in
// the test package (seam OFF via TestMain) it reproduces upstream's shipped
// behavior — the dev options with the SDET persona appended — so the upstream
// persona-contract tests pass unedited.
func qePersonaOptionsForBuild(opts []model.PersonaID) []model.PersonaID {
	if !model.QEInstallerFlow {
		return append(opts, model.PersonaSDET)
	}
	return qeFilterPersonaOptions(opts)
}

func init() {
	personaDescriptions[model.PersonaSDET] = "Persona SDET: testing senior, ISTQB y shift-left"
}
