package cli

import (
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

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
	switch p {
	case model.PersonaSDET, model.PersonaDevFullStack:
		return true
	default:
		return false
	}
}

// isQEPreset indica si el valor es un preset válido del overlay.
func isQEPreset(p model.PresetID) bool {
	switch p {
	case model.PresetQESDET, model.PresetQEFront, model.PresetQEAPI, model.PresetQEPerf, model.PresetDevFullStack:
		return true
	default:
		return false
	}
}

// qeUseTestDesignSDD decide, según la persona seleccionada, si la instalación
// usa el SDD test-design QE (override ON) o el SDD de desarrollo upstream
// (override OFF). Solo la persona SDET (perfil QE) recibe el test-design SDD;
// las demás —incluida Dev FullStack— reciben el SDD de desarrollo original.
func qeUseTestDesignSDD(sel model.Selection) bool {
	return sel.Persona == model.PersonaSDET
}

// qeCouplePersonaPreset acopla la persona Dev FullStack a su preset dev en el
// flujo CLI (paridad con el auto-select del TUI): si se pidió
// --persona dev-fullstack SIN --preset explícito, deriva PresetDevFullStack
// (foundationSkills upstream) en vez del default QE (SDET). Con --preset
// explícito se respeta la elección del usuario.
func qeCouplePersonaPreset(persona model.PersonaID, resolved model.PresetID, rawPresetFlag string) model.PresetID {
	if persona == model.PersonaDevFullStack && strings.TrimSpace(rawPresetFlag) == "" {
		return model.PresetDevFullStack
	}
	return resolved
}
