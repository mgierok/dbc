# Overview

We believe that replacing panel-switch shortcut combinations with `Enter` and `Esc` transitions for keyboard-first DBC users will improve navigation clarity and reduce context-switch confusion.
We will know this is true when panel transitions are validated as a two-key model with full release-evidence coverage during implementation verification.

# Metadata

- Status: READY

# Problem Statement

Current panel switching behavior depends on `Ctrl+w` combinations (`h`, `l`, `w`), which increases cognitive load in the most common runtime flow: selecting a table, inspecting records, and returning to table selection. Users need a simpler, predictable transition model that keeps right-panel context safety intact.

# Current State (As-Is)

- Panel switching is explicitly supported via `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w`.
- `Enter` is context-dependent and is not the canonical left-to-right panel transition from table selection.
- `Esc` exits local right-panel contexts (for example field focus and popup states), but neutral right-panel `Esc` is not defined as the canonical return to left-panel table selection.
- Existing runtime test cases were authored around the prior panel-switch model and may include implicit dependencies on `Ctrl+w` navigation.

# Target State After Release (To-Be)

- In the left panel, `Enter` confirms the selected table and transitions to the right panel in Records view.
- In right-panel neutral runtime state, `Esc` transitions back to left-panel table selection.
- In right-panel nested contexts, `Esc` remains context-first (exits popup/field-focus before any panel transition).
- `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w` are no longer supported as panel-transition controls.
- Release validation includes an updated `TC-002` and a regression audit of runtime cases `TC-003..TC-008` for hidden dependencies on old panel switching.

# Business Rationale and Strategic Fit

DBC positions itself as a keyboard-first terminal workflow. Simplifying a frequent panel transition path reduces onboarding friction, lowers interaction ambiguity, and improves day-to-day navigation efficiency without broad functional expansion.

# Goals

- G1: Establish an explicit and predictable two-key panel transition model (`Enter` and `Esc`) for runtime navigation.
- G2: Remove dependence on multi-key panel-switch shortcuts for routine left-right panel movement.
- G3: Preserve right-panel context safety by retaining context-first `Esc` behavior in nested states.
- G4: Ensure regression confidence by requiring test-case updates and runtime testcase audit evidence for this navigation shift.

# Non-Goals

- NG1: No changes to startup selector interactions.
- NG2: No changes to filtering semantics, save transaction behavior, or staged data lifecycle rules.
- NG3: No introduction of configurable navigation modes.
- NG4: No redesign of records/schema rendering beyond transition entry behavior.

# Scope (In Scope / Out of Scope)

In Scope:
- Define `Enter` in left panel as table confirmation plus transition to right-panel Records view.
- Define right-panel neutral `Esc` as transition back to left-panel table selection.
- Preserve context-first `Esc` precedence for right-panel nested contexts.
- Remove runtime support for `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w` for panel switching.
- Update `test-cases/TC-002-main-layout-focus-switching-remains-predictable.md` to validate new panel navigation behavior and unsupported `Ctrl+w` transitions.
- Audit runtime test cases `test-cases/TC-003` through `test-cases/TC-008` and update any case with implicit old-navigation dependencies.
- Update `test-cases/suite-coverage-matrix.md` if scenario/assertion mappings change due to testcase updates.

Out of Scope:
- Startup/config selector shortcuts and startup flow behavior.
- Data model changes, schema operations, SQL behavior, and persistence engine behavior.
- Additional navigation shortcuts beyond the explicit `Enter`/`Esc` model.

# Functional Requirements

FR-001:
- When the user is in the left panel and presses `Enter`, the product must confirm the selected table and transition focus to the right panel in Records view for that table.
- Acceptance: Given an active left-panel table selection, pressing `Enter` opens Records view for that table and marks the right panel as active.

FR-002:
- When the user is in right-panel neutral state and presses `Esc`, the product must transition focus to the left panel while preserving the current table selection context.
- Acceptance: Given right-panel neutral state, pressing `Esc` makes the left panel active and the selected table remains visible as current context.

FR-003:
- In right-panel nested contexts, `Esc` must remain context-first and must not immediately switch panels in the same step.
- Acceptance: Given an active popup or field-focus state, pressing `Esc` exits the nested context first; panel transition occurs only from neutral right-panel state.

FR-004:
- The product must discontinue `Ctrl+w h`, `Ctrl+w l`, and `Ctrl+w w` as supported panel-transition controls.
- Acceptance: Runtime panel transition does not occur when these combinations are pressed during release validation scenarios.

FR-005:
- User-visible runtime key guidance must reflect the new panel transition model.
- Acceptance: Release-validation evidence shows runtime key hints and related documentation surfaces list `Enter`/`Esc` transitions and do not list removed `Ctrl+w` panel-switch behavior.

FR-006:
- Release validation must include an updated `TC-002` that verifies the new panel model and context-safe behavior.
- Acceptance: Updated `TC-002` passes assertions for `Enter` left-to-right transition, neutral-state right-panel `Esc` return, and context-first `Esc` precedence in nested right-panel contexts.

FR-007:
- Release validation must include a runtime testcase regression audit for `TC-003..TC-008` to detect and resolve hidden dependencies on removed `Ctrl+w` panel-switch behavior.
- Acceptance: Audit evidence confirms all six runtime cases were reviewed, and any impacted case was updated and passed with deterministic assertions.

# Non-Functional Product Requirements

