# Overview

Define one standardized manual validation scenario tied to the canonical fixture and one startup method so functional checks can be reproduced consistently.

## Metadata

- Status: DONE
- PRD: PRD-004-agent-testability-tmp-startup-fixture.md
- Task ID: 03
- Task File: PRD-004-TASK-03-manual-scenario-reproducibility.md
- PRD Requirements: FR-009, NFR-001, NFR-004
- PRD Metrics: M3

## Objective

Publish one end-to-end manual scenario with explicit steps, expected observations, and pass/fail criteria using fixture-backed startup.

## Working Software Checkpoint

After this task, application behavior remains unchanged, and a reusable manual scenario is available to validate expected navigation/inspection outcomes.

## Technical Scope

### In Scope

- Add one documented manual scenario under `docs/` that uses `docs/test.db`.
- Bind the scenario to one defined startup path from Task 2.
- Include explicit step list, expected observations, and pass/fail criteria.
- Include reproducibility notes for reruns in tmp context.

### Out of Scope

- Multiple scenario suites.
- Automated interactive test harness implementation.
- Runtime feature changes.

## Implementation Plan

1. Select one startup variant from Task 2 as the scenario entry path.
2. Define deterministic preconditions using `docs/test.db` and tmp environment setup.
3. Write ordered scenario steps covering startup, navigation checkpoints, and expected data observations.
4. Add explicit pass/fail decision points and failure-reporting notes.
5. Execute the scenario once end-to-end and record evidence.

## Verification Plan

- FR-009 happy-path check: execute the full scenario and confirm every expected observation is met with final pass result.
- FR-009 negative-path check: alter one expected observation to an incorrect value and confirm scenario validation flags failure.
- NFR-001 check: have an operator follow the scenario using only documented instructions; confirm no hidden setup knowledge is required.
- NFR-004 check: copy-paste all listed commands and confirm they execute without manual rewriting.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file containing execution notes for one full run.
  - Threshold/expected value: one standardized manual scenario executed end-to-end with expected outcomes.
  - Check procedure: run the documented scenario and record command path, observed checkpoints, and final pass/fail in `Completion Summary`.

## Acceptance Criteria

- One manual scenario is documented with deterministic setup, steps, expected observations, and pass/fail criteria.
- Scenario is explicitly linked to canonical fixture data and one startup method.
- Scenario can be rerun in tmp context with consistent results.
- Scenario failure mode is explicit and actionable.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-004-TASK-01-fixture-foundation-and-coverage-contract](.tasks/PRD-004-TASK-01-fixture-foundation-and-coverage-contract.md), [PRD-004-TASK-02-tmp-startup-playbook-variants](.tasks/PRD-004-TASK-02-tmp-startup-playbook-variants.md)
- blocks: [PRD-004-TASK-04-integration-hardening](.tasks/PRD-004-TASK-04-integration-hardening.md)

## Completion Summary

Completed deliverables:

1. Updated `docs/test-fixture.md` with one standardized manual validation scenario bound to startup `Variant 1: Direct Launch via -d`.
2. Added deterministic execution contract for the scenario:
   - precondition reference to `Shared Tmp Bootstrap`,
   - ordered step list,
   - explicit expected observations,
   - explicit pass/fail criteria,
   - actionable failure-reporting and rerun notes.

Verification evidence (FR/NFR + M3):

- FR-009:
  - happy: executed `HOME="$TMP_HOME" "$DBC_BIN" -d "$TMP_DB"` in tmp context; observed direct main-view startup with fixture table list (`categories`, `customers`, `order_items`, `orders`, `products`), then `j` + `Enter` switched to `customers` records view showing rows for `alice@example.com`, `bob@example.com`, `charlie@example.com`; `q` exited with status `0`.
  - negative: intentionally altered one expected observation (`dave@example.com` instead of `charlie@example.com`) in scenario validation check; validation failed as expected (`actual_emails=alice@example.com|bob@example.com|charlie@example.com`).
- NFR-001:
  - scenario run used only instructions documented in `docs/test-fixture.md` (`Shared Tmp Bootstrap` + scenario steps) without hidden setup steps.
- NFR-004:
  - scenario command path (`HOME="$TMP_HOME" "$DBC_BIN" -d "$TMP_DB"`) executed copy-paste without manual rewriting.

Metric checkpoint:

- M3 PASS: one standardized manual scenario was executed end-to-end with expected outcomes and explicit failure-mode validation.

Project validation:

- `golangci-lint run ./...` -> `0 issues.`
- `go test ./...` -> pass for all packages.
