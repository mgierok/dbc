# Overview

Implement mandatory first-database creation when no configured entries exist, with optional additional entry creation in the same setup context before entering normal browsing.

## Metadata

- Status: READY
- PRD: PRD-1-database-config-management.md
- Task ID: 3
- Task File: PRD-1-TASK-3-forced-first-database-setup.md

## Objective

Enforce first-run gating that blocks normal browsing until at least one valid database entry is created in-app.

## Working Software Checkpoint

Startup remains functional for existing users with configured entries, and first-run users get a guided setup flow that guarantees at least one valid entry before app entry.

## Technical Scope

### In Scope

- Detect zero-entry config state at startup.
- Route to mandatory creation flow before selector/browsing progression.
- Allow optional loop to add additional entries before completion.
- Ensure completion criterion is at least one valid persisted entry.

### Out of Scope

- In-session `:config` command workflow.
- Dirty-state decision flow for mid-session navigation.

## Implementation Plan

1. Add startup branch for empty config state.
2. Reuse selector-management form components for first-entry creation.
3. Add continue/finish controls for optional additional entry creation.
4. Prevent exit to main browsing until valid first entry exists.
5. Add startup flow tests for empty and non-empty config paths.

## Verification Plan

- Verify app cannot enter main browsing with zero entries.
- Verify first valid entry unlocks normal flow.
- Verify optional additional entry loop persists each confirmed entry.
- Verify existing non-empty configs bypass forced setup.

## Acceptance Criteria

- Zero-entry startup always enforces first-entry creation.
- At least one valid entry is required before main browsing starts.
- Users can add additional entries in the same setup context.
- Flow remains keyboard-first and aligned with selector interaction style.
- Project validation requirement: affected tests pass and planned verification commands are defined (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-1-TASK-1-config-management-foundation](.tasks/PRD-1-TASK-1-config-management-foundation.md), [PRD-1-TASK-2-selector-crud-and-persistence](.tasks/PRD-1-TASK-2-selector-crud-and-persistence.md)
- blocks: [PRD-1-TASK-4-in-session-config-command-routing](.tasks/PRD-1-TASK-4-in-session-config-command-routing.md)

## Completion Summary

Not started
