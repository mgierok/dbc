# Overview

Define and render runtime help popup content with required section structure and keyboard scrolling so users can discover supported commands and keyboard keywords even when content exceeds visible popup height.

## Metadata

- Status: DONE
- PRD: PRD-008-runtime-help-popup-reference.md
- Task ID: 02
- Task File: PRD-008-TASK-02-runtime-help-popup-content-and-scroll.md
- PRD Requirements: FR-002, FR-003, FR-004, NFR-002
- PRD Metrics: M2, M3

## Objective

Implement help popup content composition and overflow scrolling behavior that satisfies required sections and one-line description readability.

## Working Software Checkpoint

After this task, the runtime help popup presents required `Supported Commands` and `Supported Keywords` sections with concise one-line descriptions, and keyboard scrolling can reach final entries when content overflows.

## Technical Scope

### In Scope

- Help popup rendering for `Supported Commands` section.
- Help popup rendering for `Supported Keywords` section.
- One-line descriptions for listed commands and keyboard keywords.
- Keyboard scrolling state and rendering for overflow content within help popup.

### Out of Scope

- Runtime command routing/lifecycle logic for opening or closing help popup.
- Startup/selector help behavior.
- Any redesign unrelated to required help popup content and scroll behavior.

## Implementation Plan

1. Define deterministic help popup content model containing commands and keyboard keywords with one-line descriptions.
2. Implement renderer output with explicit section headers `Supported Commands` and `Supported Keywords`.
3. Add scroll offset and key-handling integration for help popup content overflow.
4. Ensure scrolling allows reaching final listed item while preserving existing popup layout conventions.
5. Update/add focused TUI rendering and interaction tests for section completeness and overflow navigation.

## Verification Plan

- FR-002 happy-path check: Open help popup and verify `Supported Commands` section is visible with one-line descriptions.
- FR-002 negative-path check: Validate test fails when `Supported Commands` section header or descriptions are missing.
- FR-003 happy-path check: Open help popup and verify `Supported Keywords` section is visible with one-line descriptions.
- FR-003 negative-path check: Validate test fails when `Supported Keywords` section header or descriptions are missing.
- FR-004 happy-path check: With overflow content, use keyboard scrolling and verify final item becomes reachable.
- FR-004 negative-path check: Use non-scroll key path and verify scroll offset does not change unexpectedly.
- Metric checkpoint (M2):
  - Metric ID: M2
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: FR coverage checkpoints in this task mapped to FR-002, FR-003, and FR-004 are PASS.
  - Check procedure: Record per-FR assertion results in this task file `Completion Summary` when task is marked DONE.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Required popup sections completeness is 2/2 with one-line description presence PASS.
  - Check procedure: Record section/description assertion outcomes in this task file `Completion Summary` when task is marked DONE.

## Acceptance Criteria

1. Help popup includes `Supported Commands` and `Supported Keywords` sections with one-line descriptions.
2. Keyboard scrolling enables navigation to final help item when content exceeds popup height.
3. Popup content remains scannable with clear section separation and concise line items.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-008-TASK-01-runtime-help-command-and-popup-lifecycle](.tasks/PRD-008-TASK-01-runtime-help-command-and-popup-lifecycle.md)
- blocks: [PRD-008-TASK-03-runtime-help-release-evidence-test-update](.tasks/PRD-008-TASK-03-runtime-help-release-evidence-test-update.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

- Implemented deterministic runtime help content rendering in `internal/interfaces/tui/view.go`:
  - Added required `Supported Commands` and `Supported Keywords` sections.
  - Added concise one-line descriptions for each listed command/keyword entry.
  - Preserved popup layout conventions with section separation and overflow indicator.
- Implemented overflow scrolling behavior for help popup in `internal/interfaces/tui/model.go`:
  - Added help popup scroll offset state.
  - Added keyboard scrolling handlers for `j/k`, `down/up`, `Ctrl+f`/`Ctrl+b`, `g`/`G`, and `home`/`end`.
  - Kept non-scroll key paths stable (no unexpected scroll movement).
- Added focused tests in `internal/interfaces/tui/view_test.go`:
  - `TestRenderHelpPopup_IncludesRequiredSectionsAndOneLineDescriptions` (FR-002/FR-003 happy path) PASS.
  - `TestRenderHelpPopup_ScrollCanReachFinalItemWhenOverflowing` (FR-004 happy path) PASS.
  - `TestHandleHelpPopupKey_NonScrollKeyDoesNotChangeRenderedWindow` (FR-004 negative path) PASS.
- Updated source-of-truth docs for current behavior:
  - `docs/product-documentation.md` now documents required help sections and overflow scrolling interaction.
  - `docs/technical-documentation.md` now documents sectioned help rendering and help-scroll mechanics.
- Verification evidence:
  - `go test ./internal/interfaces/tui -run 'TestRenderHelpPopup_IncludesRequiredSectionsAndOneLineDescriptions|TestRenderHelpPopup_ScrollCanReachFinalItemWhenOverflowing|TestHandleHelpPopupKey_NonScrollKeyDoesNotChangeRenderedWindow'` PASS.
  - `golangci-lint run ./...` PASS (`0 issues`).
  - `go test ./...` PASS.
- Metric checkpoint results:
  - M2 PASS: FR-002, FR-003, and FR-004 checkpoints in this task are PASS.
  - M3 PASS: required popup sections completeness is 2/2 with one-line description presence PASS.
