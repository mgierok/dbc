# Overview

We believe that defining a cost-optimized regression scenario suite with assertion-level precision for terminal users and maintainers will improve release confidence and reduce regression escape risk.
We will know this is true when all critical flow areas are covered by approved scenarios with full step-to-assertion traceability during the PRD-005 execution window.

# Metadata

- Status: READY

# Problem Statement

Current regression verification for critical TUI journeys is not standardized in a single, auditable scenario suite. This increases the risk that behavior regressions in startup, configuration, navigation, staging, and save flows are discovered late or missed.

# Current State (As-Is)

- There is no consolidated regression scenario pack in `docs/test-cases/` for critical-path execution.
- Verification is vulnerable to inconsistency because assertion-level reporting and step-to-assertion traceability are not enforced by a single suite contract.
- Scenario overlap handling is not explicitly optimized for execution cost.
- A mandatory first-task process gate for updating TDD guidance is not defined for this initiative.

# Target State After Release (To-Be)

- A merged-journey regression suite exists and covers critical startup, config, runtime, save, and navigation behaviors.
- Every scenario reports outcomes at assertion level using unique IDs and explicit step mapping.
- Scenario outcomes remain deterministic (`PASS`/`FAIL`) even when multiple checks are combined in one user journey.
- Redundant checks are present only when unavoidable in the journey and are explicitly justified.
- Existing test cases are reviewed and mapped into proposed PRD-005 scenarios before approval.
- Legacy scenarios covered by PRD-005 are explicitly replaced by merged equivalents or removed from active use.
- The first downstream task updates `docs/test-driven-development.md` with the merged-scenario and assertion-precision standard, and completion of later tasks depends on this gate.

# Business Rationale and Strategic Fit

This initiative reduces release risk and manual verification cost at the same time. It aligns with product reliability goals by making regression checks repeatable, auditable, and decision-oriented, while avoiding unnecessary execution overhead from duplicated journey runs.

# Goals

- G1: Establish one regression suite contract for critical TUI journeys.
- G2: Reduce execution cost through journey merging without losing diagnostic precision.
- G3: Enforce assertion-level evidence for every failed check.
- G4: Make the TDD guidance update a blocking first step for all downstream work from this PRD.

# Non-Goals

- NG1: Introducing new runtime features or changing current product behavior.
- NG2: Defining implementation details for automation tooling.
- NG3: Covering non-SQLite engines or schema-management workflows.
- NG4: Replacing product or technical documentation with process content outside this PRD scope.

# Scope (In Scope / Out of Scope)

In Scope:
- Define a merged-journey regression scenario set for critical startup, selector/config, runtime browsing, filtering, staging, save, and dirty navigation behavior.
- Define assertion-level reporting rules for merged scenarios (IDs, mapping, deterministic outcomes, failure evidence).
- Define anti-redundancy policy and explicit justification requirement for unavoidable repeated checks.
- Require review of existing test cases and incorporation of relevant coverage into proposed PRD-005 scenarios.
- Require disposition of legacy scenarios in PRD-005 scope (`replace` or `remove`) to avoid parallel active catalogs.
- Define downstream gating rule: first generated task must update `docs/test-driven-development.md`.
- Require scenario authoring to follow `docs/test-case-specification.md` and `docs/test-case-template.md`.

Out of Scope:
- Implementing automated test runners or CI orchestration changes.
- Modifying application architecture, persistence model, or runtime interaction model.
- Creating non-regression exploratory test catalogs.

# Functional Requirements

FR-001:
- Requirement: The regression suite must cover these flow areas: startup, config management/recovery, runtime browsing, filtering, staging lifecycle, save flow, and dirty-state navigation.
- Acceptance: A scenario-to-flow mapping table exists and each listed flow area is covered by at least one approved scenario.

FR-002:
- Requirement: Scenarios that share the same primary user journey must be merged when merging materially reduces execution cost.
- Acceptance: The scenario set contains merged journeys where overlap is high, and no duplicate journey runs remain without justification.

FR-003:
- Requirement: Each scenario must bind exactly one startup script and command from the startup script catalog defined in `docs/test-case-specification.md`.
- Acceptance: Every scenario metadata block contains one startup script path and one exact startup command.

FR-004:
- Requirement: Every scenario must define unique assertion IDs (`A-001`, `A-002`, ...) and map each step to one assertion ID.
- Acceptance: Scenario content includes a complete step-to-assertion mapping with no unmapped steps.

FR-005:
- Requirement: Assertion outcomes must be deterministic and binary (`PASS` or `FAIL` only).
- Acceptance: Scenario execution result tables use only `PASS`/`FAIL` values for every assertion.

FR-006:
- Requirement: Scenario outcome must be derived from assertion outcomes (`FAIL` if any assertion fails, otherwise `PASS`).
- Acceptance: Scenario result rule is documented and applied consistently in execution reports.

FR-007:
- Requirement: Every failed assertion must include a concrete reason and observable evidence.
- Acceptance: Each `FAIL` assertion includes both `reason` and `evidence` fields in execution output.

FR-008:
- Requirement: Redundant verification is allowed only for unavoidable flow dependencies and must be explicitly justified.
- Acceptance: Any repeated check includes a written "unavoidable flow dependency" justification note.

FR-009:
- Requirement: The first downstream task generated from this PRD must update `docs/test-driven-development.md` with merged-scenario and assertion-level precision rules.
- Acceptance: Task sequencing shows this documentation task as first and marked complete before any later task is marked done.

