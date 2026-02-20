# Dirty Config Navigation Requires Explicit Decision

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-006` |
| Functional Behavior Reference | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` |
| Startup Script | `scripts/start-selector-from-config.sh` |
| Startup Command | `bash scripts/start-selector-from-config.sh` |

## 2. Scenario

- Subject under test: staging lifecycle behavior with undo/redo and explicit `:config` dirty-decision routing (`cancel`, `save`, `discard`).
- Expected result: staged changes are reversible via undo/redo and `:config` navigation executes deterministic decision outcomes.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-selector-from-config.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-selector-from-config.sh`. | Startup selector opens with configured database entry. | `A1` |
| S2 | Press `Enter` on selected entry to open runtime. | Main two-panel runtime view opens. | `A2` |
| S3 | Open records and stage one valid edit in selected table. | A staged change is created for the selected record. | `A3` |
| S4 | Press `u`. | Most recent staged edit is undone. | `A4` |
| S5 | Press `Ctrl+r`. | Undone staged edit is restored by redo. | `A5` |
| S6 | Run `:config`, choose `cancel`, and confirm. | Decision popup closes and runtime remains active with staged changes intact. | `A6` |
| S7 | Run `:config` again, choose `save`, and confirm. | Product saves pending changes, then opens selector only after successful save. | `A7` |
| S8 | Press `Enter` on selector entry to re-open runtime. | Runtime opens again from selector with clean state. | `A8` |
| S9 | Stage another valid edit in runtime. | Staged changes are present again for decision testing. | `A9` |
| S10 | Run `:config`, choose `discard`, and confirm. | Staged changes are cleared and selector view opens without save. | `A10` |
| S11 | Press `q` from selector. | Application exits cleanly to terminal prompt. | `A11` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | Startup command reaches selector-first flow with configured entry available. | `PASS` | Selector list is first visible screen and contains configured row. |
| A2 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | Selector entry opens runtime context where staging lifecycle can be exercised. | `PASS` | Main two-panel runtime layout appears after `Enter`. |
| A3 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | Valid edit is staged before any save execution. | `PASS` | Edited value remains pending in session after confirm. |
| A4 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | Undo removes most recent staged change from current table state. | `PASS` | Previously staged value reverts after `u`. |
| A5 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | Redo reapplies the most recently undone staged change. | `PASS` | Reverted value returns after `Ctrl+r`. |
| A6 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | `cancel` in dirty `:config` decision keeps runtime active and preserves staged changes. | `PASS` | Selector does not open; runtime remains active with staged value still pending. |
| A7 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | `save` decision persists staged changes first and opens selector only after success. | `PASS` | Save confirmation path completes and selector opens after successful save. |
| A8 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | Runtime can be re-entered from selector after successful save path. | `PASS` | `Enter` on selector row reopens runtime normally. |
| A9 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | New staged change can be created after re-entry for discard-path validation. | `PASS` | New pending edit is created in current runtime session. |
| A10 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | `discard` in dirty `:config` decision clears staged changes and opens selector without save. | `PASS` | Selector appears immediately and pending edit is not retained in runtime state. |
| A11 | `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | Quit from selector exits process normally after decision-path checks. | `PASS` | Terminal prompt returns after `q`. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally constrained to area 4.7 staging lifecycle, undo/redo behavior, and dirty-navigation decisions.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`
2. Confirm temporary directory under `<TMP_ROOT>` is removed.
