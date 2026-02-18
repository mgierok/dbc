# Direct Launch Opens Main View

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-001` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: direct launch startup path (`-d`) with fixture database.
- Expected result: application opens main two-panel view immediately, without selector screen.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | App starts directly in main view (no selector visible). | `A1` |
| S2 | Observe left panel table list after startup. | Table list contains fixture tables (`categories`, `customers`, `order_items`, `orders`, `products`). | `A2` |
| S3 | Press `q`. | Application exits cleanly to terminal. | `A3` |

## 5. Assertions

| Assertion ID | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| A1 | First visible screen is main runtime layout, not startup selector. | `PASS` | Main two-panel layout rendered immediately after command. |
| A2 | Left table panel includes all expected fixture table names. | `PASS` | Visible table list includes: categories, customers, order_items, orders, products. |
| A3 | App process closes after `q` and shell prompt returns. | `PASS` | Terminal returns to prompt with no blocking UI. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `TMP_ROOT value from startup output should be kept for cleanup.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
2. Confirm temporary directory under `<TMP_ROOT>` is removed.
