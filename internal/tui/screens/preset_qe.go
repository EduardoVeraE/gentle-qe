package screens

import "github.com/gentleman-programming/gentle-ai/internal/model"

// preset_qe.go — opciones de preset del overlay Gentle-QE para el picker.
//
// Las descripciones/labels se inyectan a los maps upstream vía init() (cero
// edición inline). PresetOptions delega a qeFilterPresetOptions(opts) con una
// sola línea; el parámetro opts (la lista dev-only upstream) se ignora a
// propósito — el build QE siempre devuelve únicamente los presets QE, nunca
// un append.

func qeFilterPresetOptions(_ []model.PresetID) []model.PresetID {
	return []model.PresetID{
		model.PresetQESDET,
		model.PresetQEFront,
		model.PresetQEAPI,
		model.PresetQEPerf,
	}
}

// qePresetOptionsForBuild is the seam-aware entry the PresetOptions anchor
// calls. Seam ON (prod) → QE-only presets; seam OFF (tests) → upstream's
// shipped behavior: the dev presets with the QE presets appended, so the
// upstream preset tests pass unedited.
func qePresetOptionsForBuild(opts []model.PresetID) []model.PresetID {
	if !model.QEInstallerFlow {
		return insertBeforeCustomPreset(opts,
			model.PresetQESDET,
			model.PresetQEFront,
			model.PresetQEAPI,
			model.PresetQEPerf,
		)
	}
	return qeFilterPresetOptions(opts)
}

// insertBeforeCustomPreset inserts qePresets into opts immediately before
// model.PresetCustom rather than at the tail. PresetCustom is the manual
// "choose every component yourself" escape hatch and upstream tests (and
// upstream's own PresetOptions() callers) reasonably treat it as the LAST
// selectable preset — a plain append would push it into the middle of the
// list once QE presets are added, breaking that invariant. If Custom is not
// present in opts, qePresets are appended at the end as before.
func insertBeforeCustomPreset(opts []model.PresetID, qePresets ...model.PresetID) []model.PresetID {
	idx := -1
	for i, p := range opts {
		if p == model.PresetCustom {
			idx = i
			break
		}
	}
	if idx < 0 {
		return append(append([]model.PresetID{}, opts...), qePresets...)
	}
	result := make([]model.PresetID, 0, len(opts)+len(qePresets))
	result = append(result, opts[:idx]...)
	result = append(result, qePresets...)
	result = append(result, opts[idx:]...)
	return result
}

func init() {
	presetLabels[model.PresetQESDET] = "QE · SDET Full"
	presetLabels[model.PresetQEFront] = "QE · Frontend E2E"
	presetLabels[model.PresetQEAPI] = "QE · API"
	presetLabels[model.PresetQEPerf] = "QE · Performance"
	presetLabels[model.PresetDevFullStack] = "Dev · FullStack"

	presetDescriptions[model.PresetQESDET] = "Stack SDET completo: todas las QE skills + persona SDET"
	presetDescriptions[model.PresetQEFront] = "E2E frontend con Playwright + accesibilidad"
	presetDescriptions[model.PresetQEAPI] = "Testing de API / contract testing"
	presetDescriptions[model.PresetQEPerf] = "Performance/carga con k6"
	presetDescriptions[model.PresetDevFullStack] = "Perfil dev: skills default del upstream (auto-seleccionado con la persona Dev FullStack)"
}
