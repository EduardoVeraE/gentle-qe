# Design: Simplify TUI Installer Flow for QE Users

## Technical Approach

Unconditional QE flow for the fork build. All skip/default/filter logic lives in
net-new files that mirror the existing `preset_qe.go`/`persona_qe.go` convention;
upstream `model.go`/`welcome.go` receive only 1-line delegating anchors. Two
package boundaries: `internal/tui/model_qe.go` (package `tui`, owns pickerFlow +
predicate + welcome-cursor logic) and `internal/tui/screens/welcome_qe.go`
(package `screens`, owns the Welcome list collapse; list filters for
Persona/Preset extend the existing `*_qe.go` files). Zero upstream logic edits →
re-appliable after upstream merge; `overlay.json` `mustContain` checks guard drift.

Two independent levers, per the explore:
- **Screen skip** — filter `pickerFlowSlice` + suppress 2 predicates (hide 6 screens).
- **List filter** — shrink Persona/Preset option lists (screens stay, lists shrink).

All QE filtering is gated by a **testing seam** (`qeInstallerFlow`, default ON in the
production binary, disabled per test package by a net-new `TestMain`) so the ~93 upstream
dev-flow tests pass unedited while the fork binary stays unconditional. See "Testing Seam".

## Architecture Decisions

### Decision: SDDMode default = explicit `single` literal (recommended; not blocking)
**Choice**: Set `SDDMode: model.SDDModeSingle` **explicitly** in the `NewModel`
Selection literal (its own registered anchor), then hide `ScreenSDDMode` via the
pickerFlow filter. **Alternatives**: leave `SDDMode` at zero value `""`; or default `multi`.
**Rationale (the strong reason — auto-promotion, not the type argument)**: `single`
and `""` are functionally identical at injection time (`inject.go` treats both as the
single-orchestrator path), so "`""` ≠ single" is a *weak* justification. The real reason
the explicit literal is mandatory: an empty `SDDMode` is **silently auto-promoted to
`SDDModeMulti`** whenever OpenCode profiles are present or detected —
`internal/app/app.go:664-666` (`if selection.SDDMode == "" { selection.SDDMode = model.SDDModeMulti }`
when `len(overrides.Profiles) > 0`) and `internal/cli/sync.go:681-682`
(`else if len(profiles) > 0 && sddMode == "" { sddMode = model.SDDModeMulti }`). Without
the explicit `single` literal, a QE with OpenCode profiles would silently land in the
exact multi-mode complexity dl9 is hiding, and `SDDModeMulti` also re-adds a
`ScreenModelPicker` to `pickerFlowSlice` when the OpenCode model cache exists
(model.go:3961-3968). Pinning `single` closes the auto-promotion path deterministically.
**Why not multi**: `SDDMode` is **not** what gentle-qe-589 built. 589's multi-agent SDD
cycle (per-phase `sdd-*` sub-agents) rides on **ComponentSDD skill/prompt injection**,
intact under `single`; the installer `SDDMode` knob only governs per-phase OpenCode model
assignment. `single` keeps the 589 orchestration, drops zero QE capability.
**Open-question posture**: firmly recommend `single`; escalate only if product intent is
that a QE should default into per-phase multi-agent model tuning — in which case set the
`NewModel` field to `SDDModeMulti` AND extend `qeFilterPickerFlow` to also strip the
resurfaced `ScreenModelPicker`. Absent that intent, treat as decided.

### Decision: Screen skip via pickerFlowSlice filter (not component suppression)
**Choice**: Drop the 6 screens from the flow slice / predicates; keep ComponentSDD.
**Alternatives**: Suppress ComponentSDD so `shouldShow*` evaluate false.
**Rationale**: Suppressing ComponentSDD would strip the real SDD skill injection the
QE preset must deliver. QE presets fall to `componentsForPreset` default branch
(model.go:4050) → include ComponentSDD → this is *why* SDDMode/StrictTDD/pickers show
today. Principle: **hiding a screen never removes preset functionality** — hide SDDMode
via a default, keep the component. StrictTDD screen stays visible (SDET discipline).

