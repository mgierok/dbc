# Insert, Edit, and Delete Operations Stay Deterministic

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-005` |
| Functional Behavior Reference | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: runtime data operations for insert, edit, and delete in records context.
- Expected result: insert, edit, and delete interactions each produce deterministic operation outcomes without requiring app restart.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | App opens main two-panel runtime view directly (no selector). | `A1` |
| S2 | In table list, select `categories` and press `Enter` to open records view. | Records view opens for `categories`. | `A2` |
| S3 | Press `Enter` again to enter field-focus mode in records. | Cell-level navigation is active for record operations. | `A3` |
| S4 | Edit one existing `categories.name` value to a unique value and confirm the edit popup. | Edit is accepted and staged for the selected persisted row. | `A4` |
| S5 | On a persisted row, press `d`. | Delete marker is toggled on for the selected persisted row. | `A5` |
| S6 | Press `d` again on the same persisted row. | Delete marker is removed (toggle-off behavior). | `A6` |
| S7 | Press `i` to stage a new row. | Pending insert row appears at top of records list. | `A7` |
| S8 | Populate required pending-insert fields and confirm the edit popup. | Pending insert remains valid for subsequent write workflow. | `A8` |
| S9 | With pending insert selected, press `d`. | Pending insert row is removed immediately (not converted to delete marker). | `A9` |
| S10 | Press `q`. | Application exits cleanly to terminal prompt. | `A10` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | Runtime opens in context where table-level data operations can be performed. | `PASS` | Main two-panel runtime view appears immediately after launch command. |
| A2 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | `categories` records are accessible for insert/edit/delete operations. | `PASS` | Right panel opens records content for `categories`. |
| A3 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | Field-focus mode is reachable for cell-level edit workflow. | `PASS` | Cell-level navigation is visibly active after second `Enter`. |
| A4 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | Editing an existing persisted value is accepted and staged through edit popup flow. | `PASS` | Updated unique value is confirmed in popup and reflected in records row. |
| A5 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | `d` on persisted row toggles delete marker on. | `PASS` | Selected persisted row shows delete-marked state after first `d`. |
| A6 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | Repeating `d` on same persisted row toggles delete marker off. | `PASS` | Previously delete-marked persisted row returns to non-delete state. |
| A7 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | `i` stages a new record at top of records view. | `PASS` | A new pending row is inserted at top position immediately after `i`. |
| A8 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | Pending insert accepts value entry in edit popup for required data fields. | `PASS` | Required field value is accepted and remains populated in pending insert row. |
| A9 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | `d` on pending insert removes that pending row immediately. | `PASS` | Pending insert row disappears from records list after `d`. |
| A10 | `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | Quit exits process normally after data-operation checks. | `PASS` | Terminal prompt returns after `q`. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally constrained to area 4.6 insert/edit/delete operation behavior.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
2. Confirm temporary directory under `<TMP_ROOT>` is removed.
