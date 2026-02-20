# Overview

This task closes uncovered Functional Behavior areas by adding only the minimum new scenarios needed after expand-first refactors are complete.

## Metadata

- Status: READY
- PRD: PRD-006-functional-behavior-grouped-test-case-coverage.md
- Task ID: 04
- Task File: PRD-006-TASK-04-add-missing-area-scenarios-45-48.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-006, FR-009, FR-010, NFR-002, NFR-003
- PRD Metrics: M2, M3

## Objective

Add and integrate missing area-pure scenarios for filtering (`4.5`) and visual state communication (`4.8`) to achieve full `4.1` through `4.8` Functional Behavior coverage.

## Working Software Checkpoint

The regression suite remains runnable with deterministic outcomes while adding `TC-007` and `TC-008` to complete area coverage.

## Technical Scope

### In Scope

- Create `TC-007` for area `4.5` filtering flow coverage.
- Create `TC-008` for area `4.8` visual state communication coverage.
- Keep new scenarios strictly area-pure and template-compliant.
- Update coverage matrix and release-readiness audit to full `4.1` to `4.8` mapping.
- Update expand-first evidence table and ratio for PRD-006.

### Out of Scope

- Refactoring `TC-001` to `TC-006`.
- Runtime feature or architecture changes.
- Final integration and release gate conclusion.

## Implementation Plan

1. Confirm remaining uncovered areas are only `4.5` and `4.8` after TASK-03 completion.
2. Create `TC-007` to validate filtering core flow: column selection, operator selection, value input when required, apply effect, and reset on table switch.
3. Create `TC-008` to validate visual state communication: mode indicator, status line context, row markers (`[INS]`, `[DEL]`), and edited-cell marker.
4. Update suite coverage matrix with area -> scenario -> assertion mappings for all `4.1` to `4.8`.
5. Update release-readiness audit with new scenario evidence and recomputed expand-first ratio.

## Verification Plan

- FR-001 happy-path check: `TC-007` and `TC-008` each declare exactly one Functional Behavior area.
- FR-001 negative-path check: Ownership audit fails when either new scenario has missing or multiple area declarations.
- FR-002 happy-path check: Assertions in each new scenario are mapped only to its declared area.
- FR-002 negative-path check: Purity audit fails if any assertion in `TC-007` or `TC-008` maps outside declared area.
- FR-003 happy-path check: New scenarios are added only after refactor evidence for `TC-001` to `TC-006` is complete.
- FR-003 negative-path check: Audit fails if expand-first evidence is missing for prerequisite refactor scope.
- FR-004 happy-path check: Coverage matrix shows `4.1` to `4.8` all mapped with scenario and assertion IDs.
- FR-004 negative-path check: Matrix fails if any Functional Behavior area row is missing scenario or assertion evidence.
- FR-006 happy-path check: `TC-007` includes deterministic assertions for full filter core flow and reset behavior.
- FR-006 negative-path check: Audit fails when any required filter flow element is unverified.
- FR-009 happy-path check: `TC-008` includes deterministic assertions for required mode/status/marker visibility behaviors.
- FR-009 negative-path check: Audit fails when any required visual communication indicator is missing from assertions.
- FR-010 happy-path check: Governance artifacts are synchronized with new scenarios and mappings.
- FR-010 negative-path check: Cross-artifact consistency checks fail on any mapping drift.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Functional Behavior coverage completeness reaches `8/8` areas with PASS status.
  - Check procedure: Validate matrix rows for `4.1` to `4.8` each include scenario IDs, assertion IDs, and PASS status.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Expand-first adherence ratio remains >=70% after adding new scenarios.
  - Check procedure: Compute ratio from evidence table (`expanded existing` vs `new scenario`) and verify threshold pass.

## Acceptance Criteria

1. New scenarios `TC-007` (`4.5`) and `TC-008` (`4.8`) are created, area-pure, deterministic, and template-compliant.
2. Functional Behavior area coverage is complete for `4.1` through `4.8` with assertion-level traceability.
3. Governance artifacts and expand-first evidence are updated and internally consistent.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-006-TASK-03-refactor-existing-scenarios-areas-46-47](.tasks/PRD-006-TASK-03-refactor-existing-scenarios-areas-46-47.md)
- blocks: [PRD-006-TASK-05-integration-hardening](.tasks/PRD-006-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Not started
