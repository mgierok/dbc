# Overview

Standardize startup failure signaling so usage errors and runtime failures are clearly separated by exit code and message contract for predictable automation behavior.

## Metadata

- Status: DONE
- PRD: PRD-3-cli-help-version-and-startup-cli-standards.md
- Task ID: 4
- Task File: PRD-3-TASK-4-startup-exit-code-and-usage-error-standardization.md
- PRD Requirements: FR-006, FR-007, NFR-002, NFR-004

## Objective

Enforce startup exit-code and usage-error messaging contract aligned to CLI standards.

## Working Software Checkpoint

After this task, startup users receive consistent actionable guidance on invalid arguments, and runtime failures keep existing failure semantics with distinct exit-code classes.

## Technical Scope

### In Scope

- Map startup invalid usage/argument validation failures to exit code `2`.
- Keep runtime/operational startup failures mapped to exit code `1`.
- Standardize usage-error message shape with error summary, hint, and usage line.
- Add tests for exit-code partitioning and message contract consistency.

### Out of Scope

- New startup flags beyond current PRD scope.
- Redesign of runtime operational error internals outside startup contract boundary.
- Non-startup command-mode exit-code policies.

## Implementation Plan

1. Introduce explicit startup error classification for usage vs runtime failure paths.
2. Route argument-validation and unsupported-argument failures to usage-error output contract and exit code `2`.
3. Ensure runtime startup failures remain mapped to exit code `1`.
4. Add/update tests for unsupported args, missing values, malformed usage, and runtime failure mapping.
5. Ensure guidance messages remain concise and actionable for terminal-first use.

## Verification Plan

- FR-006 happy-path check: run scoped invalid usage scenarios (unsupported option, missing value, malformed combination); confirm code `2` with actionable guidance and usage line.
- FR-006 negative-path check: confirm no usage-error scenario in scope returns code `1`.
- FR-007 happy-path check: trigger representative runtime/operational startup failure; confirm code `1` is preserved.
- FR-007 negative-path check: confirm informational success and usage-validation failures are not misclassified as runtime failure code `1`.
- NFR-002 check: review startup usage-error output for concise, actionable phrasing in all covered scenarios.

## Acceptance Criteria

- Invalid startup usage and argument-validation failures return exit code `2`.
- Usage-error output includes actionable guidance and usage contract fields.
- Runtime/operational startup failures continue to return exit code `1`.
- Startup failure signaling remains coherent with existing direct-launch and selector-first behavior.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-3-TASK-1-startup-informational-dispatch-foundation](.tasks/PRD-3-TASK-1-startup-informational-dispatch-foundation.md)
- blocks: [PRD-3-TASK-5-documentation-standards-alignment](.tasks/PRD-3-TASK-5-documentation-standards-alignment.md), [PRD-3-TASK-6-integration-hardening](.tasks/PRD-3-TASK-6-integration-hardening.md)

## Completion Summary

- Implemented startup failure classification in `cmd/dbc/main.go` with explicit separation of:
  - usage/argument-validation failures -> exit code `2`,
  - runtime/operational failures -> exit code `1`.
- Introduced `startupUsageError` for startup argument/validation parse paths and routed all parse validation failures through this type.
- Added standardized usage-error output contract for startup argument failures:
  - `Error: ...`
  - `Hint: ...` (with corrective guidance and `dbc --help` reference)
  - `Usage: dbc [options].`
- Added/updated startup tests in `cmd/dbc/main_test.go` to cover:
  - usage-error type classification for representative invalid startup scenarios,
  - exit-code partitioning (`2` usage, `1` runtime),
  - usage-error message contract tokens and actionable guidance.
- Verification executed:
  - `go test ./cmd/dbc -run "(ParseStartupOptions|ClassifyStartupFailure|RunStartupDispatch)"` passed.
  - `go test ./...` passed.
  - `golangci-lint run ./...` passed (`0 issues`).
- Downstream context for next tasks:
  - Task 5 can align documentation language to the implemented startup usage-error contract and exit-code semantics without changing this behavior.
