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
// qeTestDesignSDDEnabled gates the QE test-design SDD override. Default true
// preserves the fork's historical behavior (override ON) for every caller that
// does not opt out — including the direct-Inject tests in this package and the
// production SDET install. The cli layer flips it per install according to the
// selected persona (SetQETestDesignSDD): SDET keeps the QE test-design SDD;
// dev personas (Dev FullStack) fall back to the upstream dev SDD.
//
// It is a package global mirroring model.QEInstallerFlow. The cli layer sets it
// deterministically on every run, so it never depends on a previous run's
// value; tests that flip it MUST restore it via t.Cleanup.
var qeTestDesignSDDEnabled = true

// SetQETestDesignSDD toggles the QE test-design SDD override for subsequent
// QESDDTestingContent calls. Called by the cli install/sync flow.
func SetQETestDesignSDD(on bool) {
	qeTestDesignSDDEnabled = on
}

func QESDDTestingContent(skillID, fileName string) (string, bool) {
	if !qeTestDesignSDDEnabled {
		return "", false // gated OFF (e.g. dev persona) → fall open to upstream dev SDD
	}
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
