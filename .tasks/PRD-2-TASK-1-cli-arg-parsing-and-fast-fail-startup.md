# Overview

Add direct-launch CLI argument handling and startup branching so known-target sessions can bypass selector and fail fast when invalid.

## Metadata

- Status: DONE
- PRD: PRD-2-cli-direct-database-launch.md
- Task ID: 1
- Task File: PRD-2-TASK-1-cli-arg-parsing-and-fast-fail-startup.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, NFR-001, NFR-002

## Objective

Implement direct-launch CLI parsing and fail-fast startup behavior.

## Working Software Checkpoint

The app still supports selector-first startup when direct-launch parameter is absent, and direct-launch startup works independently when the parameter is provided.

## Technical Scope

### In Scope

- Parse `-d` and `--database` direct-launch aliases.
- Branch startup flow to direct-launch path before selector loop.
- Validate direct-launch target before runtime UI starts.
- Skip selector for successful direct-launch startup.
- Print clear failure output and exit non-zero on direct-launch validation failure.
- Add tests for argument parsing and startup branch behavior.

### Out of Scope

- Session-scoped temporary selector entry behavior.
- Normalized path matching and configured-entry reuse logic.
- Changes to in-session `:config` behavior.

## Implementation Plan

1. Add argument parsing for direct-launch aliases in startup entrypoint.
2. Add startup orchestration branch for direct-launch when parameter is present.
3. Reuse existing database open/ping validation path before runtime start.
4. Return readable startup failure output with non-zero exit when validation fails.
5. Keep existing selector-first flow unchanged when direct-launch parameter is absent.

## Verification Plan

- FR-001 happy-path check: start app with `-d <valid-db-path>` and with `--database <valid-db-path>`; confirm both are recognized as direct-launch requests.
- FR-001 negative-path check: start app with missing direct-launch value; confirm startup reports parameter error and exits non-zero.
- FR-002 happy-path check: provide reachable target; confirm runtime starts only after successful connectivity validation.
- FR-002 negative-path check: provide unreachable target; confirm runtime does not start.
- FR-003 happy-path check: on validation failure, confirm output includes clear failure reason and corrective guidance.
- FR-003 negative-path check: on validation failure, confirm selector is not opened as fallback and process exits non-zero.
- FR-004 happy-path check: successful direct-launch reaches main runtime without selector interaction.
- FR-004 negative-path check: no direct-launch parameter still uses existing selector-first startup.
- NFR-001/NFR-002 check: verify direct-launch startup messaging remains understandable and unambiguous.

## Acceptance Criteria

- Direct-launch aliases `-d` and `--database` are accepted and trigger direct-launch mode.
- Main runtime opens only when direct-launch connectivity validation succeeds.
- Direct-launch validation failure outputs clear user-facing message and exits with non-zero status.
- Successful direct-launch bypasses startup selector.
- Selector-first startup behavior remains unchanged when direct-launch parameter is not provided.
- Project validation requirement: affected tests and quality checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: none
- blocks: [PRD-2-TASK-2-launch-target-identity-and-config-reuse](.tasks/PRD-2-TASK-2-launch-target-identity-and-config-reuse.md), [PRD-2-TASK-3-session-scoped-selector-entry-for-direct-launch](.tasks/PRD-2-TASK-3-session-scoped-selector-entry-for-direct-launch.md)

## Completion Summary

Implemented direct-launch startup path in `cmd/dbc/main.go` with explicit parsing for `-d` and `--database`, including fail-fast handling for invalid argument usage and unsupported startup arguments.

Added startup branch resolution so direct-launch selection is attempted before selector flow, while preserving selector-first behavior when no direct-launch argument is provided.

Reused existing `connectSelectedDatabase`/`engine.OpenSQLiteDatabase` validation path for direct-launch connectivity checks; on direct-launch failure startup now exits non-zero with a clear, actionable error message and no selector fallback.

Added/updated tests in `cmd/dbc/main_test.go` for:
- direct-launch alias parsing (`-d`, `--database`),
- missing-value and unsupported-argument error handling with clear messaging,
- startup branch resolution behavior (direct-launch bypasses selector callback, selector path preserved when flag absent),
- direct-launch failure message clarity and guidance text.

Verification executed:
- `go test ./...` (PASS)
- `golangci-lint run ./...` (PASS)
