# Overview

Ensure direct-launched sessions preserve `:config` switching behavior and expose CLI-launched target in selector only for current process without persisting it.

## Metadata

- Status: DONE
- PRD: PRD-002-cli-direct-database-launch.md
- Task ID: 03
- Task File: PRD-002-TASK-03-session-scoped-selector-entry-for-direct-launch.md
- PRD Requirements: FR-005, FR-006, FR-007, NFR-003, NFR-004, NFR-005

## Objective

Add session-scoped selector entry behavior for direct-launched targets with strict non-persistence.

## Working Software Checkpoint

Users can direct-launch, work normally, invoke `:config`, and switch safely, while configuration file contents remain unchanged unless explicit selector CRUD actions are taken.

## Technical Scope

### In Scope

- Track active direct-launch target in process-lifetime session state.
- Extend selector launch state to carry session-scoped additional options.
- Apply unified UTF source markers to all selector options:
  - `⚙` for config-backed entries.
  - `⌨` for CLI session-scoped entries.
- Inject temporary selector option for CLI-launched target when direct-launch target is not already reused from config.
- Preserve current dirty-state decision behavior before `:config` navigation.
- Ensure temporary direct-launch entry is not auto-persisted to config.
- Add tests for selector visibility, switching behavior, and non-persistence across restart.

### Out of Scope

- New configuration schema fields for source tagging.
- Changes to selector CRUD persistence semantics.
- Non-SQLite connection behavior extensions.

## Implementation Plan

1. Extend selector launch-state contract for process-scoped additional options.
2. Populate launch state with direct-launch temporary option when opening selector from runtime.
3. Update selector rendering so every option includes source marker (`⚙` config, `⌨` CLI).
4. Reuse configured identity path from Task 2 to suppress temporary option on match.
5. Ensure temporary option exists only in memory and is never written through config store.
6. Add tests for in-session visibility, marker correctness, and post-restart absence.

## Verification Plan

- FR-005 happy-path check: direct-launched session executes `:config` and switches to another available database.
- FR-005 negative-path check: invoke `:config` with dirty staged changes; confirm save/discard/cancel decision remains mandatory.
- FR-006 happy-path check: in direct-launched session, selector includes current target option for re-selection and marks it with `⌨`.
- FR-006 negative-path check: when direct-launch target reuses configured entry identity, temporary duplicate option is not shown.
- FR-007 happy-path check: direct-launched temporary entry is not present after app restart unless explicitly saved by existing flows.
- FR-007 negative-path check: opening selector during session without explicit save does not mutate config file.
- Marker consistency check: all config-backed options are marked with `⚙` and all CLI session-scoped options with `⌨`.
- NFR-003/NFR-004/NFR-005 check: verify session navigation and startup-path coherence stay predictable and safe.

## Acceptance Criteria

- `:config` remains available and functional in direct-launched sessions.
- Session-scoped direct-launch target is available in selector during the same process when needed.
- Every selector option has a source marker: `⚙` for config entries and `⌨` for CLI entries.
- No duplicate temporary entry appears when configured identity reuse applies.
- Temporary direct-launch entries are never auto-persisted to config.
- Existing dirty-state prompt behavior for `:config` remains unchanged.
- Project validation requirement: affected tests and quality checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-002-TASK-01-cli-arg-parsing-and-fast-fail-startup](.tasks/PRD-002-TASK-01-cli-arg-parsing-and-fast-fail-startup.md), [PRD-002-TASK-02-launch-target-identity-and-config-reuse](.tasks/PRD-002-TASK-02-launch-target-identity-and-config-reuse.md)
- blocks: [PRD-002-TASK-04-integration-hardening](.tasks/PRD-002-TASK-04-integration-hardening.md)

## Completion Summary

Implemented session-scoped selector entry support for direct-launch flow without config persistence side effects.

Delivered changes:
- Extended selector launch contract with `AdditionalOptions` and source-aware option model (`config` vs `CLI session`).
- Added in-memory session tracking in `cmd/dbc/main.go` to carry direct-launch temporary option across `:config` selector re-entry only when startup used non-config direct identity.
- Added selector rendering markers for all options (`⚙` config, `⌨` CLI session).
- Kept config mutations scoped to config-backed entries only; CLI session entries are selectable but blocked for edit/delete.
- Preserved existing dirty `:config` decision flow (`save` / `discard` / `cancel`) without behavioral regression.

Verification executed:
- `go test ./cmd/dbc ./internal/interfaces/tui` (PASS)
- `go test ./...` (PASS)
- `golangci-lint run ./...` (PASS)
