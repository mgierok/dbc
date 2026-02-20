# Filtering Core Flow Applies and Resets on Table Switch

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-007` |
| Functional Behavior Reference | `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: guided filtering flow in records context, including column/operator/value selection, apply effect, and reset on table switch.
- Expected result: filter flow behaves deterministically, applies one active filter for selected table, and resets when switching to another table.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | App opens main two-panel runtime view directly (no selector). | `A1` |
| S2 | Select table `products`, press `Enter` to open records view, then press `F`. | Filter popup opens and starts at column-selection step. | `A2` |
| S3 | Select column `name` and press `Enter`, then select operator `LIKE` and press `Enter`. | Filter flow advances to value-input step because `LIKE` requires value input. | `A3` |
| S4 | Enter value `Tea%` and confirm filter. | Filter is applied and records list reflects `name LIKE 'Tea%'` constraint. | `A4` |
| S5 | Open filter popup again (`F`), choose `category_id`, operator `=`, value `1`, then confirm. | New filter replaces prior filter and only one active filter remains for `products`. | `A5` |
| S6 | Switch to table `categories` in left panel. | Active filter from `products` is reset on table switch. | `A6` |
| S7 | Switch back to `products` and open records view. | Previous `products` filter is no longer active after table-switch reset. | `A7` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | Runtime opens in context where filtering can be triggered from records workflow. | `PASS` | Main runtime view is visible and selected table can enter records/filter flow. |
| A2 | `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | `F` opens guided filter popup and starts with column selection. | `PASS` | Filter popup appears and shows column-choice step first. |
| A3 | `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | Column and operator selection steps execute in sequence; `LIKE` requires value input step. | `PASS` | Flow progresses from column to operator to value input after selecting `LIKE`. |
| A4 | `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | Confirmed value-input filter is applied to currently selected table records. | `PASS` | Records list updates and visible rows satisfy `name LIKE 'Tea%'`. |
| A5 | `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | Applying another filter replaces previous one, preserving one-active-filter rule for selected table. | `PASS` | Active-filter summary reflects only `category_id = 1` and prior filter is no longer active. |
| A6 | `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | Switching to different table resets active filter state. | `PASS` | Filter summary for new table does not carry over `products` filter. |
| A7 | `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | Returning to original table after switch keeps filter reset behavior deterministic. | `PASS` | `products` records reopen without previously active filter summary. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally constrained to area 4.5 filtering flow and reset semantics.`

## 7. Cleanup

1. Exit app using `q`.
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
