package screens

import (
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// persona_label_qe_test.go — tests del helper de label de persona (overlay QE).
// Net-new: cubre el comportamiento del renombre visible (SDET en mayúsculas) y
// protege contra regresión a string(persona) crudo. Registrado en overlay.json.

// TestQEPersonaLabel_SeamOn_UppercasesSDET: con el seam ON (build QE), el label
// de la persona SDET se muestra en mayúsculas.
func TestQEPersonaLabel_SeamOn_UppercasesSDET(t *testing.T) {
	enableQESeam(t)
	if got := qePersonaLabel(model.PersonaSDET); got != "SDET" {
		t.Fatalf("qePersonaLabel(SDET) seam ON = %q, want %q", got, "SDET")
	}
}

// TestQEPersonaLabel_SeamOff_ReturnsRaw: con el seam OFF (parity upstream), el
// helper devuelve el string crudo — para SDET y para las personas upstream, así
// los contract tests que esperan labels literales no se rompen.
func TestQEPersonaLabel_SeamOff_ReturnsRaw(t *testing.T) {
	// seam OFF es el default del paquete (TestMain); no llamamos enableQESeam.
	if got := qePersonaLabel(model.PersonaSDET); got != string(model.PersonaSDET) {
		t.Fatalf("qePersonaLabel(SDET) seam OFF = %q, want raw %q", got, string(model.PersonaSDET))
	}
	if got := qePersonaLabel(model.PersonaGentlemanNeutralArtifacts); got != "gentleman-neutral-artifacts" {
		t.Fatalf("qePersonaLabel(gentleman-neutral) = %q, want raw literal", got)
	}
}

// TestRenderPersona_SeamOn_ShowsUppercaseSDET: la pantalla renderizada en build
// QE muestra "SDET" y NO el id crudo "sdet" como label (guard de regresión).
func TestRenderPersona_SeamOn_ShowsUppercaseSDET(t *testing.T) {
	enableQESeam(t)
	out := RenderPersona(model.PersonaSDET, 0)
	if !strings.Contains(out, "SDET") {
		t.Fatalf("RenderPersona seam ON debe contener %q, got:\n%s", "SDET", out)
	}
	if strings.Contains(out, "sdet") {
		t.Fatalf("RenderPersona seam ON no debe contener el label crudo %q, got:\n%s", "sdet", out)
	}
}
