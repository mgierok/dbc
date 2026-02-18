# Overview

Create the canonical repository fixture database and its explicit coverage contract so local validation can run on deterministic, lightweight SQLite data.

## Metadata

- Status: READY
- PRD: PRD-4-agent-testability-tmp-startup-fixture.md
- Task ID: 1
- Task File: PRD-4-TASK-1-fixture-foundation-and-coverage-contract.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, NFR-002
- PRD Metrics: M2

## Objective

Deliver `docs/test.db` with documented relational and edge-case coverage plus explicit small-fixture thresholds.

## Working Software Checkpoint

After this task, the application startup and runtime behavior remain unchanged, and a deterministic fixture is available for local validation workflows.

## Technical Scope

### In Scope

- Add canonical fixture file `docs/test.db`.
- Seed related tables with coherent cross-table records.
- Define and document required edge-case categories in fixture contract.
- Define and document small-fixture thresholds for file size and row volume.

### Out of Scope

- Any runtime behavior changes.
- New CLI flags or startup modes.
- Automated TUI-driving test framework changes.

## Implementation Plan

1. Create `docs/test.db` with a compact relational schema and deterministic seed data.
2. Ensure fixture data includes required edge-case categories (null/default/not-null/unique/foreign-key/check, empty values, long values, and varied SQLite types).
3. Document relation map and edge-case coverage contract in `docs/`.
4. Define fixture limits for this PRD: file size `<= 1 MiB`, total rows `<= 300`, max rows per table `<= 120`.
5. Verify fixture integrity and thresholds before task closure.

## Verification Plan

- FR-001 happy-path check: confirm `docs/test.db` exists and is explicitly identified in docs as canonical fixture.
- FR-001 negative-path check: rename or remove fixture path in local check; confirm fixture-presence validation fails.
- FR-002 happy-path check: run FK integrity check (`PRAGMA foreign_key_check;`) and confirm no violations.
- FR-002 negative-path check: run relation-consistency checklist against intentionally invalid sample expectation and confirm mismatch is detected.
- FR-003 happy-path check: run edge-case checklist and confirm every required category is present.
- FR-003 negative-path check: remove one required category from checklist expectation and confirm verification fails.
- FR-004 happy-path check: verify thresholds are met (`<= 1 MiB`, `<= 300` total rows, `<= 120` rows per table).
- FR-004 negative-path check: execute threshold check with stricter temporary limit and confirm failure is reported.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file with checklist outcomes.
  - Threshold/expected value: all required fixture coverage categories pass and threshold limits pass.
  - Check procedure: execute fixture coverage and threshold checks, then record concrete pass/fail evidence in `Completion Summary`.

## Acceptance Criteria

- `docs/test.db` exists and is defined as the canonical fixture.
- Fixture data includes relational links and coherent cross-table records.
- Fixture edge-case coverage contract is documented and fully satisfied.
- Fixture size and row thresholds are explicitly documented and verified as passing.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: none
- blocks: [PRD-4-TASK-2-tmp-startup-playbook-variants](.tasks/PRD-4-TASK-2-tmp-startup-playbook-variants.md), [PRD-4-TASK-3-manual-scenario-reproducibility](.tasks/PRD-4-TASK-3-manual-scenario-reproducibility.md), [PRD-4-TASK-4-integration-hardening](.tasks/PRD-4-TASK-4-integration-hardening.md)

## Completion Summary

Not started.
