package persona

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// TestInjectSDETPersonaPreservesSpanishFirstContract es el contrato de idioma
// end-to-end del overlay Gentle-QE: al instalar la persona SDET, el AGENTS.md
// resultante debe conservar el contrato español-first (tuteo, sin voseo ni slang
// regional). Si un merge de upstream o un cambio en el wiring de persona rompe
// la inyección de la sección Language, este test falla en vez de degradar en
// silencio el posicionamiento español-first del fork.
func TestInjectSDETPersonaPreservesSpanishFirstContract(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter(), model.PersonaSDET)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false; la persona SDET no se instaló")
	}

	content, err := os.ReadFile(filepath.Join(home, ".config", "opencode", "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)

	// El contrato español-first debe sobrevivir intacto en el artefacto instalado.
	for _, want := range []string{
		"## Language",
		"neutral, formal Spanish",
		`tuteo: "tú", "puedes", "mira"`,
		"No regional slang or voseo",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("persona SDET instalada perdió el contrato de idioma: falta %q\ncontenido:\n%s", want, text)
		}
	}

	// Guarda explícita contra el voseo rioplatense en el contrato instalado: la
	// presencia de estas formas indicaría que se filtró la persona equivocada.
	for _, voseo := range []string{"querés", "podés", "tenés", "Rioplatense"} {
		if strings.Contains(text, voseo) {
			t.Fatalf("persona SDET instalada contiene voseo/registro equivocado %q:\n%s", voseo, text)
		}
	}
}
