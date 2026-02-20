# Overview

This task refactors the first four existing scenarios into area-pure Functional Behavior coverage for startup/access, layout/focus, table/schema, and records/navigation.

## Metadata

- Status: READY
- PRD: PRD-006-functional-behavior-grouped-test-case-coverage.md
- Task ID: 02
- Task File: PRD-006-TASK-02-refactor-existing-scenarios-areas-41-44.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-010, NFR-001, NFR-002, NFR-004
- PRD Metrics: M3

## Objective

Refactor `TC-001` through `TC-004` so each scenario declares exactly one Functional Behavior area and all assertions remain area-pure for `4.1`, `4.2`, `4.3`, and `4.4`.

## Working Software Checkpoint

After this task, the suite still runs with four refactored scenarios that remain deterministic and executable while preserving existing scenario IDs.

## Technical Scope

### In Scope

- Refactor `TC-001` to area `4.1` (including informational startup behavior coverage using approved script binding).
- Refactor `TC-002` to area `4.2`.
- Refactor `TC-003` to area `4.3`.
- Refactor `TC-004` to area `4.4`.
- Update governance artifacts and traceability fields required by these scenario refactors.
- Record expand-first evidence for all behavior additions delivered in this task.

### Out of Scope

- Refactoring `TC-005` and `TC-006`.
- Creating new `TC-*` files.
- Final `4.1` to `4.8` closure and release decision.

## Implementation Plan

1. Assign final area ownership declarations and assertion-ID mapping plans for `TC-001` to `TC-004`.
2. Refactor scenario content so each assertion maps only to its declared area and remove cross-area assertions.
3. Ensure startup informational behavior (`help`/`version`) is covered under area `4.1` with one approved startup script binding.
4. Update coverage matrix and release-readiness audit artifacts for the four refactored areas and assertion mapping evidence.
5. Update expand-first evidence table with explicit classification for each changed behavior item.

## Verification Plan

- FR-001 happy-path check: `TC-001` to `TC-004` each declare exactly one Functional Behavior area (`4.1`, `4.2`, `4.3`, `4.4` respectively).
- FR-001 negative-path check: Ownership audit fails if any of `TC-001` to `TC-004` has missing or multiple area declarations.
- FR-002 happy-path check: Assertion rows in each scenario map only to the scenario's declared area.
- FR-002 negative-path check: Purity audit fails if any assertion ID for `TC-001` to `TC-004` maps outside declared area.
- FR-003 happy-path check: Coverage additions in this task are delivered by refactoring existing `TC-001` to `TC-004` only.
- FR-003 negative-path check: Audit fails if a new `TC-*` file is introduced in this task scope.
- FR-004 happy-path check: Coverage matrix shows complete mapping for areas `4.1` to `4.4` with scenario and assertion IDs.
- FR-004 negative-path check: Coverage matrix marks FAIL if any of areas `4.1` to `4.4` lacks scenario or assertion mapping.
- FR-005 happy-path check: Informational startup assertions are present under area `4.1` and script binding is from approved catalog.
- FR-005 negative-path check: Startup informational coverage fails when not bound to approved startup script.
- FR-010 happy-path check: All touched governance artifacts remain synchronized with refactored scenario ownership/mapping.
- FR-010 negative-path check: Release-audit consistency checks fail when governance artifacts diverge from scenario content.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: 100% of behavior additions in this task are delivered through expansion/refactor of existing scenarios.
  - Check procedure: Count behavior additions in task evidence table and verify each entry is classified as `Expanded existing TC`.

## Acceptance Criteria

1. `TC-001` to `TC-004` are area-pure and each declares exactly one Functional Behavior area.
2. Informational startup behavior coverage is represented under area `4.1` with approved script binding.
3. Governance artifacts and traceability rows for areas `4.1` to `4.4` are updated and auditable.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-006-TASK-01-functional-area-governance-foundation](.tasks/PRD-006-TASK-01-functional-area-governance-foundation.md)
- blocks: [PRD-006-TASK-03-refactor-existing-scenarios-areas-46-47](.tasks/PRD-006-TASK-03-refactor-existing-scenarios-areas-46-47.md), [PRD-006-TASK-05-integration-hardening](.tasks/PRD-006-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Not started
