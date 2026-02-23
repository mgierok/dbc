# Overview

Add runtime `:help` command routing and popup lifecycle behavior so active-session users can open help deterministically, re-open idempotently, and close only with `Esc` while preserving existing unsupported-command behavior.

## Metadata

- Status: DONE
- PRD: PRD-008-runtime-help-popup-reference.md
- Task ID: 01
- Task File: PRD-008-TASK-01-runtime-help-command-and-popup-lifecycle.md
- PRD Requirements: FR-001, FR-005, FR-006, FR-007, NFR-001
- PRD Metrics: M2, M4

## Objective

Implement runtime command handling and popup lifecycle state transitions for `:help` without changing existing behavior for unsupported commands.

## Working Software Checkpoint

After this task, users can run `:help` in active runtime contexts to open a help popup, repeated `:help` keeps it open, `Esc` closes it, and unsupported commands still show unknown-command status without breaking session flow.

## Technical Scope

### In Scope

- Runtime command-entry handling branch for `:help`.
- Help popup lifecycle state management (open, already-open idempotence, close).
- Close behavior constrained to `Esc` for help popup lifecycle.
- Regression preservation for unsupported runtime command handling.

### Out of Scope

- Final help popup content sections and scrolling behavior.
- Startup or selector-context help behavior.
- Adding new runtime command families beyond `:help`.

## Implementation Plan

1. Extend runtime command submission logic in TUI model to recognize `:help` in active main-session contexts.
2. Add help popup lifecycle state fields and transitions for open and already-open idempotent handling.
3. Ensure help popup close path is triggered by `Esc` and unrelated keys do not dismiss popup.
4. Keep unsupported-command fallback unchanged and verify no routing regressions for unknown commands.
5. Update focused runtime model tests for command lifecycle and unsupported-command guardrail behavior.

## Verification Plan

- FR-001 happy-path check: Enter `:help` in runtime command entry and verify help popup opens with no unknown-command status.
- FR-001 negative-path check: Enter `help` without `:` and verify command is not executed as runtime command.
- FR-005 happy-path check: Enter `:help` while help popup is already open and verify popup remains open.
- FR-005 negative-path check: Re-enter `:help` and verify it does not close/reset popup unexpectedly.
- FR-006 happy-path check: Press `Esc` while help popup is open and verify popup closes.
- FR-006 negative-path check: Press unrelated keys while help popup is open and verify popup remains open.
- FR-007 happy-path check: Enter unsupported runtime command and verify unknown-command status is shown while session remains active.
- FR-007 negative-path check: Enter misspelled `:help` variant (for example `:helpp`) and verify unknown-command handling path is unchanged.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: FR coverage checkpoints in this task mapped to FR-001, FR-005, FR-006, and FR-007 are PASS.
  - Check procedure: Record per-FR assertion results in this task file `Completion Summary` when task is marked DONE.
- Metric checkpoint (M4):
  - Metric ID: M4
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: 0 regressions in unsupported-command handling.
  - Check procedure: Record unsupported-command regression assertion results in this task file `Completion Summary` when task is marked DONE.

## Acceptance Criteria

1. Runtime command entry accepts `:help` in active main-session contexts and opens help popup.
2. Re-running `:help` while popup is open is idempotent and does not dismiss popup.
3. Help popup closes on `Esc` and remains open for unrelated keys.
4. Unsupported-command behavior remains unchanged with unknown-command status and usable session.
5. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: none
- blocks: [PRD-008-TASK-02-runtime-help-popup-content-and-scroll](.tasks/PRD-008-TASK-02-runtime-help-popup-content-and-scroll.md), [PRD-008-TASK-03-runtime-help-release-evidence-test-update](.tasks/PRD-008-TASK-03-runtime-help-release-evidence-test-update.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

- Implemented runtime `:help` command routing in `internal/interfaces/tui/model.go`:
  - `:help` opens runtime help popup.
  - Re-running `:help` while popup is open keeps popup open (idempotent open).
  - Help popup closes only on `Esc`; unrelated keys keep it open.
  - Existing unknown-command fallback remains unchanged.
- Added focused lifecycle/regression tests in `internal/interfaces/tui/model_test.go`:
  - `TestHandleKey_CommandHelpOpensPopupWithoutUnknownStatus` (FR-001 happy path) PASS.
  - `TestHandleKey_HelpCommandRequiresExplicitPrefix` (FR-001 negative path) PASS.
  - `TestHandleKey_CommandHelpReenterKeepsPopupOpen` (FR-005 happy path) PASS.
  - `TestHandleKey_CommandHelpReenterDoesNotResetStatusUnexpectedly` (FR-005 negative path) PASS.
  - `TestHandleKey_HelpPopupEscClosesPopup` (FR-006 happy path) PASS.
  - `TestHandleKey_HelpPopupUnrelatedKeysDoNotClosePopup` (FR-006 negative path) PASS.
  - `TestHandleKey_InvalidCommandShowsErrorAndKeepsSessionActive` (FR-007 happy path) PASS.
  - `TestHandleKey_MisspelledHelpCommandUsesUnknownCommandFallback` (FR-007 negative path) PASS.
- Added help popup rendering/shortcut state in `internal/interfaces/tui/view.go` to reflect runtime lifecycle state.
- Updated source-of-truth docs:
  - `docs/product-documentation.md` with runtime `:help` command behavior and help-popup interaction.
  - `docs/technical-documentation.md` with runtime `:help` routing/lifecycle mechanics.
- Verification evidence:
  - `go test ./internal/interfaces/tui` PASS.
  - `golangci-lint run ./...` PASS (`0 issues`).
  - `go test ./...` PASS.
- Metric checkpoint results:
  - M2 PASS: FR-001, FR-005, FR-006, FR-007 verification checkpoints are all PASS.
  - M4 PASS: unsupported-command fallback remains unchanged with 0 regressions.
