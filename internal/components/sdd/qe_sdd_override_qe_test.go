package sdd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/components/skills"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// qePhaseHeadings maps each of the 7 QE-covered SDD phases to the required
// structural section headings for its asset (design verification step 2,
// "SDD QE Phase Contract" spec). These are exact heading LINES, not loose
// keywords — this is what makes the oracle structural instead of a keyword
// search (see TestQEStructuralOracleRejectsKeywordStuffing below).
var qePhaseHeadings = map[string][]string{
	"sdd-explore": {
		"## Risk Assessment",
		"## Testability Assessment",
		"## Defect Clustering (80/20)",
		"## Oracle Inventory",
	},
	"sdd-propose": {
		"## Test Requirement",
	},
	"sdd-spec": {
		"## Oracle-First Requirement Format",
	},
	"sdd-design": {
		"## Test Levels",
		"## ISTQB Techniques",
		"## Test Pyramid",
		"## Risk-Based Prioritization And Defect Clustering",
	},
	"sdd-tasks": {
		"## Scenarios By Test Level",
	},
	"sdd-apply": {
		"## Test Code Deliverable",
	},
	"sdd-verify": {
		"## Execution And Coverage",
		"## Flaky Test Policy",
	},
}

var qeAllSevenPhases = []string{
	"sdd-explore", "sdd-propose", "sdd-spec", "sdd-design",
	"sdd-tasks", "sdd-apply", "sdd-verify",
}

// hasHeading reports whether content contains heading as an exact,
// whitespace-trimmed line. This is the structural primitive: a document that
// merely mentions the keywords in a footer paragraph (no real heading line)
// does NOT satisfy this check.
func hasHeading(content, heading string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == heading {
			return true
		}
	}
	return false
}

func requireHeadings(t *testing.T, content string, headings []string, label string) {
	t.Helper()
	for _, want := range headings {
		if !hasHeading(content, want) {
			t.Errorf("%s: missing required structural heading %q", label, want)
		}
	}
}

func requireAbsent(t *testing.T, content, marker, label string) {
	t.Helper()
	if strings.Contains(content, marker) {
		t.Errorf("%s: dev-only marker %q leaked into QE content", label, marker)
	}
}

// --- Scenario: Oracle test proves QE override across all injector paths ---

// TestQEStructuralOracleAllSevenPhases drives the shared QESDDTestingContent
// helper (the function every one of the 3 injector paths calls) across the 7
// phases that ship a QE asset, and asserts the structural headings required
// by the "SDD QE Phase Contract" spec are present.
func TestQEStructuralOracleAllSevenPhases(t *testing.T) {
	for _, phase := range qeAllSevenPhases {
		t.Run(phase, func(t *testing.T) {
			content, ok := skills.QESDDTestingContent(phase, "SKILL.md")
			if !ok {
				t.Fatalf("QESDDTestingContent(%q, SKILL.md) ok = false, want true", phase)
			}
			requireHeadings(t, content, qePhaseHeadings[phase], phase)
		})
	}
}

// TestQEStructuralOracleRejectsKeywordStuffing proves the oracle is
// structural, not a keyword search: a candidate document that contains the
// required keywords only loosely (pasted into a footer, no real heading
// line) MUST fail the same check that a legitimate QE asset passes.
func TestQEStructuralOracleRejectsKeywordStuffing(t *testing.T) {
	stuffed := "# Some Document\n\n" +
		"This document mentions Test Levels, ISTQB Techniques, Test Pyramid and " +
		"Risk-Based Prioritization And Defect Clustering only as loose words in a " +
		"footer paragraph, never as real section headings.\n"

	for _, want := range qePhaseHeadings["sdd-design"] {
		if hasHeading(stuffed, want) {
			t.Fatalf("hasHeading matched a keyword-stuffed footer for %q — oracle is not structural", want)
		}
	}

	legit, ok := skills.QESDDTestingContent("sdd-design", "SKILL.md")
	if !ok {
		t.Fatal("QESDDTestingContent(sdd-design, SKILL.md) ok = false")
	}
	requireHeadings(t, legit, qePhaseHeadings["sdd-design"], "sdd-design (legit)")
}

// --- Negative markers: phase-specific and dev-verified ---

