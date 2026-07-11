# Tasks: SDD cycle produces test-design artifacts (Gentle-QE)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~950-1250 (9 assets ~120-250 lines each + 2 helpers ~60 lines + 4 anchors ~5-15 lines each + cross-path test ~250 lines + unit test ~80 lines + overlay.json) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (assets+helpers) -> PR 2 (4 anchors+overlay) -> PR 3 (tests) |
| Delivery strategy | ask-on-risk |
| Chain strategy | feature-branch-chain |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | 9 net-new QE assets + 2 helpers (Phases 1-2) | PR 1 | Base = feature/tracker branch. No upstream file touched. Independently mergeable/reviewable content unit. |
| 2 | 4 minimal anchors + overlay.json registration (Phase 3) | PR 2 | Base = PR 1 branch. Small, mechanical, highest scrutiny (touches upstream files). |
| 3 | Cross-path test + parser unit tests + `qe-overlay verify` (Phase 4) | PR 3 | Base = PR 2 branch. Oracle-heavy; verifies units 1-2 end-to-end. |

## Phase 1: Net-New QE Phase Assets (Foundation)

Traceability: `sdd-qe-phase-contract` spec (all requirements), design Layer 2.

- [x] 1.1 Create `internal/assets/skills/_qe-sdd/explore.md` — risk (likelihood x impact), testability, defect-clustering (80/20), oracle inventory (existing vs. missing). No feature-feasibility section. (Req: Explore-Phase Content Frames Quality Risk And Testability)
- [x] 1.2 Create `internal/assets/skills/_qe-sdd/propose.md` — capability -> test-requirement + candidate oracle + risk statement; no "capability as shippable feature" section. (Req: Propose-Phase Content Frames Test-Requirements)
- [x] 1.3 Create `internal/assets/skills/_qe-sdd/spec.md` — oracle-first GIVEN/WHEN/THEN; every requirement states its oracle explicitly. (Req: Spec-Phase Requirements MUST State A Verifiable Oracle)
- [x] 1.4 Create `internal/assets/skills/_qe-sdd/design.md` — Testing Strategy as the whole document: named test levels (unit/integration/system/acceptance), named ISTQB technique per non-trivial requirement, test-pyramid reference, risk+defect-clustering prioritization. (Req: Design-Phase Content Is An ISTQB Test Strategy)
- [x] 1.5 Create `internal/assets/skills/_qe-sdd/tasks.md` — GIVEN/WHEN/THEN scenarios/cases grouped by test level; no implementation-subtask framing. (Req: Tasks-Phase Content Enumerates Automatable Scenarios)
- [x] 1.6 Create `internal/assets/skills/_qe-sdd/apply.md` — test code deliverable (specs/fixtures/POMs) referencing fork QA skills (`playwright-e2e-testing`, `a11y-playwright-testing`, `k6-load-test`, `api-testing`, `qa-owasp-security`, `qa-manual-istqb`); no production-capability framing. (Req: Apply/Verify-Phase Content Targets Test Code)
- [x] 1.7 Create `internal/assets/skills/_qe-sdd/verify.md` — execution + coverage-by-risk + flaky-test detection; flaky test MUST be fixed or removed, never ignored. (Req: Apply/Verify-Phase Content Targets Test Code And Flakiness)
- [x] 1.8 Create `internal/assets/skills/_qe-sdd/apply.strict-tdd.md` — RED = failing assertion/oracle; GREEN = minimal page-object/fixture/helper; REFACTOR = clean test-support code without breaking the oracle; zero production-code instructions. (Req: Strict-TDD Modules Reframe To Test-Automation TDD)
- [x] 1.9 Create `internal/assets/skills/_qe-sdd/verify.strict-tdd-verify.md` — confirms each RED oracle genuinely failed pre-GREEN, no vacuous pass; zero production-code instructions. (Req: Strict-TDD Modules Reframe To Test-Automation TDD)
- [x] 1.10 Do NOT create `archive.md`, `onboard.md`, or `init.md` in `_qe-sdd/` — confirm directory has exactly 9 files. (Req: Archive-Phase Content Remains Mechanical; fail-open scenario)

