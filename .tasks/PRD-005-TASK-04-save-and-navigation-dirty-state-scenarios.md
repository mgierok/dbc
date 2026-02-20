# Overview

This task defines regression scenarios for save-failure handling and dirty-state navigation safety so write-path risk is covered with explicit decision-path behavior and recovery expectations.

## Metadata

- Status: DONE
- PRD: PRD-005-full-quality-regression-scenarios.md
- Task ID: 04
- Task File: PRD-005-TASK-04-save-and-navigation-dirty-state-scenarios.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-008, NFR-002, NFR-003
- PRD Metrics: none

## Objective

Create save and navigation dirty-state regression scenarios that verify deterministic outcomes for success/failure and explicit user decision paths.

## Working Software Checkpoint

Save and dirty-navigation safety behaviors are covered by executable manual scenarios with clear PASS/FAIL assertions and evidence-ready result fields.

## Technical Scope

### In Scope

- Add or update `TC-*` files for save operation behavior including failure retention and user-visible feedback.
- Add or update `TC-*` files for dirty-state navigation decisions (`save`, `discard`, `cancel`) during context transitions.
- Ensure deterministic final result behavior at assertion and scenario levels.
- Ensure evidence fields are concrete and auditable for save/navigation outcomes.

### Out of Scope

- Startup-only and selector/config-only journeys.
- Generic runtime browsing interactions not tied to save or dirty-state navigation.
- Full-suite metric rollup and release decision finalization.

## Implementation Plan

1. Define scenarios that exercise staged-change save flows, including a failure path with retained staged state.
2. Define scenarios for dirty-state navigation decisions and resulting behavior.
3. Bind each scenario to one approved startup script and exact command.
4. Add deterministic step and assertion mappings with explicit evidence expectations.
5. Run governance checks from TASK-01 and remediate any compliance or determinism findings.

## Verification Plan

- FR-001 happy-path check: Coverage matrix maps `save` and `navigation` journey areas to one or more scenarios.
- FR-001 negative-path check: Coverage matrix flags FAIL when `save` or `navigation` areas are missing scenario coverage.
- FR-002 happy-path check: Each scenario includes exactly one startup script and one startup command.
- FR-002 negative-path check: Any startup binding mismatch fails compliance audit.
- FR-003 happy-path check: All scenario files pass strict heading order and metadata field checks.
- FR-003 negative-path check: Structure audit fails if required sections/fields are missing.
- FR-004 happy-path check: Every step has one action, one expected outcome, and one assertion ID.
- FR-004 negative-path check: Step-level mapping audit fails if any row is incomplete.
- FR-005 happy-path check: Assertion criteria and final scenario result are binary and deterministic.
- FR-005 negative-path check: Determinism audit fails any ambiguous result criteria or forbidden final state.
- FR-006 happy-path check: Save and navigation journeys include explicit failure triggers and user-visible recovery behavior.
- FR-006 negative-path check: Missing documented recovery path for failure behavior fails audit.
- FR-008 happy-path check: Scenario-level result is PASS only when all assertions pass.
- FR-008 negative-path check: Any failed assertion or unmet precondition forces scenario FAIL with reason.
- Metric checkpoints: none in this task (`PRD Metrics: none`).

## Acceptance Criteria

1. Save and dirty-state navigation scenarios exist with deterministic and auditable assertions.
2. Failure handling and user-visible recovery for save/navigation journeys are explicitly covered.
3. Governance checks from TASK-01 pass for all scenario files in this task.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-005-TASK-01-suite-governance-and-coverage-foundation](.tasks/PRD-005-TASK-01-suite-governance-and-coverage-foundation.md)
- blocks: [PRD-005-TASK-05-integration-hardening](.tasks/PRD-005-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Delivered save and dirty-navigation scenario coverage artifacts:

- Added `test-cases/TC-005-save-failure-retains-staged-changes-until-corrected.md` for save failure/recovery behavior, including retained staged state after failed save and successful corrected retry.
- Added `test-cases/TC-006-dirty-config-navigation-requires-explicit-decision.md` for dirty-state `:config` decision behavior across `cancel`, `discard`, and `save` outcomes.
- Updated `test-cases/suite-coverage-matrix.md` to map `save` and `navigation` journey areas to the new scenarios and mark both areas as `PASS` with explicit failure/recovery scenario IDs.

Verification executed against this task verification plan:

- FR-001: coverage matrix now maps both required areas (`save: TC-005`, `navigation: TC-006`) and sets both coverage statuses to `PASS`; negative-path unmapped conditions are no longer present.
- FR-002: each new scenario metadata block contains exactly one startup script and one startup command from the approved startup script catalog.
- FR-003: both scenario files include required heading order (`## 1` through `## 7`) and required metadata fields/table structures from the template contract.
- FR-004: all step rows in `TC-005` and `TC-006` maintain one action, one expected outcome, and one assertion ID mapping.
- FR-005: assertion result fields and final scenario result fields are deterministic and binary (`PASS`/`FAIL`) with no third-state outcomes.
- FR-006: explicit failure/recovery behavior is covered for both target journeys:
  - save: failure on constraint-violating staged edit followed by in-session correction and successful save,
  - navigation: guarded dirty-state decision flow with deterministic `cancel`, `discard`, and `save` outcomes.
- FR-008: scenario-level result contract remains binary, and final `PASS` is consistent with all listed assertions marked `PASS`.
- Metric checkpoints: none (`PRD Metrics: none`).

Downstream decision context:

- Parent PRD remains open because `PRD-005-TASK-05-integration-hardening` is still `READY` and must complete full-suite integration hardening before PRD closure.
