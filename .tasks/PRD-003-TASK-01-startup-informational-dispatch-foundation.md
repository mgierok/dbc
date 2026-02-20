# Overview

Establish startup informational-flag dispatch for help/version requests so `dbc` can short-circuit runtime initialization deterministically while preserving existing selector and direct-launch startup behavior.

## Metadata

- Status: DONE
- PRD: PRD-003-cli-help-version-and-startup-cli-standards.md
- Task ID: 01
- Task File: PRD-003-TASK-01-startup-informational-dispatch-foundation.md
- PRD Requirements: FR-001, FR-003, FR-005, FR-006, NFR-004

## Objective

Introduce startup informational dispatch and precedence rules that prevent runtime side effects for help/version requests.

## Working Software Checkpoint

After this task, selector-first startup and direct-launch startup still work as before when informational flags are not requested, and startup can route informational requests without opening databases.

## Technical Scope

### In Scope

- Extend startup argument parsing to recognize informational intents for `--help`/`-h` and `--version`/`-v`.
- Define deterministic startup precedence and conflict handling between informational and non-informational startup options.
- Add a startup dispatch branch that can return informational outcomes before config and database startup work begins.
- Add tests for informational dispatch routing and no-side-effect guarantees.

### Out of Scope

- Final help text contract details and examples.
- Version token generation and output formatting details.
- Final exit-code/message standardization for all startup failure classes.

## Implementation Plan

1. Add informational startup options representation alongside existing direct-launch parsing contract.
2. Define parser rules for repeated logically identical informational flags and mixed invalid flag combinations.
3. Add a startup dispatch decision layer that resolves informational command handling before config/database initialization.
4. Ensure no DB/config selector initialization occurs when dispatch resolves to informational output path.
5. Add focused tests covering dispatch routing, precedence, and side-effect blocking.

## Verification Plan

- FR-005 happy-path check: run startup with `--help`, `-h`, `--version`, and `-v`; confirm each path is handled before config/database startup work.
- FR-005 negative-path check: run startup with no informational flag; confirm normal selector/direct-launch startup path is preserved.
- FR-001 negative-path check: run startup with repeated logical help flags (for example `--help -h`); confirm usage-error classification and no runtime startup side effects.
- FR-003 negative-path check: run startup with repeated logical version flags (for example `--version -v`); confirm usage-error classification and no runtime startup side effects.
- FR-006 negative-path check: run startup with unsupported/malformed argument combinations under informational dispatch context; confirm usage-error path is selected.

## Acceptance Criteria

- Startup can classify and route informational requests before config/database startup logic executes.
- Startup dispatch behavior is deterministic for informational flag inputs.
- Repeated logical informational aliases are treated as usage errors according to PRD assumptions.
- Non-informational startup behavior remains unchanged after dispatch integration.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: none
- blocks: [PRD-003-TASK-02-startup-help-output-contract](.tasks/PRD-003-TASK-02-startup-help-output-contract.md), [PRD-003-TASK-03-startup-version-output-contract](.tasks/PRD-003-TASK-03-startup-version-output-contract.md), [PRD-003-TASK-04-startup-exit-code-and-usage-error-standardization](.tasks/PRD-003-TASK-04-startup-exit-code-and-usage-error-standardization.md)

## Completion Summary

- Added informational startup parsing and dispatch foundation in `cmd/dbc/main.go`:
  - recognized `-h`/`--help` and `-v`/`--version`,
  - enforced deterministic conflict rules:
    - repeated logical informational aliases are rejected,
    - help/version combination is rejected,
    - informational and direct-launch flags in one invocation are rejected,
  - introduced `runStartupDispatch` so informational paths return before config-path resolution and runtime/DB initialization.
- Added informational command rendering placeholders for dispatch wiring:
  - help path emits temporary startup-help placeholder text,
  - version path emits `dev` token placeholder.
- Added/updated tests in `cmd/dbc/main_test.go` for:
  - informational alias parsing,
  - repeated logical alias rejection,
  - mixed informational/direct-launch rejection,
  - mixed help/version rejection,
  - dispatch no-side-effect behavior (informational handler called, runtime startup handler skipped).
- Updated documentation to reflect current factual startup behavior:
  - `docs/product-documentation.md`,
  - `docs/technical-documentation.md`,
  - `README.md`.
- Verification executed:
  - `go test ./cmd/dbc` passed during Red/Green cycle.
  - Full-project validation passed:
    - `go test ./...`
    - `golangci-lint run ./...`
- Downstream note:
  - Task 2 and Task 3 should replace placeholder informational outputs with final help/version contracts while reusing the dispatch foundation.
