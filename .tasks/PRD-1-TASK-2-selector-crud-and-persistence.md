# Overview

Extend startup selector into a configuration management surface so users can add, edit, and delete database entries entirely in-app with persisted changes and visible config path.

## Metadata

- Status: DONE
- PRD: PRD-1-database-config-management.md
- Task ID: 2
- Task File: PRD-1-TASK-2-selector-crud-and-persistence.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-009, FR-010, NFR-001, NFR-002, NFR-004, NFR-005

## Objective

Implement selector-level CRUD management for database entries with confirmation on destructive actions and immediate persisted results.

## Working Software Checkpoint

Users can still select and open a database from startup selector, and now can manage entries in-app with confirmed deletion and persisted updates.

## Technical Scope

### In Scope

- Add selector actions for add/edit/delete entries.
- Integrate selector with config management use cases from Task 1.
- Show active config file path in selector/management UI.
- Add delete confirmation UX and error/status messaging for invalid actions.

### Out of Scope

- Forced first-database gating flow when config is empty.
- In-session `:config` command behavior.

## Implementation Plan

1. Extend selector model state and key handling for CRUD actions.
2. Add modal/form interactions for entry creation and editing.
3. Add confirmation prompt for deletion and apply persistence call.
4. Refresh selector list after each successful mutation.
5. Add selector tests covering add/edit/delete and confirmation flows.

## Verification Plan

- Verify add/edit/delete updates selector list in the same session.
- Verify persisted changes survive app restart.
- Verify deletion requires explicit confirmation.
- Verify active config path is always displayed in selector management context.

## Acceptance Criteria

- Users can add a valid entry (`name`, `db_path`) in selector and select it immediately.
- Users can edit existing entry values and see updates immediately.
- Users can delete an entry only after explicit confirmation.
- Selector displays active config file path while managing entries.
- Project validation requirement: affected tests pass and planned verification commands are defined (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-1-TASK-1-config-management-foundation](.tasks/PRD-1-TASK-1-config-management-foundation.md)
- blocks: [PRD-1-TASK-3-forced-first-database-setup](.tasks/PRD-1-TASK-3-forced-first-database-setup.md), [PRD-1-TASK-4-in-session-config-command-routing](.tasks/PRD-1-TASK-4-in-session-config-command-routing.md)

## Completion Summary

Implemented selector-level configuration CRUD with use-case integration and explicit delete confirmation.

- Reworked startup selector integration:
  - `cmd/dbc/main.go` now passes config management use cases (`list/create/update/delete/active-path`) into selector startup flow.
  - `internal/interfaces/tui/selector.go` now uses a config-management adapter over application use cases instead of static options.
- Added selector CRUD interactions in startup UI:
  - `a` opens add form.
  - `e` opens edit form for selected entry.
  - `d` opens explicit delete confirmation (delete only on confirm).
  - Form supports field switching (`Tab`), field clear (`Ctrl+u`), submit (`Enter`), cancel (`Esc`).
  - Selector refreshes entries after successful mutations and keeps selection stable.
- Added active config-path visibility:
  - Selector renders `Config: <active-path>` in startup management view.
- Added selector tests for new behavior:
  - `internal/interfaces/tui/selector_test.go` now covers add/edit/delete flows, explicit delete confirmation, and active path rendering.
- Updated documentation to reflect delivered behavior:
  - `docs/product-documentation.md`
  - `docs/technical-documentation.md`

Verification run:
- `go test ./...` passed
- `golangci-lint run ./...` passed

Downstream context for Task 3/4:
- Selector now supports empty-list rendering and in-place creation/editing/deletion mechanics, providing reusable building blocks for forced first-database setup and later `:config` navigation integration.
