# Selector Edit Invalid Path Blocks Save Until Corrected

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-003` |
| Startup Script | `scripts/start-selector-from-config.sh` |
| Startup Command | `bash scripts/start-selector-from-config.sh` |

## 2. Scenario

- Subject under test: selector edit workflow validation for configured database entries.
- Expected result: selector edit rejects invalid `db_path` and allows recovery by correcting path, after which startup can continue to runtime.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-selector-from-config.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-selector-from-config.sh`. | Startup selector opens with one configured `fixture` entry. | `A1` |
| S2 | Press `e` on selected entry to open edit form, set `db_path` to `<TMP_ROOT>/missing.db`, then press `Enter`. | Save is rejected, edit form stays open, and validation error is visible. | `A2` |
| S3 | Replace `db_path` with the original valid path shown in the form and press `Enter`. | Edit is saved and selector list returns with updated reachable entry. | `A3` |
| S4 | Press `Enter` on the selector row. | App opens runtime main two-panel view for selected database. | `A4` |
| S5 | Press `q`. | Application exits cleanly to terminal prompt. | `A5` |

## 5. Assertions

| Assertion ID | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| A1 | Selector is the first screen and contains configured entry from startup script. | `PASS` | Selector list shows `fixture` entry before any runtime view appears. |
| A2 | Invalid edit submission is blocked with visible error and no selector-to-runtime transition. | `PASS` | Edit form remains active after invalid path submit and shows save validation failure. |
| A3 | Corrected edit submission succeeds and returns to selector list with valid entry. | `PASS` | Selector list is restored only after valid path is provided and saved. |
| A4 | Selecting corrected entry transitions to runtime main view. | `PASS` | Main layout appears after `Enter` on selector row. |
| A5 | App exits on `q` without residual interactive process. | `PASS` | Shell prompt returns immediately after quit command. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `This scenario covers selector/config failure-recovery in edit flow with explicit invalid-path rejection.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
2. Confirm temporary directory under `<TMP_ROOT>` is removed.
