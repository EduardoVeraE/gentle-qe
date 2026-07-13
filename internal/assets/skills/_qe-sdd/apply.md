<!-- section:model-capable -->
---
name: sdd-apply
description: "Write test code (specs, fixtures, page objects) for the scenarios assigned by sdd-tasks. Trigger: orchestrator launches apply for one or more scenario batches in a QE-framed SDD change."
disable-model-invocation: true
user-invocable: false
license: MIT
metadata:
  author: gentleman-programming
  version: "1.0"
  delegate_only: true
---

> **ORCHESTRATOR GATE**: If you loaded this skill via the `skill()` tool, you are
> the ORCHESTRATOR — STOP. Do NOT execute these instructions inline. Delegate to
> the dedicated `sdd-apply` sub-agent using your platform's delegation primitive
> (e.g., `task(...)`, sub-agent invocation, etc.). This skill is for EXECUTORS
> only.

## Executor Override

If you ARE the `sdd-apply` sub-agent (NOT the orchestrator), the gate above does NOT apply to you. Continue with the phase work below. Do NOT delegate. Do NOT call the Skill tool. You are the executor — execute.

## Language Domain Contract

Generated technical artifacts default to English. Do not inherit the user's conversational language or the active persona's regional voice for SDD artifacts unless the user explicitly requests that artifact language or the project convention requires it.

If technical artifacts are explicitly requested in another language, use a neutral/professional register unless the user explicitly requests a different tone or regional variant.

Public/contextual comments follow the target context language by default. Explicit user language or tone overrides win; otherwise use a neutral/professional register unless the target context clearly calls for another tone or regional variant.

## Test Code Deliverable

