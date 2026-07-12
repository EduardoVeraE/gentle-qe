# Tasks: Simplify TUI Installer Flow for QE Users

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~550-850 (2 net-new files ~110-150; 2 modified `_qe.go` bodies ~10-20; 8 upstream anchors ~10-15; overlay.json ~15-20; 2 net-new test files w/ 7x2 table + boundary + invariant + Bubbletea integration ~350-550) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (net-new + modified bodies) -> PR 2 (8 anchors + overlay.json + state.yaml) -> PR 3 (tests) |
| Delivery strategy | ask-on-risk (assumed default — not set by orchestrator this session; confirm before apply) |
| Chain strategy | feature-branch-chain |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | `model_qe.go` + `welcome_qe.go` + modified `persona_qe.go`/`preset_qe.go` bodies (Phase 1) | PR 1 | Base = feature/tracker branch. Zero upstream file touched. |
| 2 | 8 upstream anchors + `overlay.json` + `state.yaml` (Phases 2-3) | PR 2 | Base = PR 1 branch. Small, mechanical, highest scrutiny (touches upstream `model.go`/`welcome.go`/`persona.go`/`preset.go`). |
| 3 | Unit + integration tests (Phases 4-6) | PR 3 | Base = PR 2 branch. Oracle-heavy; verifies units 1-2 end-to-end, incl. `qe-overlay verify`. |

## APPLY BLOCKER — read before continuing (Phases 5-6 paused)

Applying Phases 1-4 exactly as specified surfaced a **structural regression not
caught by design/judgment-day review**: anchors 2.1/2.6/2.7/2.8 make
`pickerFlowSlice()`, `WelcomeOptions()`, `PersonaOptions()`, `PresetOptions()`
**unconditionally** QE-only (no dev/Gentleman path left), which breaks **47
pre-existing upstream tests** across `internal/tui` and `internal/tui/screens`
— including `pickerFlowSlice()`'s OWN dedicated unit test suite
(`TestPickerFlowSlice`, `TestPickerNextScreen`, `TestPickerPreviousScreen`) and
`TestInstallNavigationRoundTrips`, the exact test task 6.2 names as the
regression oracle. This is not a coding bug — it is the literal, correct
result of the approved anchors — but it makes task 6.2's acceptance criterion
("no regression in existing navigation/golden tests") impossible to satisfy
without editing ~47 upstream test files (far outside the declared 8-anchor
scope and the "zero upstream content edits" constraint).
**Phases 5 and 6.2 are paused pending a user/design decision.** See the apply
return summary for full details and options. Phases 1-4, 3.1-3.3, 6.1, 6.3 are
complete and green.

## Phase 1: Net-New Files & Modified QE Bodies (Foundation)

