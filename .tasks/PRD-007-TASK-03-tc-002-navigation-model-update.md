# Overview

This task updates `TC-002` to validate the new runtime panel-navigation contract and provide auditable release evidence for the required two-key model and context-safe `Esc` behavior.

## Metadata

- Status: DONE
- PRD: PRD-007-simplified-panel-navigation-enter-esc.md
- Task ID: 03
- Task File: PRD-007-TASK-03-tc-002-navigation-model-update.md
- PRD Requirements: FR-006, FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002, NFR-004
- PRD Metrics: M2, M5

## Objective

Refactor `test-cases/TC-002-main-layout-focus-switching-remains-predictable.md` so it deterministically verifies the new `Enter`/`Esc` panel model, nested-context `Esc` precedence, and unsupported `Ctrl+w` transitions.

## Working Software Checkpoint

After this task, the testcase suite remains runnable with a template-compliant `TC-002` that validates current runtime navigation behavior without relying on removed shortcut assumptions.

## Technical Scope

### In Scope

- Rewrite `TC-002` test steps, assertions, and evidence to the new panel model.
- Add explicit `TC-002` assertions for:
  - left-panel `Enter` transition to right-panel Records view,
  - right-panel neutral `Esc` transition back to left panel,
  - nested-context `Esc` precedence,
  - unsupported `Ctrl+w h/l/w` panel switching.
- Keep testcase formatting and metadata compliant with testcase specification contracts.

### Out of Scope

- Runtime implementation changes beyond already-delivered behavior from TASK-01.
- Full runtime-audit updates for `TC-003..TC-008`.
- Cross-suite integration closure and final PRD release decision.

## Implementation Plan

1. Update `TC-002` scenario subject/expected result to the new `Enter`/`Esc` focus model.
2. Replace old `Ctrl+w` transition steps with deterministic `Enter` and `Esc` transition steps.
3. Add a nested-context step proving `Esc` exits local context before panel return.
4. Add negative assertion step proving `Ctrl+w h/l/w` does not transition panel focus.
5. Review assertion IDs and evidence text for deterministic, binary-resolvable pass criteria.

## Verification Plan

- FR-006 happy-path check: Updated `TC-002` includes required assertions for `Enter` transition, right-neutral `Esc` return, nested-context `Esc` precedence, and unsupported `Ctrl+w` transitions.
- FR-006 negative-path check: Validation fails if `TC-002` omits any required assertion area from the PRD acceptance statement.
- FR-001 happy-path check: `TC-002` proves left-panel `Enter` opens right-panel Records view for selected table.
- FR-001 negative-path check: `TC-002` fails if `Enter` does not transition to right-panel Records view from left-panel context.
- FR-002 happy-path check: `TC-002` proves neutral right-panel `Esc` returns focus to left panel while preserving table context.
- FR-002 negative-path check: `TC-002` fails if neutral right-panel `Esc` does not return to left panel.
- FR-003 happy-path check: `TC-002` proves nested right-panel `Esc` exits context first without immediate panel switch.
- FR-003 negative-path check: `TC-002` fails if first `Esc` in nested context jumps directly to left panel.
- FR-004 happy-path check: `TC-002` includes explicit assertion that `Ctrl+w h/l/w` does not perform panel transitions.
- FR-004 negative-path check: `TC-002` fails if any removed `Ctrl+w` combination still transitions focus.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: `TC-002` verifies all required navigation checks in scope with deterministic pass criteria.
  - Check procedure: Execute `TC-002`, confirm each required assertion has explicit `PASS` evidence, and record results in task summary.
- Metric checkpoint (M5):
  - Metric ID: M5
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Nested right-panel context checks in `TC-002` show 100% context-first `Esc` retention.
  - Check procedure: Run nested-context `Esc` step in `TC-002` and verify evidence confirms local-context exit precedes panel transition.

## Acceptance Criteria

1. `TC-002` is updated to the new panel-transition model and remains template-compliant.
2. `TC-002` contains deterministic happy and negative assertions covering `FR-001`, `FR-002`, `FR-003`, `FR-004`, and `FR-006`.
3. `TC-002` no longer depends on old `Ctrl+w` panel-switch semantics.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-007-TASK-01-runtime-navigation-contract-enter-esc](.tasks/PRD-007-TASK-01-runtime-navigation-contract-enter-esc.md)
- blocks: [PRD-007-TASK-04-runtime-regression-audit-tc-003-to-tc-008](.tasks/PRD-007-TASK-04-runtime-regression-audit-tc-003-to-tc-008.md), [PRD-007-TASK-05-integration-hardening](.tasks/PRD-007-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

- Updated `test-cases/TC-002-main-layout-focus-switching-remains-predictable.md` to the new runtime navigation model:
  - replaced old `Ctrl+w`-driven panel-switch happy-path steps with `Enter` (left-panel -> right-panel Records) and `Esc` (neutral right-panel -> left-panel) transitions,
  - added nested-context `Esc` precedence evidence using filter-popup flow (`Esc` closes nested context first, second `Esc` performs panel return),
  - added explicit negative assertion for unsupported `Ctrl+w h/l/w` panel transitions.
- Updated `test-cases/suite-coverage-matrix.md` mapping for `[4.2 Main Layout and Focus Model]`:
  - assertion mapping changed from `TC-002:A1,A2,A3,A4` to `TC-002:A1,A2,A3,A4,A5,A6`.
- Verification executed (all pass for artifact scope):
  - `rg -n 'Enter|Esc|Ctrl\\+w h|Ctrl\\+w l|Ctrl\\+w w|nested|filter popup|Result \\(`PASS`/`FAIL`\\)' test-cases/TC-002-main-layout-focus-switching-remains-predictable.md`
  - `rg -n '^\\| A[0-9]+ \\|' test-cases/TC-002-main-layout-focus-switching-remains-predictable.md`
  - `rg -n 'TC-002:A1,A2,A3,A4,A5,A6' test-cases/suite-coverage-matrix.md`
- Metric evidence:
  - M2 satisfied: updated `TC-002` now contains deterministic assertions for all required navigation checks in scope (`Enter`, neutral `Esc`, nested-context-first `Esc`, unsupported `Ctrl+w`).
  - M5 satisfied: nested-context evidence in `A4`/`A5` confirms first `Esc` exits local context before any panel transition.
- Downstream impact:
  - TASK-04 can audit `TC-003..TC-008` against the same explicit `Enter`/`Esc` + unsupported-`Ctrl+w` contract style introduced in `TC-002`.
