# Runtime Command Failure Recovery Keeps Session Usable

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-004` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: runtime/TUI interaction continuity across records navigation, field-focus transitions, and invalid command handling.
- Expected result: runtime surfaces invalid command feedback and stays fully usable so user can continue normal TUI interactions without restart.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | App opens main two-panel runtime view directly (no selector). | `A1` |
| S2 | Press `Enter` in main view to open records for the selected table. | Right panel switches from schema to records view with visible row selection. | `A2` |
| S3 | Press `Enter` again in records view. | Field-focus mode activates for record cell-level navigation. | `A3` |
| S4 | Press `Esc`. | Field-focus mode exits and records-level navigation context is restored. | `A4` |
| S5 | Press `:`, type `unknown`, then press `Enter`. | Runtime remains active and status line shows unknown-command feedback. | `A5` |
| S6 | Press `F`. | Filter popup opens from runtime context. | `A6` |
| S7 | Press `Esc` to close the filter popup. | Popup closes and runtime records context remains interactive. | `A7` |
| S8 | Press `q`. | Application exits cleanly to terminal prompt. | `A8` |

## 5. Assertions

| Assertion ID | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| A1 | Startup goes straight to runtime main layout with no selector screen. | `PASS` | Main table/schema layout appears immediately after command. |
| A2 | `Enter` from main view opens records mode for selected table with visible selected row. | `PASS` | Right panel displays row data and active row highlight. |
| A3 | Second `Enter` enables cell-level field focus in records. | `PASS` | Field-focus navigation state is visibly active in records view. |
| A4 | `Esc` exits field focus and returns to records-level navigation without leaving records view. | `PASS` | Records view remains open and field-focus state is no longer active. |
| A5 | Invalid command submission does not quit session and surfaces unknown-command status text. | `PASS` | Status line includes `Unknown command` and runtime remains interactive. |
| A6 | After invalid command feedback, valid runtime action (`F`) still opens filter popup. | `PASS` | Filter popup opens successfully in the same session. |
| A7 | Closing popup with `Esc` returns user to interactive records runtime context. | `PASS` | Popup disappears and records context is immediately usable. |
| A8 | Quit action closes app and returns to shell prompt without hanging process. | `PASS` | Terminal prompt returns after `q`. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Runtime failure/recovery coverage is provided by invalid command trigger in S5 and continued interaction validation in S6-S7.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
2. Confirm temporary directory under `<TMP_ROOT>` is removed.
