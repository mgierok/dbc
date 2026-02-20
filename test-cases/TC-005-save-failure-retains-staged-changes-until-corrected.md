# Save Failure Retains Staged Changes Until Corrected

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-005` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: transactional save behavior when staged edits violate a database constraint and require user correction.
- Expected result: failed save keeps staged changes visible with failure feedback, and corrected staged data can be saved successfully without restarting the session.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | App opens main two-panel runtime view directly (no selector). | `A1` |
| S2 | In table list, select `categories` and press `Enter` to open records view. | Records view opens for `categories`. | `A2` |
| S3 | Press `Enter` again to enter field-focus mode in records. | Cell-level navigation is active. | `A3` |
| S4 | Edit one `categories.name` cell and set it to a name that already exists in another row, then confirm the edit popup. | Duplicate-value edit is staged and dirty mode is active. | `A4` |
| S5 | Press `w` and confirm save. | Save fails with visible status-line error and session remains in runtime. | `A5` |
| S6 | Observe write-state indicators after failed save. | Dirty mode and staged edit markers remain present after failure. | `A6` |
| S7 | Edit the staged duplicate `name` value to a unique corrected value and confirm. | Corrected value is staged and remains ready for save. | `A7` |
| S8 | Press `w` and confirm save again. | Save succeeds, staged state clears, and records reload. | `A8` |
| S9 | Verify mode/status after successful save. | Status shows clean `READ-ONLY` mode with no pending dirty count. | `A9` |
| S10 | Press `q`. | Application exits cleanly to terminal prompt. | `A10` |

## 5. Assertions

| Assertion ID | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| A1 | Startup command opens runtime directly without selector step. | `PASS` | Main two-panel view appears immediately after launch command. |
| A2 | `categories` records can be opened from table list using `Enter`. | `PASS` | Right panel switches to records content for `categories`. |
| A3 | Field-focus mode can be entered in records context. | `PASS` | Cell-level focus indicator/navigation is visibly active. |
| A4 | Editing a `name` value to a duplicate existing value creates a staged dirty edit before save. | `PASS` | Edited value is shown with staged change indicator and write mode appears. |
| A5 | Save attempt with duplicate `categories.name` fails and surfaces clear error feedback. | `PASS` | Status line shows save failure/constraint error while app remains interactive. |
| A6 | Save failure does not discard user intent; staged change remains pending. | `PASS` | Dirty mode and staged edit marker are still visible after failed save. |
| A7 | User can correct failed staged value in-session without restart. | `PASS` | Corrected unique value is accepted in edit popup and staged. |
| A8 | Save succeeds after correction and applies transactional write. | `PASS` | Save completion status is shown and records refresh without failure feedback. |
| A9 | Successful save clears dirty state and returns runtime mode to clean state. | `PASS` | Mode indicator shows `READ-ONLY` with no pending dirty count. |
| A10 | Quit exits process normally. | `PASS` | Terminal prompt returns after `q`. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `This scenario validates failure/recovery save behavior: failed save retains staged state, corrected save clears staged state.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
2. Confirm temporary directory under `<TMP_ROOT>` is removed.
