# Overview

This task implements the runtime keyboard-navigation contract change from panel-switch shortcuts to an explicit `Enter`/`Esc` transition model while keeping existing context safety behavior intact.

## Metadata

- Status: DONE
- PRD: PRD-007-simplified-panel-navigation-enter-esc.md
- Task ID: 01
- Task File: PRD-007-TASK-01-runtime-navigation-contract-enter-esc.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002
- PRD Metrics: M1, M5

## Objective

Implement deterministic runtime panel-transition behavior where left-panel `Enter` opens right-panel Records view, right-neutral `Esc` returns to left-panel table selection, and `Ctrl+w h/l/w` no longer performs panel transitions.

## Working Software Checkpoint

After this task, the runtime remains fully usable for table browsing, records interaction, and nested-context exits, with panel transitions performed through the new two-key model.

## Technical Scope

### In Scope

- Update runtime key handling for panel transitions in TUI model code.
- Remove `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w` transition support.
- Preserve context-first `Esc` behavior in nested right-panel contexts.
- Add or update focused TUI unit tests for the new key contract.

### Out of Scope

- User-facing documentation and README keybinding updates.
- `TC-002` scenario file rewrite.
- Regression-audit updates for `TC-003..TC-008`.

## Implementation Plan

1. Update runtime input handling so left-panel `Enter` confirms table selection and transitions focus to right-panel Records view.
2. Update right-panel `Esc` handling to return to left-panel table selection only when right-panel state is neutral.
3. Keep nested right-panel `Esc` precedence unchanged for field-focus and popup contexts.
4. Remove runtime behavior that treats `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w` as panel-transition shortcuts.
5. Add and update TUI unit tests that verify deterministic repeated transitions and non-transition behavior for removed shortcuts.

## Verification Plan

- FR-001 happy-path check: With left panel active and a selected table, pressing `Enter` opens Records view and makes right panel active.
- FR-001 negative-path check: Pressing `Enter` in non-left-panel contexts does not break existing records/edit flow behavior.
- FR-002 happy-path check: In right-panel neutral runtime state, pressing `Esc` returns focus to left panel and preserves selected table context.
- FR-002 negative-path check: Pressing `Esc` outside right-panel neutral state does not trigger immediate panel return.
- FR-003 happy-path check: In nested right-panel context (field focus or popup), pressing `Esc` exits nested context first.
- FR-003 negative-path check: A single `Esc` in nested right-panel context does not jump directly to left-panel selection.
- FR-004 happy-path check: `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w` do not cause panel transitions after the change.
- FR-004 negative-path check: Pressing removed `Ctrl+w` combinations does not silently alter active-panel state.
- Metric checkpoint (M1):
  - Metric ID: M1
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Runtime panel transitions use only `Enter` and `Esc`, and removed `Ctrl+w` combinations are unsupported.
  - Check procedure: Execute focused runtime key-flow tests and record observed panel states for `Enter`, `Esc`, and `Ctrl+w` combinations.
- Metric checkpoint (M5):
  - Metric ID: M5
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Context-first nested `Esc` behavior remains 100% retained in covered right-panel nested states.
  - Check procedure: Execute nested-context regression tests and confirm first `Esc` exits local context before any panel transition.

## Acceptance Criteria

1. Runtime panel transitions follow the new `Enter`/`Esc` model without regressions in nested-context safety.
2. `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w` no longer transition panel focus.
3. Unit tests cover happy and negative transition paths for the updated key contract.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: none
- blocks: [PRD-007-TASK-02-runtime-key-guidance-and-doc-sync](.tasks/PRD-007-TASK-02-runtime-key-guidance-and-doc-sync.md), [PRD-007-TASK-03-tc-002-navigation-model-update](.tasks/PRD-007-TASK-03-tc-002-navigation-model-update.md), [PRD-007-TASK-05-integration-hardening](.tasks/PRD-007-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

- Implemented runtime navigation contract in `internal/interfaces/tui/model.go`:
  - `Enter` now transitions from left-panel table selection to right-panel Records view with content focus.
  - Neutral right-panel `Esc` now returns focus to left-panel table selection.
  - Nested-context `Esc` precedence remains intact (`recordFieldFocus` exits first, no immediate panel return in the same keypress).
  - Removed `Ctrl+w h/l/w` panel-transition behavior (deleted pending `Ctrl+w` handling and focus-toggle path).
- Added/updated focused unit coverage in `internal/interfaces/tui/model_test.go`:
  - `TestHandleKey_EnterFromTablesSwitchesToRecordsAndContentFocus`
  - `TestHandleKey_EscInRightPanelNeutralReturnsToTables`
  - `TestHandleKey_CtrlWPanelShortcutsAreUnsupported`
  - strengthened nested `Esc` assertion in `TestHandleKey_EscClearsFieldFocus`
- Verification executed:
  - `go test ./internal/interfaces/tui -run 'TestHandleKey_(EnterFromTablesSwitchesToRecordsAndContentFocus|EscInRightPanelNeutralReturnsToTables|CtrlWPanelShortcutsAreUnsupported|EscClearsFieldFocus)'`
  - `go test ./internal/interfaces/tui`
  - `gofmt -w internal/interfaces/tui/model.go internal/interfaces/tui/model_test.go`
  - `golangci-lint run ./...`
  - `go test ./...`
- Downstream impact for dependent tasks:
  - TASK-02 can now align user-visible key guidance with implemented `Enter`/`Esc` model.
  - TASK-03/TASK-04 can validate and audit runtime cases against removed `Ctrl+w` transitions.
