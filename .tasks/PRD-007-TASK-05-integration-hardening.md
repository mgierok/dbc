# Overview

This task closes PRD-007 with full cross-task integration hardening, regression verification, and release-evidence consolidation for requirements and metrics.

## Metadata

- Status: READY
- PRD: PRD-007-simplified-panel-navigation-enter-esc.md
- Task ID: 05
- Task File: PRD-007-TASK-05-integration-hardening.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, NFR-001, NFR-002, NFR-003, NFR-004
- PRD Metrics: M1, M2, M3, M4, M5

## Objective

Validate end-to-end consistency of runtime behavior, key guidance, testcase artifacts, and metric evidence so PRD-007 can be closed with auditable confidence.

## Working Software Checkpoint

After this task, all PRD-007 deliverables are integrated, validated, and regression-checked with no unresolved cross-task conflicts.

## Technical Scope

### In Scope

- Execute cross-task verification across TASK-01 through TASK-04 outputs.
- Validate final requirement traceability and metric checkpoint evidence for PRD-007.
- Run final repository quality gates required by project policy.
- Capture final release-readiness evidence in completion summary.

### Out of Scope

- New runtime behavior or UI scope beyond PRD-007.
- New testcase scenario creation unrelated to PRD-007 requirements.
- Post-release monitoring implementation.

## Implementation Plan

1. Aggregate artifacts produced by TASK-01 to TASK-04 and verify dependency closure.
2. Execute final validation for behavior, guidance consistency, testcase updates, and regression-audit completeness.
3. Run required quality commands for final verification (`gofmt`, `golangci-lint run ./...`, `go test ./...`).
4. Consolidate FR/NFR and metric evidence into this task file and resolve any detected integration drift before closure.

## Verification Plan

- FR-001 happy-path check: Integrated runtime behavior confirms left-panel `Enter` transitions to right-panel Records view.
- FR-001 negative-path check: Integration review fails if any path prevents deterministic left-to-right transition from table-selection context.
- FR-002 happy-path check: Integrated runtime behavior confirms neutral right-panel `Esc` returns focus to left panel with table context retained.
- FR-002 negative-path check: Integration review fails if neutral right-panel `Esc` behavior regresses or becomes inconsistent.
- FR-003 happy-path check: Integrated runtime behavior confirms nested-context `Esc` exits local context before any panel transition.
- FR-003 negative-path check: Integration review fails if nested `Esc` handling jumps directly to panel transition.
- FR-004 happy-path check: Integrated behavior and test evidence confirm `Ctrl+w h/l/w` panel switching remains unsupported.
- FR-004 negative-path check: Integration review fails if any runtime path or testcase still treats removed `Ctrl+w` shortcuts as valid transitions.
- FR-005 happy-path check: Runtime hints, README, and documentation remain synchronized with implemented key behavior.
- FR-005 negative-path check: Integration review fails if guidance conflicts appear across runtime/docs surfaces.
- FR-006 happy-path check: Updated `TC-002` provides full required coverage and passes deterministically.
- FR-006 negative-path check: Integration review fails if `TC-002` misses required assertions or evidence quality.
- FR-007 happy-path check: Runtime audit evidence for `TC-003..TC-008` is complete with impacted-case updates and deterministic outcomes.
- FR-007 negative-path check: Integration review fails if audit completeness is below `6/6` or any impacted case remains unresolved.
- Metric checkpoint (M1):
  - Metric ID: M1
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Runtime panel transitions are limited to `Enter` and `Esc` model and removed shortcuts are unsupported.
  - Check procedure: Reconcile runtime behavior, tests, and guidance artifacts against PRD target definition.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: `5/5` critical navigation checks pass in release evidence.
  - Check procedure: Confirm assertion pass evidence across updated `TC-002` and linked runtime validation artifacts.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Canonical transition flow is represented with two keys (`Enter`, `Esc`) only.
  - Check procedure: Validate final guidance and scripted flow artifacts for key-count target conformance.
- Metric checkpoint (M4):
  - Metric ID: M4
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Runtime regression audit completeness is `6/6` with all impacted cases passing.
  - Check procedure: Verify audit evidence table from TASK-04 and final testcase states.
- Metric checkpoint (M5):
  - Metric ID: M5
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Context-first nested `Esc` retention remains at 100% in validated right-panel nested contexts.
  - Check procedure: Confirm nested-context regression evidence from implementation and `TC-002` validation artifacts.

## Acceptance Criteria

1. All PRD-007 requirements (`FR-001..FR-007`, `NFR-001..NFR-004`) have verified evidence and no orphan traceability gaps.
2. All PRD-007 metrics (`M1..M5`) meet target thresholds with evidence recorded in completion summary.
3. Cross-task integration produces no unresolved behavioral or documentation conflicts.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-007-TASK-01-runtime-navigation-contract-enter-esc](.tasks/PRD-007-TASK-01-runtime-navigation-contract-enter-esc.md), [PRD-007-TASK-02-runtime-key-guidance-and-doc-sync](.tasks/PRD-007-TASK-02-runtime-key-guidance-and-doc-sync.md), [PRD-007-TASK-03-tc-002-navigation-model-update](.tasks/PRD-007-TASK-03-tc-002-navigation-model-update.md), [PRD-007-TASK-04-runtime-regression-audit-tc-003-to-tc-008](.tasks/PRD-007-TASK-04-runtime-regression-audit-tc-003-to-tc-008.md)
- blocks: none

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Not started
