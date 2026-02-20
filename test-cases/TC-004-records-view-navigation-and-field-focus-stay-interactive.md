# Records View Navigation and Field Focus Stay Interactive

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-004` |
| Functional Behavior Reference | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: records view entry, row navigation, paging motions, and field-focus transitions for selected table.
- Expected result: records view remains interactive with visible row/cell navigation state throughout records-level and field-focus actions.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | App opens main two-panel runtime view directly (no selector). | `A1` |
| S2 | Press `Enter` in main view to open records for the selected table. | Right panel switches from schema to records view with visible row selection. | `A2` |
| S3 | Press `j` and `k` in records view. | Row selection moves down/up while remaining in records view. | `A3` |
| S4 | Press `Ctrl+f`, then `Ctrl+b`. | Records list pages down/up and remains interactive. | `A4` |
| S5 | Press `Enter` again in records view. | Field-focus mode activates for cell-level navigation. | `A5` |
| S6 | Press `h` and `l` in field-focus mode. | Active cell focus moves left/right within row context. | `A6` |
| S7 | Press `Esc`. | Field-focus mode exits and records-level row navigation context is restored. | `A7` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Startup reaches runtime context that allows opening records for selected table. | `PASS` | Main runtime layout is visible and selected table is ready for records entry. |
| A2 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `Enter` opens records view with visible selected row. | `PASS` | Right panel switches to record rows and highlights active row. |
| A3 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Row navigation keys move selected record while staying in records view. | `PASS` | Selection changes as `j`/`k` are pressed with no context loss. |
| A4 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Paging motions keep records browsing interactive and stable. | `PASS` | Records list responds to page motions and remains navigable. |
| A5 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Field-focus mode can be entered from records view. | `PASS` | Cell-level focus state becomes active after `Enter`. |
| A6 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Field-focus left/right navigation works at cell level. | `PASS` | Focused cell changes horizontally with `h`/`l`. |
| A7 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `Esc` exits field-focus mode and returns to records-level navigation. | `PASS` | Row-level navigation state is restored and records view remains open. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally limited to records view and navigation ownership.`

## 7. Cleanup

1. Exit app using `q`.
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
