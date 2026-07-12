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
		return append(opts,
			model.PresetQESDET,
			model.PresetQEFront,
			model.PresetQEAPI,
			model.PresetQEPerf,
		)
	}
	return qeFilterPresetOptions(opts)
}

func init() {
	presetLabels[model.PresetQESDET] = "QE · SDET Full"
	presetLabels[model.PresetQEFront] = "QE · Frontend E2E"
	presetLabels[model.PresetQEAPI] = "QE · API"
	presetLabels[model.PresetQEPerf] = "QE · Performance"

	presetDescriptions[model.PresetQESDET] = "Stack SDET completo: todas las QE skills + persona SDET"
	presetDescriptions[model.PresetQEFront] = "E2E frontend con Playwright + accesibilidad"
	presetDescriptions[model.PresetQEAPI] = "Testing de API / contract testing"
	presetDescriptions[model.PresetQEPerf] = "Performance/carga con k6"
}
