# Proposal: Simplify TUI Installer Flow for QE Users

## Intent

The Gentle-QE fork's installer inherits ~20+ upstream screens, including dev-only concepts (per-agent model pickers, SDD Mode, Community Tools, OpenCode Plugins) that overwhelm QE users. Goal: a fast, coherent QE path — agents → persona SDET → preset QE → skills → install — with sensible silent defaults for everything else. This extends the "pure QE replacement" philosophy of gentle-qe-589: the fork build ALWAYS shows the simplified QE flow, unconditionally.

## Scope

### In Scope
- **Hide 6 screens** (apply silent defaults): Claude/Kiro/Codex model pickers, SDDMode, CommunityTools, OpenCodePlugins.
- **Keep StrictTDD visible** — valuable SDET discipline.
- **Filter Persona/Preset option lists** to QE-only (persona SDET; QE presets) — screens stay, lists shrink.
- **Collapse Welcome menu** to QE essentials (install + must-haves).
- Net-new `internal/tui/model_qe.go` owns all skip/default/filter logic (mirrors `preset_qe.go`/`persona_qe.go`).
- Register all new anchors in `tools/qe-overlay/overlay.json`.

### Out of Scope
- SDD-cycle override (already done in gentle-qe-589).
- Any non-installer change.

## Capabilities

### New Capabilities
- `qe-installer-flow`: unconditional QE-simplified installer behavior (screen hiding, silent defaults, list filtering).

### Modified Capabilities
- None (upstream logic untouched; behavior added via overlay anchors).

## Approach

Extend the EXISTING lever — `pickerFlowSlice` + `shouldShow*` predicates already skip screens conditionally (precedent: `isPiOnlyAgents()` full skip). Add ONE delegating line per touch point (pickerFlowSlice return, CommunityTools/OpenCodePlugins predicates, WelcomeOptions, Persona/Preset list filters); all logic lives in `model_qe.go`. Hiding a screen NEVER removes functionality the QE preset needs (e.g., keep ComponentSDD — hide SDDMode via a default, do not strip the component).

**Proposed silent defaults (to confirm in design):**
- SDDMode → single (leaner QE path; recommend, justify in design).
- Per-phase model → reuse existing system defaults.
- CommunityTools → none.
- OpenCodePlugins → none (OpenCode-only anyway).

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| pickerFlowSlice invariant: filter must NOT read `m.Screen`, only `m.Selection`/new QE fields | High | Keep predicate logic in model_qe.go reading Selection only |
| List filtering (Persona/Preset) is a DIFFERENT lever than screen hiding | Med | Treat as separate anchors; document in design |
| Upstream model.go drift breaks anchors | Med | Minimal 1-line anchors; overlay.json mustContain checks catch drift |

## Rollback Plan

Remove `model_qe.go` and revert the 1-line anchors (pickerFlowSlice, 2 predicates, WelcomeOptions, list filters); drop the new entries from overlay.json. Upstream behavior restores fully — zero upstream logic was edited.

## Dependencies
- gentle-qe-589 (QE preset/persona overlay) already merged.

## Success Criteria
- [ ] Fork build shows only QE-essential screens; 6 dev screens hidden with working defaults.
- [ ] Persona/Preset lists show QE-only options; Welcome collapsed.
- [ ] Zero upstream content edits; overlay re-appliable after upstream merge.
