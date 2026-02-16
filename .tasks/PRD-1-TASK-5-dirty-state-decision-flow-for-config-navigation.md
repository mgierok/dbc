# Overview

Protect staged data integrity during `:config` navigation by requiring explicit save, discard, or cancel decisions when unsaved changes exist.

## Metadata

- Status: DONE
- PRD: PRD-1-database-config-management.md
- Task ID: 5
- Task File: PRD-1-TASK-5-dirty-state-decision-flow-for-config-navigation.md
- PRD Requirements: FR-007, NFR-001, NFR-002, NFR-003, NFR-005

## Objective

Add mandatory tri-option dirty-state prompt (`save`, `discard`, `cancel`) before allowing `:config` navigation from an active dirty session.

## Working Software Checkpoint

`:config` navigation remains usable, and dirty-state paths now require explicit user decision so no staged changes are silently lost.

## Technical Scope

### In Scope

- Detect dirty staged state when `:config` is requested.
- Present explicit tri-option prompt with keyboard-first interactions.
- Implement behaviors:
  - `save`: run save flow, then navigate on success.
  - `discard`: clear staged state, then navigate.
  - `cancel`: remain in current context, no state loss.
- Add regression tests for data-loss prevention.

### Out of Scope

- Generalized multi-context unsaved-change framework.
- Additional command behaviors beyond `:config` safety flow.

## Implementation Plan

1. Extend confirm popup model to support tri-option action for config navigation.
2. Integrate dirty-state check into `:config` handling path.
3. Wire save/discard/cancel branches to existing staged-change mechanisms.
4. Add tests for each branch and failure handling (save error retains state, no navigation).
5. Update product/technical docs for new command safety behavior and shortcuts.

## Verification Plan

- Verify dirty `:config` always prompts and blocks navigation until explicit choice.
- Verify `save` branch navigates only after successful persistence.
- Verify `discard` branch clears staged state then navigates.
- Verify `cancel` branch keeps user in place with unchanged staged state.
- Verify no silent staged-data loss in regression tests.

## Acceptance Criteria

- Dirty-state `:config` flow enforces explicit `save`, `discard`, or `cancel`.
- Navigation cannot proceed without explicit decision when dirty.
- Save failure path preserves staged state and reports clear error.
- Data-loss guard behavior is documented in product and technical documentation.
- Project validation requirement: affected tests pass and planned verification commands are defined (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-1-TASK-4-in-session-config-command-routing](.tasks/PRD-1-TASK-4-in-session-config-command-routing.md)
- blocks: none

## Completion Summary

Implemented dirty-state decision flow for `:config` with explicit `save`, `discard`, and `cancel` paths.

- Updated `internal/interfaces/tui/model.go`:
  - `:config` now checks dirty state and opens tri-option decision popup when staged changes exist.
  - Added keyboard selection support for multi-option confirm popups (`j/k`, `Enter`, `Esc`).
  - Implemented behaviors:
    - `save`: execute save flow, then open selector only after successful save.
    - `discard`: clear staged state and open selector immediately.
    - `cancel`: close popup and keep current session with staged state preserved.
  - Save error path now blocks navigation and preserves staged state.
- Updated `internal/interfaces/tui/view.go`:
  - confirm popup renders selectable options for tri-option decisions,
  - status shortcut hints include multi-option confirm controls.
- Added regression tests in `internal/interfaces/tui/model_test.go` covering:
  - dirty `:config` prompt gating,
  - cancel keeps staged state and blocks navigation,
  - discard clears state and navigates,
  - save success navigates only after persistence,
  - save failure keeps staged state and blocks navigation.
- Updated documentation:
  - `docs/product-documentation.md`
  - `docs/technical-documentation.md`

Verification run:
- `go test ./...` passed.
- `golangci-lint run ./...` passed with `0 issues`.
