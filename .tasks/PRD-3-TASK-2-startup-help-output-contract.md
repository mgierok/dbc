# Overview

Define and implement deterministic startup help output so users can discover supported startup options and examples consistently from the CLI.

## Metadata

- Status: READY
- PRD: PRD-3-cli-help-version-and-startup-cli-standards.md
- Task ID: 2
- Task File: PRD-3-TASK-2-startup-help-output-contract.md
- PRD Requirements: FR-001, FR-002, NFR-001

## Objective

Implement the full `--help`/`-h` output contract for startup usage discoverability.

## Working Software Checkpoint

After this task, help output is available and stable via both help aliases with exit code `0`, while other startup behavior remains functional.

## Technical Scope

### In Scope

- Build deterministic help output payload for startup CLI surface.
- Ensure help output includes startup description, canonical usage, documented aliases, and practical examples.
- Ensure both `--help` and `-h` resolve to equivalent output.
- Add tests that lock help output structure and required content.

### Out of Scope

- Version output token generation.
- Runtime-vs-usage exit-code mapping beyond help success behavior.
- Non-startup command-mode help expansion.

## Implementation Plan

1. Create startup help renderer with stable ordering and wording for required sections.
2. Wire help renderer to informational dispatch result for `--help`/`-h`.
3. Include practical examples for direct launch and version invocation in help output.
4. Ensure output stream and formatting remain deterministic and testable.
5. Add/update tests to validate alias equivalence and required help sections.

## Verification Plan

- FR-001 happy-path check: run `dbc --help` and `dbc -h`; confirm output equivalence and success exit behavior.
- FR-002 happy-path check: confirm help contains startup description, canonical usage line, alias documentation, and at least two practical examples.
- FR-002 negative-path check: remove or alter any required help section token in test fixtures; confirm contract tests fail.
- NFR-001 check: run repeated executions for the same build artifact; confirm help output is deterministic.

## Acceptance Criteria

- `--help` and `-h` produce equivalent deterministic help output.
- Help output contains required startup usage sections and practical examples, including direct launch and version invocation.
- Help output remains stable for the same build artifact.
- Help output does not trigger database open/validation side effects.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-3-TASK-1-startup-informational-dispatch-foundation](.tasks/PRD-3-TASK-1-startup-informational-dispatch-foundation.md)
- blocks: [PRD-3-TASK-5-documentation-standards-alignment](.tasks/PRD-3-TASK-5-documentation-standards-alignment.md), [PRD-3-TASK-6-integration-hardening](.tasks/PRD-3-TASK-6-integration-hardening.md)

## Completion Summary

Not started.
