# Design: SDD cycle produces test-design artifacts (Gentle-QE)

## Technical Approach

The SDD phase content that reaches any assistant is the per-phase SKILL/prompt
markdown. Three independent injector paths embed and write that markdown. This
change ships **net-new, full-content QE phase assets** and swaps upstream dev
content for QE content **at read time**, anchored in the three injector paths
(plus a fourth uninstall-robustness anchor) through a single fork-owned helper.
No upstream prompt/artifact CONTENT is edited; the fork touches only minimal Go
anchors plus net-new assets, so every upstream merge re-applies cleanly.

Two layers: **(1)** the override architecture (the technical HOW), **(2)** the
ISTQB/SDET substance each phase asset must carry (the WHAT-it-says).

---

## Layer 1 — Override Architecture

### Decision: Full-content per-phase override, not string-diff

| Option | Tradeoff | Decision |
|--------|----------|----------|
| String-map (`qeNeutralizeRegionalVoice` pattern) | Fine for 2-line voice tweaks; brittle and unmaintainable against 150–250-line SKILL.md that evolve to v2/v3 | Rejected |
| Full-content asset swap at read time | One source of truth per phase, decoupled from upstream text drift | **Chosen** |

**Rationale**: the proposal locks a pure REPLACEMENT (test replaces dev, no mode
selector). A whole-document swap is the only approach that scales and survives
upstream reformatting of the dev prompts.

### Decision: Net-new assets under the already-embedded `skills/` tree

**Choice**: place QE assets FLAT at `internal/assets/skills/_qe-sdd/`. Two kinds:
- **7 phase SKILL overrides**: `{explore,propose,spec,design,tasks,apply,verify}.md`.
- **2 strict-TDD sibling overrides** (see next decision): `apply.strict-tdd.md`,
  `verify.strict-tdd-verify.md`.

`archive`, `onboard`, `init` get **no** asset → fail-open to upstream.
**Alternatives considered**: a new top-level `skills-qe/` dir (would force editing
the upstream `//go:embed all:skills …` line in `assets.go` — forbidden fork
touch); reusing `skills/sdd-*/` (collides with upstream skill dirs).
**Rationale**: the `_qe-sdd` prefix mirrors `_shared`, is already captured by
`all:skills`, and is **never treated as an installable skill** (skill IDs are an
explicit list; catalog never registers it). **Correction from review**: `_qe-sdd`
IS enumerated by raw `skills/` directory walkers — specifically the uninstaller
(`uninstall/service.go`, which only skips `sdd-` and `_shared`). That is Path 4
below. Registration goes in `overlayFiles`, **not** `netNewDirs`:
`qe-overlay verify`'s `verifyNetNewInstallable` (`tools/qe-overlay/main.go:222`)
requires a `SKILL.md` at a netNewDir root and would FAIL for `_qe-sdd`.

### Decision: Cover strict-TDD siblings — reframe TDD to test automation