### Decision: Welcome collapse needs a TWO-GAP cursor-remap anchor (THE fragile point)
**Choice**: Filter `WelcomeOptions` (auto-syncs render + `optionCount` bound) **plus**
a second anchor that remaps `m.Cursor` from the collapsed index to the upstream
*canonical* index at the FIRST line of the `ScreenWelcome` confirm case, before the
untouched switch. **Alternatives**: (a) filter the list only; (b) reverse-label-lookup remap.

**Why a single uniform offset is WRONG (verified against model.go:1483-1569)**:
`confirmSelection` ScreenWelcome dispatches by **hardcoded index** — cases 0-5 static,
then a dynamic 6+ counter. The counter has **two non-uniform gaps**, not one:
1. **Leader gap** — canonical 4 (Configure models), 5 (Create Agent), 6 (OpenCode
   Community Plugins), and conditionally 7 (OpenCode SDD Profiles, inserted only when
   `hasDetectedOpenCode()` is true) are all excluded from the collapsed keep-set.
2. **Tail gap** — canonical `CommunityTools` (Community Tools/Plugins) sits *between*
   Managed uninstall and Quit and is also excluded.

A flat offset would send collapsed "Quit" to the `CommunityTools` case. The remap must
step over **both** gaps. Full canonical order (mirrors `WelcomeOptions`):

| canonical (no OpenCode) | canonical (OpenCode detected) | label |
|---|---|---|
| 0 | 0 | Start installation |
| 1 | 1 | Upgrade tools |
| 2 | 2 | Sync configs |
| 3 | 3 | Upgrade + Sync |
| 4 | 4 | Configure models *(excluded)* |
| 5 | 5 | Create your own Agent *(excluded)* |
| 6 | 6 | OpenCode Community Plugins *(excluded)* |
| — | 7 | OpenCode SDD Profiles *(excluded; present only if `hasDetectedOpenCode`)* |
| 7 | 8 | Manage backups |
| 8 | 9 | Managed uninstall |
| 9 | 10 | Community Tools/Plugins *(excluded — the tail gap)* |
| 10 | 11 | Quit |

**Collapsed keep-set → canonical remap table** (7 entries × 2 variants):

| collapsed | label | canonical (no OpenCode) | canonical (OpenCode detected) |
|---|---|---|---|
| 0 | Start installation | 0 | 0 |
| 1 | Upgrade tools | 1 | 1 |
| 2 | Sync configs | 2 | 2 |
| 3 | Upgrade + Sync | 3 | 3 |
| 4 | Manage backups | 7 | 8 |
| 5 | Managed uninstall | 8 | 9 |
| 6 | Quit | 10 | 11 |

Offsets are **+3/+3/+4** (no OpenCode) and **+4/+4/+5** (OpenCode) — the extra +1 on Quit
is the tail gap (skipping `CommunityTools`).

**`qeWelcomeCanonicalCursor(m, collapsed)` mechanism** — replays the same conditional
counter the upstream switch uses, stepping both gaps:
```
collapsed 0..3            → canonical == collapsed        (static leaders)
base := 7; if hasDetectedOpenCode() { base = 8 }         (leader gap; +1 for Profiles)
collapsed 4 (Backups)    → base
collapsed 5 (Uninstall)  → base + 1
collapsed 6 (Quit)       → base + 3                        (tail gap: skip CommunityTools at base+2)
```

**Rejected — reverse-label-lookup** (`indexOf(fullList, collapsed[cursor])`, self-
synchronizing over arbitrary gaps): the collapse anchor replaces `WelcomeOptions`' return,
so the full upstream list is not available at the confirm site without reconstructing it
in the overlay (label duplication + drift risk). The two-gap replay is self-contained in
`model_qe.go`, zero duplication, and is fully pinned by the tests below. Upstream
reordering of the Welcome menu is the residual risk; the structural invariant + the
end-to-end "Quit after collapse" boundary test fail loudly on it, and overlay.json
`mustContain` pins the anchor. This is the most fragile point of the feature — treat the
remap table above as the authoritative oracle.

