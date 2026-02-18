# Overview

Run final integration hardening across fixture, startup playbooks, and manual scenario artifacts to confirm cross-task consistency and no startup regressions before PRD closure.

## Metadata

- Status: DONE
- PRD: PRD-4-agent-testability-tmp-startup-fixture.md
- Task ID: 4
- Task File: PRD-4-TASK-4-integration-hardening.md
- PRD Requirements: NFR-003
- PRD Metrics: M4

## Objective

Validate that PRD-4 deliverables work together coherently and do not introduce regressions in existing startup behaviors.

## Working Software Checkpoint

After this task, documented workflows remain executable, existing startup behavior is preserved, and release-readiness evidence is consolidated.

## Technical Scope

### In Scope

- Cross-check consistency between fixture docs, startup playbooks, and manual scenario.
- Re-verify startup behaviors for direct launch, config-driven launch, and no-parameter launch.
- Validate that deliverables from Tasks 1-3 satisfy PRD traceability without behavior expansion.
- Consolidate final release-readiness evidence for PRD-4.

### Out of Scope

- New feature development.
- Refactors unrelated to PRD-4 scope.
- Expanding supported database engines or startup modes.

## Implementation Plan

1. Review outputs from Tasks 1-3 for internal consistency and broken references.
2. Re-run startup verification for the three startup paths covered by PRD-4 documentation.
3. Confirm no runtime behavior changes were introduced beyond fixture/documentation assets.
4. Validate complete FR/NFR/M traceability closure for PRD-4.
5. Record final hardening and regression evidence for closure decision.

## Verification Plan

- Cross-task integration check: verify all PRD-4 task outputs reference consistent fixture path, startup commands, and scenario expectations.
- Regression check for `-d` startup: run startup using valid and invalid direct-launch targets and confirm expected existing behavior.
- Regression check for config-file startup: run startup with valid and invalid config entries and confirm expected existing behavior.
- Regression check for no-parameter startup: run startup with and without configured DB and confirm expected existing behavior.
- NFR-003 check: confirm user-visible startup/runtime behavior remains unchanged by comparing observed results against current product/technical docs.
- Metric checkpoint (M4):
  - Metric ID: M4
  - Evidence source artifact: `Completion Summary` entry in this task file with startup regression verification outcomes.
  - Threshold/expected value: `0` startup regressions introduced by PRD-4 changes.
  - Check procedure: execute regression checks for all startup paths and record observed outcomes and regression count in `Completion Summary`.

## Acceptance Criteria

- Outputs from Tasks 1-3 are internally consistent and executable as documented.
- Regression verification for all relevant startup paths is completed with zero regressions.
- PRD-4 traceability for requirements and metrics is complete and auditable.
- Final release-readiness hardening evidence is recorded for PRD closure.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-4-TASK-1-fixture-foundation-and-coverage-contract](.tasks/PRD-4-TASK-1-fixture-foundation-and-coverage-contract.md), [PRD-4-TASK-2-tmp-startup-playbook-variants](.tasks/PRD-4-TASK-2-tmp-startup-playbook-variants.md), [PRD-4-TASK-3-manual-scenario-reproducibility](.tasks/PRD-4-TASK-3-manual-scenario-reproducibility.md)
- blocks: none

## Completion Summary

Completed deliverables:

1. Cross-task integration consistency check passed across Task 1-3 outputs and `docs/test-fixture.md`:
   - canonical fixture path `docs/test.db` is consistently referenced,
   - startup playbook variants are complete (`variant_sections=3`, `specific_behavior_sections=3`, `when_to_use_sections=3`),
   - manual scenario remains explicitly bound to `Variant 1: Direct Launch via -d`.
2. Startup regression verification re-run across required paths:
   - `-d` startup:
     - happy: `go test ./cmd/dbc -run 'TestResolveStartupSelection_UsesDirectLaunchWithoutSelectorCall|TestConnectSelectedDatabase_ReturnsDatabaseForExistingReachableConnection'` passed.
     - negative: `HOME="$TMP_HOME" go run ./cmd/dbc -d "$TMP_ROOT/missing.db"` exited with `1` and emitted expected direct-launch guidance (`Cannot start DBC with direct launch target ... retry with -d/--database`).
   - config-file startup:
     - happy: selector/config path behavior remained green via `go test ./cmd/dbc -run 'TestResolveStartupSelection_UsesSelectorWhenDirectLaunchMissing'` and `go test ./internal/interfaces/tui -run 'TestDatabaseSelector_ViewShowsActiveConfigPath|TestDatabaseSelector_EnterSelects'`.
     - negative: malformed config startup check (`name = "fixture` without closing quote) exited with `1` and emitted expected parse failure (`toml: basic strings cannot have new lines`).
   - no-parameter startup:
     - happy/selector-first path remained green via `go test ./cmd/dbc -run 'TestResolveStartupSelection_UsesSelectorWhenDirectLaunchMissing'`.
     - missing-config mandatory-setup behavior remained green via `go test ./internal/interfaces/tui -run 'TestDatabaseSelector_EmptyConfigStartsInForcedSetupForm|TestDatabaseSelector_ForcedSetupEscCancelsStartup'`.
3. NFR-003 and release hardening check passed:
   - no runtime/startup behavior expansion was introduced in this task; verification confirmed existing startup contracts still match `docs/product-documentation.md` and `docs/technical-documentation.md`.

Metric checkpoint:

- M4 PASS: startup regression count is `0` across direct-launch, config-driven, and no-parameter verification paths.

Project validation:

- `golangci-lint run ./...` -> `0 issues.`
- `go test ./...` -> pass for all packages.
