package skills

import "github.com/gentleman-programming/gentle-ai/internal/model"

// presets_qe.go — skills y presets del overlay SDET/QE (Gentle-QE).
//
// Aislado de presets.go (upstream). SkillsForPreset y AllSkillIDs delegan aquí
// con una sola línea cada una (puntos de anclaje vigilados por tools/qe-overlay).

// qaSkills es el set completo de skills QA/SDET del fork.
var qaSkills = []model.SkillID{
	model.SkillQAManualISTQB,
	model.SkillQAOWASPSecurity,
	model.SkillAPITesting,
	model.SkillPlaywrightE2E,
	model.SkillA11yPlaywright,
	model.SkillK6LoadTest,
	model.SkillSeleniumE2EJava,
}

// qePresetSkills mapea cada preset QE a su set de skills.
//
// PresetDevFullStack es el único preset del fork orientado a DESARROLLO: trae
// las foundationSkills default del upstream (no las QA). Las skills SDD se
// instalan aparte por el componente SDD; su flavor (dev vs test-design) lo
// decide el gate por persona, no este mapa.
var qePresetSkills = map[model.PresetID][]model.SkillID{
	model.PresetQEFront:      {model.SkillPlaywrightE2E, model.SkillA11yPlaywright},
	model.PresetQEPerf:       {model.SkillK6LoadTest},
	model.PresetQEAPI:        {model.SkillAPITesting},
	model.PresetQESDET:       qaSkills,         // stack SDET completo
	model.PresetDevFullStack: foundationSkills, // perfil dev: skills default upstream
}

// qeAllSkills devuelve todas las skills QE (para AllSkillIDs).
func qeAllSkills() []model.SkillID {
	return copySkills(qaSkills)
}

// qeSkillsForPreset devuelve las skills de un preset QE, o (nil, false) si el
// preset no es del overlay.
func qeSkillsForPreset(preset model.PresetID) ([]model.SkillID, bool) {
	ids, ok := qePresetSkills[preset]
	if !ok {
		return nil, false
	}
	return copySkills(ids), true
}