### Decision: Suppress CommunityTools without touching CommunityToolsStandalone
**Choice**: `shouldShowCommunityToolsScreen` early-returns false via a QE constant.
**Rationale**: The predicate (`InstallFlowActive && !CommunityToolsStandalone`) is not
Selection-gated. The QE guard only suppresses **install-flow auto-navigation**; the
standalone entry uses `setScreen` directly (model.go:1554) and is unaffected — and the
Welcome collapse removes its menu item anyway.

## Data Flow (QE install path)

    Welcome(collapsed) → Detection → Agents → Persona[SDET] → Preset[QE] →
      StrictTDD → DependencyTree → Install
    (skipped: Claude/Kiro/Codex pickers · SDDMode · CommunityTools · OpenCodePlugins)

## File Changes / Anchor Map

**Canonical anchor count = 8** upstream 1-line anchors (rows 3-10 below), plus 2 net-new
source files and 2 modified overlay `*_qe.go` bodies. Keep this "8" consistent everywhere.

| # | File | Action | Anchor (1 line) → delegates to |
|---|------|--------|-------------------------------|
| — | `internal/tui/model_qe.go` | Create | seam var `qeInstallerFlow` (+`qeFlowDefault=true`); owns `qeFilterPickerFlow`, `qeSuppressCommunityTools`, `qeSuppressOpenCodePlugins`, `qeWelcomeCanonicalCursor` (all seam-gated) |
| — | `internal/tui/screens/welcome_qe.go` | Create | seam var `qeInstallerFlow` (+`qeFlowDefault=true`); `qeWelcomeOptions()` collapse (seam-gated) |
| — | `internal/tui/screens/persona_qe.go` | Modify body | pure `qeFilterPersonaOptions(opts)` (always QE-only) + seam-aware `qePersonaOptionsForBuild(opts)` |
| — | `internal/tui/screens/preset_qe.go` | Modify body | pure `qeFilterPresetOptions(opts)` (always QE-only) + seam-aware `qePresetOptionsForBuild(opts)` |
| — | `internal/tui/qe_seam_test.go` | Create (test) | `TestMain` sets `qeInstallerFlow=false` for the tui package |
| — | `internal/tui/screens/qe_seam_test.go` | Create (test) | `TestMain` sets `qeInstallerFlow=false` for the screens package |
| 1 | `model.go:3976` `return s` | Modify | `return qeFilterPickerFlow(s)` (drops 4 pickers+SDDMode; reads slice, never `m.Screen`) |
| 2 | `model.go:3749` CommunityTools | Modify | `if qeSuppressCommunityTools() { return false }` |
| 3 | `model.go:3733` OpenCodePlugins | Modify | `if qeSuppressOpenCodePlugins() { return false }` |
| 4 | `model.go:1484` ScreenWelcome | **Insert** (new line before `switch m.Cursor`) | `m.Cursor = qeWelcomeCanonicalCursor(m, m.Cursor)` |
| 5 | `model.go:594-601` `NewModel` Selection literal | Modify | add field `SDDMode: model.SDDModeSingle,` (closes the `""`→multi auto-promotion path — see SDDMode decision) |
| 6 | `welcome.go:55` `return opts` | Modify | `return qeWelcomeOptions(opts, showProfiles, hasEngines)` |
| 7 | `persona.go:13` PersonaOptions | Modify | `return qePersonaOptionsForBuild(opts)` (seam-aware: OFF→append dev+SDET, ON→[SDET]) |
| 8 | `preset.go:17` PresetOptions | Modify | `return qePresetOptionsForBuild(opts)` (seam-aware: OFF→append dev+QE, ON→4 QE) |

**Silent defaults**: SDDMode→`single` via an **explicit** `NewModel` field anchor —
required because an empty `SDDMode` auto-promotes to `SDDModeMulti` when OpenCode profiles
exist (app.go:664-666, sync.go:681-682), not merely a type nicety; per-phase models→keep
`NewModel` installState-derived assignments (model.go:599-602); CommunityTools→nil;
OpenCodePlugins→nil. For pickers/CommunityTools/OpenCodePlugins, skipping the screen leaves
Selection at these nil/existing defaults with no extra anchor.

## Invariants
- `qeFilterPickerFlow` filters slice **elements by Screen identity**; it reads
  `m.Selection` only, never live `m.Screen` (model.go:3942 invariant).