const (
	// negMarkerImplementCodeChanges is dev-verified present in
	// internal/assets/claude/agents/sdd-apply.md (line 4, the frontmatter
	// description). It must never appear in Path 2's QE-swapped output for
	// sdd-apply.
	negMarkerImplementCodeChanges = "Implement code changes"

	// negMarkerProductionCode is dev-verified present in
	// internal/assets/skills/sdd-apply/strict-tdd.md and
	// internal/assets/skills/sdd-verify/strict-tdd-verify.md. It must never
	// appear in the QE strict-TDD sibling assets.
	negMarkerProductionCode = "production code"
)

// TestQENegativeMarkersAreDevVerified encodes the "Negative markers are
// phase-specific and dev-verified" scenario: each marker used elsewhere in
// this file must actually occur in the real upstream dev asset it guards,
// and must be absent from the corresponding QE asset.
func TestQENegativeMarkersAreDevVerified(t *testing.T) {
	devWrapper, err := os.ReadFile(filepath.Join("..", "..", "assets", "claude", "agents", "sdd-apply.md"))
	if err != nil {
		t.Fatalf("read upstream sdd-apply wrapper: %v", err)
	}
	if !strings.Contains(string(devWrapper), negMarkerImplementCodeChanges) {
		t.Fatalf("negative marker %q not verified present in upstream sdd-apply wrapper — drop it per spec", negMarkerImplementCodeChanges)
	}

	devApplyTDD, err := os.ReadFile(filepath.Join("..", "..", "assets", "skills", "sdd-apply", "strict-tdd.md"))
	if err != nil {
		t.Fatalf("read upstream sdd-apply strict-tdd.md: %v", err)
	}
	if !strings.Contains(string(devApplyTDD), negMarkerProductionCode) {
		t.Fatalf("negative marker %q not verified present in upstream sdd-apply/strict-tdd.md", negMarkerProductionCode)
	}

	devVerifyTDD, err := os.ReadFile(filepath.Join("..", "..", "assets", "skills", "sdd-verify", "strict-tdd-verify.md"))
	if err != nil {
		t.Fatalf("read upstream sdd-verify strict-tdd-verify.md: %v", err)
	}
	if !strings.Contains(string(devVerifyTDD), negMarkerProductionCode) {
		t.Fatalf("negative marker %q not verified present in upstream sdd-verify/strict-tdd-verify.md", negMarkerProductionCode)
	}

	qeApplyTDD, ok := skills.QESDDTestingContent("sdd-apply", "strict-tdd.md")
	if !ok {
		t.Fatal("QESDDTestingContent(sdd-apply, strict-tdd.md) ok = false")
	}
	requireAbsent(t, qeApplyTDD, negMarkerProductionCode, "sdd-apply strict-tdd (QE)")

	qeVerifyTDD, ok := skills.QESDDTestingContent("sdd-verify", "strict-tdd-verify.md")
	if !ok {
		t.Fatal("QESDDTestingContent(sdd-verify, strict-tdd-verify.md) ok = false")
	}
	requireAbsent(t, qeVerifyTDD, negMarkerProductionCode, "sdd-verify strict-tdd-verify (QE)")
}

// --- Path 1: internal/components/skills/inject.go -> InjectWithCapability ---