NFR-001:
- Panel transition behavior must remain deterministic across repeated transitions in one runtime session.

NFR-002:
- The navigation update must not degrade right-panel context safety (popup and field-focus exits must remain predictable).

NFR-003:
- The simplified model must be learnable from in-product cues in a single short usage session.

NFR-004:
- Release sign-off evidence must be auditable through explicit testcase artifacts and coverage mapping updates.

# Success Metrics and Release Criteria

Primary Outcome Metric:
- M1: Panel-transition key model simplification
  - Baseline: Runtime supports 3 dedicated panel-switch shortcuts (`Ctrl+w h`, `Ctrl+w l`, `Ctrl+w w`).
  - Target: Runtime panel transitions use only 2 keys (`Enter` left->right, `Esc` right->left neutral state), and `Ctrl+w` panel-switch shortcuts are unsupported.
  - Measurement window: Implementation verification and final release validation for this PRD.
  - Measurement method: Runtime behavior evidence from automated/manual validation checklist and updated `TC-002` assertions confirming supported/unsupported panel transitions.

Leading Indicators:
- M2: Critical navigation scenario pass coverage
  - Baseline: 0 of 5 target panel-navigation checks pass against the new expected model before implementation.
  - Target: 5 of 5 checks pass (`Enter` transition, neutral `Esc` return, context-first nested `Esc`, unsupported `Ctrl+w`, key-hint consistency).
  - Measurement window: Implementation verification cycle and release validation gate.
  - Measurement method: Scenario checklist artifact and testcase assertion results from updated navigation coverage.

- M3: Canonical transition keystroke count
  - Baseline: 3 panel-transition shortcut combinations are currently documented for direct panel switching.
  - Target: 2 transition keys (`Enter`, `Esc`) are sufficient for the canonical table-select -> records-inspect -> return flow.
  - Measurement window: Release validation scripted usability run.
  - Measurement method: Scripted key-sequence audit artifact for canonical runtime flow.

- M4: Runtime regression-audit completeness for impacted cases
  - Baseline: 0 of 6 runtime cases (`TC-003..TC-008`) audited against the new panel model for hidden dependencies.
  - Target: 6 of 6 runtime cases audited; impacted cases updated and passing with deterministic evidence.
  - Measurement window: Implementation verification and final release validation.
  - Measurement method: Audit checklist artifact plus updated testcase files and `test-cases/suite-coverage-matrix.md` mapping evidence.

Guardrail Metric:
- M5: Context-safe `Esc` retention in nested right-panel states
  - Baseline: 100% of existing nested right-panel contexts use context-first `Esc` behavior.
  - Target: 100% retention of context-first `Esc` behavior after panel-navigation simplification.
  - Measurement window: Implementation verification and release validation.
  - Measurement method: Focused regression assertions in updated `TC-002` (and any impacted runtime cases) proving nested-context exit precedes panel transition.

Release Criteria:
- FR-001 through FR-007 acceptance statements are all validated and evidenced.
- M1 target is achieved exactly.
- M2, M4, and M5 targets are fully met.
- Updated `TC-002` passes and runtime audit for `TC-003..TC-008` is complete with recorded evidence.
- No unresolved regression remains in runtime testcase suite and coverage mapping for functional behavior sections stays consistent.

# Risks and Dependencies

Risks:
- Users accustomed to `Ctrl+w` panel switching may initially attempt removed shortcuts and perceive navigation friction.
- Neutral-state versus nested-state `Esc` handling can regress if context boundaries are implemented inconsistently.
- Runtime testcase suite may contain implicit dependencies on old panel transitions that are not obvious from current assertion text.

Dependencies:
- Reliable right-panel context-state detection is required to preserve context-first `Esc` behavior.
- Runtime key-hint surfaces and related validation documentation must be updated to match the new model.
- Testcase maintenance workflow must complete `TC-002` update and full `TC-003..TC-008` audit with auditable evidence artifacts.

# State & Failure Matrix

| Flow Area | Trigger / Failure Mode | Expected Product Response | User-Visible Recovery Path |
| --- | --- | --- | --- |
| startup | Out of scope for this feature | Startup behavior remains unchanged | Continue existing startup flow; no new action required |
| config | Out of scope for this feature | Config/selector behavior remains unchanged | Continue existing `:config` and selector interaction model |
| save | Out of scope for this feature | Save semantics and dirty-state handling remain unchanged | Continue existing save/discard/cancel flows |
| navigation | Trigger: `Enter` in left panel, `Esc` in right neutral state, `Esc` in nested right-panel state; Failure mode: removed `Ctrl+w` shortcuts attempted | `Enter` transitions to right Records view; neutral right `Esc` returns left; nested `Esc` exits local context first; `Ctrl+w` no longer transitions panels | User uses `Enter` and `Esc` model for panel transitions; when in nested context, press `Esc` to exit context first, then `Esc` again from neutral right panel to return left |

# Assumptions

- A1 (High): Scope is limited to main runtime two-panel interaction and excludes startup selector behavior.
- A2 (Medium): Existing data-operation and filtering behavior remains unchanged unless directly affected by panel transition entry points.
- A3 (Medium): Runtime key-hint and related user-facing references can be updated in the same release scope as behavior changes.
- A4 (Medium): The existing runtime testcase set (`TC-003..TC-008`) is sufficient to detect hidden old-navigation dependencies after mandated audit and updates.