- `qeWelcomeCanonicalCursor` reconstructs the same two-gap `hasDetectedOpenCode` counter
  the upstream switch uses, so collapsed index → canonical action.
- **Structural fail-fast**: `len(qeWelcomeOptions(...))` MUST equal the number of collapsed
  cases `qeWelcomeCanonicalCursor` handles (7). A test asserts this so any future change to
  the keep-set that is not reflected in the remap fails immediately, rather than silently
  mis-routing a menu action.

## Testing Seam (unconditional in prod, disable-able in tests)

**Problem (discovered at apply, missed by design + judgment-day)**: making the 8 anchors
return QE-only *unconditionally* breaks ~93 pre-existing upstream dev-flow tests
(`TestPickerFlowSlice`, `TestPickerNextScreen`, `TestWelcomeMenu_*`, `TestSDDMode*`,
`TestInstallNavigationRoundTrips`, `TestPersonaOptionsIncludeGentlemanNeutralArtifacts`, …).
The upstream tests drive the *dev* flow (cursor indices into the full Welcome menu, all
pickers present, dev personas/presets listed). "Unconditional QE binary" and "dev-flow tests
green" are contradictory unless the QE filter can be **disabled inside the test binary**
without editing each upstream test.

**Verified constraints (grep, not assumed)**:
- `internal/tui` and `internal/tui/screens` have **no** `TestMain` and **zero** `t.Parallel()`
  calls → tests run serially within each package; a package-scoped var is safe.
- The 93 tests build the model via the production `NewModel(...)` constructor (same path as
  production) → the on/off differentiator MUST be test-context, not a constructor argument.
- A contradiction already exists in the `screens` package: `persona_language_contract_test.go`
  (upstream) asserts `PersonaOptions()` **contains** a dev persona, while the net-new
  `persona_preset_qe_test.go` asserts it is **QE-only**. Only a seam lets both pass.

**Chosen mechanism — package-scoped seam var, prod-default ON, disabled per test package by a
net-new `TestMain`; QE tests opt back in locally (never `t.Parallel`)**:

- Two independent unexported vars, one per package (the anchored functions live in two
  packages), each backed by an immutable production default:
  - `internal/tui/model_qe.go`:      `const qeFlowDefault = true` · `var qeInstallerFlow = qeFlowDefault`
  - `internal/tui/screens/welcome_qe.go`: `const qeFlowDefault = true` · `var qeInstallerFlow = qeFlowDefault`
- **Production** (no `_test.go` compiled): the var stays `true` → filter always on → the fork
  binary is unconditional. No runtime flag/env/preset can turn it off (satisfies Req 5).
- **Tests**: two net-new `TestMain` files flip the var **off** for the whole package before
  `m.Run()`, so all 93 upstream dev-flow tests see the dev flow **unedited**:
  - `internal/tui/qe_seam_test.go`:            `func TestMain(m *testing.M){ qeInstallerFlow = false; os.Exit(m.Run()) }`
  - `internal/tui/screens/qe_seam_test.go`:    same, for the screens var.
- **QE tests** re-enable the seam locally and restore it: `qeInstallerFlow = true;
  defer func(){ qeInstallerFlow = false }()`. These tests MUST NOT call `t.Parallel()`.

**How each anchored function consults the seam** (OFF path reproduces upstream exactly):