func TestQEOverride_Path1_SkillsInjectionSwapsAllPhasesAndSiblings(t *testing.T) {
	home := t.TempDir()

	ids := []model.SkillID{
		model.SkillSDDExplore, model.SkillSDDPropose, model.SkillSDDSpec,
		model.SkillSDDDesign, model.SkillSDDTasks, model.SkillSDDApply,
		model.SkillSDDVerify,
	}

	result, err := skills.InjectWithCapability(home, claudeAdapter(), ids, "capable")
	if err != nil {
		t.Fatalf("InjectWithCapability() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("InjectWithCapability() changed = false")
	}

	for _, id := range ids {
		phase := string(id)
		path := filepath.Join(home, ".claude", "skills", phase, "SKILL.md")
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("ReadFile(%s) error = %v", path, readErr)
		}
		requireHeadings(t, string(content), qePhaseHeadings[phase], "Path1 "+phase)
	}

	// Strict-TDD siblings are walked alongside SKILL.md for sdd-apply/sdd-verify.
	applyTDDPath := filepath.Join(home, ".claude", "skills", "sdd-apply", "strict-tdd.md")
	applyTDD, err := os.ReadFile(applyTDDPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", applyTDDPath, err)
	}
	requireAbsent(t, string(applyTDD), negMarkerProductionCode, "Path1 sdd-apply/strict-tdd.md")
	if !strings.Contains(string(applyTDD), "RED") || !strings.Contains(string(applyTDD), "GREEN") || !strings.Contains(string(applyTDD), "REFACTOR") {
		t.Fatalf("Path1 sdd-apply/strict-tdd.md missing RED/GREEN/REFACTOR automation framing:\n%s", applyTDD)
	}

	verifyTDDPath := filepath.Join(home, ".claude", "skills", "sdd-verify", "strict-tdd-verify.md")
	verifyTDD, err := os.ReadFile(verifyTDDPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", verifyTDDPath, err)
	}
	requireAbsent(t, string(verifyTDD), negMarkerProductionCode, "Path1 sdd-verify/strict-tdd-verify.md")
}

// --- Path 2: internal/components/sdd/inject.go step 3c (native sub-agent wrappers) ---

func TestQEOverride_Path2_NativeSubAgentBodySwapPreservesFrontmatter(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject() changed = false")
	}

	path := filepath.Join(home, ".claude", "agents", "sdd-apply.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	text := string(content)

	if !strings.Contains(text, "name: sdd-apply") {
		t.Errorf("Path2 wrapper missing preserved name: line:\n%s", text)
	}
	if strings.Contains(text, "{{CLAUDE_MODEL}}") {
		t.Errorf("Path2 wrapper still contains unresolved {{CLAUDE_MODEL}} placeholder")
	}
	if !strings.Contains(text, "model: ") {
		t.Errorf("Path2 wrapper missing resolved model: line:\n%s", text)
	}
	if !strings.Contains(text, "tools: ") {
		t.Errorf("Path2 wrapper missing preserved tools: line:\n%s", text)
	}

	wantDesc := qeSubAgentDescription("sdd-apply")
	if !strings.Contains(text, strings.Trim(wantDesc, `"`)) {
		t.Errorf("Path2 wrapper description not rewritten to QE description; want substring %q in:\n%s", wantDesc, text)
	}

	requireAbsent(t, text, negMarkerImplementCodeChanges, "Path2 sdd-apply wrapper")
	requireHeadings(t, text, qePhaseHeadings["sdd-apply"], "Path2 sdd-apply body")
}

// TestQEOverride_Path2_NoDuplicateFrontmatterOrSectionMarkers is the
// regression oracle for the BLOCKER: Path 2 must never paste a QE asset's OWN
// frontmatter as raw body text (which produces a wrapper with two frontmatter
// blocks), and must never leak the <!-- section:model-capable/small -->
// markers into the served output — those markers must be resolved by
// extractModelSection before qeSwapNativeAgentBody runs, exactly like
// Path 1 (skills.InjectWithCapability) and Path 3 (WriteSharedPromptFiles).
func TestQEOverride_Path2_NoDuplicateFrontmatterOrSectionMarkers(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject() changed = false")
	}

	agentsDir := filepath.Join(home, ".claude", "agents")

	for _, phase := range []string{"sdd-explore", "sdd-propose", "sdd-spec", "sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify"} {
		path := filepath.Join(agentsDir, phase+".md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", path, err)
		}
		text := string(content)

		fenceCount := 0
		for _, line := range strings.Split(text, "\n") {
			if line == "---" {
				fenceCount++
			}
		}
		if fenceCount != 2 {
			t.Errorf("%s: expected exactly 2 frontmatter fence lines (one pair, no duplicated QE frontmatter), got %d:\n%s", phase, fenceCount, text)
		}

		for _, marker := range []string{
			"<!-- section:model-capable -->", "<!-- /section:model-capable -->",
			"<!-- section:model-small -->", "<!-- /section:model-small -->",
		} {
			if strings.Contains(text, marker) {
				t.Errorf("%s: leaked model-section marker %q into served output; extractModelSection must run before qeSwapNativeAgentBody:\n%s", phase, marker, text)
			}
		}
	}

	// sdd-apply and sdd-verify are the two QE assets that carry BOTH a
	// model-capable and model-small variant. The default capability
	// ("capable", used when no Claude phase assignment overrides it) must
	// select exactly ONE variant — not both variants concatenated.
	applyPath := filepath.Join(agentsDir, "sdd-apply.md")
	applyContent, err := os.ReadFile(applyPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", applyPath, err)
	}
	applyText := string(applyContent)
	if got := strings.Count(applyText, "## Test Code Deliverable"); got != 1 {
		t.Errorf("sdd-apply: expected exactly 1 occurrence of the model-capable-only heading %q (proves only one variant was served), got %d:\n%s", "## Test Code Deliverable", got, applyText)
	}

	verifyPath := filepath.Join(agentsDir, "sdd-verify.md")
	verifyContent, err := os.ReadFile(verifyPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", verifyPath, err)
	}
	verifyText := string(verifyContent)
	if got := strings.Count(verifyText, "## Execution And Coverage"); got != 1 {
		t.Errorf("sdd-verify: expected exactly 1 occurrence of the model-capable-only heading %q (proves only one variant was served), got %d:\n%s", "## Execution And Coverage", got, verifyText)
	}
}

