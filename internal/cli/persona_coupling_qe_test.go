package cli

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// persona_coupling_qe_test.go — acople persona→preset del overlay Gentle-QE en
// el flujo CLI y validez de la persona/preset Dev FullStack. Net-new overlay.

// TestQECouplePersonaPreset_DevFullStackWithoutPreset: --persona dev-fullstack
// SIN --preset debe derivar PresetDevFullStack (foundationSkills), no el default
// QE (SDET).
func TestQECouplePersonaPreset_DevFullStackWithoutPreset(t *testing.T) {
	got := qeCouplePersonaPreset(model.PersonaDevFullStack, model.PresetQESDET, "")
	if got != model.PresetDevFullStack {
		t.Fatalf("qeCouplePersonaPreset(DevFullStack, _, \"\") = %q, want %q", got, model.PresetDevFullStack)
	}
}

// TestQECouplePersonaPreset_ExplicitPresetRespected: con --preset explícito se
// respeta la elección del usuario aunque la persona sea Dev FullStack.
func TestQECouplePersonaPreset_ExplicitPresetRespected(t *testing.T) {
	got := qeCouplePersonaPreset(model.PersonaDevFullStack, model.PresetQEFront, "qe-front")
	if got != model.PresetQEFront {
		t.Fatalf("qeCouplePersonaPreset(DevFullStack, qe-front, \"qe-front\") = %q, want %q", got, model.PresetQEFront)
	}
}

// TestQECouplePersonaPreset_SDETUnaffected: la persona SDET no se acopla; su
// preset resuelto pasa intacto.
func TestQECouplePersonaPreset_SDETUnaffected(t *testing.T) {
	got := qeCouplePersonaPreset(model.PersonaSDET, model.PresetQESDET, "")
	if got != model.PresetQESDET {
		t.Fatalf("qeCouplePersonaPreset(SDET, qe-sdet, \"\") = %q, want %q", got, model.PresetQESDET)
	}
}

// TestIsQEPersona_DevFullStack: la persona dev es válida en el CLI
// (--persona dev-fullstack).
func TestIsQEPersona_DevFullStack(t *testing.T) {
	if !isQEPersona(model.PersonaDevFullStack) {
		t.Fatalf("isQEPersona(PersonaDevFullStack) = false, want true")
	}
}
