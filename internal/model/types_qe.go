package model

// types_qe.go — identificadores del overlay SDET/QE (Gentle-QE).
//
// Vive aparte de types.go (upstream) para que los merges del upstream nunca
// generen conflictos sobre estas definiciones. El upstream nunca crea ni toca
// este archivo. Ver el plan de overlay sostenible (Capa 1).

// Skills QA/SDET net-new (sus directorios viven en internal/assets/skills/).
const (
	SkillQAManualISTQB     SkillID = "qa-manual-istqb"
	SkillQAOWASPSecurity   SkillID = "qa-owasp-security"
	SkillAPITesting        SkillID = "api-testing"
	SkillPlaywrightE2E     SkillID = "playwright-e2e-testing"
	SkillA11yPlaywright    SkillID = "a11y-playwright-testing"
	SkillK6LoadTest        SkillID = "k6-load-test"
	SkillSeleniumE2EJava   SkillID = "selenium-e2e-java"
)

// Persona SDET (su contenido vive en internal/assets/generic/persona-sdet.md).
const (
	PersonaSDET PersonaID = "sdet"
)

// Presets QE.
const (
	PresetQEFront PresetID = "qe-front" // E2E frontend: Playwright + a11y
	PresetQEPerf  PresetID = "qe-perf"  // Performance: k6
	PresetQEAPI   PresetID = "qe-api"   // API/contract testing
	PresetQESDET  PresetID = "qe-sdet"  // Stack SDET completo: todas las QE skills + PersonaSDET
)