// --- Path 3: internal/components/sdd/prompts.go -> WriteSharedPromptFiles ---

func TestQEOverride_Path3_SharedPromptFilesSwapsBeforeCodeGraphGuidance(t *testing.T) {
	home := t.TempDir()

	guidance := "## CodeGraph Guidance\nUse codegraph_explore before broad reads."
	changed, err := WriteSharedPromptFiles(home, nil, guidance)
	if err != nil {
		t.Fatalf("WriteSharedPromptFiles() error = %v", err)
	}
	if !changed {
		t.Fatal("WriteSharedPromptFiles() changed = false")
	}

	path := filepath.Join(SharedPromptDir(home), "sdd-design.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	text := string(content)

	requireHeadings(t, text, qePhaseHeadings["sdd-design"], "Path3 sdd-design")
	if !strings.Contains(text, "CodeGraph Guidance") {
		t.Errorf("Path3 sdd-design.md missing injected CodeGraph guidance (must run AFTER the QE swap):\n%s", text)
	}
}

// --- Gate / fail-open: archive, onboard, init, and non-SDD ids ---

func TestQEOverride_GateFailOpenForNonQEPhases(t *testing.T) {
	tests := []string{"sdd-archive", "sdd-onboard", "sdd-init", "judgment-day", "go-testing"}
	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			content, ok := skills.QESDDTestingContent(id, "SKILL.md")
			if ok {
				t.Fatalf("QESDDTestingContent(%q, SKILL.md) ok = true, want false (fail-open to upstream); got content len=%d", id, len(content))
			}
			if content != "" {
				t.Fatalf("QESDDTestingContent(%q, SKILL.md) content = %q, want empty on ok=false", id, content)
			}
		})
	}
}

// --- Strict-TDD oracle ---

func TestQEOverride_StrictTDDOracle(t *testing.T) {
	applyTDD, ok := skills.QESDDTestingContent("sdd-apply", "strict-tdd.md")
	if !ok {
		t.Fatal("QESDDTestingContent(sdd-apply, strict-tdd.md) ok = false, want true")
	}
	for _, want := range []string{"RED", "GREEN", "REFACTOR"} {
		if !strings.Contains(applyTDD, want) {
			t.Errorf("sdd-apply strict-tdd.md missing %q automation framing", want)
		}
	}
	requireAbsent(t, applyTDD, negMarkerProductionCode, "sdd-apply strict-tdd.md")

	verifyTDD, ok := skills.QESDDTestingContent("sdd-verify", "strict-tdd-verify.md")
	if !ok {
		t.Fatal("QESDDTestingContent(sdd-verify, strict-tdd-verify.md) ok = false, want true")
	}
	requireAbsent(t, verifyTDD, negMarkerProductionCode, "sdd-verify strict-tdd-verify.md")

	if _, ok := skills.QESDDTestingContent("sdd-design", "strict-tdd.md"); ok {
		t.Fatal("QESDDTestingContent(sdd-design, strict-tdd.md) ok = true, want false (no such QE sibling asset)")
	}
}

// --- qeSwapNativeAgentBody parser unit tests (design: Path 2 fragility risk) ---

