# Overview

This task establishes the governance baseline for PRD-006 so every scenario can be audited by exactly one Functional Behavior area with deterministic and traceable evidence.

## Metadata

- Status: DONE
- PRD: PRD-006-functional-behavior-grouped-test-case-coverage.md
- Task ID: 01
- Task File: PRD-006-TASK-01-functional-area-governance-foundation.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-010, NFR-001, NFR-002, NFR-003
- PRD Metrics: M3, M4

## Objective

Define and align governance contracts, templates, and startup-script catalog entries required to enforce one-area scenario ownership and area-pure assertion auditing.

## Working Software Checkpoint

Existing `TC-001` through `TC-006` remain executable and auditable while new governance rules are introduced for PRD-006.

## Technical Scope

### In Scope

- Update test-case governance contracts to require exactly one Functional Behavior area declaration (`4.1` to `4.8`) per scenario.
- Define area-purity audit rules for scenario assertions.
- Add startup informational script support to cover `help` and `version` behavior using approved startup-script binding.
- Update suite governance artifacts to support area-to-scenario-to-assertion traceability.
- Introduce expand-first evidence contract for `TC-001` through `TC-006` refactor tracking.

### Out of Scope

- Refactoring existing scenario files.
- Adding new `TC-*` scenario files.
- Final release-readiness integration decision.

## Implementation Plan

1. Update `docs/test-case-specification.md` with mandatory Functional Behavior ownership and assertion-purity audit requirements.
2. Update `docs/test-case-template.md` so each scenario explicitly declares one Functional Behavior area in a template-prescribed location.
3. Add one approved startup script for informational startup behavior and register it in the startup scripts catalog.
4. Update governance artifacts in `test-cases/` (`suite-coverage-matrix`, `scenario-structure-and-metadata-checklist`, `deterministic-result-audit-checklist`, `full-suite-release-readiness-audit`) for PRD-006 contracts.
5. Add baseline mapping/evidence structure for expand-first tracking and PRD-006 metric checkpoints.

## Verification Plan

- FR-001 happy-path check: Governance contracts require exactly one declared Functional Behavior area for every active `TC-*` case.
- FR-001 negative-path check: Audit contract marks a scenario as FAIL when zero or multiple Functional Behavior areas are declared.
- FR-002 happy-path check: Governance contracts define a deterministic area-purity pass condition for assertion ownership.
- FR-002 negative-path check: Audit contract marks a scenario as FAIL when assertion IDs map to more than one Functional Behavior area.
- FR-003 happy-path check: Governance contract enforces explicit expand-first evidence before new `TC-*` creation.
- FR-003 negative-path check: Audit contract marks additions as FAIL when new scenario creation occurs without expand-first evidence.
- FR-004 happy-path check: Coverage matrix contract supports `4.1` to `4.8` mapping with scenario and assertion traceability fields.
- FR-004 negative-path check: Coverage matrix contract fails when any area row has missing scenario or assertion mapping.
- FR-005 happy-path check: Startup catalog includes an approved informational startup script and command mapping.
- FR-005 negative-path check: Informational startup coverage fails audit when script binding is missing or non-catalog.
- FR-010 happy-path check: All required governance artifacts are updated in one synchronized change set.
- FR-010 negative-path check: Cross-artifact mismatch rules are defined and produce audit FAIL.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Expand-first evidence model is fully defined and usable for all subsequent PRD-006 tasks.
  - Check procedure: Validate that release-readiness audit includes explicit expanded-vs-new classification fields and ratio calculation inputs.
- Metric checkpoint (M4):
  - Metric ID: M4
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Governance contracts define zero-tolerance determinism violations with binary PASS/FAIL rules.
  - Check procedure: Validate deterministic-result checklist and release audit include explicit violation-count field and FAIL triggers.

## Acceptance Criteria

1. Governance and template contracts enforce one-area ownership and area-purity auditing for scenarios.
2. Startup informational behavior coverage is script-bindable through approved startup catalog contract.
3. PRD-006 governance artifacts are synchronized and auditable.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: none
- blocks: [PRD-006-TASK-02-refactor-existing-scenarios-areas-41-44](.tasks/PRD-006-TASK-02-refactor-existing-scenarios-areas-41-44.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

- Updated governance contracts in `docs/test-case-specification.md` to require exactly one markdown-based Functional Behavior reference per scenario and assertion-purity equality checks.
- Updated `docs/test-case-template.md` so metadata and assertion tables include `Functional Behavior Reference`, using Product Documentation as the single source of truth for selectable areas.
- Added informational startup script support by introducing `scripts/start-informational.sh` (`help`/`version`) and registering it in startup scripts catalog.
- Updated suite governance artifacts:
  - `test-cases/suite-coverage-matrix.md` now provides Functional Behavior reference -> scenario -> assertion traceability.
  - `test-cases/scenario-structure-and-metadata-checklist.md` now audits one-reference ownership and reference equality.
  - `test-cases/deterministic-result-audit-checklist.md` now includes ownership purity checks and explicit violation-count contract.
  - `test-cases/full-suite-release-readiness-audit.md` now includes cross-artifact mismatch fail policy and expand-first evidence table model with ratio inputs.
- Verification evidence:
  - `go test ./...` passed.
  - `golangci-lint run ./...` passed.
