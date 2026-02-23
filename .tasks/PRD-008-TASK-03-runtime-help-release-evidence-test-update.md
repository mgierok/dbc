# Overview

Update one existing runtime-oriented test case into the authoritative release evidence artifact for PRD-008 so all functional requirements and guardrails are validated deterministically in a single maintained test entry point.

## Metadata

- Status: DONE
- PRD: PRD-008-runtime-help-popup-reference.md
- Task ID: 03
- Task File: PRD-008-TASK-03-runtime-help-release-evidence-test-update.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, NFR-003
- PRD Metrics: M1, M2, M3, M4

## Objective

Expand one existing runtime test case to provide deterministic release validation evidence for all PRD-008 requirements and metrics.

## Working Software Checkpoint

After this task, one updated existing runtime test case serves as the single release-evidence artifact for `:help` behavior coverage and unsupported-command regression protection while the broader suite remains runnable.

## Technical Scope

### In Scope

- Select and extend one existing runtime test case as the primary evidence test.
- Add structured subcases/assertions in that existing test for FR-001 through FR-007.
- Capture metric-oriented checks (M1-M4) through explicit test assertions and completion evidence notes.
- Keep unsupported-command regression assertions inside the same updated existing test case.

### Out of Scope

- Creating separate evidence documents outside task `Completion Summary`.
- Introducing unrelated new command behavior tests not tied to PRD-008.
- Refactoring unrelated test infrastructure.

## Implementation Plan

1. Identify one existing runtime test case that already exercises command-entry behavior and extend it as the canonical PRD-008 evidence test.
2. Add explicit happy-path and negative-path assertion coverage for each FR-001 through FR-007 within the updated existing test case.
3. Ensure unsupported-command regression assertions are retained in the same updated test case.
4. Execute targeted test run for the updated case and affected scope checks required by repository workflow.
5. Record FR/NFR/metric evidence outcomes in this task file `Completion Summary` when marking DONE.

## Verification Plan

- FR-001 happy-path check: Updated existing test case asserts `:help` opens popup in runtime context.
- FR-001 negative-path check: Updated existing test case asserts invalid command prefix/path does not satisfy `:help` behavior.
- FR-002 happy-path check: Updated existing test case asserts `Supported Commands` section is present.
- FR-002 negative-path check: Updated existing test case asserts section absence/mislabel is detected as failure condition.
- FR-003 happy-path check: Updated existing test case asserts `Supported Keywords` section is present.
- FR-003 negative-path check: Updated existing test case asserts missing keyword descriptions are detected as failure condition.
- FR-004 happy-path check: Updated existing test case asserts keyboard scrolling reaches final help item for overflow content.
- FR-004 negative-path check: Updated existing test case asserts non-scroll path does not incorrectly advance scroll state.
- FR-005 happy-path check: Updated existing test case asserts repeated `:help` is idempotent while popup is open.
- FR-005 negative-path check: Updated existing test case asserts repeated `:help` does not dismiss popup.
- FR-006 happy-path check: Updated existing test case asserts `Esc` closes help popup.
- FR-006 negative-path check: Updated existing test case asserts unrelated key does not close help popup.
- FR-007 happy-path check: Updated existing test case asserts unsupported command behavior remains unknown-command status with active session.
- FR-007 negative-path check: Updated existing test case asserts unsupported-command regression conditions remain blocked.
- Metric checkpoint (M1):
  - Metric ID: M1
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Updated existing test case pass rate is 100% during release validation.
  - Check procedure: Run the updated existing test case and record PASS result in this task file `Completion Summary`.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: 7/7 FR validation mappings are PASS.
  - Check procedure: Record FR-to-assertion mapping and PASS outcomes in this task file `Completion Summary`.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Required popup section completeness is 2/2 with description presence PASS.
  - Check procedure: Record section completeness assertion outcomes in this task file `Completion Summary`.
- Metric checkpoint (M4):
  - Metric ID: M4
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: 0 regressions accepted in unsupported-command handling.
  - Check procedure: Record unsupported-command guardrail assertion outcomes in this task file `Completion Summary`.

## Acceptance Criteria

1. One existing runtime test case is updated as the single release-evidence artifact for PRD-008.
2. The updated existing test case covers happy-path and negative-path validations for FR-001 through FR-007.
3. Metric evidence mappings for M1-M4 are represented by assertions and ready for completion logging in this task file.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-008-TASK-01-runtime-help-command-and-popup-lifecycle](.tasks/PRD-008-TASK-01-runtime-help-command-and-popup-lifecycle.md), [PRD-008-TASK-02-runtime-help-popup-content-and-scroll](.tasks/PRD-008-TASK-02-runtime-help-popup-content-and-scroll.md)
- blocks: none

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

- Updated existing runtime command-entry test `TestHandleKey_CommandHelpOpensPopupWithoutUnknownStatus` in `internal/interfaces/tui/model_test.go` into the canonical PRD-008 release-evidence test with structured subcases.
- Added explicit happy-path and negative-path assertions in the same test for FR-001 through FR-007:
  - FR-001: `:help` opens popup; missing `:` prefix does not execute help command.
  - FR-002: `Supported Commands` section exists; command rows enforce canonical section/description shape.
  - FR-003: `Supported Keywords` section exists; keyword rows enforce one-line descriptions.
  - FR-004: keyboard scrolling reaches final help item; non-scroll key keeps scroll offset stable.
  - FR-005: repeated `:help` remains idempotent and does not dismiss popup.
  - FR-006: `Esc` closes popup; unrelated key does not close popup.
  - FR-007: unsupported and misspelled commands retain unknown-command fallback with active session.
- Metric evidence:
  - M1 PASS: updated existing release-evidence test passes (`go test ./internal/interfaces/tui -run '^TestHandleKey_CommandHelpOpensPopupWithoutUnknownStatus$'`).
  - M2 PASS: FR mapping coverage is 7/7 PASS (FR-001..FR-007).
  - M3 PASS: required popup section completeness is 2/2 with one-line description checks.
  - M4 PASS: unsupported-command guardrail assertions remain PASS with 0 accepted regressions.
- Verification evidence:
  - `gopls check internal/interfaces/tui/model_test.go` PASS.
  - `go test ./internal/interfaces/tui -run '^TestHandleKey_CommandHelpOpensPopupWithoutUnknownStatus$'` PASS.
  - `golangci-lint run ./...` PASS (`0 issues`).
  - `go test ./...` PASS.
