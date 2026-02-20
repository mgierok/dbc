# Overview

This task refactors existing write-path scenarios so data operations and staging/save behaviors are represented as separate area-pure Functional Behavior scenarios.

## Metadata

- Status: DONE
- PRD: PRD-006-functional-behavior-grouped-test-case-coverage.md
- Task ID: 03
- Task File: PRD-006-TASK-03-refactor-existing-scenarios-areas-46-47.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-007, FR-008, FR-010, NFR-002, NFR-003, NFR-004
- PRD Metrics: M3

## Objective

Refactor `TC-005` and `TC-006` so `TC-005` is pure area `4.6` (insert/edit/delete) and `TC-006` is pure area `4.7` (staging/undo/redo/save decisions).

## Working Software Checkpoint

After this task, write-path regression coverage remains executable and deterministic with separated ownership for data operations and staging/save lifecycle behavior.

## Technical Scope

### In Scope

- Refactor `TC-005` to cover area `4.6` only with deterministic insert, edit, and delete assertions.
- Refactor `TC-006` to cover area `4.7` only with deterministic staging, undo, redo, and save/decision assertions.
- Update area-to-assertion mapping evidence for both scenarios.
- Update governance and release-audit artifacts for these refactors.
- Record expand-first evidence for behavior additions in this task.

### Out of Scope

- Adding new scenarios for missing areas (`4.5`, `4.8`).
- Final full-suite release decision.
- Runtime feature code changes.

## Implementation Plan

1. Re-scope `TC-005` assertions to area `4.6`, ensuring insert/edit/delete are all covered with deterministic outcomes.
2. Re-scope `TC-006` assertions to area `4.7`, ensuring staged lifecycle and decision paths are covered with deterministic outcomes.
3. Remove or relocate any cross-area assertions that violate one-area purity.
4. Update coverage matrix and release-readiness audit mappings for areas `4.6` and `4.7`.
5. Update expand-first evidence classification for all behavior additions/changes in this task.

## Verification Plan

- FR-001 happy-path check: `TC-005` and `TC-006` each declare exactly one Functional Behavior area (`4.6` and `4.7`).
- FR-001 negative-path check: Ownership audit fails if either scenario has missing or multiple area declarations.
- FR-002 happy-path check: Assertion IDs in `TC-005` and `TC-006` map only to their declared areas.
- FR-002 negative-path check: Purity audit fails if any assertion in these scenarios maps to another area.
- FR-003 happy-path check: All behavior additions in this task are delivered via expansion/refactor of existing `TC-005` and `TC-006`.
- FR-003 negative-path check: Audit fails if a new scenario is added in this task scope.
- FR-007 happy-path check: Area `4.6` assertions include deterministic checks for insert, edit, and delete behaviors.
- FR-007 negative-path check: Audit fails when any of insert/edit/delete coverage is missing in area `4.6`.
- FR-008 happy-path check: Area `4.7` assertions include deterministic staged-change lifecycle and undo/redo/save decision outcomes.
- FR-008 negative-path check: Audit fails when undo/redo or save decision outcome coverage is missing/non-deterministic.
- FR-010 happy-path check: All touched governance artifacts remain synchronized with updated area ownership/mappings.
- FR-010 negative-path check: Release-audit consistency checks fail when matrix and scenario mappings diverge.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Cumulative expand-first adherence remains >=70% after refactoring `TC-005` and `TC-006`.
  - Check procedure: Recompute cumulative ratio in evidence table (`expanded existing` / `all additions`) and verify threshold pass.

## Acceptance Criteria

1. `TC-005` is area-pure for `4.6` and deterministically covers insert/edit/delete behavior.
2. `TC-006` is area-pure for `4.7` and deterministically covers staging/undo/redo/save decision behavior.
3. Governance and traceability artifacts are updated and consistent for areas `4.6` and `4.7`.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-006-TASK-02-refactor-existing-scenarios-areas-41-44](.tasks/PRD-006-TASK-02-refactor-existing-scenarios-areas-41-44.md)
- blocks: [PRD-006-TASK-04-add-missing-area-scenarios-45-48](.tasks/PRD-006-TASK-04-add-missing-area-scenarios-45-48.md), [PRD-006-TASK-05-integration-hardening](.tasks/PRD-006-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

- Refactored existing scenarios without creating new `TC-*` files:
  - `test-cases/TC-005-save-failure-retains-staged-changes-until-corrected.md` is now area-pure `4.6` and includes deterministic insert/edit/delete operation assertions.
  - `test-cases/TC-006-dirty-config-navigation-requires-explicit-decision.md` is now area-pure `4.7` and includes deterministic staged lifecycle, undo/redo, and `:config` decision-path assertions (`cancel`, `save`, `discard`).
- Updated governance/traceability artifacts for this refactor set:
  - `test-cases/suite-coverage-matrix.md`
  - `test-cases/full-suite-release-readiness-audit.md`
- Verification evidence:
  - FR-001 happy-path passed: both scenarios contain exactly one metadata `Functional Behavior Reference` (verified with `rg -n '\| Functional Behavior Reference \|' test-cases/TC-005... test-cases/TC-006...`).
  - FR-002 happy/negative checks passed: no cross-area references found in `TC-005` and `TC-006` outside their declared areas (no-match audits for foreign area anchors).
  - FR-003 negative check passed: `git diff --name-only | rg '^test-cases/TC-'` returned only `TC-005` and `TC-006` (no new `TC-*` files).
  - FR-007 check passed: `TC-005` includes deterministic operation steps/assertions for edit (`S4/A4`), delete toggle/remove (`S5,S6,S9`), and insert (`S7,S8`).
  - FR-008 check passed: `TC-006` includes deterministic staged lifecycle coverage for undo/redo (`S4,S5`) and dirty decision outcomes (`S6,S7,S10`).
  - FR-010 checks passed: matrix rows for `4.6` and `4.7` are synchronized to area-pure ownership (`PASS`) and release audit baseline/evidence was updated to reflect refactor completion.
- Metric checkpoint (M3):
  - Expand-first evidence table now includes `REF-005` and `REF-006`, both classified `Expanded Existing TC`.
  - Cumulative ratio result: `6/6 = 100%` (meets `>=70%` threshold).
- Project quality gates:
  - `gofmt`: no changed Go files.
  - `golangci-lint run ./...`: passed (`0 issues`).
  - `go test ./...`: passed.
