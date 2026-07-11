# qe-installer-flow Specification

## Purpose

Defines the unconditional QE-simplified installer flow for the Gentle-QE fork build: which upstream screens are hidden with silent defaults, which stay visible, how Persona/Preset option lists are filtered, and the zero-upstream-edit constraint. The picker chain is a state machine (`pickerFlowSlice` + `shouldShow*` predicates gating `Screen` transitions); hiding a screen means its predicate/list never yields that state for the QE build, not that functionality is removed.

## Requirements

### Requirement: Dev-Only Screens Hidden With Silent Defaults

The system MUST NOT present the Claude Model Picker, Kiro Model Picker, Codex Model Picker, SDDMode, CommunityTools, or OpenCodePlugins screens in the QE build's navigation flow (`pickerFlowSlice` and its cross-cutting guards). Each hidden screen MUST still apply its default value, and no feature reachable only via a hidden screen's underlying component (e.g. `ComponentSDD`) MAY be silently dropped as a result.

**Test technique**: state-transition testing — `Screen` is the state, `pickerFlowSlice`/`shouldShow*`/guard predicates are the transition function. A hidden screen means the transition function never returns that state for the QE build; the oracle is the resulting `Model.Screen` and `Model.Selection` after driving `Update()` with the same key sequence used by existing navigation tests (e.g. `TestInstallNavigationRoundTrips`).

Note on oracle placement: the 3 model pickers and SDDMode are members of `pickerFlowSlice()` upstream, so their oracle is slice non-membership. `CommunityTools` and `OpenCodePlugins` are gated by call-site guards (`shouldShowCommunityToolsScreen`/`shouldShowOpenCodePluginsScreen`), never by slice membership — asserting slice non-membership for these two would pass identically on unmodified upstream and prove nothing. Their oracle MUST instead be that `Model.Screen` never becomes `ScreenCommunityTools`/`ScreenOpenCodePlugins` while driving `Update()` through the full flow.

#### Scenario: Slice-gated screens are skipped and their defaults are applied

- GIVEN a QE-build `Model` with `ComponentSDD` and `AgentClaudeCode` selected (conditions that would show `ScreenSDDMode` and `ScreenClaudeModelPicker` upstream)
- WHEN a test drives `Model.Update()` through the picker flow from `ScreenPreset` to `ScreenDependencyTree`
- THEN `pickerFlowSlice()` never contains `ScreenSDDMode`, `ScreenClaudeModelPicker`, `ScreenKiroModelPicker`, or `ScreenCodexModelPicker`
- AND `Model.Selection.SDDMode` equals the documented QE default after the flow completes
- AND `Model.Selection.Components` still contains `ComponentSDD`

#### Scenario: Guard-gated screens never become the active screen

- GIVEN a QE-build `Model` with `AgentOpenCode` selected and `InstallFlowActive` true (conditions that would satisfy `shouldShowOpenCodePluginsScreen`/`shouldShowCommunityToolsScreen` upstream)
- WHEN a test drives `Model.Update()` with the full key sequence from `ScreenPreset` through installation completion
- THEN `Model.Screen` is never observed to equal `ScreenCommunityTools` or `ScreenOpenCodePlugins` at any step
- AND the CommunityTools/OpenCodePlugins default values are applied to `Model.Selection` without the user visiting either screen

#### Scenario: Backward navigation never lands on a hidden screen

- GIVEN the same QE-build `Model` as above, positioned after the picker flow
- WHEN a test drives Esc / "← Back" backward through the flow
- THEN the visited screen sequence contains none of the 6 hidden screens
- AND the sequence is the exact reverse of the forward sequence produced in the happy-path scenario

### Requirement: StrictTDD Screen Remains Visible

The system MUST continue to present the StrictTDD screen in the QE build whenever `shouldShowStrictTDDScreen()` (or its QE-build equivalent) is true, unchanged from upstream visibility conditions.

#### Scenario: StrictTDD appears when SDD is selected

- GIVEN a QE-build `Model` with `ComponentSDD` selected
- WHEN a test computes `pickerFlowSlice()`
- THEN `ScreenStrictTDD` is present in the returned slice

### Requirement: QE-Only Persona and Preset Options

The system MUST restrict `PersonaOptions()` to QE persona(s) only (e.g. `PersonaSDET`) and `PresetOptions()` to QE preset(s) only (e.g. `PresetQESDET`, `PresetQEFront`, `PresetQEAPI`, `PresetQEPerf`) in the QE build. Dev-only persona and preset entries (e.g. `PersonaGentleman`, `PresetFullGentleman`, `PresetMinimal`) MUST NOT appear in either list.

#### Scenario: Persona list contains only QE personas

- GIVEN the QE build
- WHEN a test calls `screens.PersonaOptions()`
- THEN the returned slice contains only QE persona IDs
- AND no dev persona ID (Gentleman, Neutral, Custom) is present

#### Scenario: Preset list contains only QE presets

- GIVEN the QE build
- WHEN a test calls `screens.PresetOptions()`
- THEN the returned slice contains only QE preset IDs
- AND no dev preset ID (FullGentleman, EcosystemOnly, Minimal) is present

### Requirement: Collapsed Welcome Menu

The system MUST present a Welcome menu limited to QE-essential options (installation plus must-have actions) in the QE build. Dev-only or QE-irrelevant Welcome options MUST NOT appear.

#### Scenario: Welcome menu omits non-essential dev options

- GIVEN the QE build
- WHEN a test calls `screens.WelcomeOptions(...)` with representative arguments
- THEN the returned option list contains only the documented QE-essential entries
- AND excludes options the design marks as dev-only or non-essential for QE

### Requirement: Unconditional QE Flow

The system MUST show the QE-simplified installer flow (hidden screens, filtered lists, collapsed Welcome) for every invocation of the fork build, regardless of CLI flags, environment variables, or persisted state. No flag, preset selection, or environment condition MAY restore the full upstream dev flow within the fork build.

#### Scenario: No flag restores the dev flow

- GIVEN the fork build binary, tested across the flag/env/preset combinations exercised by the existing installer test suite
- WHEN a test drives the picker flow under each combination
- THEN the 6 hidden screens never appear and Persona/Preset lists never include dev-only entries in any combination

### Requirement: Zero Upstream Content Edits

The fork's `internal/tui/model.go` MUST touch upstream only through single-line delegating anchors to `model_qe.go` (or equivalent QE-overlay file), each registered in `tools/qe-overlay/overlay.json`. This requirement's verifiable guarantee is scoped to what `qe-overlay verify` actually checks today: that every registered anchor string is present (`mustContain`) in `model.go`. It does NOT claim or require a line-by-line diff against upstream — `qe-overlay verify` performs `strings.Contains` presence checks, not a diff, and so cannot detect spurious non-anchor edits. Extending `qe-overlay` to perform real diff verification is out of scope for this change and is tracked as a separate cross-cutting improvement (bead).

#### Scenario: All registered QE anchors are present

- GIVEN the fork's `model.go` after the dl9 anchors are added
- WHEN `tools/qe-overlay/overlay.json`'s `mustContain` checks run via `qe-overlay verify`
- THEN every anchor string registered for this change is found present in `model.go`
- AND the check does not assert anything about non-anchor lines, since presence-only verification cannot detect unregistered upstream edits
