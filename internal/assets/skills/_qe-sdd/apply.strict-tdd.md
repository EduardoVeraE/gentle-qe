# Strict TDD Module — Apply Phase (Test-Automation Framing)

> **This module is loaded ONLY when Strict TDD Mode is enabled AND a test runner is available.**
> If you are reading this, the orchestrator already verified both conditions. Follow every instruction.

## TDD Philosophy (Test-Automation Framing)

This module reframes the classic RED-GREEN-REFACTOR cycle for **test-automation deliverables** — Page Objects, fixtures, API clients, load-test scripts. There is no application feature behind this cycle to build; the deliverable IS the test-support code that automates a scenario. You write the failing oracle first, then the minimal test-support helper that satisfies it, then you clean the test-support code up.

### The Three Laws (Test-Automation Version)

1. **Do NOT write test-support code (Page Object, fixture, helper)** until you have a failing assertion/oracle
2. **Do NOT write more of the failing assertion** than is necessary to fail meaningfully
3. **Do NOT write more test-support code** than is necessary to make that assertion pass

## TDD Implementation Cycle (Automation)

For EVERY scenario assigned to you, follow this cycle strictly:

```
FOR EACH SCENARIO:
├── 0. SAFETY NET (only if modifying existing test-support files)
│   ├── Run existing tests for the fixtures/POMs being modified
│   ├── Capture baseline: "{N} tests passing"
│   ├── If any FAIL → STOP, report as "pre-existing failure"
│   └── This baseline proves you did not break what already worked
│
├── 1. UNDERSTAND
│   ├── Read the scenario's GIVEN/WHEN/THEN and its explicit **Oracle** line
│   ├── Read the Testing Strategy design decision for this scenario's layer
│   ├── Read existing fixtures/Page Objects/helpers (match the style)
│   └── Determine test layer (unit/integration/system/acceptance)
│
├── 2. RED — Write the failing assertion/oracle FIRST
│   ├── Write the assertion that encodes the scenario's THEN clause
│   ├── The assertion MUST reference a Page Object, fixture, or client method
│   │   that does NOT exist yet (this guarantees failure — no need to execute
│   │   to confirm)
│   ├── If the Page Object/fixture/client already exists:
│   │   └── Write an assertion for the NEW behavior not yet supported
│   └── GATE: Do NOT proceed to GREEN until the failing assertion is written
│
├── 3. GREEN — Write the MINIMUM test-support code to pass
│   ├── Implement ONLY the Page Object method / fixture / helper the failing
│   │   assertion needs
│   ├── Fake It is VALID here (hardcoded fixture data is OK for a first pass)
│   ├── EXECUTE the test → must PASS
│   │   ├── ✅ Passed → proceed to TRIANGULATE or REFACTOR
│   │   └── ❌ Failed → fix the test-support code, NOT the assertion
│   └── GATE: Do NOT proceed until GREEN is confirmed by execution
│
├── 4. TRIANGULATE (MANDATORY for most scenarios)
│   ├── DEFAULT: triangulation is REQUIRED. You need a compelling reason to skip it.
│   ├── Add a second case with DIFFERENT inputs/expected outputs
│   ├── EXECUTE → if Fake It breaks (hardcoded fixture no longer works):
│   │   └── Generalize the Page Object/fixture to handle real variability
│   ├── WATCH OUT for GREEN that passes trivially:
│   │   ├── If the assertion passes because the element/response isn't
│   │   │   exercised → NOT a real GREEN
│   │   ├── If a loop iterates 0 times → NOT a real GREEN
│   │   └── A real GREEN means: the automated interaction RAN and produced
│   │       the expected output
│   └── GATE: All spec scenarios for this batch must have tests before REFACTOR
│
├── 5. REFACTOR — Clean the test-support code without breaking the oracle
│   ├── Extract shared Page Object methods / fixtures
│   ├── Remove duplication across specs
│   ├── Improve naming, remove magic strings/selectors
│   ├── EXECUTE tests after EACH refactor step → must STILL PASS
│   │   ├── ✅ Still passing → refactor is safe, continue
│   │   └── ❌ Failed → REVERT that refactor step, try smaller
│   └── GATE: The oracle (the original failing assertion) must still pass
│       green after every refactor
│
├── 6. Mark scenario complete [x]
└── 7. Note any deviations or issues discovered
```

No step in this cycle instructs writing, completing, or modifying feature/application logic. If a scenario appears to require that, STOP and report it to the orchestrator — that work belongs to a dev SDD apply cycle, not this QE cycle.

## Choosing Test Layer

```
Determine test layer by WHAT the scenario exercises:
├── Pure logic, data transformation, oracle computation
│   └── Unit test
├── API contract, service boundary, module-to-module interaction
│   └── Integration test (use `api-testing` skill conventions)
├── Full user journey through the running system
│   └── System/E2E test (use `playwright-e2e-testing` skill conventions)
├── Non-functional (load, security, a11y)
│   └── k6-load-test / qa-owasp-security / a11y-playwright-testing conventions
└── Default: Unit test (always the fallback)
```

## Assertion Quality Rules (MANDATORY)

**Every assertion must verify REAL behavior exercised by the automation.** A test that passes without exercising the interaction under test gives false confidence.

### Banned Assertion Patterns (NEVER write these)

```
expect(true).toBe(true)              # ❌ Tautology
assert 1 == 1                        # ❌ Always passes
expect(result).toHaveLength(0)       # ❌ ONLY valid with an explicit empty-state setup
expect(result).toBeDefined()         # ❌ Alone is useless — assert the actual value
```

### What Makes a REAL Assertion

1. **Exercises the automated interaction** — the test drives a Page Object, calls an API client, or runs a load script
2. **Asserts a specific output** — compares against a concrete expected value derived from the scenario's THEN clause
3. **Would FAIL if the interaction were broken** — if the underlying behavior regresses, THIS test breaks

### Mock Hygiene Rules

**If you need more mocks than assertions, you are testing at the WRONG level.** Extract fixture/setup logic to a reusable helper before reaching for another mock. 7+ mocks in one spec file means the scenario belongs at a different layer.

## Return Summary Extension

When Strict TDD Mode is active, your return summary MUST include:

```markdown
### TDD Cycle Evidence (Test-Automation)
| Scenario | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|----------|-----------|-------|------------|-----|-------|-------------|----------|
| U1.1 | `path/spec.ext` | Unit | ✅ 5/5 | ✅ Written | ✅ Passed | ✅ 2 cases | ✅ Clean |

### Test Summary
- **Total scenarios automated**: {N}
- **Total tests passing**: {N}
- **Layers used**: Unit ({N}), Integration ({N}), System ({N})
- **Fixtures/Page Objects created**: {N}
```

## Rules (Strict TDD specific, test-automation framing)

- NEVER write test-support code (Page Object/fixture/helper) before writing its failing assertion — the ONE rule that cannot be broken
- NEVER skip the GREEN execution gate — you MUST run the test and confirm it passes
- NEVER skip triangulation when the scenario has multiple partitions
- NEVER write trivial assertions (see Banned Assertion Patterns above)
- ALWAYS verify that every assertion exercises the real automated interaction and asserts a SPECIFIC expected value
- ALWAYS report the TDD Cycle Evidence table — the verify phase will check it
- No step in this module instructs writing or completing feature/application logic — only test-automation code (specs, fixtures, Page Objects, helpers)
