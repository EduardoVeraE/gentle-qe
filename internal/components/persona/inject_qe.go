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

// qePersonaContent devuelve el contenido de la persona del overlay Gentle-QE, o
// "" si la persona no es del fork. Cada persona QE resuelve a su propio asset
// generic net-new; se compone con el slot opcional del maintainer
// (lineamientos-personales.md), anexándolo solo si tiene contenido real.
func qePersonaContent(_ model.AgentID, persona model.PersonaID) string {
	var asset string
	switch persona {
	case model.PersonaSDET:
		asset = "generic/persona-sdet.md"
	case model.PersonaDevFullStack:
		asset = "generic/persona-devfullstack.md"
	default:
		return ""
	}
	base := assets.MustRead(asset)
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

// qeRegionalVoiceReplacements es el mapa exacto viejo→nuevo que neutraliza la
// directiva rioplatense de los assets gentleman del upstream. Los .md del
// upstream NO se editan: la política del fork se aplica en runtime para que
// los syncs mergeen limpio. Si el upstream reformula alguna de estas líneas,
// el reemplazo deja de matchear y regional_voice_qe_test.go falla indicando
// actualizar este mapa.
var qeRegionalVoiceReplacements = [][2]string{
	{
		"When replying to the user in Spanish, use warm natural Rioplatense Spanish (voseo) without overloading the reply with slang.",
		"When replying to the user in Spanish, mirror the user's own Spanish register and regional voice; if there is no clear signal, default to neutral Latin American Spanish (tuteo: \"tú\", \"puedes\") without regional slang.",
	},
	{
		"Never inject Rioplatense slang, voseo,",
		"Never inject regional slang, voseo,",
	},
}

// qeNeutralizeRegionalVoice aplica la política de idioma del fork Gentle-QE al
// contenido de persona/output-style antes de instalarlo: la persona espeja el
// registro español del usuario y cae a español latinoamericano neutro (tuteo),
// nunca impone voseo rioplatense por defecto.
func qeNeutralizeRegionalVoice(content string) string {
	for _, repl := range qeRegionalVoiceReplacements {
		content = strings.ReplaceAll(content, repl[0], repl[1])
	}
	return content
}