## Phase 2: Helpers (depends on Phase 1 assets existing)

- [x] 2.1 Create `internal/components/skills/inject_qe.go` with `QESDDTestingContent(skillID, fileName string) (string, bool)`: gate on `IsSDDSkill`, map `skillID` -> phase (strip `sdd-` prefix), map `fileName` (`SKILL.md` -> `phase.md`; else `phase.fileName`) to `skills/_qe-sdd/...`, read via `assets.Read`, `ok=false` on error/empty (fail-open). (Req: QE Content Replaces Dev Content In All Injector Paths — fail-open scenario)
- [x] 2.2 Create `internal/components/sdd/inject_qe.go` with `qeSwapNativeAgentBody(wrapper, qeBody, qeDescription string) string`: locate frontmatter fence (`"---\n" ... "\n---"`), defensive no-op if no fence found, preserve `name:`/`model:`/`effort:`/`tools:`, replace only the `description:` entry (including folded `>` scalar form) with `qeDescription`, replace body after the closing fence with `qeBody`. (Req: No Upstream Prompt Content Is Edited — anchor-only constraint)
- [x] 2.3 Add `qeSubAgentDescription(phase string) string` to `internal/components/sdd/inject_qe.go` returning the QE-framed one-line description per phase for Path 2 frontmatter rewrite.

## Phase 3: Injector Anchors (depends on Phase 2 helpers)

- [x] 3.1 Anchor Path 1 in `internal/components/skills/inject.go` inside the `WalkDir` closure (~line 64-88, before/around `extractModelSection` at line 88): add `if qe, ok := QESDDTestingContent(string(id), filepath.ToSlash(relPath)); ok { content = qe }` so every walked `.md` (SKILL.md + strict-tdd siblings) is swapped for all `SupportsSkills()` adapters. (Req: QE Content Replaces Dev Content In All Injector Paths)
- [x] 3.2 Anchor Path 2 in `internal/components/sdd/inject.go` step 3c (~line 703, inside `isMarkdownSubAgentPromptFile(entry.Name())` block, BEFORE `injectCodeGraphGuidanceIntoPrompt` at line 704): derive `phase` from `strings.TrimSuffix(entry.Name(), ".md")`, call `skills.QESDDTestingContent(phase, "SKILL.md")`, on `ok` call `qeSwapNativeAgentBody(contentStr, qe, qeSubAgentDescription(phase))` before the CodeGraph call. (Req: QE Content Replaces Dev Content In All Injector Paths; Path 2 frontmatter-preserving scenario)
- [x] 3.3 Anchor Path 3 in `internal/components/sdd/prompts.go` `WriteSharedPromptFiles` (~line 73, immediately after `extractModelSection(skillContent, capability)` and BEFORE `injectCodeGraphGuidanceIntoPrompt`): swap `content` via `skills.QESDDTestingContent(phase, "SKILL.md")` when `ok`. (Req: QE Content Replaces Dev Content In All Injector Paths)
- [x] 3.4 Anchor Path 4 in `internal/components/uninstall/service.go` (~line 532): extend the skip condition to `entry.Name() == "_shared" || entry.Name() == "_qe-sdd"` so `ComponentSkills` uninstall never deletes the QE override assets. (Req: overlay/robustness — Path 4, design correction)
- [x] 3.5 Register all 9 assets + `internal/components/skills/inject_qe.go` + `internal/components/sdd/inject_qe.go` under `overlayFiles` in `tools/qe-overlay/overlay.json` (NOT `netNewDirs` — `_qe-sdd` lacks a root `SKILL.md` and would fail `verifyNetNewInstallable`). (Req: Overlay Registers Net-New Assets And Anchors)
- [x] 3.6 Add the 4 `inlineAnchors` entries to `tools/qe-overlay/overlay.json`: `inject.go` mustContain `QESDDTestingContent`; `sdd/inject.go` mustContain `qeSwapNativeAgentBody`; `sdd/prompts.go` mustContain `QESDDTestingContent`; `uninstall/service.go` mustContain `_qe-sdd`. (Req: Overlay Registers Net-New Assets And Anchors)

## Phase 4: Tests / Verification (depends on Phases 1-3)

