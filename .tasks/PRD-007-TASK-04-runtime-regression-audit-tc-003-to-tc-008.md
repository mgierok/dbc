# Overview

This task performs the required runtime regression audit for `TC-003..TC-008` to detect and resolve hidden dependencies on removed panel-switch shortcuts.

## Metadata

- Status: READY
- PRD: PRD-007-simplified-panel-navigation-enter-esc.md
- Task ID: 04
- Task File: PRD-007-TASK-04-runtime-regression-audit-tc-003-to-tc-008.md
- PRD Requirements: FR-007, NFR-001, NFR-004
- PRD Metrics: M4, M2

## Objective

Audit and update runtime testcase files `TC-003` through `TC-008` so they are deterministic under the new panel-navigation model and contain no implicit dependency on removed `Ctrl+w` transitions.

## Working Software Checkpoint

After this task, all six runtime regression scenarios remain executable and aligned with current navigation behavior, with explicit audit completeness evidence.

## Technical Scope

### In Scope

- Review `test-cases/TC-003` through `test-cases/TC-008` for direct and implicit panel-switch assumptions.
- Update only impacted testcase steps/assertions/evidence to align with `Enter`/`Esc` model.
- Keep testcase metadata and template structure valid after edits.
- Update `test-cases/suite-coverage-matrix.md` where assertion or scenario mapping changes require synchronization.
- Record per-file audit outcome evidence for all six audited scenarios.

### Out of Scope

- Additional runtime feature behavior changes.
- Rewriting `TC-002` (handled in TASK-03).
- Final full-suite release-hardening conclusion.

## Implementation Plan

1. Execute structured audit pass over `TC-003` through `TC-008` and mark each file as impacted or non-impacted.
2. For each impacted file, replace old or implicit panel-switch assumptions with explicit `Enter`/`Esc`-compatible interactions.
3. Recheck scenario determinism and assertion clarity after updates.
4. Update `test-cases/suite-coverage-matrix.md` if assertion IDs, references, or scenario mappings changed.
5. Record complete `6/6` audit evidence in this task file completion summary.

## Verification Plan

- FR-007 happy-path check: Audit report includes all six runtime files (`TC-003..TC-008`) with explicit status and deterministic post-audit assertions.
- FR-007 negative-path check: Audit fails if any runtime testcase in the required range lacks an explicit reviewed outcome.
- NFR-001 happy-path check: Updated runtime scenarios preserve deterministic navigation outcomes across repeated executions.
- NFR-001 negative-path check: Determinism check fails if updated steps depend on ambiguous focus state transitions.
- Metric checkpoint (M4):
  - Metric ID: M4
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Audit completeness is `6/6`, and every impacted case is updated and passing with deterministic evidence.
  - Check procedure: Record per-file audit table for `TC-003..TC-008`, list modifications, and confirm pass results for impacted files.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Required navigation checks remain fully represented after audit adjustments.
  - Check procedure: Validate updated testcase assertions and `suite-coverage-matrix` mappings still cover intended navigation behaviors.

## Acceptance Criteria

1. `TC-003..TC-008` are fully audited with explicit evidence for each file.
2. Any impacted runtime testcase is updated to remove hidden dependency on removed `Ctrl+w` panel switching.
3. `test-cases/suite-coverage-matrix.md` remains synchronized with testcase mapping changes.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-007-TASK-03-tc-002-navigation-model-update](.tasks/PRD-007-TASK-03-tc-002-navigation-model-update.md)
- blocks: [PRD-007-TASK-05-integration-hardening](.tasks/PRD-007-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Not started
