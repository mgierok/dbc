# Overview

This task synchronizes user-visible key guidance with the new runtime panel-transition model so product cues and documentation match factual behavior.

## Metadata

- Status: DONE
- PRD: PRD-007-simplified-panel-navigation-enter-esc.md
- Task ID: 02
- Task File: PRD-007-TASK-02-runtime-key-guidance-and-doc-sync.md
- PRD Requirements: FR-005, NFR-003, NFR-004
- PRD Metrics: M1, M3

## Objective

Align runtime key hints and documentation surfaces with the `Enter`/`Esc` panel-transition model and remove guidance that suggests `Ctrl+w` panel switching.

## Working Software Checkpoint

After this task, runtime behavior remains unchanged from TASK-01 and users see accurate in-product and repository documentation guidance for panel transitions.

## Technical Scope

### In Scope

- Update runtime key-hint text to reflect new panel transitions.
- Update `README.md` keyboard controls for panel navigation.
- Update `docs/product-documentation.md` interaction model to remove old panel-switch shortcuts and reflect the new model.
- Update `docs/technical-documentation.md` runtime interaction notes that reference panel-switch behavior.

### Out of Scope

- Additional runtime key-behavior implementation changes.
- Scenario testcase content updates in `test-cases/TC-002`.
- Regression-audit edits for `TC-003..TC-008`.

## Implementation Plan

1. Update runtime status shortcut/hint rendering for panel-transition guidance.
2. Remove `Ctrl+w` panel-switch references from user-visible guidance surfaces and replace with `Enter`/`Esc` model descriptions.
3. Update `README.md` keyboard control section to match runtime truth.
4. Update `docs/product-documentation.md` and `docs/technical-documentation.md` where panel-switch behavior is described.
5. Perform consistency scan to ensure no remaining panel-transition guidance conflicts across runtime and docs.

## Verification Plan

- FR-005 happy-path check: Runtime hints and documentation surfaces show `Enter`/`Esc` panel transitions and no longer present removed `Ctrl+w` panel switching as supported behavior.
- FR-005 negative-path check: Repository scan fails if any panel-transition guidance still claims `Ctrl+w h/l/w` is supported.
- NFR-003 happy-path check: A short key-guidance walkthrough allows users to infer the canonical transition path (`Enter` to records, `Esc` to tables).
- NFR-003 negative-path check: Guidance is considered non-learnable if runtime or docs provide conflicting panel-transition instructions.
- Metric checkpoint (M1):
  - Metric ID: M1
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Documented and in-product panel-transition guidance matches runtime support exactly.
  - Check procedure: Compare runtime hint strings and documented shortcuts against implemented key behavior and record evidence references.
- Metric checkpoint (M3):
  - Metric ID: M3
  - Evidence source artifact: `Completion Summary` entry in this task file.
  - Threshold/expected value: Canonical transition guidance requires exactly two keys (`Enter`, `Esc`) and does not include dedicated `Ctrl+w` panel-switch shortcuts.
  - Check procedure: Audit guidance sections and verify only two canonical transition keys are listed for panel movement.

## Acceptance Criteria

1. User-visible runtime key guidance reflects `Enter`/`Esc` transitions and excludes removed `Ctrl+w` panel switching.
2. `README.md`, `docs/product-documentation.md`, and `docs/technical-documentation.md` are synchronized with runtime behavior.
3. No conflicting panel-transition guidance remains across in-product and documentation surfaces.
4. Project validation requirement: affected tests and quality checks pass for the changed scope.

## Dependencies

- blocked-by: [PRD-007-TASK-01-runtime-navigation-contract-enter-esc](.tasks/PRD-007-TASK-01-runtime-navigation-contract-enter-esc.md)
- blocks: [PRD-007-TASK-05-integration-hardening](.tasks/PRD-007-TASK-05-integration-hardening.md)

Format rule:
- Use `none` when there is no dependency.
- When there are dependencies, use Markdown links separated by comma+space on the same line.
- Example: `- blocked-by: [PRD-012-TASK-01-config-foundation](.tasks/PRD-012-TASK-01-config-foundation.md), [PRD-012-TASK-02-schema-migration](.tasks/PRD-012-TASK-02-schema-migration.md)`

## Completion Summary

- Updated runtime shortcut guidance in `internal/interfaces/tui/view.go`:
  - `Tables: Enter records | F filter`
  - `Schema: Esc tables | F filter`
  - `Records: Esc tables | Enter edit | i insert | d delete | u undo | Ctrl+r redo | w save | F filter`
- Updated unit tests in `internal/interfaces/tui/view_test.go`:
  - adjusted existing table/records shortcut expectations,
  - added `TestStatusShortcuts_SchemaPanel` for right-panel neutral guidance.
- Synced user-facing keyboard documentation:
  - `README.md` keyboard controls now describe panel transitions with `Enter`/`Esc` and keep concise runtime wording,
  - `docs/product-documentation.md` updated in:
    - `4.2 Main Layout and Focus Model` (explicit panel-transition rules),
    - `6.2 Global/Main Navigation` (panel transition shortcuts),
    - `6.3 View and Record Actions` (records-context `Enter`/`Esc` actions),
  - `docs/technical-documentation.md` updated in `5.2 Main Read Flow` with factual runtime transition handling in `internal/interfaces/tui/model.go`.
- Verification executed (all pass):
  - `gofmt -w internal/interfaces/tui/view.go internal/interfaces/tui/view_test.go`
  - `go test ./internal/interfaces/tui`
  - `golangci-lint run ./...`
  - `go test ./...`
  - `rg -n 'Ctrl\\+w h|Ctrl\\+w l|Ctrl\\+w w' README.md docs/product-documentation.md docs/technical-documentation.md internal/interfaces/tui || true` (no matches)
- Metric evidence:
  - M1 satisfied: runtime hints and documentation surfaces now describe the same `Enter`/`Esc` transition model.
  - M3 satisfied: canonical panel-transition guidance uses exactly two keys (`Enter`, `Esc`) in the updated guidance surfaces.
