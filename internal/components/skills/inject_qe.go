package skills

import (
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// QESDDTestingContent returns QE test-design override content for a file
// inside an SDD skill dir. skillID e.g. "sdd-apply"; fileName e.g. "SKILL.md"
// or "strict-tdd.md" (a sibling file walked alongside SKILL.md).
//
// ok=false for non-SDD ids or when no matching QE asset exists — fail-open so
// the caller falls back to serving the existing upstream (dev) content
// unchanged. This is how archive/onboard/init (no QE asset) stay
// upstream-neutral even though they are SDD-gated.
func QESDDTestingContent(skillID, fileName string) (string, bool) {
	if !IsSDDSkill(model.SkillID(skillID)) {
		return "", false
	}

	phase := strings.TrimPrefix(skillID, "sdd-") // sdd-apply -> apply
	asset := "skills/_qe-sdd/" + phase + ".md"   // SKILL.md -> phase.md
	if fileName != "SKILL.md" {
		asset = "skills/_qe-sdd/" + phase + "." + fileName // apply.strict-tdd.md
	}

	content, err := assets.Read(asset)
	if err != nil || len(content) == 0 {
		return "", false // fail-open to upstream
	}
	return content, true
}
