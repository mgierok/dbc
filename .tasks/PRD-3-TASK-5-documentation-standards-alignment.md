# Overview

Align user-facing and technical documentation with startup CLI help/version and exit-code behavior changes while referencing standards documentation without duplicating normative text.

## Metadata

- Status: READY
- PRD: PRD-3-cli-help-version-and-startup-cli-standards.md
- Task ID: 5
- Task File: PRD-3-TASK-5-documentation-standards-alignment.md
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

- Broad documentation restructuring unrelated to PRD-3 behavior.
- Introducing new standards not already defined in `docs/cli-parameter-and-output-standards.md`.
- Non-startup documentation cleanup unrelated to scoped changes.

## Implementation Plan

1. Identify startup behavior sections in `README.md`, `docs/product-documentation.md`, and `docs/technical-documentation.md` affected by PRD-3.
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

- blocked-by: [PRD-3-TASK-2-startup-help-output-contract](.tasks/PRD-3-TASK-2-startup-help-output-contract.md), [PRD-3-TASK-3-startup-version-output-contract](.tasks/PRD-3-TASK-3-startup-version-output-contract.md), [PRD-3-TASK-4-startup-exit-code-and-usage-error-standardization](.tasks/PRD-3-TASK-4-startup-exit-code-and-usage-error-standardization.md)
- blocks: [PRD-3-TASK-6-integration-hardening](.tasks/PRD-3-TASK-6-integration-hardening.md)

## Completion Summary

Not started.
