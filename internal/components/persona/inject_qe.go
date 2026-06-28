package persona

import (
	"regexp"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// inject_qe.go — contenido de la persona SDET del overlay Gentle-QE.
//
// Aislado de inject.go (upstream). personaContent delega aquí con un solo case
// (ancla vigilada por tools/qe-overlay). El upstream nunca toca este archivo.

var htmlCommentRE = regexp.MustCompile(`(?s)<!--.*?-->`)

// qePersonaContent devuelve el contenido de la persona SDET, o "" si la persona
// no es del overlay. Compone persona-sdet.md con el slot opcional del maintainer
// (lineamientos-personales.md), anexándolo solo si tiene contenido real.
func qePersonaContent(_ model.AgentID, persona model.PersonaID) string {
	if persona != model.PersonaSDET {
		return ""
	}
	base := assets.MustRead("generic/persona-sdet.md")
	if extra := personalGuidelines(); extra != "" {
		base = strings.TrimRight(base, "\n") +
			"\n\n## Lineamientos personales\n\n" + extra + "\n"
	}
	return base
}

// personalGuidelines lee el slot del maintainer y devuelve su contenido real
// (sin comentarios HTML ni espacios). Devuelve "" si el slot está vacío.
func personalGuidelines() string {
	raw, err := assets.Read("generic/lineamientos-personales.md")
	if err != nil {
		return ""
	}
	stripped := strings.TrimSpace(htmlCommentRE.ReplaceAllString(raw, ""))
	return stripped
}
