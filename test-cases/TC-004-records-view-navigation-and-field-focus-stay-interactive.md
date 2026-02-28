# Records View Navigation and Field Focus Stay Interactive

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-004` |
| Functional Behavior Reference | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: records view entry, row navigation, fixed-page pagination motions (`Ctrl+f`/`Ctrl+b`), guided sort flow, field-focus transitions, and single-row detail inspection for selected table.
- Expected result: records view remains interactive with deterministic single-sort behavior, bounded pagination behavior, visible row/cell navigation state, and selected-row detail that closes cleanly back to records list.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | App opens main two-panel runtime view directly (no selector). | `A1` |
| S2 | Select table `products`, then press `Enter` to open records. | Right panel switches from schema to records view with visible row selection. | `A2` |
| S3 | Press `j` and `k` in records view. | Row selection moves down/up while remaining in records view. | `A3` |
| S4 | Press `Ctrl+f`, then `Ctrl+b` in records view for `products`. | Pagination shortcuts stay bounded for single-page data (`Page 1/1`) and records context remains interactive. | `A4` |
| S5 | Press `Shift+S`, choose column `name`, choose direction `ASC`, then confirm. | Sort is applied for selected table and records header marks active sorted column with `↑`. | `A5` |
| S6 | Open sort again with `Shift+S`, choose column `name`, choose direction `DESC`, then confirm. | New sort replaces previous sort so only one active sort remains, and header indicator switches to `↓`. | `A6` |
| S7 | Press `i` to stage a pending insert, then apply sort again with `Shift+S` (for example `id ASC`). | Pending insert row remains at top and is not reordered by SQL sort. | `A7` |
| S8 | Press `d` on the pending insert row. | Pending insert row is removed immediately. | `A8` |
| S9 | Press `Esc` to return to left panel, switch to `categories`, then switch back to `products` and open records again. | Active sort is reset after table switch and previous sort indicator is not carried over. | `A9` |
| S10 | Press `e` in records view. | Field-focus mode activates for cell-level navigation. | `A10` |
| S11 | Press `h` and `l` in field-focus mode. | Active cell focus moves left/right within row context. | `A11` |
| S12 | Press `Esc`. | Field-focus mode exits and records-level row navigation context is restored. | `A12` |
| S13 | Press `Enter` in records view. | Right panel opens selected-row detail in vertical column/value layout without truncating field content. | `A13` |
| S14 | Press `Esc` while row detail is open. | Row detail closes and records list context is restored for continued navigation. | `A14` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Startup reaches runtime context that allows opening records for selected table. | `PASS` | Main runtime layout is visible and selected table is ready for records entry. |
| A2 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `Enter` opens records view with visible selected row. | `PASS` | Right panel switches to record rows and highlights active row. |
| A3 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Row navigation keys move selected record while staying in records view. | `PASS` | Selection changes as `j`/`k` are pressed with no context loss. |
| A4 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `Ctrl+f`/`Ctrl+b` execute pagination motions without leaving records context, and page navigation stays bounded when only one page exists. | `PASS` | For fixture table `products` (`5` persisted rows), `Ctrl+f`/`Ctrl+b` keep records view interactive and remain on `Page 1/1`. |
| A5 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Sort popup supports column and direction steps, and applying sort marks selected column with `↑` for `ASC`. | `PASS` | `Shift+S` opens guided flow; after apply, header shows active sort indicator `↑` on `name`. |
| A6 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Re-running sort replaces previous sort and keeps one active sort with updated direction indicator. | `PASS` | Second apply changes active sort to `name DESC`; header indicator is `↓` and previous `ASC` state is gone. |
| A7 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Pending insert rows remain pinned at top even when SQL sort is applied. | `PASS` | Pending `✚` row stays first after sort apply and persisted rows are ordered below it. |
| A8 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Pending insert can be removed immediately without leaving records context. | `PASS` | `d` removes pending insert row and list returns to persisted rows only. |
| A9 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Sort state is reset on table switch and does not carry over when returning to original table. | `PASS` | After switching `products -> categories -> products`, previous sort indicator is not present in `products` header. |
| A10 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Field-focus mode can be entered from records view. | `PASS` | Cell-level focus state becomes active after `e`. |
| A11 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | Field-focus left/right navigation works at cell level. | `PASS` | Focused cell changes horizontally with `h`/`l`. |
| A12 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `Esc` exits field-focus mode and returns to records-level navigation. | `PASS` | Row-level navigation state is restored and records view remains open. |
| A13 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `Enter` opens vertical selected-row detail and preserves full field visibility. | `PASS` | Right panel shows column/value blocks and long values are wrapped instead of truncated. |
| A14 | `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `Esc` closes row detail and returns to records list context. | `PASS` | Records list selection is visible again and row navigation remains active. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally limited to area 4.4 records navigation, bounded fixed-page motions, guided sort behavior, and detail-view ownership.`

## 7. Cleanup

1. Exit app using `:q`.
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
