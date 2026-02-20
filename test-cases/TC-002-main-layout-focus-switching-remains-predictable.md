# Main Layout Focus Switching Remains Predictable

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-002` |
| Functional Behavior Reference | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: runtime two-panel focus model and keyboard-driven focus switches between left and right panels.
- Expected result: layout remains stable and active panel indication follows focus-change shortcuts deterministically.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | Main two-panel runtime view opens with active panel indication. | `A1` |
| S2 | Press `Ctrl+w l`. | Focus moves to right panel and active-panel indicator updates. | `A2` |
| S3 | Press `Ctrl+w h`. | Focus moves back to left panel and active-panel indicator updates. | `A3` |
| S4 | Press `Ctrl+w w`. | Focus cycles to the other panel while layout remains two-panel. | `A4` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | Startup lands in stable two-panel runtime layout with clear active-panel indicator. | `PASS` | Left panel and right panel are visible simultaneously and one panel is visually marked active. |
| A2 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | `Ctrl+w l` moves focus to right panel and updates active-state indicator. | `PASS` | Right panel becomes active after shortcut. |
| A3 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | `Ctrl+w h` moves focus back to left panel and updates active-state indicator. | `PASS` | Left panel becomes active after shortcut. |
| A4 | `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | `Ctrl+w w` cycles focus to the opposite panel without breaking two-panel layout. | `PASS` | Focus switches panel and both panels remain present. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally limited to layout/focus ownership and excludes startup-access assertions.`

## 7. Cleanup

1. Exit app using `q`.
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
