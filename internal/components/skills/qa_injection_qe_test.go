package skills

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/agents"
	"github.com/gentleman-programming/gentle-ai/internal/agents/claude"
	"github.com/gentleman-programming/gentle-ai/internal/agents/opencode"
	"github.com/gentleman-programming/gentle-ai/internal/agents/qwen"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// qaSkillIDs are the seven testing skills the QE overlay ships on top of
// upstream. See internal/model/types_qe.go and internal/catalog/skills_qe.go.
var qaSkillIDs = []model.SkillID{
	model.SkillQAManualISTQB,
	model.SkillQAOWASPSecurity,
	model.SkillAPITesting,
	model.SkillPlaywrightE2E,
	model.SkillA11yPlaywright,
	model.SkillK6LoadTest,
	model.SkillSeleniumE2EJava,
}

// TestInjectQASkillsToAgentsRealFilesystem is an integration-level regression
// guard for the installer bug where empty embedded assets in qa-owasp-security
// aborted the whole skills component mid-install (observed for "qwen-code").
//
// The upstream E2E shell suite only exercises the engram component for
// qwen-code and never installs the fork's QA skills, so it did not catch that
// break. This test injects every QA skill for each skills-capable agent into a
// temp HOME and asserts the on-disk oracle: every skill lands with a non-empty
// SKILL.md, no zero-byte file is written, and qa-owasp-security's sub-tree
// (references/scripts/templates) actually installs. It runs in the Unit Tests
// CI job on every PR — no Docker, no nightly gate.
func TestInjectQASkillsToAgentsRealFilesystem(t *testing.T) {
	agentsUnderTest := []struct {
		name    string
		adapter agents.Adapter
	}{
		{"qwen-code", qwen.NewAdapter()},
		{"opencode", opencode.NewAdapter()},
		{"claude-code", claude.NewAdapter()},
	}

	for _, a := range agentsUnderTest {
		t.Run(a.name, func(t *testing.T) {
			if !a.adapter.SupportsSkills() {
				t.Skipf("%s does not support skills", a.name)
			}
			home := t.TempDir()
			skillsDir := a.adapter.SkillsDir(home)
			if skillsDir == "" {
				t.Skipf("%s has no skills dir", a.name)
			}

			result, err := Inject(home, a.adapter, qaSkillIDs)
			if err != nil {
				t.Fatalf("Inject QA skills into %s failed: %v", a.name, err)
			}
			if len(result.Skipped) == len(qaSkillIDs) {
				t.Fatalf("%s: all QA skills were skipped, none installed", a.name)
			}

			// Oracle 1: each QA skill installs a non-empty SKILL.md.
			for _, id := range qaSkillIDs {
				p := filepath.Join(skillsDir, string(id), "SKILL.md")
				fi, statErr := os.Stat(p)
				if statErr != nil {
					t.Errorf("%s: missing SKILL.md for %q: %v", a.name, id, statErr)
					continue
				}
				if fi.Size() == 0 {
					t.Errorf("%s: empty SKILL.md for %q", a.name, id)
				}
			}

			// Oracle 2 (the actual regression): no installed file is zero bytes.
			// A single empty embedded asset is what aborted the install.
			_ = filepath.WalkDir(skillsDir, func(p string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				info, infoErr := d.Info()
				if infoErr != nil {
					t.Errorf("%s: cannot stat installed file %q: %v", a.name, p, infoErr)
					return nil
				}
				if info.Size() == 0 {
					t.Errorf("%s: installed empty file %q", a.name, p)
				}
				return nil
			})

			// Oracle 3: qa-owasp-security's sub-tree installs with real content.
			// This is the skill that shipped the empty .gitkeep placeholders.
			owaspDir := filepath.Join(skillsDir, string(model.SkillQAOWASPSecurity))
			for _, sub := range []string{"references", "scripts", "templates"} {
				entries, readErr := os.ReadDir(filepath.Join(owaspDir, sub))
				if readErr != nil {
					t.Errorf("%s: qa-owasp-security/%s not installed: %v", a.name, sub, readErr)
					continue
				}
				if len(entries) == 0 {
					t.Errorf("%s: qa-owasp-security/%s installed empty", a.name, sub)
				}
			}
		})
	}
}
