# Overview

Run end-to-end integration hardening for selector-first and direct-launch startup modes, including regression and cross-flow checks before PRD closure.

## Metadata

- Status: READY
- PRD: PRD-2-cli-direct-database-launch.md
- Task ID: 4
- Task File: PRD-2-TASK-4-integration-hardening.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, NFR-001, NFR-002, NFR-003, NFR-004, NFR-005

## Objective

Verify cross-task interactions and regression safety for direct-launch and selector workflows.

## Working Software Checkpoint

Software remains fully usable through both startup paths with no regressions in switching behavior, safety prompts, and persistence rules.

## Technical Scope

### In Scope

- Add integration-oriented checks across direct-launch startup, selector startup, and in-session `:config` switching.
- Validate cross-task interactions and regression safety.
- Validate direct-launch failure behavior, messaging clarity, and non-zero exit behavior.
- Validate temporary-entry lifecycle and non-persistence behavior end-to-end.
- Validate final requirement coverage evidence for PRD-2.

### Out of Scope

- New feature additions beyond PRD-2 requirements.
- Architectural refactors unrelated to PRD-2 behavior.

## Implementation Plan

1. Add integration-level tests for startup mode branching and session switching.
2. Execute end-to-end scenarios covering launch, switch, and relaunch paths.
3. Validate failure and recovery behavior for invalid direct-launch targets.
4. Validate temporary-entry lifecycle and duplicate-suppression behavior.
5. Produce final requirement traceability evidence for PRD closure readiness.

## Verification Plan

- FR-001 happy-path check: both aliases work in integrated startup flow.
- FR-001 negative-path check: malformed/missing direct-launch value fails clearly and exits non-zero.
- FR-002 happy-path check: successful validation gates runtime start.
- FR-002 negative-path check: failed validation blocks runtime start.
- FR-003 happy-path check: error output is readable and actionable.
- FR-003 negative-path check: no ambiguous fallback flow on failure.
- FR-004 happy-path check: direct-launch skips selector on success.
- FR-004 negative-path check: selector-first behavior remains for no direct arg.
- FR-005 happy-path check: `:config` switching works in direct-launched session.
- FR-005 negative-path check: dirty-state decision gate still enforced before selector navigation.
- FR-006 happy-path check: direct-launch target can be reselected during same process and is marked with `⌨`.
- FR-006 negative-path check: no duplicate option when configured identity reuse applies.
- FR-007 happy-path check: temporary entries disappear after restart.
- FR-007 negative-path check: no implicit config persistence from temporary entry visibility.
- Source-marker check: selector consistently marks config entries with `⚙` and CLI entries with `⌨`.
- FR-008 happy-path check: normalized match reuses configured identity.
- FR-008 negative-path check: non-matching path remains separate identity.
- Run full project gates: `go test ./...` and `golangci-lint run ./...`.

## Acceptance Criteria

- Integrated evidence exists for all PRD-2 functional requirements and non-functional requirements.
- Selector-first and direct-launch modes are behaviorally coherent and regression-safe.
- Selector source markers are unified and unambiguous (`⚙` config, `⌨` CLI).
- Temporary-entry lifecycle and dedup behavior meet PRD-2 expectations.
- No unintended config persistence incidents are observed in integrated scenarios.
- Project validation requirement: full verification commands pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-2-TASK-1-cli-arg-parsing-and-fast-fail-startup](.tasks/PRD-2-TASK-1-cli-arg-parsing-and-fast-fail-startup.md), [PRD-2-TASK-2-launch-target-identity-and-config-reuse](.tasks/PRD-2-TASK-2-launch-target-identity-and-config-reuse.md), [PRD-2-TASK-3-session-scoped-selector-entry-for-direct-launch](.tasks/PRD-2-TASK-3-session-scoped-selector-entry-for-direct-launch.md)
- blocks: none

## Completion Summary

Not started.