`skills/sdd-apply/strict-tdd.md` (18KB) and `skills/sdd-verify/strict-tdd-verify.md`
(12KB) OVERRIDE Step 4 with "production code" framing when `StrictTDD=true`. A
SKILL.md-only swap would leave the fork emitting dev code under strict mode. Ship
two sibling QE assets that reframe **RED → GREEN → REFACTOR** to automation:
RED = write the failing assert/**oracle**; GREEN = the minimal Page Object/fixture
to make it pass; REFACTOR = clean the test without breaking it.
**Consequence**: the helper can no longer map phase→single file. It maps
`(skillID, fileName)` to a sibling asset, and **Path 1 must override EVERY `.md`
in the skill dir**, not just `SKILL.md`.

### Decision: One fork-owned helper, gated by `IsSDDSkill`

New file `internal/components/skills/inject_qe.go` (net-new `_qe.go`):

```go
// QESDDTestingContent returns QE override content for a file inside an SDD skill
// dir. skillID e.g. "sdd-apply"; fileName e.g. "SKILL.md" or "strict-tdd.md".
// ok=false for non-SDD ids or when no matching QE asset exists (fail-open).
func QESDDTestingContent(skillID, fileName string) (string, bool) {
    if !IsSDDSkill(model.SkillID(skillID)) {        // reuse existing gate
        return "", false
    }
    phase := strings.TrimPrefix(skillID, "sdd-")     // sdd-apply → apply
    asset := "skills/_qe-sdd/" + phase + ".md"       // SKILL.md → phase.md
    if fileName != "SKILL.md" {
        asset = "skills/_qe-sdd/" + phase + "." + fileName // apply.strict-tdd.md
    }
    content, err := assets.Read(asset)
    if err != nil || len(content) == 0 {
        return "", false                             // fail-open to upstream
    }
    return content, true
}
```

Exported because paths 2 and 3 live in package `sdd`, which already imports
`skills`. `IsSDDSkill` already exists (`skills/inject.go`) — reused.

### The four anchor points (verified against current code)

**Path 1 — `internal/components/skills/inject.go` → `InjectWithCapability`.**
Fan-out: **every adapter with `SupportsSkills()`** (Claude, Cursor, Kiro, Kimi and
any future skill-capable adapter — not only those four). Inside the existing
`WalkDir`, override EVERY walked file (covers `SKILL.md` AND `strict-tdd*.md`):

```go
if qe, ok := QESDDTestingContent(string(id), filepath.ToSlash(relPath)); ok {
    content = qe
}
```
Full swap. The subsequent `extractModelSection` is a harmless no-op — the QE
assets carry no `model-capable`/`model-small` markers, so `capable` and `small`
installs receive identical QE content. **Accepted trade-off**: QE test-design
substance does not vary by model tier (unlike some upstream SKILLs).

**Path 2 — `internal/components/sdd/inject.go` → step 3c** (native sub-agent
wrappers, e.g. `claude/agents/sdd-apply.md`). The wrapper carries dev frontmatter
(`description: "Implement code changes… writes code"`) **and** a dev body. A
marked-section append would leave both intact — that is NOT a pure replacement.
**Body-swap** instead: preserve `name`/`model` (`{{CLAUDE_MODEL}}` already
resolved earlier in the loop)/`effort`/`tools`, rewrite `description:` to a QE
line, and replace the entire body with QE content. Gate with the existing
`isMarkdownSubAgentPromptFile(entry.Name())` (already used at ~line 703) so it
applies ONLY to `.md` prompt wrappers and never to Kimi's sibling `.yaml` agents.
Run it **before** `injectCodeGraphGuidanceIntoPrompt` so CodeGraph guidance lands
in the new QE body:

```go
if isMarkdownSubAgentPromptFile(entry.Name()) {
    phase := strings.TrimSuffix(entry.Name(), ".md")           // sdd-apply
    if qe, ok := skills.QESDDTestingContent(phase, "SKILL.md"); ok {
        contentStr = qeSwapNativeAgentBody(contentStr, qe, qeSubAgentDescription(phase))
    }
    contentStr = injectCodeGraphGuidanceIntoPrompt(contentStr, opts.CodeGraphGuidanceMarkdown)
}
```

`qeSwapNativeAgentBody` is a net-new fork helper (`internal/components/sdd/inject_qe.go`).
There is no reusable split helper: `skillregistry.parseFrontmatter` lives in
another package and only EXTRACTS name/description (it cannot reconstruct). The
helper reuses the same fence detection (`"---\n" … "\n---"`), keeps the
frontmatter block, replaces only the `description:` entry (folded `>` scalar
included) with the QE description, and swaps the body:

```go
// qeSwapNativeAgentBody preserves YAML frontmatter (name/model/effort/tools),
// rewrites description: to qeDescription, replaces the body with qeBody.
// Defensive no-op if the input has no frontmatter fence.
func qeSwapNativeAgentBody(wrapper, qeBody, qeDescription string) string
```

Native-agent strict-TDD is served through Path 1 (the skills dir the wrapper
references), so Path 2 only body-swaps the phase document.

**Path 3 — `internal/components/sdd/prompts.go` → `WriteSharedPromptFiles`.**
Serves OpenCode / Kilocode shared prompt files. The swap MUST run **before
`injectCodeGraphGuidanceIntoPrompt` (line 74)** — not merely before write — or the
CodeGraph guidance is discarded:

```go
content := extractModelSection(skillContent, capability)
if qe, ok := skills.QESDDTestingContent(phase, "SKILL.md"); ok {   // BEFORE guidance
    content = qe
}
content = injectCodeGraphGuidanceIntoPrompt(content, guidance)
```
Path 3 reads only `SKILL.md`; strict-TDD siblings are handled by Path 1 wherever
skills are installed.

**Path 4 — `internal/components/uninstall/service.go` (~line 532)** (correction).
The `ComponentSkills` uninstall walks the embedded `skills/` dir and removes every
entry except those prefixed `sdd-` or named `_shared`. `_qe-sdd` would be treated
as a QA skill and removed. Add it to the skip set for robustness:

```go
if !entry.IsDir() || strings.HasPrefix(entry.Name(), "sdd-") ||
    entry.Name() == "_shared" || entry.Name() == "_qe-sdd" {
    continue
}
```

### Overlay registration (`tools/qe-overlay/overlay.json`)

- `overlayFiles` (**not** `netNewDirs`): add all 9 QE assets
  (`internal/assets/skills/_qe-sdd/*.md`), the two helper files
  (`internal/components/skills/inject_qe.go`, `internal/components/sdd/inject_qe.go`),
  and the verification test file. `overlayFiles` uses an `isFile` check, so no
  `SKILL.md`-at-root requirement applies.
- `inlineAnchors`: **four** entries, one per path:
  - `internal/components/skills/inject.go` → mustContain `QESDDTestingContent`
  - `internal/components/sdd/inject.go` → mustContain `qeSwapNativeAgentBody`
  - `internal/components/sdd/prompts.go` → mustContain `QESDDTestingContent`
  - `internal/components/uninstall/service.go` → mustContain `_qe-sdd`

`qe-overlay verify` then fails loudly if any anchor is dropped by a future merge.

### Data flow

```
internal/assets/skills/_qe-sdd/*.md   (net-new, //go:embed all:skills)
  7 phase SKILL overrides: explore,propose,spec,design,tasks,apply,verify
  2 strict-TDD siblings:   apply.strict-tdd.md, verify.strict-tdd-verify.md
  (archive/onboard/init: NO asset → fail-open to upstream)
                    │
                    ▼
   skills.QESDDTestingContent(skillID, fileName)   ── gate: IsSDDSkill
        ┌───────────┬───────────────┬──────────────┐
        ▼           ▼               ▼              (robustness)
   Path 1        Path 2          Path 3           Path 4
 InjectWith-   inject.go 3c    WriteShared-      uninstall
 Capability    body-swap       PromptFiles       skip _qe-sdd
 (ALL .md,     (.md only,      (SKILL.md,        (do not remove
  full swap:   frontmatter-     swap BEFORE       QE assets on
  SKILL +      preserving,      CodeGraph)        skills uninstall)
  strict-tdd)  desc rewrite)
        │           │               │
  all SupportsSkills  native .md    OpenCode/
  adapters           wrappers       Kilocode
```

### File changes

| File | Action | Description |
|------|--------|-------------|
| `internal/assets/skills/_qe-sdd/{explore,propose,spec,design,tasks,apply,verify}.md` | Create | 7 phase SKILL overrides (Layer 2). No `archive.md` |
| `internal/assets/skills/_qe-sdd/apply.strict-tdd.md` | Create | RED-GREEN-REFACTOR reframed to automation |
| `internal/assets/skills/_qe-sdd/verify.strict-tdd-verify.md` | Create | Strict-TDD verify reframed to automation |
| `internal/components/skills/inject_qe.go` | Create | `QESDDTestingContent(skillID, fileName)` helper + gate |
| `internal/components/sdd/inject_qe.go` | Create | `qeSwapNativeAgentBody` + `qeSubAgentDescription` |
| `internal/components/skills/inject.go` | Modify (anchor) | Path 1: QE swap for every walked `.md` in `WalkDir` |
| `internal/components/sdd/inject.go` | Modify (anchor) | Path 2: body-swap in step 3c, before CodeGraph |
| `internal/components/sdd/prompts.go` | Modify (anchor) | Path 3: QE swap before `injectCodeGraphGuidanceIntoPrompt` |
| `internal/components/uninstall/service.go` | Modify (anchor) | Path 4: skip `_qe-sdd` in skills uninstall |
| `internal/components/sdd/qe_sdd_override_qe_test.go` | Create | Cross-path verification test |
| `tools/qe-overlay/overlay.json` | Modify | overlayFiles (9 assets + 2 helpers + test) + 4 inlineAnchors |

---

## Layer 2 — Per-phase ISTQB / SDET content contract

Every asset opens with the same mandate and closes every section with the
**oracle question**. Encoded quality principles (verbatim spirit across assets):
testing shows the PRESENCE of defects, never absence; exhaustive testing is
impossible → risk-based prioritization + equivalence partitioning; shift-left;
the test pyramid is never inverted; a flaky test is a BROKEN test. Each phase
must force: **What ORACLE are we using? How do we know a test caught a real
defect?**

### explore (most shift-left — risk, testability, defect clustering, oracle gaps)
The earliest phase, viewed through an SDET/ISTQB lens. Map the area under change:
- **Risk analysis**: enumerate quality risks and rank by **likelihood × impact**;
  this ranking drives every downstream prioritization decision.
- **Testability**: assess how observable/controllable the area is (seams, hooks,
  determinism, hidden state) and flag testability debt that will force E2E where
  a lower layer should suffice.
- **Defect clustering**: apply the **20/80 rule** — locate the modules/paths where
  defects historically concentrate; that is where automation earns its cost.
- **Oracle inventory**: which oracles ALREADY exist (assertions, contracts,
  monitors, golden data) vs. which are MISSING and must be built. No oracle → no
  meaningful test. Testing shows presence of defects, never absence.
Output is a risk-and-testability map that seeds propose, not a solution sketch.

### propose (`capability → test-requirement / oracle`)
Reframe each product capability as a **quality risk** and the **test requirement**
that addresses it. For every capability: risk it carries, the oracle that would
detect a failure, and the acceptance boundary. Output is a capability→risk→oracle
contract, not a feature list.

### spec (oracle-first, GIVEN/WHEN/THEN)
Test requirements written oracle-first: each requirement states the observable
oracle before the steps. Scenarios in GIVEN/WHEN/THEN. Cover happy path plus the
boundary and negative partitions. Every scenario names its oracle and the defect
class it targets.

### design (the heaviest asset — Testing Strategy becomes the WHOLE document)
The dev "Testing Strategy" subsection is promoted to the entire document:
- **Test levels**: unit / integration / system / acceptance — what belongs at each.
- **ISTQB techniques**, explicitly chosen per requirement: equivalence
  partitioning, boundary value analysis, decision tables, state-transition
  testing, use-case testing; exploratory charters where scripted tests are weak.
- **Test pyramid**: majority unit, fewer integration, minimal E2E — never
  inverted; justify any E2E as irreplaceable.
- **Risk-based prioritization**: rank by likelihood × impact; spend automation
  where risk concentrates.
- **Defect clustering**: target the ~20% of modules holding ~80% of defects.
- Non-functional strategy where risk warrants (performance/k6, security/OWASP,
  a11y). Each layer names its oracle.

### tasks (scenarios/cases to automate, mapped to pyramid + technique)
Each task is an automatable scenario/case tagged with its **pyramid layer** and
the **ISTQB technique** that derived it. Prioritize by risk. No task is "write a
test" — it is "write THIS oracle-bearing case at THIS layer for THIS risk."

### apply (write TEST CODE — with the QA skills)
Produce test code (specs, fixtures, Page Object Models) using the fork QA skills:
`playwright-e2e-testing`, `a11y-playwright-testing`, `k6-load-test`, `api-testing`,
`qa-owasp-security`, `qa-manual-istqb`. Each artifact carries an explicit oracle
and a stable, deterministic assertion — no flaky waits, no incidental sleeps.
**Strict-TDD sibling (`apply.strict-tdd.md`)**: when strict mode is active it
overrides Step 4 and reframes the cycle to test automation —
**RED** = write the failing assert/oracle first; **GREEN** = the minimal Page
Object/fixture that makes it pass; **REFACTOR** = clean the test without breaking
it. Never "write production code."

### verify (execution + coverage + flakiness)
Run the suite; report coverage against risk (not vanity 100%). **A flaky test is
a broken test** — quarantine/fix/delete, never ignore. For every passing scenario
state the runtime oracle that proves it exercised the real behavior, and confirm
the defect it guards would actually make it fail.
**Strict-TDD sibling (`verify.strict-tdd-verify.md`)**: reframes strict-mode
verification to confirm each RED oracle genuinely failed before GREEN and that no
test passes vacuously — evidence the suite catches real defects, not that code was
shipped.

### archive (mechanical — NO QE asset)
`sdd-archive` stays purely mechanical and gets **no** `_qe-sdd` asset. The helper
returns `ok=false` for it (fail-open) and the upstream archive content is served
unchanged: link artifacts, record the final coverage/oracle summary, move the
change to archived. No new test-design decisions. `onboard` and `init` fall back
to upstream the same way.

---

## Design verification (required)

New Go test `internal/components/sdd/qe_sdd_override_qe_test.go` asserts the
override reaches all four paths. The oracle is layered — **string-matching alone
is a smoke check, not a content-quality oracle** (that is judgment-day / human
review):

1. **Path coverage**: drive each path for a representative adapter (skills path
   for a `SupportsSkills` adapter; step-3c for a native-subagent adapter;
   `WriteSharedPromptFiles` for OpenCode) and assert QE content is written.
2. **Structural oracle (primary)**: for phases
   `{explore,propose,spec,design,tasks,apply,verify}`, assert the phase-specific
   REQUIRED section headings exist — e.g. `design` must contain `## Test levels`,
   `Test pyramid`, `risk-based`; `explore` must contain risk / testability /
   defect-clustering / oracle-inventory headings. Structure proves the asset is
   the intended QE document, not merely that some keyword appears.
3. **Negative markers (verified against real upstream dev text)**: assert absence
   of dev-only strings that ACTUALLY occur in the corresponding upstream asset —
   e.g. the `sdd-apply` wrapper `description` "Implement code changes" / "writes
   code", and SKILL bodies "production code". Do NOT use `capability` (absent in
   5/7 dev phases → empty check) or "build the" (already present in dev content →
   false red). Each negative marker must be confirmed present in the upstream file
   it guards, otherwise it is dropped.
4. **Path 2 frontmatter oracle**: assert the body-swapped wrapper still contains
   `name:` and the resolved `model:` line and a QE `description:`, proving
   frontmatter was preserved and description rewritten.
5. **Gate/fail-open oracle**: `QESDDTestingContent` returns `ok=false` for a
   non-SDD id, for `judgment-day` (no `sdd-` prefix), and for
   `sdd-archive`/`sdd-onboard`/`sdd-init` (SDD-gated but no asset).
6. **Strict-TDD oracle**: `("sdd-apply","strict-tdd.md")` and
   `("sdd-verify","strict-tdd-verify.md")` return `ok=true` with RED/GREEN/REFACTOR
   automation framing; `("sdd-design","strict-tdd.md")` returns `ok=false`.

## Migration / Rollout

No migration. Rollback per proposal: delete `_qe-sdd/`, remove the two
`inject_qe.go` helpers, revert the 4 anchors (Path 1–3 swaps + Path 4 uninstall
skip), drop the overlay.json entries — SDD returns to upstream dev behavior.

## Open Questions

- [ ] Confirm the exact negative-marker set with the maintainer — each must be
      verified present in the upstream dev file it guards (see verification step 3).
- [x] RESOLVED: `sdd-explore` receives a QE asset (most shift-left phase).
- [x] RESOLVED (review): Path 2 uses BODY-SWAP (frontmatter-preserving,
      description-rewriting), not marked-section append.
- [x] RESOLVED (review): strict-TDD covered via 2 sibling assets; Path 1 overrides
      every `.md` in the skill dir; helper keyed by `(skillID, fileName)`.
- [x] RESOLVED (review): assets registered under `overlayFiles` (not `netNewDirs`),
      and `_qe-sdd` excluded from the uninstaller enumeration.
- [x] ACCEPTED: `sdd-archive`/`sdd-onboard`/`sdd-init` stay upstream-neutral via
      the helper's fail-open (SDD-gated but no `_qe-sdd` asset).
