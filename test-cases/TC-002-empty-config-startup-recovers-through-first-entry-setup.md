# Empty Config Startup Recovers Through First Entry Setup

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-002` |
| Startup Script | `scripts/start-without-database.sh` |
| Startup Command | `bash scripts/start-without-database.sh` |

## 2. Scenario

- Subject under test: startup behavior when config has no databases and mandatory first-entry setup must gate runtime start.
- Expected result: user-visible startup recovery succeeds only after adding one valid database entry, then runtime main view opens.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-without-database.sh` is executable in current environment.
3. Keep `TMP_ROOT` and `TMP_DB` values printed by startup script output for later steps.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-without-database.sh`. | App opens mandatory first-entry setup (no runtime table view is visible yet). | `A1` |
| S2 | In add form, enter `name=fixture-invalid`, set `db_path` to `<TMP_ROOT>/missing.db`, and press `Enter` to save. | Save is rejected, form stays open, and validation error is visible. | `A2` |
| S3 | Replace `db_path` with printed `<TMP_DB>` (keep non-empty name), then press `Enter`. | Entry is saved and selector list view is shown with the new database option. | `A3` |
| S4 | Press `Enter` on the saved selector row. | App transitions from selector/setup to main two-panel runtime view. | `A4` |
| S5 | Press `q`. | Application exits cleanly to terminal prompt. | `A5` |

## 5. Assertions

| Assertion ID | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| A1 | First screen after command is mandatory first-entry setup, not runtime table view. | `PASS` | Startup opens setup flow that requires adding a first database entry before continuing. |
| A2 | Invalid `db_path` submission is blocked with visible error and no transition out of form. | `PASS` | Form remains active and shows validation failure for missing/unreachable SQLite target. |
| A3 | Valid `db_path` submission succeeds and returns to selector list with saved entry visible. | `PASS` | Selector shows saved entry after form submit using `<TMP_DB>`. |
| A4 | Selecting saved entry opens runtime main view. | `PASS` | Main layout with table panel is rendered after pressing `Enter` on selector row. |
| A5 | App closes on `q` and shell prompt returns. | `PASS` | Terminal prompt returns with no hanging process. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `This scenario covers startup failure/recovery for empty config plus invalid first-entry target correction.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
2. Confirm temporary directory under `<TMP_ROOT>` is removed.
