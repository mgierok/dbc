# Overview

This final task integrates outputs from all PRD-005 behavior tasks, validates cross-task consistency, and produces the complete metric and release-readiness evidence package.

## Metadata

- Status: READY
- PRD: PRD-005-full-quality-regression-scenarios.md
- Task ID: 05
- Task File: PRD-005-TASK-05-integration-hardening.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, NFR-001, NFR-002, NFR-003, NFR-004
- PRD Metrics: M1, M2, M3, M4

## Objective

Perform full-suite integration hardening that verifies requirement coverage, dependency consistency, deterministic behavior, and metric target attainment before PRD closure.

## Working Software Checkpoint

The regression suite is complete, auditable, and release-decision-ready with all required evidence artifacts and deterministic suite-level PASS/FAIL outcomes.

## Technical Scope

### In Scope

- Validate full traceability across FR/NFR/M mappings to completed PRD-005 tasks and `TC-*` scenarios.
- Execute full-suite coverage, compliance, and determinism audits using TASK-01 governance artifacts.
- Produce release-readiness evidence with metric results and go/no-go conclusion.
- Verify cross-task interaction consistency and remove coverage gaps or conflicting assertions.

### Out of Scope

- Creating new product features or runtime code changes.
- Defining automation frameworks or CI implementation.
- Expanding PRD scope beyond current DBC behavior.

## Implementation Plan

1. Collect scenario artifacts produced by TASK-02, TASK-03, and TASK-04 and confirm dependency completion.
2. Execute full coverage matrix audit and verify all required journey areas and failure/recovery expectations are covered.
3. Execute structure/metadata compliance audit and determinism integrity audit across full suite.
4. Produce consolidated release-readiness decision summary with metric outcomes and pass/fail status.
5. Capture all metric evidence and thresholds in this task file `Completion Summary` for PRD closure traceability.

## Verification Plan

- FR-001 happy-path check: Coverage matrix confirms `startup`, `selector/config`, `runtime/TUI`, `save`, and `navigation` are all covered by mapped scenarios.
- FR-001 negative-path check: Audit fails if any required journey area has no mapped scenario.
- FR-002 happy-path check: Every scenario has exactly one startup script and one exact startup command.
- FR-002 negative-path check: Any scenario violating one-script/one-command rule fails compliance audit.
- FR-003 happy-path check: All scenarios pass strict template heading and required-field compliance.
- FR-003 negative-path check: Any missing required heading or metadata field fails structure audit.
- FR-004 happy-path check: All step rows maintain one action, one expected outcome, and one assertion ID mapping.
- FR-004 negative-path check: Any broken step mapping fails step-level audit.
- FR-005 happy-path check: Assertions and final results are fully deterministic and binary (`PASS`/`FAIL`).
- FR-005 negative-path check: Any ambiguous pass criteria or forbidden result state fails determinism audit.
- FR-006 happy-path check: Every critical journey includes explicit failure trigger and user-visible recovery validation where applicable.
- FR-006 negative-path check: Any critical journey missing failure/recovery validation fails audit.
- FR-007 happy-path check: Scenario quality review confirms context-rich scenarios with no redundant low-value fragmentation.
- FR-007 negative-path check: Fragmentation or low-value split findings fail quality review.
- FR-008 happy-path check: Suite-level result is PASS only when every scenario and assertion is PASS.
- FR-008 negative-path check: Suite-level result is FAIL if any scenario result is FAIL.
- Metric checkpoint (M1):
  - Metric ID: M1
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: 100% journey-area coverage (`5/5`).
  - Check procedure: Validate coverage matrix area counts and mapped scenario IDs.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: 100% critical-journey failure/recovery coverage (`5/5` where applicable).
  - Check procedure: Validate failure/recovery mapping checklist against all critical journeys.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: 100% template/spec compliance across full suite (`N/N`).
  - Check procedure: Execute compliance audit and verify every scenario passes.
- Metric checkpoint (M4):
  - Metric ID: M4
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: 0 determinism integrity violations.
  - Check procedure: Execute determinism audit and confirm zero violations.

## Acceptance Criteria

1. Full suite passes coverage, compliance, determinism, and dependency consistency audits.
2. Metric targets for M1, M2, M3, and M4 are evidenced and evaluated against defined thresholds.
3. Release-readiness summary exists with deterministic PASS/FAIL decision logic.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-005-TASK-02-startup-and-selector-config-scenarios](.tasks/PRD-005-TASK-02-startup-and-selector-config-scenarios.md), [PRD-005-TASK-03-runtime-tui-scenarios](.tasks/PRD-005-TASK-03-runtime-tui-scenarios.md), [PRD-005-TASK-04-save-and-navigation-dirty-state-scenarios](.tasks/PRD-005-TASK-04-save-and-navigation-dirty-state-scenarios.md)
- blocks: none

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Not started
