# Main Layout Focus Switching Remains Predictable

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-002` |
| Functional Behavior Reference | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: runtime two-panel navigation transitions using `Enter` (left -> right) and `Esc` (right -> left) with nested-context safety.
- Expected result: panel transitions are deterministic, nested right-panel contexts consume the first `Esc`, and removed `Ctrl+w` shortcuts do not switch panels.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | Main two-panel runtime view opens with active panel indication, independent left/right panel frames, and a 3-row framed status bar. | `A1` |
| S2 | From left-panel table selection, press `Enter`. | Focus transitions to right panel in Records view for selected table. | `A2` |
| S3 | In neutral right-panel Records state, press `Esc`. | Focus returns to left-panel table selection and right panel switches to Table Discovery (Schema view) for the selected table. | `A3` |
| S4 | Press `Enter` again to return to right-panel Records view, then press `F` to open filter popup. | Nested right-panel context opens (filter popup visible) while right panel remains active. | `A4` |
| S5 | Press `Esc` once with filter popup open. | Popup closes first and runtime remains in right-panel neutral state (no panel switch on first `Esc`). | `A4` |
| S6 | Press `Esc` again from right-panel neutral state. | Focus returns to left-panel table selection and right panel switches to Table Discovery (Schema view). | `A5` |
| S7 | From left panel, press `Ctrl+w l`, `Ctrl+w h`, and `Ctrl+w w`. | None of these shortcuts switches panel focus; left panel remains active. | `A6` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | Startup lands in stable two-panel runtime layout with clear active-panel indicator and independent section framing. | `PASS` | Left panel and right panel are visible simultaneously in separate boxes, status bar is rendered in its own 3-row box, and one panel is visually marked active. |
| A2 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | Pressing `Enter` from left-panel table selection opens the selected table in right-panel Records view and marks right panel as active. | `PASS` | `Enter` from tables switches focus to records content in right panel for the selected table. |
| A3 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | Pressing `Esc` from neutral right-panel Records state returns focus to left panel and forces right-panel Table Discovery for the selected table. | `PASS` | Left panel becomes active after neutral-state `Esc`, right panel renders schema/table-discovery content, and current table context remains consistent. |
| A4 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | In nested right-panel context, first `Esc` exits local context only; it must not switch panels in the same keypress. | `PASS` | Filter popup closes on first `Esc` and focus remains in right-panel runtime context. |
| A5 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | After nested context is closed, `Esc` from neutral right-panel state returns focus to left panel and forces right-panel Table Discovery. | `PASS` | Second `Esc` (from neutral right panel) returns focus to left-panel table selection and right panel shows schema/table-discovery content. |
| A6 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | `Ctrl+w l`, `Ctrl+w h`, and `Ctrl+w w` do not trigger runtime panel transitions. | `PASS` | Executing each removed `Ctrl+w` combination leaves panel ownership unchanged (no transition observed). |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally limited to layout/focus ownership and excludes startup-access assertions.`

## 7. Cleanup

1. Exit app using `:q`.
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
