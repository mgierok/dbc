# Dirty Config Navigation Requires Explicit Decision

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-006` |
| Startup Script | `scripts/start-selector-from-config.sh` |
| Startup Command | `bash scripts/start-selector-from-config.sh` |

## 2. Scenario

- Subject under test: dirty-state `:config` navigation guard requiring explicit `cancel`, `discard`, or `save` decision.
- Expected result: runtime blocks navigation until a decision is chosen and executes each decision path deterministically (`cancel` stays in runtime, `discard` opens selector without save, `save` persists then opens selector).

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-selector-from-config.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-selector-from-config.sh`. | Startup selector opens with configured database entry. | `A1` |
| S2 | Press `Enter` on selected entry to open runtime. | Main two-panel runtime view opens. | `A2` |
| S3 | Open records and stage one valid edit in selected table. | Dirty mode is active with staged changes visible. | `A3` |
| S4 | Run `:config`, choose `cancel`, and confirm. | Dirty decision popup closes and runtime remains active with staged changes intact. | `A4` |
| S5 | Run `:config` again, choose `discard`, and confirm. | Staged changes are cleared and selector view opens. | `A5` |
| S6 | Press `Enter` on selector entry to re-open runtime. | Runtime opens again from selector with clean state. | `A6` |
| S7 | Stage another valid edit in runtime. | Dirty mode is active again with pending changes. | `A7` |
| S8 | Run `:config`, choose `save`, and confirm. | Product saves pending changes, then opens selector only after successful save. | `A8` |
| S9 | Press `q` from selector. | Application exits cleanly to terminal prompt. | `A9` |

## 5. Assertions

| Assertion ID | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| A1 | Startup command opens selector-first flow with configured entry available. | `PASS` | Selector list is first visible screen and contains configured row. |
| A2 | Selecting selector entry transitions to runtime main view. | `PASS` | Main two-panel table/schema layout appears after `Enter`. |
| A3 | A valid runtime edit creates dirty write state before save. | `PASS` | Mode indicator switches to write/dirty and staged marker appears. |
| A4 | `cancel` decision blocks navigation and preserves runtime context and staged state. | `PASS` | Selector does not open; runtime remains interactive with dirty indicators unchanged. |
| A5 | `discard` decision clears staged changes and opens selector. | `PASS` | Selector screen appears after decision and previous staged markers are removed. |
| A6 | Runtime can be reopened from selector after discard path. | `PASS` | `Enter` on selector row reopens runtime normally. |
| A7 | Dirty state can be re-created after returning from discard path. | `PASS` | New staged edit is visible and dirty indicator appears again. |
| A8 | `save` decision persists changes first and opens selector only on successful save. | `PASS` | Save success feedback appears and selector opens without dirty-state error. |
| A9 | Quit from selector exits process normally. | `PASS` | Terminal prompt returns after `q`. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `This scenario validates all dirty-navigation decision paths (`cancel`, `discard`, `save`) with explicit user-visible outcomes.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
2. Confirm temporary directory under `<TMP_ROOT>` is removed.