FR-010:
- Requirement: Downstream tasks that author or review scenarios must enforce the rules introduced by FR-004 to FR-008.
- Acceptance: Task acceptance criteria for scenario work explicitly include assertion IDs, mapping, deterministic result rule, and failure evidence checks.

FR-011:
- Requirement: Before proposing or approving PRD-005 scenarios, teams must review existing test cases and reflect applicable checks in the proposed merged scenarios.
- Acceptance: PRD-005 evidence includes an old-to-new mapping table showing each relevant existing test case and where it is covered in the proposed scenario set.

FR-012:
- Requirement: Existing scenarios within PRD-005 scope must not remain active unchanged; each must be either replaced by an approved PRD-005 scenario or removed.
- Acceptance: PRD-005 closure evidence includes a disposition list (`replaced` or `removed`) for all in-scope legacy scenarios.

# Non-Functional Product Requirements

NFR-001:
- Reproducibility: Scenario definitions and execution outputs must be reproducible by different operators using the same fixture and startup script inputs.

NFR-002:
- Auditability: A reviewer must be able to trace any scenario `FAIL` to one or more failed assertion IDs with evidence.

NFR-003:
- Execution efficiency: Scenario design must minimize repeated end-to-end runs by merging high-overlap journeys.

NFR-004:
- Consistency: Scenario structure must remain compliant with the test-case contract and deterministic pass/fail policy.

# Success Metrics and Release Criteria

Primary Outcome Metric:
- M1: Critical flow coverage by approved regression scenarios.
  - Baseline: `0/7` flow areas covered by an approved PRD-005 regression suite (`0%`).
  - Target: `7/7` flow areas covered (`100%`) by approved scenarios.
  - Measurement window: From start of PRD-005 execution until PRD-005 closure decision.
  - Measurement method: Coverage audit table produced in PRD-005 task completion evidence, cross-checking flow areas against finalized scenario inventory.

Leading Indicators:
- M2: Scenario files with complete step-to-assertion mapping.
  - Baseline: `0%` (no approved PRD-005 scenario files currently exist).
  - Target: `100%` of PRD-005 scenarios include full mapping with unique assertion IDs.
  - Measurement window: During scenario-authoring tasks and final regression suite audit.
  - Measurement method: Checklist review artifact attached to PRD-005 scenario tasks, validating mapping presence per file.
- M3: Scenario files with single startup-script binding compliance.
  - Baseline: `0%` (no approved PRD-005 scenario files currently exist).
  - Target: `100%` of PRD-005 scenarios bind exactly one startup script and command.
  - Measurement window: During scenario-authoring tasks and final audit.
  - Measurement method: Metadata compliance checklist generated from `docs/test-cases/TC-*.md` files in PRD-005 scope.

Guardrail Metric:
- M4: Redundancy justification compliance for repeated checks.
  - Baseline: `0%` (no approved PRD-005 scenario files currently exist).
  - Target: `100%` of repeated checks include explicit unavoidable-flow justification.
  - Measurement window: During scenario review and closure audit.
  - Measurement method: Redundancy audit table in PRD-005 review evidence, mapping repeated checks to justification notes.

Release Criteria:
- RC1: All PRD-005 functional requirements are satisfied with evidence.
- RC2: Primary metric M1 and all supporting metrics (M2, M3, M4) meet targets.
- RC3: First downstream task updates `docs/test-driven-development.md` and is completed before any later PRD-005 task is closed.
- RC4: Final scenario suite review confirms deterministic `PASS`/`FAIL` reporting and assertion-level failure evidence.

# Risks and Dependencies

Risks:
- Over-merging could reduce readability if assertion mapping is not strictly enforced.
- Scenario quality may drift if redundancy justification is not reviewed consistently.
- Incomplete evidence capture could reduce decision confidence during release validation.

Dependencies:
- Availability and stability of fixture and startup artifacts defined in `docs/test-case-specification.md`.
- Consistent use of `docs/test-case-template.md` for downstream scenario files.
- Stakeholder review of PRD-005 task gating and acceptance evidence before closure.

# State & Failure Matrix

| Flow Area | Trigger / Failure Mode | Expected Product Response | User-Visible Recovery Path |
| --- | --- | --- | --- |
| startup | Startup mode mismatch, invalid usage contract, or startup path not covered by scenario inventory | Regression suite records assertion-level `FAIL` with explicit reason and evidence; scenario result follows deterministic rule | Operator reruns the defined startup path using bound script/command and updates evidence after corrective action |
| config | Selector/config management behavior diverges from expected contract during merged journey | Affected config assertions fail while unrelated assertions remain independently reported | Operator applies corrective config step (edit/select valid entry) and re-executes only the impacted scenario journey |
| save | Save flow branch deviates (including dirty decision branch or failed save handling) | Save-related assertions fail with retained diagnostic context and evidence requirements | Operator follows defined recovery branch (`save`, `discard`, or `cancel`) and revalidates save-specific assertions |
| navigation | Context switch behavior breaks journey continuity or expected gating order | Navigation assertions fail without collapsing assertion-level reporting for other checks | Operator returns to the last valid journey state and reruns navigation steps with assertion mapping preserved |

# Assumptions

- A1 (High): Critical flow coverage for PRD-005 is represented by seven flow areas listed in FR-001.
- A2 (High): Downstream scenario files for this PRD will be created under `docs/test-cases/` using the mandated template contract.
- A3 (Medium): Stakeholders will accept assertion-level evidence artifacts as the release decision input for this regression suite.
- A4 (Medium): Merged journeys remain operationally manageable when assertion mapping is enforced as specified.
