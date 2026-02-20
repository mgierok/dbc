# Overview

This task establishes the governance and evidence foundation for PRD-005 so every following scenario task can be created and reviewed with consistent coverage, structure, and deterministic result rules.

## Metadata

- Status: READY
- PRD: PRD-005-full-quality-regression-scenarios.md
- Task ID: 01
- Task File: PRD-005-TASK-01-suite-governance-and-coverage-foundation.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-005, NFR-001, NFR-003, NFR-004
- PRD Metrics: M3

## Objective

Create suite-level governance artifacts that define coverage mapping, structure compliance checks, and deterministic result auditing for all PRD-005 regression scenarios.

## Working Software Checkpoint

The repository remains runnable and the existing regression case `TC-001` can be audited against the new governance artifacts without changing application behavior.

## Technical Scope

### In Scope

- Define a suite coverage matrix artifact contract for required journey areas.
- Define a compliance checklist contract for template and metadata conformance.
- Define a deterministic result audit checklist contract for PASS/FAIL-only validation.
- Define evidence recording rules that downstream tasks must follow.

### Out of Scope

- Creating new journey-specific regression scenarios.
- Executing runtime feature or architecture changes.
- Final release-readiness aggregation across the full suite.

## Implementation Plan

1. Create a suite coverage matrix artifact in `test-cases/` with required journey-area mapping fields.
2. Create a structure and metadata compliance checklist aligned to `docs/test-case-template.md` and `docs/test-case-specification.md`.
3. Create a deterministic-result audit checklist enforcing PASS/FAIL-only outcomes and evidence presence.
4. Validate the new governance artifacts against `test-cases/TC-001-direct-launch-opens-main-view.md`.
5. Record governance baseline outcomes and usage notes for downstream tasks.

## Verification Plan

- FR-001 happy-path check: Coverage matrix includes all required journey areas (`startup`, `selector/config`, `runtime/TUI`, `save`, `navigation`) with scenario mapping columns.
- FR-001 negative-path check: Matrix review explicitly fails when any required journey area has no mapped scenario.
- FR-002 happy-path check: Compliance checklist enforces exactly one startup script and exactly one startup command per scenario.
- FR-002 negative-path check: Checklist returns fail when a scenario has zero or multiple startup script bindings.
- FR-003 happy-path check: Structure checklist passes only when section headings and order match `docs/test-case-template.md`.
- FR-003 negative-path check: Structure checklist fails when any required section or required metadata field is missing.
- FR-005 happy-path check: Determinism checklist passes only when assertion and final result states are binary (`PASS`/`FAIL`).
- FR-005 negative-path check: Determinism checklist fails when ambiguous or third-state outcomes appear.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Governance artifacts fully define checks needed to audit 100% scenario compliance.
  - Check procedure: Run governance checks against `TC-001` and record whether all required checks are evaluable and deterministic.

## Acceptance Criteria

1. Suite governance artifacts exist and define coverage, compliance, determinism, and evidence rules required by PRD-005.
2. Governance artifacts can be applied to existing `TC-001` and produce unambiguous pass/fail audit outcomes.
3. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: none
- blocks: [PRD-005-TASK-02-startup-and-selector-config-scenarios](.tasks/PRD-005-TASK-02-startup-and-selector-config-scenarios.md), [PRD-005-TASK-03-runtime-tui-scenarios](.tasks/PRD-005-TASK-03-runtime-tui-scenarios.md), [PRD-005-TASK-04-save-and-navigation-dirty-state-scenarios](.tasks/PRD-005-TASK-04-save-and-navigation-dirty-state-scenarios.md), [PRD-005-TASK-05-integration-hardening](.tasks/PRD-005-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Not started
