# Overview

Extend startup selector into a configuration management surface so users can add, edit, and delete database entries entirely in-app with persisted changes and visible config path.

## Metadata

- Status: READY
- PRD: PRD-1-database-config-management.md
- Task ID: 2
- Task File: PRD-1-TASK-2-selector-crud-and-persistence.md

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

Not started
