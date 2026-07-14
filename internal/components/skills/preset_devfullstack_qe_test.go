package skills

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// preset_devfullstack_qe_test.go — el preset dev del fork (Dev FullStack) debe
// resolver EXACTAMENTE a las foundationSkills del upstream, sin ninguna skill
// QA. Net-new overlay Gentle-QE.

func TestQESkillsForPreset_DevFullStack_IsFoundationOnly(t *testing.T) {
	got, ok := qeSkillsForPreset(model.PresetDevFullStack)
	if !ok {
		t.Fatalf("qeSkillsForPreset(PresetDevFullStack) ok=false, want true")
	}

	if len(got) != len(foundationSkills) {
		t.Fatalf("got %d skills %v, want %d foundationSkills %v", len(got), got, len(foundationSkills), foundationSkills)
	}
	for i := range foundationSkills {
		if got[i] != foundationSkills[i] {
			t.Fatalf("skill[%d] = %q, want %q", i, got[i], foundationSkills[i])
		}
	}

	// Ninguna skill QA debe filtrarse en el perfil dev.
	for _, qa := range qaSkills {
		for _, s := range got {
			if s == qa {
				t.Fatalf("PresetDevFullStack must not include QA skill %q, got %v", qa, got)
			}
		}
	}
}

// TestSkillsForPreset_DevFullStack_RoutesThroughOverlay verifica que el punto de
// entrada público (usado por run.go) resuelve el preset dev vía el overlay.
func TestSkillsForPreset_DevFullStack_RoutesThroughOverlay(t *testing.T) {
	got := SkillsForPreset(model.PresetDevFullStack)
	if len(got) != len(foundationSkills) {
		t.Fatalf("SkillsForPreset(PresetDevFullStack) = %v, want foundationSkills %v", got, foundationSkills)
	}
}