- [x] 1.1 Create `internal/tui/model_qe.go` (pkg `tui`): `qeFilterPickerFlow(s []Screen) []Screen` drops the 3 model pickers + `ScreenSDDMode`; reads only slice elements, never live `m.Screen`. (Req: Dev-Only Screens Hidden; design anchor #1)
- [x] 1.2 Same file: `qeSuppressCommunityTools() bool` / `qeSuppressOpenCodePlugins() bool` — QE-build constants returning `true`. (Req: Dev-Only Screens Hidden; anchors #2-#3)
- [x] 1.3 Same file: `qeWelcomeCanonicalCursor(m Model, collapsed int) int` — replays the two-gap counter (leader gap base 7/8, tail gap +3 for Quit) per design remap table. (Req: Collapsed Welcome Menu; anchor #4 — most fragile point)
- [x] 1.4 Create `internal/tui/screens/welcome_qe.go` (pkg `screens`): `qeWelcomeOptions(opts []string, showProfiles, hasEngines bool) []string` collapses to the 7 QE-essential entries. (Req: Collapsed Welcome Menu; anchor #6)
- [x] 1.5 Modify `persona_qe.go` body: rename `qePersonaOptions()` -> `qeFilterPersonaOptions(opts []model.PersonaID) []model.PersonaID` returning `[PersonaSDET]` only. (Req: QE-Only Persona and Preset Options; anchor #7)
- [x] 1.6 Modify `preset_qe.go` body: rename `qePresetOptions()` -> `qeFilterPresetOptions(opts []model.PresetID) []model.PresetID` returning the 4 QE presets only. (Req: QE-Only Persona and Preset Options; anchor #8)

## Phase 2: Upstream 1-Line Anchors (depends on Phase 1 symbols existing)

- [x] 2.1 `model.go:3976` `return s` -> `return qeFilterPickerFlow(s)`.
- [x] 2.2 `model.go:3749` `shouldShowCommunityToolsScreen` -> add `if qeSuppressCommunityTools() { return false }`.
- [x] 2.3 `model.go:3733` `shouldShowOpenCodePluginsScreen` -> add `if qeSuppressOpenCodePlugins() { return false }`.
- [x] 2.4 `model.go:1484` **insert** `m.Cursor = qeWelcomeCanonicalCursor(m, m.Cursor)` as the first line before `switch m.Cursor` in the `ScreenWelcome` confirm case.
- [x] 2.5 `model.go:594-601` `NewModel` Selection literal -> add explicit `SDDMode: model.SDDModeSingle` assignment (closes `""`->multi auto-promotion; design SDDMode decision). **Deviation**: implemented as a standalone `selection.SDDMode = model.SDDModeSingle` statement immediately after the struct literal, not as a literal field, to stay immune to gofmt's automatic struct-field column alignment (which would otherwise make the qe-overlay `mustContain` anchor's exact byte spacing fragile to unrelated field-name-length changes in the same literal). Semantically identical; `overlay.json`'s anchor text updated to match (`selection.SDDMode = model.SDDModeSingle`).
- [x] 2.6 `welcome.go:55` `return opts` -> `return qeWelcomeOptions(opts, showProfiles, hasEngines)`.
- [x] 2.7 `persona.go:13` `return append(opts, qePersonaOptions()...)` -> `return qeFilterPersonaOptions(opts)`.
- [x] 2.8 `preset.go:17` `return append(opts, qePresetOptions()...)` -> `return qeFilterPresetOptions(opts)`.

## Phase 3: Overlay Manifest & DAG State (depends on Phase 1-2 files/anchors existing)

- [x] 3.1 `tools/qe-overlay/overlay.json` `overlayFiles`: add `internal/tui/model_qe.go`, `internal/tui/screens/welcome_qe.go`, `internal/tui/model_qe_test.go`, `internal/tui/screens/welcome_qe_test.go`. **Deviation**: also added a 5th net-new file, `internal/tui/screens/persona_preset_qe_test.go` — task 4.6's QE-only Persona/Preset assertions could not be added to the existing (upstream, non-overlay) `persona_preset_test.go` without violating "zero upstream content edits," so they live in a new dedicated overlay test file instead. (Req: Zero Upstream Content Edits)
- [x] 3.2 Same file `inlineAnchors`: add 5 `model.go` entries — `mustContain` `qeFilterPickerFlow`, `qeSuppressCommunityTools`, `qeSuppressOpenCodePlugins`, `qeWelcomeCanonicalCursor`, and its own dedicated SDDMode entry (`"selection.SDDMode = model.SDDModeSingle"`, see 2.5 deviation). Added `welcome.go` entry `mustContain: "qeWelcomeOptions"`.
- [x] 3.3 Same file: updated existing `persona.go` entry `qePersonaOptions`->`qeFilterPersonaOptions` and `preset.go` entry `qePresetOptions`->`qeFilterPresetOptions`.
- [x] 3.4 `openspec/changes/gentle-qe-dl9/state.yaml` — already existed (created during the design phase); updated (not recreated) to reflect apply progress and the Phase 5/6 blocker.

## Phase 4: Unit Tests (depends on Phase 1; ref skill `go-testing`)

- [x] 4.1 `internal/tui/model_qe_test.go`: table-driven `qeWelcomeCanonicalCursor` — all 7 collapsed indices x 2 `hasDetectedOpenCode` variants -> canonical index, per design remap table. (Spec oracle: state-transition testing)
- [x] 4.2 Same file: explicit boundary case — collapsed 6 (Quit) -> canonical 10 (no OC) / 11 (OC), asserting the tail gap is stepped and NOT routed to `ScreenCommunityTools`, for both variants.
- [x] 4.3 **Deviation**: relocated to `internal/tui/screens/welcome_qe_test.go` — structural fail-fast `len(qeWelcomeOptions(...)) == 7`. `qeWelcomeOptions` is unexported in package `screens`; `model_qe_test.go` is package `tui` and cannot invoke it cross-package. Both sides independently hardcode the same `7` (documented via a `qeWelcomeCollapsedCount` constant in the screens test file with a comment cross-referencing `qeWelcomeCanonicalCursor`).
- [x] 4.4 `internal/tui/model_qe_test.go`: `qeFilterPickerFlow` excludes the 4 slice-gated screens (3 pickers + SDDMode). **Deviation**: exercised via a synthetic raw slice (not `m.pickerFlowSlice()`) because anchor 2.1 makes `pickerFlowSlice()` already return the filtered result — calling it directly would test the function against its own output. Added a companion integration-style test, `TestQEModel_PickerFlowSliceNeverExposesDevOnlyScreens`, that drives the real (post-anchor) `m.pickerFlowSlice()` with Claude+Kiro+Codex+OpenCode agents and ComponentSDD to prove the anchor is actually wired end-to-end. (Req: Dev-Only Screens Hidden, scenario 1)
- [x] 4.5 `internal/tui/screens/welcome_qe_test.go`: `qeWelcomeOptions` returns exactly the 7 QE-essential labels, excludes dev-only entries. (Req: Collapsed Welcome Menu, scenario)
- [x] 4.6 **Deviation**: new file `internal/tui/screens/persona_preset_qe_test.go` (not an extension of the existing `persona_test.go`/`preset_test.go` — those don't exist as separate files; the closest upstream file, `persona_preset_test.go`, is not a registered overlay file, so editing it would violate "zero upstream content edits"). `PersonaOptions()==[PersonaSDET]`, `PresetOptions()==4 QE presets`, no dev IDs present, plus triangulation tests for `qeFilterPersonaOptions`/`qeFilterPresetOptions` ignoring dev input. (Req: QE-Only Persona and Preset Options, both scenarios)

## Phase 5: Integration Tests — Bubbletea (depends on Phase 1-2 wired) — BLOCKED, not started

- [ ] 5.1 `internal/tui/model_qe_test.go`: drive `NewModel`->`Update()` forward through the flow; assert the 6 hidden screens never equal `m.Screen`, `Model.Selection.SDDMode`/CommunityTools/OpenCodePlugins defaults applied, `ComponentSDD` retained, `ScreenStrictTDD` still reached. Oracle = `Model.Screen` non-equality, NOT slice membership for the 2 guard-gated screens. (Req: Dev-Only Screens Hidden, scenarios 1-2; Req: StrictTDD Remains Visible)
- [ ] 5.2 Same file: drive Esc/"<- Back" backward from post-flow; assert visited sequence is the exact reverse of 5.1's forward sequence and touches none of the 6 hidden screens. (Req: Dev-Only Screens Hidden, scenario 3)
- [ ] 5.3 Same file: drive `Update(Enter)` at each of the 7 collapsed Welcome cursor positions, both OC variants; assert resulting action/`Screen` per the remap table (Start->Detection, Upgrade, Sync, Upgrade+Sync, Backups, Uninstall->UninstallMode, Quit->`tea.Quit`). End-to-end confirmation of 4.1/4.2's unit oracle.

**Not started**: writing new integration tests on top of a baseline where 47
pre-existing navigation tests are red would produce misleading green
signal for this feature while masking the regression. Needs the Phase
5/6 blocker resolved first (see top of file and apply return summary).

## Phase 6: Overlay & Full Verification (depends on Phases 1-5)

- [x] 6.1 Run `go run ./tools/qe-overlay verify` — confirm all Phase 3 anchors/files report present, exit 0. (Req: Zero Upstream Content Edits, "All registered QE anchors are present") — **PASSED**: `✓ qe-overlay: overlay intacto, sin drift.`
- [ ] 6.2 Run `go test ./internal/tui/... ./internal/tui/screens/...` — confirm no regression in existing navigation/golden tests (e.g. `TestInstallNavigationRoundTrips`). — **FAILED as specified**: 47 pre-existing tests red (see APPLY BLOCKER above); all new QE-specific tests (Phase 4) pass. `TestInstallNavigationRoundTrips` itself is among the 47 failures.
- [x] 6.3 Run `go vet ./...`. — **PASSED**, no findings.
