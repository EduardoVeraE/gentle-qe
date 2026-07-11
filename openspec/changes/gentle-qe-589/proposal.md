# Proposal: SDD cycle produces test-design artifacts (Gentle-QE)

## Intent

Gentle-QE ships an SDET/QE persona and QA skills, but its SDD cycle still emits
software-DEVELOPMENT artifacts (build a capability, write feature code). QE users
running SDD get dev prompts, not test design. This change makes the fork's SDD
phases emit TEST-DESIGN artifacts by default, so `propose → spec → design → tasks
→ apply → verify` yields test requirements, oracles, ISTQB strategy, scenarios,
and test code (specs/fixtures/POMs) instead of production features.

## Scope

### In Scope
- Pure REPLACEMENT of dev-oriented SDD phase content with test-oriented content
  (no mode selector, no dev/test coexistence — product decision is locked).
- Net-new per-phase markdown override assets, embedded via the existing
  `go:embed` skills glob (same shape as the 7 existing QA skills).
- New `_qe.go` helper (e.g. `qeSDDTestingContent(skillID)`) returning QE content
  when `skills.IsSDDSkill(id)` is true.
- Anchor the override in ALL THREE injector paths so no assistant silently runs
  dev prompts (see Approach).
- Register net-new asset dirs (`netNewDirs`) and the 3 Go anchors
  (`inlineAnchors`) in `tools/qe-overlay/overlay.json`.
- Per-phase dev→test reframing: propose (capability→test-requirement/oracle),
  spec (oracle-first GIVEN/WHEN/THEN), design (Testing Strategy becomes the WHOLE
  doc: levels, ISTQB techniques, pyramid, risk-based prioritization), tasks
  (scenarios/cases to automate), apply (write TEST code), verify (coverage/flaky
  framing), archive (mechanical).

### Out of Scope
- Fixing `qa-orchestrator.md` / `playwright-test-generator.md` not registered in
  overlay.json — separate task.
- Any TUI installer redesign — tracked under `gentle-qe-dl9`.
- Detailed technical design of the override mechanism — that is `sdd-design`.

## Approach

Full-content override PER PHASE (not string-diff): the `qeNeutralizeRegionalVoice`
string-map pattern does NOT scale to 150–250-line SKILL.md files that evolve to
v2.0/3.0. Ship complete QE phase assets and swap content at read time.

The override MUST be anchored in all three independent SDD injector paths:
1. `internal/components/skills/inject.go` → `InjectWithCapability` (Claude/Cursor/Kiro/Kimi).
2. `internal/components/sdd/inject.go` → step 3c (native-agent wrappers).
3. `internal/components/sdd/prompts.go` → `WriteSharedPromptFiles` (OpenCode/Kilocode).

Minimal upstream anchors only; all override entry points re-applicable via overlay.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| A missing injector path silently serves dev prompts | Med | Anchor all 3; overlay `verify` guards each anchor |
| Fork self-builds with test-only SDD (Go/TUI is real dev) | Accepted | Explicit product trade-off, accepted by user |
| Upstream SDD content drift breaks override | Med | Full-content assets decoupled from upstream text |

## Rollback

Remove net-new asset dirs and the `_qe.go` helper, revert the 3 minimal anchors,
drop their `overlay.json` entries. SDD reverts to upstream dev behavior.

## Success Criteria

- [ ] All SDD phases emit test-design artifacts across the 3 injector paths.
- [ ] `qe-overlay verify` passes with new netNewDirs + inlineAnchors registered.
- [ ] No upstream prompt/artifact CONTENT edited.
