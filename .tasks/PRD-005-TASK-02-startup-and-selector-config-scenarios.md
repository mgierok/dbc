# Overview

This task delivers manual regression scenarios for startup and selector/config management journeys, including user-visible failure and recovery behavior required for release-safe coverage.

## Metadata

- Status: READY
- PRD: PRD-005-full-quality-regression-scenarios.md
- Task ID: 02
- Task File: PRD-005-TASK-02-startup-and-selector-config-scenarios.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-006, FR-007, NFR-001, NFR-002
- PRD Metrics: none

## Objective

Create startup and selector/config regression scenarios that are template-compliant, deterministic, and include explicit failure/recovery validation.

## Working Software Checkpoint

Startup and selector/config workflows are manually verifiable through script-bound scenarios with deterministic assertions and auditable evidence fields.

## Technical Scope

### In Scope

- Add or update `TC-*` files covering startup contracts and selector/config flows.
- Enforce exactly one startup script binding per scenario from the approved startup catalog.
- Include failure and recovery validation for startup misuse or config entry issues.
- Align all scenario structure and metadata to `docs/test-case-template.md`.

### Out of Scope

- Runtime interaction flows that are not startup or selector/config related.
- Save failure and dirty-state navigation decision scenarios.
- Suite-level metric aggregation and release decision output.

## Implementation Plan

1. Define startup journey scenarios covering successful and failure startup contracts.
2. Define selector/config management scenarios covering add/edit/select failure and recovery flows.
3. Bind each scenario to one startup script and one exact startup command.
4. Populate step tables with one user action, one expected outcome, and one assertion ID per step.
5. Run governance checks from TASK-01 and adjust scenarios until all checks pass for this scope.

## Verification Plan

- FR-001 happy-path check: Startup and selector/config journey areas are mapped to one or more `TC-*` cases in the coverage matrix.
- FR-001 negative-path check: Matrix review fails if startup or selector/config area is left unmapped.
- FR-002 happy-path check: Every scenario metadata block contains exactly one valid startup script and one exact command.
- FR-002 negative-path check: Any scenario with missing or duplicate script binding is marked FAIL by the compliance checklist.
- FR-003 happy-path check: All scenario files pass section heading/order checks against `docs/test-case-template.md`.
- FR-003 negative-path check: Any missing required metadata field fails structure audit.
- FR-004 happy-path check: Every step row contains exactly one action, one expected outcome, and one assertion ID.
- FR-004 negative-path check: Any malformed step mapping fails step-level audit.
- FR-006 happy-path check: At least one startup failure/recovery scenario and one selector/config failure/recovery scenario are present.
- FR-006 negative-path check: Missing user-visible recovery path in either journey causes audit failure.
- FR-007 happy-path check: Scenarios include multiple context-relevant assertions and avoid low-value fragmentation.
- FR-007 negative-path check: Quality review flags and fails redundant single-assertion splits.
- Metric checkpoints: none in this task (`PRD Metrics: none`).

## Acceptance Criteria

1. Startup and selector/config scenarios are present, template-compliant, and script-bound with exact commands.
2. Failure and recovery behavior is explicitly validated for startup and selector/config journeys.
3. Governance checks from TASK-01 pass for all scenario files created or updated in this task.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-005-TASK-01-suite-governance-and-coverage-foundation](.tasks/PRD-005-TASK-01-suite-governance-and-coverage-foundation.md)
- blocks: [PRD-005-TASK-05-integration-hardening](.tasks/PRD-005-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Not started