- [x] 4.1 Create `internal/components/sdd/qe_sdd_override_qe_test.go` — drive Path 1 (one `SupportsSkills` adapter), Path 2 (step 3c for a native-subagent adapter), Path 3 (`WriteSharedPromptFiles`) for all 7 phases; assert QE content returned at each path. (Spec: Oracle test proves QE override across all injector paths)
- [x] 4.2 In the same test file, implement the STRUCTURAL oracle per phase: assert required section headings from `sdd-qe-phase-contract` are present (e.g. design requires named ISTQB technique + test-pyramid + risk/clustering headings; explore requires risk/testability/clustering/oracle-inventory headings) — reject substring-anywhere matching. (Spec: Oracle rejects trivial keyword-stuffed content)
- [x] 4.3 Add negative-marker assertions using ONLY dev-verified markers: `"Implement code changes"` (confirmed in `internal/assets/claude/agents/sdd-apply.md` line 4) and `"production code"` (confirmed in `internal/assets/skills/sdd-apply/strict-tdd.md` and `internal/assets/skills/sdd-verify/strict-tdd-verify.md`) — do NOT use `"capability"` or `"build the"` (design confirms both are false-positive/false-negative prone). (Spec: Negative markers are phase-specific and dev-verified)
- [x] 4.4 Add Path 2 frontmatter-preservation assertion: after `qeSwapNativeAgentBody`, the wrapper still contains `name:`, the resolved `model:` line, and a rewritten QE `description:`. (Design verification step 4)
- [x] 4.5 Add gate/fail-open assertions: `QESDDTestingContent` returns `ok=false` for a non-SDD id, for `judgment-day` (no `sdd-` prefix), and for `sdd-archive`/`sdd-onboard`/`sdd-init`. (Spec: Archive, onboard, and init fall through to upstream)
- [x] 4.6 Add strict-TDD oracle assertions: `("sdd-apply","strict-tdd.md")` and `("sdd-verify","strict-tdd-verify.md")` return `ok=true` with RED/GREEN/REFACTOR automation framing and no production-code instruction; `("sdd-design","strict-tdd.md")` returns `ok=false`. (Spec: Strict-TDD QE module never instructs production-code work)
- [x] 4.7 **Parser unit-test hardening (mandatory)**: create/extend unit tests for `qeSwapNativeAgentBody` in `internal/components/sdd/inject_qe.go` (co-located `_test.go` or within 4.1's file) covering: (a) frontmatter with a folded `description: >` multiline scalar is correctly replaced; (b) input with NO frontmatter fence is a defensive no-op (input returned unchanged); (c) single-line `description:` is correctly replaced; (d) `name:`/`model:`/`effort:`/`tools:` values remain byte-identical after the swap. (Design: Path 2 fragility risk — explicit test coverage requirement)
- [x] 4.8 Update `tools/qe-overlay/overlay.json` diff sanity: run `qe-overlay verify` locally, confirm exit 0 and all 4 anchors + overlayFiles report present. (Spec: Overlay verify passes with full registration)
- [x] 4.9 Run `go test ./internal/components/sdd/... ./internal/components/skills/...` and confirm no regression in existing SDD injector tests (Paths 1-3 pre-existing behavior for non-SDD skills unaffected). Also ran full `go test ./...` (53/53 packages green) — required 3 downstream fixes not enumerated as separate tasks: (a) added matching `model-capable`/`model-small` sections to `apply.md` (verify.md already had them) so the pre-existing `TestWriteSharedPromptFilesWithCapabilities` capability-differentiation test keeps passing for `sdd-apply`; (b) re-applied `extractModelSection` to the QE content in the Path 3 anchor (`sdd/prompts.go`) so those sections are honored there too — a deliberate, minimal deviation from the literal design pseudocode, see Deviations in the apply report; (c) updated `internal/assets/assets_test.go` `TestEmbeddedAssetCount` (30→31 dirs) and regenerated the `sdd-claude-agent-*` / `sdd-kiro-agent-*` golden fixtures for the 7 QE-overridden phases via `-update` (git diff confirms only those 14 files changed, nothing else).
