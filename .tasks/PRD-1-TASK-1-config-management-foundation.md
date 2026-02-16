# Overview

Build the configuration-management foundation so application and UI layers can safely read, write, and validate database entries across platforms with explicit active config path visibility.

## Metadata

- Status: DONE
- PRD: PRD-1-database-config-management.md
- Task ID: 1
- Task File: PRD-1-TASK-1-config-management-foundation.md
- PRD Requirements: FR-008, FR-010, NFR-004, NFR-005

## Objective

Introduce a stable config management foundation (port/use-case/infrastructure contract) for CRUD persistence and OS-aware default path resolution.

## Working Software Checkpoint

The app still starts and loads existing config entries exactly as today, while new foundation APIs are covered by tests and are ready for selector-management integration.

## Technical Scope

### In Scope

- Define application-level config management ports and DTOs for entry CRUD plus active-path lookup.
- Add infrastructure implementations for reading/writing `config.toml` with validation and safe persistence.
- Add/adjust config path resolver behavior to be OS-aware and testable.
- Add unit tests for decode/validate/write/path behaviors and failure cases.

### Out of Scope

- Selector UI interaction changes.
- Forced first-database setup UX.
- `:config` in-session navigation.

## Implementation Plan

1. Introduce config management interfaces in `internal/application/port`.
2. Add config management use cases in `internal/application/usecase`.
3. Extend `internal/infrastructure/config` with write/update/delete operations and path resolver tests.
4. Keep existing startup load path functional via adapters to new interfaces.
5. Add regression tests for legacy config compatibility.

## Verification Plan

- Run package tests for application/config infrastructure.
- Verify config read-only behavior remains unchanged on existing valid file.
- Verify add/edit/delete persistence round-trips on temporary files.
- Verify path resolver outputs expected OS-specific defaults.

## Acceptance Criteria

- Config management operations exist behind application ports and are test-covered.
- Existing `[[databases]]` files continue to load without migration.
- Persistence operations validate required fields (`name`, `db_path`) and reject invalid writes.
- Active config path can be queried from application layer.
- Project validation requirement: affected tests pass and planned verification commands are defined (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: none
- blocks: [PRD-1-TASK-2-selector-crud-and-persistence](.tasks/PRD-1-TASK-2-selector-crud-and-persistence.md), [PRD-1-TASK-3-forced-first-database-setup](.tasks/PRD-1-TASK-3-forced-first-database-setup.md)

## Completion Summary

Implemented config-management foundation across application and infrastructure layers.

- Added application contracts for config management:
  - `internal/application/port/config_store.go`
  - `internal/application/dto/config_database.go`
  - `internal/application/usecase/config_management.go`
- Added use-case tests for list/create/update/delete/path behaviors and validation:
  - `internal/application/usecase/config_management_test.go`
- Extended config infrastructure with:
  - OS-aware path resolver (`ResolvePathForOS`) using:
    - macOS/Linux: `~/.config/dbc/config.toml`
    - Windows: `%APPDATA%\\dbc\\config.toml` (fallback `%USERPROFILE%\\AppData\\Roaming\\dbc\\config.toml`)
  - persistent config store (`NewStore`) with safe temp-file write + rename for create/update/delete
  - active path lookup via store
- Added infrastructure tests for path resolver and CRUD persistence:
  - `internal/infrastructure/config/config_test.go`
- Kept startup flow working by loading configured databases through new use-case/port adapter in:
  - `cmd/dbc/main.go`
- Updated product/technical documentation to reflect path behavior and startup flow changes:
  - `docs/product-documentation.md`
  - `docs/technical-documentation.md`

Verification run:
- `go test ./...` passed
- `golangci-lint run ./...` passed