const qeWrapperFoldedDescription = `---
name: sdd-apply
description: >
  Implement code changes from task definitions. Use when tasks are ready and implementation
  should begin. Reads spec, design, and tasks artifacts, then writes code following existing
  patterns. Marks tasks complete as it goes.
model: claude-opus-4-6
effort: high
tools: Read, Edit, Write, Glob, Grep, Bash
---

You are the SDD **apply** executor. Do this phase's work yourself.
`

func TestQeSwapNativeAgentBody_FoldedDescriptionReplaced(t *testing.T) {
	out := qeSwapNativeAgentBody(qeWrapperFoldedDescription, "## QE Body\n\nAutomate the scenario.", `"QE apply description."`)

	if strings.Contains(out, "Implement code changes") {
		t.Errorf("folded description was not replaced:\n%s", out)
	}
	if !strings.Contains(out, `description: "QE apply description."`) {
		t.Errorf("new description missing:\n%s", out)
	}
	if !strings.Contains(out, "## QE Body") {
		t.Errorf("body was not replaced with qeBody:\n%s", out)
	}
	if !strings.Contains(out, "name: sdd-apply") {
		t.Errorf("name: line lost:\n%s", out)
	}
	if !strings.Contains(out, "model: claude-opus-4-6") {
		t.Errorf("model: line lost:\n%s", out)
	}
	if !strings.Contains(out, "effort: high") {
		t.Errorf("effort: line lost:\n%s", out)
	}
	if !strings.Contains(out, "tools: Read, Edit, Write, Glob, Grep, Bash") {
		t.Errorf("tools: line lost:\n%s", out)
	}
}

// qeWrapperFoldedDescriptionWithBlankLine reproduces a valid YAML folded
// scalar (`description: >`) whose value spans two indented paragraphs
// separated by a blank line — a blank line inside a folded/literal block
// scalar does NOT terminate it; it becomes a paragraph break in the folded
// value. The continuation-consuming loop must not treat that blank line as
// the end of the description and leave the second paragraph as an orphan,
// invalid-YAML line in the middle of the frontmatter block.
const qeWrapperFoldedDescriptionWithBlankLine = `---
name: sdd-apply
description: >
  Implement code changes from task definitions. Use when tasks are ready and
  implementation should begin.

  This second paragraph continues the same folded scalar after a blank line.
model: claude-opus-4-6
tools: Read, Edit, Write, Glob, Grep, Bash
---

You are the SDD **apply** executor. Do this phase's work yourself.
`

func TestQeSwapNativeAgentBody_FoldedDescriptionBlankLineContinuation(t *testing.T) {
	out := qeSwapNativeAgentBody(qeWrapperFoldedDescriptionWithBlankLine, "## QE Body\n\nAutomate the scenario.", `"QE apply description."`)

	if strings.Contains(out, "Implement code changes") {
		t.Errorf("first paragraph of folded description was not replaced:\n%s", out)
	}
	if strings.Contains(out, "This second paragraph") {
		t.Errorf("second paragraph (after blank line) leaked as an orphan frontmatter line — blank line incorrectly ended continuation:\n%s", out)
	}
	if !strings.Contains(out, `description: "QE apply description."`) {
		t.Errorf("new description missing:\n%s", out)
	}
	if !strings.Contains(out, "name: sdd-apply") || !strings.Contains(out, "model: claude-opus-4-6") || !strings.Contains(out, "tools: Read, Edit, Write, Glob, Grep, Bash") {
		t.Errorf("frontmatter fields not preserved:\n%s", out)
	}

	fenceCount := 0
	for _, line := range strings.Split(out, "\n") {
		if line == "---" {
			fenceCount++
		}
	}
	if fenceCount != 2 {
		t.Errorf("expected exactly 2 frontmatter fence lines (valid YAML), got %d — orphan line likely broke the frontmatter block:\n%s", fenceCount, out)
	}
}

func TestQeSwapNativeAgentBody_NoFenceIsDefensiveNoOp(t *testing.T) {
	noFence := "Just a plain markdown file with no frontmatter at all.\n"
	out := qeSwapNativeAgentBody(noFence, "## QE Body", `"QE description."`)
	if out != noFence {
		t.Fatalf("expected defensive no-op for input with no frontmatter fence, got:\n%s", out)
	}
}

