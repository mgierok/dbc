# Overview

Define and implement startup version introspection so users and automation can query a deterministic single-token build identity from the CLI.

## Metadata

- Status: READY
- PRD: PRD-3-cli-help-version-and-startup-cli-standards.md
- Task ID: 3
- Task File: PRD-3-TASK-3-startup-version-output-contract.md
- PRD Requirements: FR-003, FR-004, NFR-001, NFR-003

## Objective

Implement deterministic `--version`/`-v` output with hash-or-dev fallback behavior.

## Working Software Checkpoint

After this task, version output is available through both aliases and prints a single automation-friendly token while startup core paths remain intact.

## Technical Scope

### In Scope

- Implement startup version value resolution for short commit hash when metadata exists.
- Implement `dev` fallback when revision metadata is unavailable.
- Wire version output to informational dispatch path for `--version` and `-v`.
- Add tests locking alias equivalence and single-token output contract.

### Out of Scope

- Help text rendering and section composition.
- Broader build metadata lifecycle/tooling changes outside startup output behavior.
- Runtime-vs-usage exit-code standardization beyond version success behavior.

## Implementation Plan

1. Add startup version resolver that reads build metadata and derives short hash output.
2. Implement deterministic fallback to `dev` when build metadata is missing.
3. Connect resolver output to informational version dispatch for both aliases.
4. Enforce single-token stdout output contract without additional prose.
5. Add/update tests for alias equivalence, hash resolution, and fallback behavior.

## Verification Plan

- FR-003 happy-path check: run `dbc --version` and `dbc -v`; confirm equivalent output and success exit behavior.
- FR-004 happy-path check: run version path with revision metadata available; confirm short commit hash token is printed.
- FR-004 negative-path check: run version path with revision metadata unavailable; confirm output is exactly `dev`.
- NFR-003 check: confirm version output is single-token stdout only, suitable for shell automation parsing.
- NFR-001 check: confirm repeated execution for the same build artifact returns stable version output.

## Acceptance Criteria

- `--version` and `-v` produce equivalent output and success behavior.
- Version output prints short commit hash when available, otherwise `dev`.
- Version output is a single token on stdout with no extra prose.
- Version output is deterministic for the same build artifact.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-3-TASK-1-startup-informational-dispatch-foundation](.tasks/PRD-3-TASK-1-startup-informational-dispatch-foundation.md)
- blocks: [PRD-3-TASK-5-documentation-standards-alignment](.tasks/PRD-3-TASK-5-documentation-standards-alignment.md), [PRD-3-TASK-6-integration-hardening](.tasks/PRD-3-TASK-6-integration-hardening.md)

## Completion Summary

Not started.