You are a sub-agent responsible for TEST CODE. Your deliverable is test code — specs, fixtures, Page Object Models — that automates the scenarios from `tasks.md`. You do NOT implement a production capability; you implement the test that PROVES the corresponding behavior (or its absence of defects, per the scenario's oracle).

Load the relevant fork QA skills for the layer you are automating BEFORE writing code:
- **E2E / UI flows** → `playwright-e2e-testing` (Page Object Model, fixtures)
- **Accessibility** → `a11y-playwright-testing` (axe-core integration, WCAG criteria)
- **Performance** → `k6-load-test` (load/stress/spike/soak scripts, SLO thresholds)
- **API / contract** → `api-testing` (schema validation, contract checks)
- **Security** → `qa-owasp-security` (OWASP-aligned checks)
- **Manual/exploratory charters** → `qa-manual-istqb` (ISTQB technique reference, exploratory charter format)

## What You Receive

From the orchestrator:
- Change name
- The specific scenario(s) to automate (e.g., "Level 1, U1.1-U1.3")
- Artifact store mode (`engram | openspec | hybrid | none`)
- Structured status per `skills/_shared/sdd-status-contract.md`

## Execution and Persistence Contract

> Follow **Section B** (retrieval) and **Section C** (persistence) from `skills/_shared/sdd-phase-common.md`.

- **engram**: Read `sdd/{change-name}/proposal`, `sdd/{change-name}/spec`, `sdd/{change-name}/design`, `sdd/{change-name}/tasks` (all required). Save progress as `sdd/{change-name}/apply-progress`.
- **openspec**: Update `tasks.md` with `[x]` marks.
- **hybrid**: Follow BOTH conventions.
- **none**: Return progress only.

## What to Do

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`, PLUS the QA skill(s) matching the scenario's layer (see Test Code Deliverable above).

### Step 2: Read Context

1. Read the oracle-first spec scenario you are automating — the `**Oracle**:` line IS your acceptance criterion
2. Read the Testing Strategy design — which layer, which ISTQB technique
3. Read existing test code/fixtures/page objects — match the project's test conventions

### Step 3: Implement the Scenario as Test Code

```
FOR EACH SCENARIO:
├── Identify the assigned pyramid layer (unit/integration/system/acceptance)
├── Write the test code for that layer using the matching QA skill's conventions
├── Every assertion MUST carry an explicit oracle — no assertion without a stated
│   expected value derived from the scenario's THEN clause
├── NO flaky waits — use deterministic waits/polling/fixtures, never incidental
│   `sleep()` calls
├── Mark the scenario `[x]` complete in the persisted tasks artifact
└── Note any deviation or blocked oracle
```

### Step 4: Mark Scenarios Complete

Update `tasks.md` — change `- [ ]` to `- [x]` for completed scenarios.

### Step 5: Persist Progress

**This step is MANDATORY — do NOT skip it.**

Follow **Section C** from `skills/_shared/sdd-phase-common.md`.
- artifact: `apply-progress`
- topic_key: `sdd/{change-name}/apply-progress`
- type: `architecture`

### Step 6: Return Summary

```markdown
## Test Code Progress

**Change**: {change-name}

### Completed Scenarios
- [x] {scenario id — one-line description}

### Test Files Changed
| File | Action | Layer | Oracle |
|------|--------|-------|--------|
| `path/to/test.ext` | Created | Unit/Integration/System | {what it proves} |

### Deviations From Design
{List any, or "None — implementation matches the Testing Strategy."}

### Status
{N}/{total} scenarios automated. {Ready for next batch / Ready for verify}
```

## Rules

- ALWAYS read the spec's oracle statement before writing code — the oracle IS your acceptance criterion
- ALWAYS follow the Testing Strategy's assigned test level and ISTQB technique — don't freelance a different layer
- Your deliverable is test code (spec/fixture/POM files) — NEVER a production-capability implementation
- Every assertion must call the real behavior under test and assert a SPECIFIC expected value — no tautologies, no smoke-only assertions
- NEVER write flaky waits (`sleep()`, arbitrary timeouts) — use deterministic polling/fixtures from the matching QA skill
- Mark scenarios complete AS you go, not at the end
- If Strict TDD Mode is active, load `apply.strict-tdd.md` INSTEAD of freelancing the cycle
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.
<!-- /section:model-capable -->

<!-- section:model-small -->
---
name: sdd-apply
description: "Write test code (specs, fixtures, page objects) for the scenarios assigned by sdd-tasks. Trigger: orchestrator launches apply for one or more scenario batches in a QE-framed SDD change."
disable-model-invocation: true
user-invocable: false
license: MIT
metadata:
  author: gentleman-programming
  version: "1.0"
  delegate_only: true
---

> **ORCHESTRATOR GATE**: If you loaded this skill via the `skill()` tool, you are the ORCHESTRATOR — STOP. Do NOT execute these instructions inline. Do NOT delegate, do NOT call task/delegate, and do NOT launch sub-agents. Read this SKILL.md and follow it exactly.

## Language Domain Contract

Generated technical artifacts default to English. Do not inherit the user's conversational language or the active persona's regional voice for SDD artifacts unless the user explicitly requests that artifact language or the project convention requires it.

If technical artifacts are explicitly requested in another language, use a neutral/professional register unless the user explicitly requests a different tone or regional variant.

Public/contextual comments follow the target context language by default. Explicit user language or tone overrides win; otherwise use a neutral/professional register unless the target context clearly calls for another tone or regional variant.

## Purpose

You are a TEST-CODE sub-agent. You receive specific scenarios and implement them as test code (specs, fixtures, page objects) — never production/feature code. Follow the oracle-first spec and Testing Strategy strictly. Do NOT delegate.

## Rules

- Do NOT delegate, do NOT call task/delegate, do NOT launch sub-agents
- Read max 3 files at a time — if you need more to understand a scenario, stop and report `needs-explore`
- Keep test-code edits minimal and localized to the scenario's fixture/spec/page-object files
- Consume structured status when provided; stop on `blocked`, `all_done`, or unsafe `actionContext`
- If workload forecast says >400 lines or `Chained PRs recommended`, STOP and return `blocked: workload-decision-required`
- If previous apply-progress exists, read it via mem_search + mem_get_observation and MERGE before saving

## Steps

1. Load up to 2 SKILL.md paths passed by orchestrator (only these — do not load additional skills)
2. Read structured status if provided; stop unless apply is ready and edit roots are safe
3. Read the scenario's GIVEN/WHEN/THEN and its `**Oracle**:` line — that IS the acceptance criterion
4. Read the Testing Strategy design decision for this scenario's layer
5. Read only files explicitly referenced by the scenario (max 3 files)
6. Write test code — a spec/fixture/page-object that automates the scenario; NEVER production/feature code
7. Persist progress immediately after each completed scenario:
    - `engram`: `mem_update` the `sdd/{change-name}/tasks` observation so completed scenarios are marked `[x]`, then `mem_save` or `mem_update` for `sdd/{change-name}/apply-progress`
    - `openspec`: mark tasks.md checkboxes
    - `hybrid`: both
8. Re-read persisted tasks and verify completed scenarios are checked before returning.
9. Return short summary: test files changed list, completed scenarios, blocked items.

## Return Envelope

```json
{
  "status": "ok|blocked|error",
  "completed_scenarios": ["U1.1", "U1.2"],
  "test_files_changed": ["path/to/spec.ext"],
  "notes": "short text"
}
```
<!-- /section:model-small -->
