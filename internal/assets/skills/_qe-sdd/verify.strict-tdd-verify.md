# Strict TDD Module — Verify Phase (Test-Automation Framing)

> **This module is loaded ONLY when Strict TDD Mode is active during verify.**

## TDD Verification Philosophy (Test-Automation Framing)

You are auditing whether the RED-GREEN-REFACTOR cycle was followed HONESTLY for
test-automation deliverables. The question is never "did code get shipped" —
it is "does every reported RED prove the assertion genuinely failed before
GREEN, and does no test pass vacuously." A suite that reports green without
ever having been red proves nothing; it is evidence of shipping, not evidence
the suite catches real defects.

## Step 5a: TDD Compliance Check

For each scenario in the apply-progress TDD Cycle Evidence table:

```
FOR EACH scenario row:
├── Confirm a RED entry exists — the failing assertion/oracle was written
│   BEFORE any Page Object/fixture/helper code
├── Confirm the RED failure was a REAL failure:
│   ├── ✅ Real: assertion referenced a Page Object method/fixture/client call
│   │   that did not exist yet, or asserted new behavior not yet supported
│   ├── ❌ Vacuous: assertion would have passed even without the fixture
│   │   (tautology, always-true condition, unreachable branch)
│   └── If vacuous → CRITICAL: "RED did not genuinely fail"
├── Confirm GREEN shows an EXECUTION result, not just "implemented"
│   └── Missing execution evidence → CRITICAL: "GREEN not confirmed by run"
├── Confirm TRIANGULATE cases exist for scenarios with multiple partitions
│   └── Missing → WARNING unless "Triangulation skipped: {reason}" is present
│       and the reason is structurally valid (single-output, no branching)
└── Confirm REFACTOR (if claimed) did not break the original oracle
    └── Oracle broken after refactor and not reverted → CRITICAL
```

## Step 5b: Assertion Quality Audit (MANDATORY)

Sample the test-automation code changed in this batch and check every
assertion against the banned patterns from `apply.strict-tdd.md`:

| Pattern Found | Verdict |
|---------------|---------|
| Tautology (`expect(true).toBe(true)`, `assert 1 == 1`) | CRITICAL |
| Type-only assertion with no specific value | CRITICAL |
| Empty-collection assertion with no explanatory setup | CRITICAL |
| Assertion that never exercises the automated interaction (no navigation, no request, no fixture call) | CRITICAL — proves nothing |
| Real assertion — specific value, exercises the interaction, would fail on regression | PASS |

## Step 5c: No Vacuous Pass Confirmation

For each scenario reporting GREEN, confirm you can answer: "if the underlying
behavior this scenario targets regressed, would THIS test catch it?" If the
answer is no or unclear, mark the scenario `UNTESTED` even if the test file
technically exists and passes.

## Report Template Extension

```markdown
### TDD Compliance (Test-Automation)
| Scenario | RED Genuine? | GREEN Executed? | Triangulated? | Refactor-Safe? | Verdict |
|----------|----------------|--------------------|-----------------|------------------|---------|
| U1.1 | ✅ | ✅ | ✅ | ✅ | PASS |

### Assertion Quality
| File | Banned Pattern Found | Verdict |
|------|-------------------------|---------|
| `path/spec.ext` | None | PASS |

### No-Vacuous-Pass Confirmation
- Scenarios confirmed to catch a real regression if reintroduced: {N}/{total}
- Scenarios marked UNTESTED despite a passing test file: {N or "None"}
```

## Rules (Strict TDD Verify specific, test-automation framing)

- Do NOT accept "implemented and green" as sufficient — RED must be proven genuine, not assumed
- A test that would pass regardless of whether the automated interaction ran is NOT evidence of coverage — mark it UNTESTED
- Flag any assertion pattern from the banned list as CRITICAL, no exceptions
- Do not fix issues found here; report them for the orchestrator/user
- Return per the Section D envelope from `../_shared/sdd-phase-common.md`