| Function (package) | seam OFF (dev tests) | seam ON (prod + QE tests) |
|---|---|---|
| `qeFilterPickerFlow(s)` (tui) | `return s` (untouched slice) | strip 3 pickers + SDDMode |
| `qeSuppressCommunityTools() bool` (tui) | `return false` → anchor no-ops → upstream predicate runs | `return true` → predicate returns false |
| `qeSuppressOpenCodePlugins() bool` (tui) | `return false` | `return true` |
| `qeWelcomeCanonicalCursor(m,cur)` (tui) | `return cur` (no remap) | two-gap remap (table above) |
| `qeWelcomeOptions(opts,…)` (screens) | `return opts` (full menu) | collapse to keep-set |
| `qePersonaOptionsForBuild(opts)` (screens) | `append(opts, PersonaSDET)` (today's shipped behavior) | `qeFilterPersonaOptions(opts)` |
| `qePresetOptionsForBuild(opts)` (screens) | `append(opts, qePresetIDs…)` | `qeFilterPresetOptions(opts)` |

Both `qeSuppress*` functions are literally `return qeInstallerFlow`, so the OFF path leaves the
upstream predicate semantics intact.

**Pure vs seam-aware split (persona/preset)**: the pure filters `qeFilterPersonaOptions(opts)`
/ `qeFilterPresetOptions(opts)` are **seam-independent** — they always return the QE-only list
and back the direct unit tests (`TestQEFilterPersonaOptions_IgnoresDevInput`, already present).
The `PersonaOptions()`/`PresetOptions()` anchors call the **seam-aware wrappers**
`qePersonaOptionsForBuild` / `qePresetOptionsForBuild`, which branch on `qeInstallerFlow`.

**Why this is race-free under `go test -race`**:
- Production: the var is written once at package init (default `true`) and never mutated →
  read-only → no race.
- Tests: `TestMain` writes the var once, before any test starts (serial, pre-`m.Run`). With
  zero `t.Parallel()` in either package, all tests execute sequentially, so a QE test's
  `true`/`defer false` writes never overlap another test's read → clean under `-race`.
- The ONLY way to introduce a race is adding `t.Parallel()` to a test that mutates the seam.
  Guard: a comment on each var forbids it, and the QE tests that flip the seam are documented
  as `t.Parallel`-free. (Rejected alternative `t.Setenv`-based seam would even panic on
  `t.Parallel`, but reading env in the hot filter path is wasteful and semantically wrong.)

**Rejected seam alternatives**:
- *Model field set true by `NewModel`* — the 93 `tui` tests call `NewModel(...)` too, so they
  would inherit QE-on and still break; and `screens` free functions (`WelcomeOptions`,
  `PersonaOptions`, `PresetOptions`) have no Model to read a field from.
- *`var qeInstallerFlow = !testing.Testing()`* (no TestMain) — elegant, but imports the
  `testing` package into production overlay files (flag registration + binary bloat + smell).
  TestMain keeps `testing` out of the production binary.
- *Add a bool param to the option functions* — changes signatures and breaks the 93 upstream
  call sites.

## Testing Strategy (ref skill: `go-testing`)

| Layer | What | Approach |
|-------|------|----------|
| Unit | `qeFilterPickerFlow` excludes the 6 screens for a QE Selection | table-driven; assert slice membership |
| Unit | `PersonaOptions`==[SDET], `PresetOptions`==4 QE, `qeWelcomeOptions` collapsed | pure-function asserts |
| Unit | `qeWelcomeCanonicalCursor` maps every collapsed index → canonical (per the remap table), for BOTH `hasDetectedOpenCode` true and false | table-driven, all 7×2 cases |
| Unit | **"Quit after collapse" boundary** — collapsed 6 → canonical 10 (no OC) / 11 (OC); asserts the tail gap is stepped (not routed to CommunityTools) | explicit boundary case |
| Unit | **Structural fail-fast** — `len(qeWelcomeOptions(...)) == 7` (== cases the remap handles) | invariant assert |
| Integration | Drive `NewModel`→`Update` key events through the flow; assert the 6 screens never become `m.Screen`, defaults applied, StrictTDD reached | Bubbletea model test |
| Integration | Drive Esc/"← Back" backward from post-flow; assert visited sequence is the exact reverse of the forward path and lands on none of the 6 screens (Req 1, scenario 2) | Bubbletea model test |
| Integration | Drive `Update(Enter)` at each collapsed Welcome cursor; assert the resulting action/`Screen` (Start→Detection, Upgrade→Upgrade, Sync→Sync, Upgrade+Sync, Backups→Backups, Uninstall→UninstallMode, **Quit→tea.Quit**), for both OC variants | Bubbletea model test (end-to-end remap oracle) |
| Overlay | Every new anchor is registered (presence check only) | `tools/qe-overlay` `mustContain` verify |

**Seam usage in tests (mandatory)**:
- The 93 upstream dev-flow tests are **unedited**; the per-package `TestMain` sets
  `qeInstallerFlow = false` so they keep exercising the dev flow.
- Every NEW QE test that drives an anchored path sets `qeInstallerFlow = true` with a
  `defer` restore, and MUST NOT call `t.Parallel()`. The existing net-new QE tests
  (`TestPersonaOptions_QEBuildContainsOnlySDET`, `TestPresetOptions_QEBuildContainsOnlyQEPresets`,
  and the new picker-flow / Welcome-collapse / remap tests) are updated to flip the seam on;
  the direct pure-filter tests (`TestQEFilter*Options_IgnoresDevInput`) need no flip (they call
  the seam-independent filters).
- **Production-ON guarantee**: assert the compile-time default `qeFlowDefault == true` in both
  packages (immutable const, unaffected by `TestMain`'s var flip) — this proves the shipped fork
  binary runs with the filter ON. Complement with a `run -race` note in CI.
- Run `go test -race ./internal/tui/... ` in CI to prove the seam is race-free.

**Req 6 scope (decided by user — do NOT extend qe-overlay in dl9)**: `tools/qe-overlay`
verify today only checks anchor **presence** via `mustContain`; it does **not** compute a
real line-by-line diff against upstream. So the "anchors-only diff" guarantee is enforced by
convention + `mustContain` presence, not by an automated full-diff gate in this change. The
complete diff-verification gate is a cross-cutting qe-overlay improvement **queued in a
separate bead**, out of scope here. The spec's Req 6 is relaxed in parallel to match.

## overlay.json Additions
- `overlayFiles`: `internal/tui/model_qe.go`, `internal/tui/screens/welcome_qe.go`,
  the net-new tests `internal/tui/model_qe_test.go`,
  `internal/tui/screens/welcome_qe_test.go`, and the two net-new seam `TestMain` files
  `internal/tui/qe_seam_test.go`, `internal/tui/screens/qe_seam_test.go` (matches the
  existing convention of registering `*_qe_test.go` overlay tests). Note the existing
  `internal/tui/screens/persona_preset_qe_test.go` is already registered.
- `inlineAnchors` (net-new, model.go — **5** entries, in addition to the existing
  `model.PresetQESDET` entry which is satisfied independently by line 589 and does NOT
  cover the SDDMode literal):
  - `{"file":"internal/tui/model.go","mustContain":"qeFilterPickerFlow"}`
  - `{"file":"internal/tui/model.go","mustContain":"qeSuppressCommunityTools"}`
  - `{"file":"internal/tui/model.go","mustContain":"qeSuppressOpenCodePlugins"}`
  - `{"file":"internal/tui/model.go","mustContain":"qeWelcomeCanonicalCursor"}`
  - `{"file":"internal/tui/model.go","mustContain":"SDDMode: model.SDDModeSingle"}`
    — its own dedicated entry (BLOCKER fix: without it the SDDMode anchor is unguarded).
- `inlineAnchors` (net-new, welcome.go): `{"file":"internal/tui/screens/welcome.go","mustContain":"qeWelcomeOptions"}`.
- `inlineAnchors` (updated from append→seam-aware wrapper): persona.go entry
  `qePersonaOptions` → `qePersonaOptionsForBuild`; preset.go entry `qePresetOptions` →
  `qePresetOptionsForBuild`. None may be omitted, especially the Welcome cursor-remap anchor
  (`qeWelcomeCanonicalCursor`) and the SDDMode literal — the two drift-sensitive additions.

## Migration / Rollout
No migration. Fork-unconditional; rollback = delete the 2 net-new files
(`model_qe.go`, `welcome_qe.go`), revert the `persona_qe.go`/`preset_qe.go` bodies to
append semantics, revert the **8** upstream anchors, and drop the overlay.json entries.

## Open Questions
- [ ] None blocking. Confirm Welcome-collapse keep-set (proposed 7 entries: Start install,
  Upgrade, Sync, Upgrade+Sync, Manage backups, Managed uninstall, Quit). If the keep-set
  changes, the remap table and the `len()==7` structural invariant MUST be updated together.
