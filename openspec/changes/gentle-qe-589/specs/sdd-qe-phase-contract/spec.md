# SDD QE Phase Contract

## ADDED Requirements

### Requirement: Explore-Phase Content Frames Quality Risk And Testability

The explore-phase QE asset MUST frame exploration as a quality-risk analysis
of the target area: a likelihood x impact risk assessment, a testability
assessment, a defect-clustering analysis (the 20% of the codebase
historically responsible for 80% of defects), and an inventory of existing
oracles versus missing oracles. It MUST NOT frame exploration as a
feasibility analysis for a new feature.

#### Scenario: Explore asset requires risk, testability, clustering, and oracle inventory

- GIVEN the explore-phase QE asset is inspected
- WHEN a reviewer checks its required sections
- THEN it requires a likelihood x impact risk assessment
- AND it requires a testability assessment of the explored area
- AND it requires a defect-clustering (80/20) analysis
- AND it requires an inventory of existing oracles vs. missing oracles
- AND it does NOT require a feature-feasibility section

### Requirement: Propose-Phase Content Frames Test-Requirements, Not Capabilities

The propose-phase QE asset MUST reframe the unit of work as a
test-requirement plus a candidate oracle and risk statement, instead of a
feature capability description.

#### Scenario: Propose asset names a candidate oracle

- GIVEN the propose-phase QE asset is inspected
- WHEN a reviewer checks its required sections
- THEN it requires a stated risk and a candidate oracle for the proposed test
  work
- AND it does NOT require a "capability" section framed as shippable feature
  scope

### Requirement: Spec-Phase Requirements MUST State A Verifiable Oracle

Every requirement produced by the spec-phase QE asset MUST state how a
resulting test detects a real defect (the oracle), in addition to
GIVEN/WHEN/THEN scenarios. A requirement without an oracle statement is
incomplete.

#### Scenario: Requirement without an oracle fails review

- GIVEN a requirement in a spec-phase QE artifact
- WHEN the requirement is checked for an oracle statement
- THEN it is rejected as incomplete if it has scenarios but no statement of
  how a test would prove a defect exists
- AND it passes only when the oracle is explicit and testable

### Requirement: Design-Phase Content Is An ISTQB Test Strategy

The design-phase QE asset MUST make the Testing Strategy the entire document:
it MUST name test levels (unit, integration, system, acceptance), name at
least one ISTQB technique (equivalence partitioning, boundary value analysis,
decision table, or state transition) per non-trivial requirement, reference
the test pyramid, and justify test prioritization using risk and defect
clustering.

#### Scenario: Design asset omits ISTQB technique naming

- GIVEN the design-phase QE asset is inspected
- WHEN a reviewer checks for named ISTQB techniques
- THEN it fails if no technique from the ISTQB set is explicitly named
- AND it fails if the test pyramid or risk-based prioritization is absent

#### Scenario: Design asset covers required test levels

- GIVEN the design-phase QE asset is inspected
- WHEN a reviewer checks test-level coverage
- THEN unit, integration, system, and acceptance levels are each addressed or
  explicitly justified as not-applicable

### Requirement: Tasks-Phase Content Enumerates Automatable Scenarios

The tasks-phase QE asset MUST enumerate GIVEN/WHEN/THEN scenarios or cases to
automate, grouped by test level, instead of implementation subtasks.

#### Scenario: Tasks asset lists scenarios per test level

- GIVEN the tasks-phase QE asset is inspected
- WHEN a reviewer checks task entries
- THEN each entry maps to a GIVEN/WHEN/THEN scenario and a test level
- AND no entry describes production feature implementation steps

### Requirement: Apply/Verify-Phase Content Targets Test Code And Flakiness

The apply-phase QE asset MUST instruct writing test code (specs, fixtures,
page objects), not production feature code. The verify-phase QE asset MUST
frame verification as execution, coverage, and flaky-test detection, treating
a flaky test as a broken test requiring fix or removal.

#### Scenario: Apply asset instructs test-code output

- GIVEN the apply-phase QE asset is inspected
- WHEN a reviewer checks its primary deliverable
- THEN it is test code (spec/fixture/POM files)
- AND it is NOT a production capability implementation

#### Scenario: Verify asset flags flaky tests as defects

- GIVEN the verify-phase QE asset is inspected
- WHEN a reviewer checks its pass/fail criteria
- THEN it requires reporting coverage and identifying flaky tests
- AND it states that a flaky test MUST be fixed or removed, never ignored

### Requirement: Strict-TDD Modules Reframe To Test-Automation TDD

When StrictTDD is active in the apply/verify QE assets, the `strict-tdd.md` /
`strict-tdd-verify.md` modules MUST reframe the RED-GREEN-REFACTOR cycle as
writing TEST-AUTOMATION code, never production code. RED MUST be an
assertion/oracle that fails first (e.g. a failing test expectation with no
implementation behind it yet); GREEN MUST be the minimal page
object/fixture/helper that makes that assertion pass; REFACTOR MUST clean up
test code (fixtures, POMs, helpers) without breaking the passing oracle. The
module MUST NOT instruct writing or completing production/feature code as
part of this cycle.

#### Scenario: Strict-TDD QE module never instructs production-code work

- GIVEN the StrictTDD apply/verify QE modules (`strict-tdd.md`,
  `strict-tdd-verify.md`) are inspected
- WHEN a reviewer checks what RED, GREEN, and REFACTOR each instruct writing
- THEN RED is defined as writing a failing assertion/oracle
- AND GREEN is defined as writing the minimal test-support code (page
  object, fixture, or helper) that makes the assertion pass
- AND REFACTOR is defined as cleaning up test-support code without breaking
  the oracle
- AND no step instructs writing, completing, or modifying production/feature
  application code

### Requirement: Archive-Phase Content Remains Mechanical

The archive-phase asset MAY remain unchanged from upstream dev behavior,
since archiving (moving files, merging deltas) is mechanical and has no
dev/QE distinction.

#### Scenario: Archive asset has no QE reframing requirement

- GIVEN the archive-phase asset is inspected
- WHEN a reviewer checks for QE-specific content
- THEN no QE reframing is required for this phase to pass review
