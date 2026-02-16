# Overview

Add in-session command routing so users in an active database session can invoke `:config` and return to selector/configuration management without restarting the app.

## Metadata

- Status: DONE
- PRD: PRD-1-database-config-management.md
- Task ID: 4
- Task File: PRD-1-TASK-4-in-session-config-command-routing.md
- PRD Requirements: FR-006, NFR-001, NFR-003

## Objective

Implement `:config` command parsing and navigation routing from active session to selector/management context.

## Working Software Checkpoint

Main browsing remains fully operational; users can invoke `:config` from session and reach selector/management context when no dirty-state blocker applies.

## Technical Scope

### In Scope

- Add command-input mechanism in TUI for colon commands.
- Implement `:config` command detection and routing behavior.
- Preserve existing keybindings and focus behavior.
- Add navigation tests for command invocation and return path.

### Out of Scope

- Save/discard/cancel decision flow for dirty staged changes (implemented in Task 5).
- Additional custom command language beyond `:config`.

## Implementation Plan

1. Introduce command-entry state in `internal/interfaces/tui/model.go`.
2. Parse and validate `:config` invocation.
3. Add transition from active model to selector/management context.
4. Ensure selected database re-open workflow remains stable after returning.
5. Add tests for command parsing and successful navigation.

## Verification Plan

- Verify `:config` is recognized only as explicit command input.
- Verify navigation from active session to selector management works repeatedly.
- Verify existing non-command keyboard controls remain unaffected.
- Verify invalid command input fails safely with status message.

## Acceptance Criteria

- Users can invoke `:config` from active session.
- `:config` transitions to selector/management context without restart.
- Existing shortcuts for table browsing/editing still behave correctly.
- Navigation flow is test-covered for success and invalid-command cases.
- Project validation requirement: affected tests pass and planned verification commands are defined (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-1-TASK-2-selector-crud-and-persistence](.tasks/PRD-1-TASK-2-selector-crud-and-persistence.md), [PRD-1-TASK-3-forced-first-database-setup](.tasks/PRD-1-TASK-3-forced-first-database-setup.md)
- blocks: [PRD-1-TASK-5-dirty-state-decision-flow-for-config-navigation](.tasks/PRD-1-TASK-5-dirty-state-decision-flow-for-config-navigation.md)

## Completion Summary

Implemented in-session command routing for `:config` and runtime return to selector management.

- Added command-entry state and parsing in `internal/interfaces/tui/model.go`:
  - `:` opens command mode,
  - `Enter` executes command,
  - `Esc` cancels command input,
  - unknown commands fail safely with `Unknown command: ...` status message.
- Added `:config` routing signal in runtime model and app loop:
  - `internal/interfaces/tui/app.go` now returns `ErrOpenConfigSelector` when runtime exits by `:config`,
  - `cmd/dbc/main.go` now loops selector -> session so users can return to selector and reopen databases without restarting.
- Preserved existing non-command keyboard behavior and added tests:
  - `internal/interfaces/tui/model_test.go` now covers success path (`:config`), invalid command handling, and explicit-prefix requirement.
- Updated documentation:
  - `docs/product-documentation.md`
  - `docs/technical-documentation.md`

Verification run:
- `go test ./internal/interfaces/tui` passed (after Red -> Green cycle).
- `go test ./...` passed.
- `golangci-lint run ./...` passed with `0 issues`.

Downstream context for Task 5:
- `:config` currently navigates immediately once invoked; staged-change decision flow (save/discard/cancel) is intentionally still pending in Task 5.
