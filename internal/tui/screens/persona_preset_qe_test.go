package screens

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// persona_preset_qe_test.go — unit tests for the Gentle-QE Persona/Preset
// option-list filters (overlay Gentle-QE; ancla qe-overlay). Net-new file:
// the upstream persona_preset_test.go is not an overlay file, so QE-only
// assertions live here instead of editing it (zero upstream content edits).

// TestPersonaOptions_QEBuildContainsOnlySDET verifies PersonaOptions() in the
// QE build returns exactly [PersonaSDET] — no dev persona ID present.
func TestPersonaOptions_QEBuildContainsOnlySDET(t *testing.T) {
	enableQESeam(t)
	got := PersonaOptions()

	want := []model.PersonaID{model.PersonaSDET}
	if len(got) != len(want) {
		t.Fatalf("PersonaOptions() = %v (len %d), want %v (len %d)", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("PersonaOptions()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	dev := []model.PersonaID{model.PersonaGentleman, model.PersonaGentlemanNeutralArtifacts, model.PersonaNeutral, model.PersonaCustom}
	for _, d := range dev {
		for _, p := range got {
			if p == d {
				t.Fatalf("PersonaOptions() = %v, must not contain dev persona %q", got, d)
			}
		}
	}
}

// TestPresetOptions_QEBuildContainsOnlyQEPresets verifies PresetOptions() in
// the QE build returns exactly the 4 QE presets — no dev preset ID present.
func TestPresetOptions_QEBuildContainsOnlyQEPresets(t *testing.T) {
	enableQESeam(t)
	got := PresetOptions()

	want := []model.PresetID{
		model.PresetQESDET,
		model.PresetQEFront,
		model.PresetQEAPI,
		model.PresetQEPerf,
	}
	if len(got) != len(want) {
		t.Fatalf("PresetOptions() = %v (len %d), want %v (len %d)", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("PresetOptions()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	dev := []model.PresetID{model.PresetFullGentleman, model.PresetEcosystemOnly, model.PresetMinimal}
	for _, d := range dev {
		for _, p := range got {
			if p == d {
				t.Fatalf("PresetOptions() = %v, must not contain dev preset %q", got, d)
			}
		}
	}
}

// TestQEFilterPersonaOptions_IgnoresDevInput triangulates: even when passed a
// non-empty dev opts slice, qeFilterPersonaOptions must return only the fixed
// QE-only list — proving the filter is not an append and does not leak dev
// entries regardless of input.
func TestQEFilterPersonaOptions_IgnoresDevInput(t *testing.T) {
	devOnly := []model.PersonaID{model.PersonaGentleman, model.PersonaCustom}
	got := qeFilterPersonaOptions(devOnly)

	if len(got) != 1 || got[0] != model.PersonaSDET {
		t.Fatalf("qeFilterPersonaOptions(%v) = %v, want [PersonaSDET]", devOnly, got)
	}
}

// TestQEFilterPresetOptions_IgnoresDevInput mirrors the persona triangulation
// for presets.
func TestQEFilterPresetOptions_IgnoresDevInput(t *testing.T) {
	devOnly := []model.PresetID{model.PresetFullGentleman, model.PresetMinimal}
	got := qeFilterPresetOptions(devOnly)

	want := []model.PresetID{model.PresetQESDET, model.PresetQEFront, model.PresetQEAPI, model.PresetQEPerf}
	if len(got) != len(want) {
		t.Fatalf("qeFilterPresetOptions(%v) = %v, want %v", devOnly, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("qeFilterPresetOptions(%v)[%d] = %q, want %q", devOnly, i, got[i], want[i])
		}
	}
}
