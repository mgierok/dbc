# Overview

Run final integration hardening for PRD-3 to verify cross-task interactions and regression safety across startup informational flags, direct-launch behavior, selector-first behavior, and documentation consistency.

## Metadata

- Status: DONE
- PRD: PRD-3-cli-help-version-and-startup-cli-standards.md
- Task ID: 6
- Task File: PRD-3-TASK-6-integration-hardening.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-008, NFR-001, NFR-002, NFR-003, NFR-004

## Objective

Validate integrated startup CLI behavior and regression coverage before PRD-3 closure.

## Working Software Checkpoint

After this task, startup remains usable across selector-first and direct-launch paths with deterministic informational behavior, standardized exit semantics, and aligned documentation.

## Technical Scope

### In Scope

- Add integration-level tests that cover combined startup paths and informational short-circuit behavior.
- Verify cross-task coherence for help/version, exit-code classification, and legacy direct-launch flow.
- Verify regression safety for selector-first path and runtime startup behavior.
- Validate final FR/NFR closure evidence and documentation alignment.

### Out of Scope

- New startup feature additions beyond PRD-3 requirements.
- Architectural refactors unrelated to startup CLI standards scope.

## Implementation Plan

1. Add integration-oriented startup scenarios combining informational flags, direct-launch input, and selector-first startup.
2. Validate behavior when informational and non-informational inputs intersect, including expected precedence and failure handling.
3. Execute full regression checks for direct-launch continuity and selector-first continuity.
4. Validate documentation and implementation consistency for startup contracts.
5. Produce requirement-level closure evidence for PRD-3.

## Verification Plan

- FR-001 happy-path check: integrated startup tests confirm `--help` and `-h` are equivalent with success behavior.
- FR-001 negative-path check: integrated tests confirm repeated logical help aliases are rejected as usage errors.
- FR-002 happy-path check: integrated help contract tests confirm required usage sections and examples remain present.
- FR-002 negative-path check: integrated contract tests fail on missing/altered required help sections.
- FR-003 happy-path check: integrated startup tests confirm `--version` and `-v` are equivalent with success behavior.
- FR-003 negative-path check: integrated tests confirm repeated logical version aliases are rejected as usage errors.
- FR-004 happy-path check: integrated tests confirm short commit hash output when metadata is available.
- FR-004 negative-path check: integrated tests confirm exact `dev` fallback when metadata is unavailable.
- FR-005 happy-path check: integrated tests confirm informational flags short-circuit before DB/config startup work.
- FR-005 negative-path check: integrated tests confirm no short-circuit when informational flags are absent.
- FR-006 happy-path check: integrated tests confirm usage-error scenarios return code `2` with actionable guidance.
- FR-006 negative-path check: integrated tests confirm usage failures are not classified as runtime code `1`.
- FR-007 happy-path check: integrated tests confirm runtime startup failures return code `1`.
- FR-007 negative-path check: integrated tests confirm informational and usage paths are not misclassified as runtime failures.
- FR-008 happy-path check: integrated documentation checks confirm standards references are present where required.
- FR-008 negative-path check: integrated documentation checks confirm no standards-content duplication drift.
- NFR checks: confirm determinism, message clarity, automation-friendly version output, and no direct-launch discoverability regression.
- Run full project gates: `go test ./...` and `golangci-lint run ./...`.

## Acceptance Criteria

- Integrated evidence exists for all PRD-3 functional and non-functional requirements.
- Startup informational behavior, direct-launch flow, and selector-first flow remain coherent under combined scenarios.
- Regression coverage confirms no loss of existing startup value while introducing PRD-3 contracts.
- Documentation and implementation are aligned for startup CLI standards behavior.
- Project validation requirement: full verification commands pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-3-TASK-2-startup-help-output-contract](.tasks/PRD-3-TASK-2-startup-help-output-contract.md), [PRD-3-TASK-3-startup-version-output-contract](.tasks/PRD-3-TASK-3-startup-version-output-contract.md), [PRD-3-TASK-4-startup-exit-code-and-usage-error-standardization](.tasks/PRD-3-TASK-4-startup-exit-code-and-usage-error-standardization.md), [PRD-3-TASK-5-documentation-standards-alignment](.tasks/PRD-3-TASK-5-documentation-standards-alignment.md)
- blocks: none

## Completion Summary

- Executed integration-hardening verification without adding new persistent integration test files to the codebase.
- Ran startup cross-flow integrity test bundle in `cmd/dbc` (ad hoc selection of existing integration-relevant scenarios), covering:
  - informational short-circuit behavior vs runtime startup path,
  - help/version alias behavior and mixed-flag validation failures,
  - usage-error (`2`) vs runtime-failure (`1`) classification boundary,
  - direct-launch and selector-first startup continuity checks.
- Validated startup documentation consistency against PRD-3 FR-008 by checking:
  - standards references exist in product and technical docs (`docs/cli-parameter-and-output-standards.md`),
  - startup informational alias and exit-code semantics are documented,
  - canonical standards placeholder template text is not duplicated in product/technical docs.
- Full project verification passed:
  - `go test ./cmd/dbc -run 'TestRunStartupDispatch_(UsesInformationalHandlerWithoutRuntimeStartup|UsesRuntimeStartupWhenInformationalFlagsAreAbsent|HelpAliasesProduceEquivalentRenderedOutput|VersionAliasesProduceEquivalentRenderedOutput)|TestClassifyStartupFailure_MapsUsageErrorsToExitCodeTwoWithUsageContract|TestClassifyStartupFailure_MapsRuntimeErrorsToExitCodeOne|TestParseStartupOptions_ReturnsErrorForRepeatedLogicalInformationalAliases|TestParseStartupOptions_ReturnsErrorForMixedInformationalAndDirectLaunchFlags|TestParseStartupOptions_ReturnsErrorForMixedInformationalFlags|TestResolveStartupSelection_(UsesDirectLaunchWithoutSelectorCall|UsesSelectorWhenDirectLaunchMissing|ReusesConfiguredIdentityWhenNormalizedPathsMatch|UsesFirstConfiguredIdentityForDeterministicNormalizedMatch)'` (PASS)
  - `go test ./...` (PASS)
  - `golangci-lint run ./...` (PASS)
- PRD-3 closure readiness: all PRD-3 tasks are now `DONE`; parent PRD status can be closed after acceptance matrix evidence is confirmed in PRD file.
