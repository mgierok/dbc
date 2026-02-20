# Overview

Align user-facing and technical documentation with startup CLI help/version and exit-code behavior changes while referencing standards documentation without duplicating normative text.

## Metadata

- Status: DONE
- PRD: PRD-003-cli-help-version-and-startup-cli-standards.md
- Task ID: 05
- Task File: PRD-003-TASK-05-documentation-standards-alignment.md
- PRD Requirements: FR-008, NFR-004

## Objective

Update documentation so startup CLI behavior remains accurate, coherent, and standards-referenced.

## Working Software Checkpoint

After this task, documentation matches shipped startup behavior and users can reliably discover help/version and exit-code expectations without implementation drift.

## Technical Scope

### In Scope

- Update startup CLI behavior sections in `README.md`, product docs, and technical docs for help/version and exit-code standards.
- Add and verify references to `docs/cli-parameter-and-output-standards.md` where relevant.
- Keep product and technical documentation complementary, avoiding duplicated standards prose.
- Add/update documentation-related verification checks used by PRD closure workflow.

### Out of Scope

- Broad documentation restructuring unrelated to PRD-003 behavior.
- Introducing new standards not already defined in `docs/cli-parameter-and-output-standards.md`.
- Non-startup documentation cleanup unrelated to scoped changes.

## Implementation Plan

1. Identify startup behavior sections in `README.md`, `docs/product-documentation.md`, and `docs/technical-documentation.md` affected by PRD-003.
2. Update behavior statements for help/version contracts, informational short-circuit semantics, and exit-code mapping.
3. Add explicit cross-references to `docs/cli-parameter-and-output-standards.md` without duplicating standards text.
4. Verify terminology consistency across product and technical documents.
5. Add/update documentation checks for PRD closure evidence.

## Verification Plan

- FR-008 happy-path check: confirm updated docs reference `docs/cli-parameter-and-output-standards.md` in relevant startup sections.
- FR-008 negative-path check: confirm no section duplicates standards text verbatim beyond concise contextual references.
- NFR-004 check: confirm doc updates preserve existing direct-launch value and do not introduce contradictory startup guidance.

## Acceptance Criteria

- Product and technical documentation are updated for scoped startup CLI behavior changes.
- Documentation includes relevant references to `docs/cli-parameter-and-output-standards.md`.
- Documentation remains complementary and avoids standards-text duplication.
- User-facing startup guidance in `README.md` reflects current implementation behavior.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-003-TASK-02-startup-help-output-contract](.tasks/PRD-003-TASK-02-startup-help-output-contract.md), [PRD-003-TASK-03-startup-version-output-contract](.tasks/PRD-003-TASK-03-startup-version-output-contract.md), [PRD-003-TASK-04-startup-exit-code-and-usage-error-standardization](.tasks/PRD-003-TASK-04-startup-exit-code-and-usage-error-standardization.md)
- blocks: [PRD-003-TASK-06-integration-hardening](.tasks/PRD-003-TASK-06-integration-hardening.md)

## Completion Summary

- Updated startup CLI behavior documentation in:
  - `docs/product-documentation.md` (startup behavior now explicitly documents usage-error/runtime exit-code mapping and references startup CLI standards),
  - `docs/technical-documentation.md` (startup flow and operational error handling now document usage/runtime exit-code classification and reference startup CLI standards).
- Verified standards-reference coverage for startup sections (`docs/cli-parameter-and-output-standards.md`) and preserved non-duplicative wording by adding concise references instead of copying standards prose.
- Preserved direct-launch startup guidance and selector-first behavior statements with no contradictory changes.
- Reviewed `README.md` startup guidance and left it unchanged because current content is still behavior-accurate and user requested to skip non-essential README wording changes.
- Verification executed and passed:
  - `go test ./...`
  - `golangci-lint run ./...`
