package cli

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// sdd_gating_qe_test.go — predicado que decide, por persona, si la instalación
// usa el SDD test-design QE o el SDD dev upstream. Net-new overlay Gentle-QE.

// TestQEUseTestDesignSDD_OnlySDET: solo la persona SDET recibe el SDD
// test-design QE; las demás (incluida Dev FullStack) reciben el SDD dev.
func TestQEUseTestDesignSDD_OnlySDET(t *testing.T) {
	if !qeUseTestDesignSDD(model.Selection{Persona: model.PersonaSDET}) {
		t.Fatalf("qeUseTestDesignSDD(SDET) = false, want true (test-design SDD)")
	}

	devPersonas := []model.PersonaID{
		model.PersonaDevFullStack,
		model.PersonaGentleman,
		model.PersonaNeutral,
		model.PersonaCustom,
		"",
	}
	for _, p := range devPersonas {
		if qeUseTestDesignSDD(model.Selection{Persona: p}) {
			t.Fatalf("qeUseTestDesignSDD(%q) = true, want false (dev SDD)", p)
		}
	}
}
