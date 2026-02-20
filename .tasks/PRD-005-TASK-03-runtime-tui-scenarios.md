# Overview

This task defines runtime and TUI interaction regression scenarios so core browsing and interaction behavior is covered with deterministic assertions and explicit failure/recovery checks.

## Metadata

- Status: DONE
- PRD: PRD-005-full-quality-regression-scenarios.md
- Task ID: 03
- Task File: PRD-005-TASK-03-runtime-tui-scenarios.md
- PRD Requirements: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, NFR-002
- PRD Metrics: none

## Objective

Create context-rich runtime/TUI regression scenarios that validate normal behavior and user-visible recovery when runtime interaction failures occur.

## Working Software Checkpoint

Runtime/TUI user journeys are represented by executable manual scenarios with deterministic PASS/FAIL outcomes and auditable evidence.

## Technical Scope

### In Scope

- Add or update runtime/TUI-focused `TC-*` files for table navigation, records view, and interaction behavior.
- Include runtime failure/recovery expectations where user-visible failures can occur.
- Ensure each scenario follows the mandatory template structure and deterministic assertion policy.
- Keep scenarios context-rich with multiple relevant assertions per scenario.

### Out of Scope

- Startup and selector/config management scenarios.
- Save failure and dirty-state `:config` decision workflows.
- Suite-level release gate and metric aggregation output.

## Implementation Plan

1. Define runtime/TUI journey scenarios covering key navigation and interaction flows from current product behavior.
2. Add runtime failure/recovery validation branches where user-visible failures are applicable.
3. Bind each scenario to a single approved startup script and exact startup command.
4. Populate scenario steps and assertions with deterministic, measurable criteria.
5. Run governance checks from TASK-01 and revise files until all checks pass for this task scope.

## Verification Plan

- FR-001 happy-path check: Coverage matrix maps runtime/TUI journey area to one or more runtime scenarios.
- FR-001 negative-path check: Coverage matrix flags runtime/TUI as FAIL if not mapped.
- FR-002 happy-path check: Each runtime scenario has exactly one startup script and one startup command in metadata.
- FR-002 negative-path check: Compliance checklist fails scenarios with invalid startup binding count.
- FR-003 happy-path check: Runtime scenario files match required heading names and order.
- FR-003 negative-path check: Any heading/metadata mismatch fails structure audit.
- FR-004 happy-path check: Each test step row contains one action, one expected outcome, and one assertion ID.
- FR-004 negative-path check: Step-level audit fails if action/outcome/assertion mapping is incomplete.
- FR-005 happy-path check: All assertion pass criteria are concrete and resolvable only to PASS or FAIL.
- FR-005 negative-path check: Determinism audit fails any ambiguous criteria or non-binary result state.
- FR-006 happy-path check: Runtime journey includes explicit failure trigger and user-visible recovery expectation where applicable.
- FR-006 negative-path check: Missing runtime recovery expectation fails coverage audit.
- FR-007 happy-path check: Scenario content is expanded and context-rich with multiple high-value assertions.
- FR-007 negative-path check: Quality review fails redundant fragmented scenarios with low incremental value.
- Metric checkpoints: none in this task (`PRD Metrics: none`).

## Acceptance Criteria

1. Runtime/TUI regression scenarios are created or updated with deterministic assertions and required metadata.
2. Runtime failure/recovery behavior is explicitly covered where applicable.
3. Governance checks from TASK-01 pass for all runtime/TUI scenario files in this task.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-005-TASK-01-suite-governance-and-coverage-foundation](.tasks/PRD-005-TASK-01-suite-governance-and-coverage-foundation.md)
- blocks: [PRD-005-TASK-05-integration-hardening](.tasks/PRD-005-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

Delivered runtime/TUI scenario coverage artifacts:

- Added `test-cases/TC-004-runtime-command-failure-recovery-keeps-session-usable.md` as a context-rich runtime interaction scenario covering records navigation, field-focus transitions, invalid command handling, and continued interaction recovery.
- Updated `test-cases/suite-coverage-matrix.md` to map `runtime/TUI` journey coverage to `TC-004` and mark runtime failure/recovery coverage as explicitly present.

Verification executed against this task verification plan:

- FR-001: coverage matrix now maps `runtime/TUI` to runtime scenario `TC-004`; negative-path condition (runtime row unmapped) is no longer present.
- FR-002: `TC-004` metadata contains exactly one startup script and one startup command from the approved script catalog.
- FR-003: `TC-004` follows required section headings and order (`## 1` through `## 7`) and required template table columns.
- FR-004: each `TC-004` test-step row includes one user action, one expected outcome, and one assertion ID.
- FR-005: assertion pass criteria in `TC-004` are concrete and binary-resolvable; assertion and final-result fields remain constrained to `PASS`/`FAIL`.
- FR-006: runtime failure/recovery is explicit in `TC-004` through invalid command trigger (`:unknown`) and recovery checks that confirm continued runtime interactivity.
- FR-007: `TC-004` is intentionally context-rich (multiple high-value assertions across one coherent flow) rather than fragmented single-assert scenarios.
- Metric checkpoints: none (`PRD Metrics: none`).

Downstream decision context:

- PRD-005 full-suite closure remains intentionally open for `save` and `navigation` journey coverage to be completed in subsequent tasks.
