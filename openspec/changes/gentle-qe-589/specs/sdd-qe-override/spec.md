# SDD QE Override

## ADDED Requirements

### Requirement: QE Content Replaces Dev Content In All Injector Paths

The system MUST serve QE test-design content, not dev content, for every SDD
phase that ships a QE asset — exactly these 7: explore, propose, spec, design,
tasks, apply, verify — across all three independent content-serving injector
paths: `internal/components/skills/inject.go` (`InjectWithCapability`),
`internal/components/sdd/inject.go` (step 3c), and
`internal/components/sdd/prompts.go` (`WriteSharedPromptFiles`). This is a pure
replacement — no mode selector, flag, or env var MUST restore dev content.
`archive` is explicitly NOT part of this set (see the fail-open scenario
below).

A fourth, uninstall-robustness anchor in
`internal/components/uninstall/service.go` MUST also exist: it does not serve
content, but MUST prevent the `_qe-sdd` overlay from being deleted by the
generic skills uninstall path (see the "Uninstall Never Removes The QE
Overlay" requirement below). Together these are the 4 anchor points this
change registers in `tools/qe-overlay/overlay.json`.

#### Scenario: Oracle test proves QE override across all injector paths (structural check)

- GIVEN a Go test that invokes each of the 3 content-serving injector paths
  for a given SDD phase
- WHEN the test requests content for that phase from each path
- THEN the returned content contains ALL of the required section
  headings/structural elements defined for that phase in the "SDD QE Phase
  Contract" spec (e.g. the design-phase asset MUST contain a named-ISTQB-
  technique section, a test-pyramid reference, and a risk/defect-clustering
  prioritization section — not merely those words present anywhere in the
  document)
- AND the returned content does NOT contain any term from that phase's
  negative-marker set, where the set is derived per the "Negative markers are
  phase-specific and dev-verified" scenario below (never a generic, unverified
  keyword list)
- AND this structural-plus-negative-marker assertion holds independently for
  exactly the 7 phases with a QE asset — explore, propose, spec, design,
  tasks, apply, verify — across all 3 content-serving paths (a missing
  anchor, a missing required section, or a verified dev marker found in the
  served content MUST fail this test)
- AND for phases whose QE asset carries its own `model-capable`/`model-small`
  sections (`apply`, `verify`), each content-serving path MUST extract the
  section matching the resolved capability before serving it: the served
  content MUST contain exactly one frontmatter fence pair and MUST NOT
  contain the `<!-- section:model-capable -->` / `<!-- section:model-small -->`
  markers or both variants concatenated

#### Scenario: Oracle rejects trivial keyword-stuffed content

- GIVEN a candidate asset that contains a phase's required ISTQB/QE keywords
  only as loose terms (e.g. pasted into a footer) without the required
  section structure defined for that phase in the "SDD QE Phase Contract"
  spec
- WHEN the oracle test evaluates that asset
- THEN the test FAILS, because it validates structural presence of the
  required headings/sections per phase, not mere occurrence of keywords
  anywhere in the document

#### Scenario: Negative markers are phase-specific and dev-verified

- GIVEN the negative-marker set used to fail an asset for a given phase
- WHEN that set is constructed
- THEN each marker MUST be verified to actually appear in that phase's real
  upstream dev asset content
- AND each marker MUST be verified absent from the corresponding legitimate
  QE asset for that phase
- AND a marker MUST NOT be applied to a phase for which it was not verified
  against that phase's actual dev content (e.g. a term present in one
  phase's dev asset but never present in another phase's dev asset MUST NOT
  be used as that other phase's negative marker)

#### Scenario: No dual-mode escape hatch exists

- GIVEN any request for SDD phase content with no explicit override flag set
- WHEN the request is served
- THEN QE content is returned unconditionally
- AND no environment variable, flag, or config switches the response back to
  dev content

#### Scenario: Archive, onboard, and init fall through to upstream (fail-open)

- GIVEN a phase with no QE asset — archive, onboard, or init
- WHEN `skills.IsSDDSkill(id)` is true but no QE asset exists for that phase
- THEN the `_qe.go` helper returns `ok=false`
- AND the injector path falls back to serving the existing upstream (dev)
  content unchanged
- AND `archive` is never asserted to carry QE/ISTQB markers by the oracle
  test above

### Requirement: Overlay Registers Net-New Assets And Anchors

`tools/qe-overlay/overlay.json` MUST register the net-new QE phase asset files
under `overlayFiles` (not `netNewDirs` — the QE assets live flat under
`_qe-sdd/` with no per-directory `SKILL.md` at the root, so `netNewDirs`'
`verifyNetNewInstallable` check does not apply) and the 4 injector-path
anchors under `inlineAnchors`, so `qe-overlay verify` can detect drift or a
missing anchor.

#### Scenario: Overlay verify passes with full registration

- GIVEN overlay.json lists the QE phase asset files under `overlayFiles` and
  all 4 anchors under `inlineAnchors`
- WHEN `qe-overlay verify` runs
- THEN it exits 0
- AND its report confirms all 4 anchors are present in their injector files

#### Scenario: Missing anchor fails verify

- GIVEN one of the 4 anchors is removed or edited in an injector file
- WHEN `qe-overlay verify` runs
- THEN it exits non-zero
- AND it identifies which injector path is missing its anchor

### Requirement: Uninstall Never Removes The QE Overlay

The generic `ComponentSkills` uninstall path in
`internal/components/uninstall/service.go` walks every directory name under
the embedded `skills/` tree and removes any entry it does not explicitly
skip. It MUST skip `_qe-sdd` exactly like it already skips `_shared` and any
`sdd-*` directory, so the QE overlay is never deleted by a skills uninstall.

#### Scenario: Skills uninstall preserves the QE overlay directory

- GIVEN a real skills uninstall operation plan built for a component-skills
  uninstall
- WHEN the plan's operations are applied against a real filesystem that has
  both an installable skill directory (e.g. `go-testing`) and a `_qe-sdd`
  directory present under the same skills root
- THEN the installable skill directory MUST be removed
- AND the `_qe-sdd` directory and its files MUST still exist afterward

### Requirement: No Upstream Prompt Content Is Edited

The override MUST NOT modify existing upstream dev prompt/artifact text.
Only minimal anchors invoking a `_qe.go` helper (e.g. `qeSDDTestingContent`)
are permitted inside the 3 injector files.

#### Scenario: Upstream diff stays anchor-only

- GIVEN a diff of the 3 injector files against upstream
- WHEN the diff is inspected
- THEN it contains only anchor insertions guarded by
  `skills.IsSDDSkill(id)`
- AND no existing dev prompt/artifact string is altered or deleted