// qeBodyWithOwnFrontmatter reproduces the shape every real _qe-sdd/{phase}.md
// asset has: its own YAML frontmatter fence, followed by the real body. This
// is the exact BLOCKER shape — qeSwapNativeAgentBody must strip this
// frontmatter, not paste it raw as wrapper body text.
const qeBodyWithOwnFrontmatter = `---
name: sdd-apply
description: "QE apply description carried by the QE asset itself."
disable-model-invocation: true
---

## Test Code Deliverable

Automate the assigned scenarios.
`

func TestQeSwapNativeAgentBody_StripsQEAssetsOwnFrontmatter(t *testing.T) {
	out := qeSwapNativeAgentBody(qeWrapperFoldedDescription, qeBodyWithOwnFrontmatter, `"QE apply description."`)

	fenceCount := 0
	for _, line := range strings.Split(out, "\n") {
		if line == "---" {
			fenceCount++
		}
	}
	if fenceCount != 2 {
		t.Errorf("expected exactly 2 frontmatter fence lines (wrapper's own, QE body's own stripped), got %d:\n%s", fenceCount, out)
	}
	if strings.Contains(out, "disable-model-invocation: true") {
		t.Errorf("QE asset's own frontmatter field leaked into output — qeStripOwnFrontmatter did not strip it:\n%s", out)
	}
	if strings.Contains(out, `description: "QE apply description carried by the QE asset itself."`) {
		t.Errorf("QE asset's own description field leaked into output body:\n%s", out)
	}
	if !strings.Contains(out, "## Test Code Deliverable") {
		t.Errorf("QE body content missing after frontmatter strip:\n%s", out)
	}
	if !strings.Contains(out, "name: sdd-apply") {
		t.Errorf("wrapper's own name: line lost:\n%s", out)
	}
}

func TestQeSwapNativeAgentBody_SingleLineDescriptionReplaced(t *testing.T) {
	wrapper := "---\n" +
		"name: sdd-design\n" +
		`description: "Create the technical design document."` + "\n" +
		"model: claude-sonnet-4-6\n" +
		"tools: Read, Edit, Write\n" +
		"---\n\n" +
		"Dev body content.\n"

	out := qeSwapNativeAgentBody(wrapper, "## QE Design Body", `"QE design description."`)

	if strings.Contains(out, "Create the technical design document.") {
		t.Errorf("single-line description was not replaced:\n%s", out)
	}
	if !strings.Contains(out, `description: "QE design description."`) {
		t.Errorf("new description missing:\n%s", out)
	}
	if !strings.Contains(out, "name: sdd-design") || !strings.Contains(out, "model: claude-sonnet-4-6") || !strings.Contains(out, "tools: Read, Edit, Write") {
		t.Errorf("frontmatter fields not preserved:\n%s", out)
	}
	if !strings.Contains(out, "## QE Design Body") {
		t.Errorf("body not replaced:\n%s", out)
	}
}

func TestQeSwapNativeAgentBody_PreservesFrontmatterFieldsByteIdentical(t *testing.T) {
	out := qeSwapNativeAgentBody(qeWrapperFoldedDescription, "## QE Body", `"QE apply description."`)

	wantLines := []string{
		"name: sdd-apply",
		"model: claude-opus-4-6",
		"effort: high",
		"tools: Read, Edit, Write, Glob, Grep, Bash",
	}
	for _, want := range wantLines {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if line == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("frontmatter field line %q not preserved byte-identical in output:\n%s", want, out)
		}
	}
}

func TestQeSubAgentDescription_CoversAllSevenPhasesAndDefault(t *testing.T) {
	for _, phase := range qeAllSevenPhases {
		if got := qeSubAgentDescription(phase); got == "" || got == `"QE test-design phase for a Gentle-QE SDD change."` {
			t.Errorf("qeSubAgentDescription(%q) returned the generic default, want a phase-specific description", phase)
		}
	}
	if got := qeSubAgentDescription("sdd-archive"); got != `"QE test-design phase for a Gentle-QE SDD change."` {
		t.Errorf("qeSubAgentDescription(sdd-archive) = %q, want the generic default (archive has no QE asset)", got)
	}
}
