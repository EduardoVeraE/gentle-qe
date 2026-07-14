package screens

import "github.com/gentleman-programming/gentle-ai/internal/model"

// persona_label_qe.go — etiqueta visible de persona del overlay Gentle-QE.
//
// La pantalla "Choose your Persona" renderiza el nombre de cada persona con
// string(persona) crudo (persona.go), que para el ID "sdet" mostraría "sdet"
// en minúsculas. Este helper mayúsculiza el label de las personas del fork solo
// para presentación, SIN tocar el PersonaID (que es key en assets, CLI y sync).
//
// Seam-aware: con el seam OFF (tests de parity upstream) devuelve el string
// crudo intacto, así los contract tests que esperan labels upstream literales
// (p.ej. "gentleman-neutral-artifacts") pasan sin editar.
func qePersonaLabel(p model.PersonaID) string {
	if !model.QEInstallerFlow {
		return string(p)
	}
	switch p {
	case model.PersonaSDET:
		return "SDET"
	default:
		return string(p)
	}
}
