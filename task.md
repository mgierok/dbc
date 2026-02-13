# Stage 2 Completion Plan: SQLite Data Operations (Remaining Work)

## Summary
- Current state: update/edit flow with staged cell edits and transactional save is implemented.
- Remaining Stage 2 scope to finish: insert, delete, and session-level undo/redo.
- Delivery strategy (chosen): two milestones.
- Undo/redo scope (chosen): table-scoped.
- Insert auto-increment UX (chosen): hidden by default with optional reveal/edit.

## Public API / Interface Changes
1. `internal/application/port/engine.go`
- Replace `ApplyRecordUpdates(ctx, tableName, updates)` with `ApplyRecordChanges(ctx, tableName, changes)`.

2. `internal/domain/model` additions
- Add `RecordInsert`:
  - `Values []ColumnValue`
  - `ExplicitAutoValues []ColumnValue` (for revealed auto-increment fields)
- Add `RecordDelete`:
  - `Identity RecordIdentity`
- Add `TableChanges`:
  - `Inserts []RecordInsert`
  - `Updates []RecordUpdate`
  - `Deletes []RecordDelete`
- Add validation errors:
  - `ErrMissingTableChanges`
  - `ErrMissingInsertValues`
  - `ErrMissingDeleteIdentity`

3. `internal/application/usecase`
- Replace `SaveRecordEdits` with `SaveTableChanges`:
  - `Execute(ctx, tableName string, changes model.TableChanges) error`
- Update wiring in `cmd/dbc/main.go` and TUI constructor.

4. Schema metadata extension
- Extend `model.Column` and `dto.SchemaColumn` with:
  - `DefaultValue *string` (from SQLite `PRAGMA table_info(...).dflt_value`)
  - `AutoIncrement bool` (derived by checking table SQL in `sqlite_master` for `AUTOINCREMENT` on PK column)

## Milestone A (Insert + Delete + Unified Save Pipeline)
1. Infrastructure (SQLite adapter)
- Implement `ApplyRecordChanges` in `internal/infrastructure/engine/sqlite_update.go`.
- Behavior: single transaction per save, order is `INSERT` -> `UPDATE` -> `DELETE`.
- Any error causes rollback of all staged changes.
- Keep identifier quoting and bound args exactly as current implementation style.

2. Application layer
- Add `SaveTableChanges` use case.
- Remove update-only assumptions from use case tests and replace with change-set tests.

3. TUI staged state model
- Replace `stagedEdits` map with table-scoped staged state:
  - `pendingInserts []pendingInsertRow`
  - `pendingUpdates map[rowKey]recordEdits`
  - `pendingDeletes map[rowKey]recordDelete`
  - `history`/`future` stacks introduced in Milestone B (struct placeholders now)
- Dirty counter becomes total staged operations:
  - update cell changes count as changed cells
  - each pending insert counts as 1
  - each pending delete counts as 1

4. Insert UX (`i`)
- In Records view, `i` creates a pending row at top and selects it.
- Prefill rules:
  - `DefaultValue` exists: prefill display with that value
  - nullable without default: `NULL`
  - required without default: empty value (must be edited before save)
- Auto-increment columns:
  - hidden from normal field navigation in pending insert rows
  - `Ctrl+a` toggles "show auto fields" for current pending row, enabling explicit edit
- Pending insert rows are visually marked (`[INS]` prefix).

5. Delete UX (`d`)
- `d` toggles delete mark on selected existing row.
- Marked rows are visually tagged (`[DEL]` prefix).
- If row is a pending insert, `d` removes that pending insert immediately (not marked).
- Deleting a row with staged updates keeps updates in memory until save/undo; save executor ignores updates for rows also marked delete.

6. Save UX (`w`)
- Reuse confirmation popup: "Save staged changes?"
- On success:
  - clear staged state
  - reload records from offset 0 with current filter
  - keep current table and view
- On failure:
  - keep all staged state
  - show error in status bar

7. Table switch with dirty state
- Keep current confirm behavior ("Discard changes and switch tables?").
- Discard clears inserts/updates/deletes for current table.

## Milestone B (Undo/Redo + Stage 2 polish)
1. Undo/redo model (table-scoped)
- Add reversible operation entries for staged actions:
  - `opInsertAdded`
  - `opInsertRemoved`
  - `opCellEdited`
  - `opDeleteToggled`
- `u` pops from undo stack and applies inverse.
- `Ctrl+r` reapplies from redo stack.
- Any new staged action clears redo stack.

2. Undo/redo boundaries
- Scope is current table only.
- On table switch with discard confirmation accepted, both stacks are cleared.
- Undo/redo unavailable when popup is active; keys handled only in normal records mode.

3. UX polish for Stage 2 parity
- Status shortcuts in records mode become:
  - `Records: Enter edit | i insert | d delete | u undo | Ctrl+r redo | w save | F filter`
- Keep `READ-ONLY` vs `WRITE (dirty: N)` mode indicator.
- Ensure field focus navigation works across normal rows and pending insert rows.

4. Documentation updates
- Update `docs/BRD.md` Stage 2 checklist to completed items when implementation lands.
- Update shortcut list in `README.md` or doc section where runtime keys are documented.

## Test Plan and Acceptance Scenarios
1. Domain/Application tests
- `SaveTableChanges` delegates full `TableChanges` payload to engine.
- Validation returns expected errors for empty/invalid change sets.

2. Infrastructure tests (`internal/infrastructure/engine/sqlite_update_test.go`)
- Transaction applies insert+update+delete together successfully.
- Rollback occurs when one operation fails.
- Delete by composite PK and single PK.
- Insert with default values and explicit auto-increment value when provided.
- Update skipped/ignored for rows marked delete in same save batch.

3. TUI model tests (`internal/interfaces/tui/model_test.go`, `view_test.go`)
- `i` creates pending insert at top with expected initial values.
- `d` toggles delete mark for persisted row.
- `d` on pending insert removes it.
- `w` builds and submits correct `TableChanges` content.
- Dirty counter reflects inserts/deletes/edits.
- `u` and `Ctrl+r` revert/reapply staged actions in order.
- Table switch with dirty state prompts discard and clears table-scoped staged state after confirmation.
- Status bar shortcuts and mode text match expected strings.

4. End-to-end smoke (manual)
- Open DB, insert a row, edit another row, mark one for delete, save, verify DB state.
- Repeat with forced SQL error to verify rollback and preserved staged state.
- Verify undo/redo before save and clear behavior after discard.

## Assumptions and Defaults
- Stage 2 remains SQLite-only and single-table editing context.
- Save is atomic for all staged changes in current table.
- Table-scoped undo/redo is sufficient for "session-level" requirement in this stage.
- `Ctrl+a` is reserved for revealing auto-increment fields in pending insert editing.
- Existing filter behavior remains unchanged and applies to reloaded records after save.
