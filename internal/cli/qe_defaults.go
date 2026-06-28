package cli

import "github.com/gentleman-programming/gentle-ai/internal/model"

// qe_defaults.go — defaults y validación del overlay SDET/QE (Gentle-QE).
//
// Aislado de validate.go (upstream). normalizePersona/normalizePreset delegan
// aquí con líneas de anclaje vigiladas por tools/qe-overlay. Mantiene paridad
// con el default del TUI (PresetQESDET / PersonaSDET).

const (
	qeDefaultPersona = model.PersonaSDET
	qeDefaultPreset  = model.PresetQESDET
)

// isQEPersona indica si el valor es una persona válida del overlay.
func isQEPersona(p model.PersonaID) bool {
	return p == model.PersonaSDET
}

// isQEPreset indica si el valor es un preset válido del overlay.
func isQEPreset(p model.PresetID) bool {
	switch p {
	case model.PresetQESDET, model.PresetQEFront, model.PresetQEAPI, model.PresetQEPerf:
		return true
	default:
		return false
	}
}
