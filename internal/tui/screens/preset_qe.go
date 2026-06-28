package screens

import "github.com/gentleman-programming/gentle-ai/internal/model"

// preset_qe.go — opciones de preset del overlay Gentle-QE para el picker.
//
// Las descripciones/labels se inyectan a los maps upstream vía init() (cero
// edición inline). PresetOptions appendea qePresetOptions() con una sola línea.

func qePresetOptions() []model.PresetID {
	return []model.PresetID{
		model.PresetQESDET,
		model.PresetQEFront,
		model.PresetQEAPI,
		model.PresetQEPerf,
	}
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
