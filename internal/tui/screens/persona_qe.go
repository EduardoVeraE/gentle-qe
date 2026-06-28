package screens

import "github.com/gentleman-programming/gentle-ai/internal/model"

// persona_qe.go — opción de persona SDET del overlay Gentle-QE para el picker.
//
// La descripción se inyecta al map upstream vía init() (cero edición inline).
// PersonaOptions appendea qePersonaOptions() con una sola línea.

func qePersonaOptions() []model.PersonaID {
	return []model.PersonaID{model.PersonaSDET}
}

func init() {
	personaDescriptions[model.PersonaSDET] = "Persona SDET: testing senior, ISTQB y shift-left"
}
