# Visual State Indicators Remain Visible During Staged Changes

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-008` |
| Functional Behavior Reference | `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: visual state communication for mode/status indicators, row markers, and edited-cell marker during staged write activity.
- Expected result: required visual indicators are shown deterministically as staged changes are created in records view.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | Main runtime view opens with no staged changes. | `A1` |
| S2 | Select table `categories`, press `Enter` to open records view, then inspect status line. | Status line shows clean-mode visual state for current table/view context. | `A2` |
| S3 | Enter field-focus mode and edit one existing cell value; confirm edit popup. | Edited value is visibly marked and mode indicator switches to dirty write mode. | `A3` |
| S4 | Press `i` to stage a new row. | Pending insert row is visibly marked with `[INS]`. | `A4` |
| S5 | Move to a persisted row and press `d`. | Selected persisted row is visibly marked with `[DEL]`. | `A5` |
| S6 | Inspect status line again after staged operations. | Status line still shows contextual mode/view/table/filter/hint communication in dirty state. | `A6` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` | Initial runtime visual state shows clean mode before staging. | `PASS` | Mode indicator displays `READ-ONLY` immediately after startup. |
| A2 | `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` | Status line communicates runtime context in records workflow. | `PASS` | Status line includes current view and selected table context with shortcut hints. |
| A3 | `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` | Edited-cell marker and dirty mode indicator appear after staged edit. | `PASS` | Edited cell shows `*` marker and mode indicator changes to `WRITE (dirty: N)`. |
| A4 | `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` | Pending insert rows are visibly marked. | `PASS` | Newly staged row is rendered with `[INS]` marker. |
| A5 | `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` | Pending delete rows are visibly marked for persisted records. | `PASS` | Selected persisted row is rendered with `[DEL]` marker after `d`. |
| A6 | `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` | Status line remains informative in dirty state, including mode and context communication. | `PASS` | Status line still communicates mode plus current view/table/filter and shortcut hints while staged markers are visible. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally constrained to area 4.8 visual communication indicators.`

## 7. Cleanup

1. Exit app using `q`.
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
