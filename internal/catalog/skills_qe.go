package catalog

import "github.com/gentleman-programming/gentle-ai/internal/model"

// skills_qe.go — registro de las skills QA/SDET del overlay Gentle-QE.
//
// Se agregan al catálogo upstream vía init()+append, sin editar el slice
// mvpSkills de skills.go. Así un merge del upstream nunca conflictúa aquí.

var qaSkills = []Skill{
	{ID: model.SkillQAManualISTQB, Name: "qa-manual-istqb", Category: "qa", Priority: "p0"},
	{ID: model.SkillQAOWASPSecurity, Name: "qa-owasp-security", Category: "qa", Priority: "p0"},
	{ID: model.SkillAPITesting, Name: "api-testing", Category: "qa", Priority: "p0"},
	{ID: model.SkillPlaywrightE2E, Name: "playwright-e2e-testing", Category: "qa", Priority: "p0"},
	{ID: model.SkillA11yPlaywright, Name: "a11y-playwright-testing", Category: "qa", Priority: "p0"},
	{ID: model.SkillK6LoadTest, Name: "k6-load-test", Category: "qa", Priority: "p0"},
	{ID: model.SkillSeleniumE2EJava, Name: "selenium-e2e-java", Category: "qa", Priority: "p0"},
}

func init() {
	mvpSkills = append(mvpSkills, qaSkills...)
}
