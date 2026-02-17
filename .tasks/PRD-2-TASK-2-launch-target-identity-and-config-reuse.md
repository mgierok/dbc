# Overview

Define and implement normalized connection identity matching so direct launch reuses configured entries instead of creating duplicates.

## Metadata

- Status: DONE
- PRD: PRD-2-cli-direct-database-launch.md
- Task ID: 2
- Task File: PRD-2-TASK-2-launch-target-identity-and-config-reuse.md
- PRD Requirements: FR-008, NFR-004, NFR-005

## Objective

Implement path-normalized matching and configured-entry reuse for direct-launch targets.

## Working Software Checkpoint

Direct-launch startup remains functional, and when direct target matches configured database identity after normalization, the canonical configured entry is reused.

## Technical Scope

### In Scope

- Add deterministic path-normalization logic for SQLite direct-launch identity comparison.
- Resolve direct-launch target against configured entries using normalized identity.
- Reuse configured entry identity when normalized match exists.
- Suppress duplicate temporary identity for already configured direct-launch targets.
- Add tests for matching and non-matching normalization cases.

### Out of Scope

- Session-scoped selector temporary entry injection.
- Changes to dirty-state `:config` decision flow.
- Changes to config file schema.

## Implementation Plan

1. Add normalization helper for direct-launch and configured-entry path comparison.
2. Add startup resolution logic that maps direct-launch target to existing configured entry when matched.
3. Use resolved configured identity in runtime/selector state when match exists.
4. Add regression tests proving duplicate suppression and stable reuse behavior.

## Verification Plan

- FR-008 happy-path check: provide direct-launch target that normalizes to an existing configured entry; confirm existing entry identity is reused.
- FR-008 negative-path check: provide target that does not normalize to any configured entry; confirm no configured identity reuse occurs.
- NFR-004 check: repeat same normalized match scenario and confirm deterministic identity selection.
- NFR-005 check: confirm selector and direct-launch paths remain coherent with one canonical entry identity when matched.

## Acceptance Criteria

- Normalized-equivalent direct-launch target reuses existing configured entry identity.
- No duplicate temporary selector entry is created when configured match exists.
- Non-matching targets remain distinct from configured entries.
- Matching behavior is deterministic across repeated runs.
- Project validation requirement: affected tests and quality checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-2-TASK-1-cli-arg-parsing-and-fast-fail-startup](.tasks/PRD-2-TASK-1-cli-arg-parsing-and-fast-fail-startup.md)
- blocks: [PRD-2-TASK-3-session-scoped-selector-entry-for-direct-launch](.tasks/PRD-2-TASK-3-session-scoped-selector-entry-for-direct-launch.md), [PRD-2-TASK-4-integration-hardening](.tasks/PRD-2-TASK-4-integration-hardening.md)

## Completion Summary

Implemented normalized direct-launch identity resolution in `cmd/dbc/main.go`:
- Added deterministic SQLite path normalization helper used for identity comparison.
- Added configured-entry lookup for direct-launch startup and reused configured identity on normalized match.
- Kept direct-launch behavior unchanged for non-matching targets.
- Preserved deterministic selection by reusing the first configured match in configuration order.

Added regression coverage in `cmd/dbc/main_test.go` for:
- normalized-match reuse of configured identity,
- non-match fallback to direct identity,
- deterministic first-match behavior when multiple normalized matches exist.

Verification executed:
- `go test ./cmd/dbc` (PASS)
- `go test ./...` (PASS)
- `golangci-lint run ./...` (PASS)
