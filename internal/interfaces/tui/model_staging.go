package tui

import (
	"fmt"
	"reflect"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) requestSaveChanges() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || !m.hasDirtyEdits() {
		return m, nil
	}
	if m.saveChanges == nil {
		m.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	m.openConfirmPopup(confirmSave, "Save staged changes?")
	return m, nil
}

func (m *Model) confirmSaveChanges() (tea.Model, tea.Cmd) {
	changes, err := m.buildTableChanges()
	if err != nil {
		m.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	if len(changes.Inserts) == 0 && len(changes.Updates) == 0 && len(changes.Deletes) == 0 {
		return m, nil
	}
	count := m.dirtyEditCount()
	return m, saveChangesCmd(m.ctx, m.saveChanges, m.currentTableName(), changes, count)
}

func (m *Model) confirmConfigSaveAndOpen() (tea.Model, tea.Cmd) {
	m.pendingConfigOpen = true
	updatedModel, cmd := m.confirmSaveChanges()
	if cmd == nil {
		m.pendingConfigOpen = false
	}
	return updatedModel, cmd
}

func (m *Model) confirmConfigDiscardAndOpen() (tea.Model, tea.Cmd) {
	m.pendingConfigOpen = false
	m.clearStagedState()
	m.openConfigSelector = true
	m.statusMessage = "Opening config manager"
	return m, tea.Quit
}

func (m *Model) confirmDiscardTableSwitch() (tea.Model, tea.Cmd) {
	if m.pendingTableIndex < 0 || m.pendingTableIndex >= len(m.tables) {
		m.pendingTableIndex = -1
		return m, nil
	}
	target := m.pendingTableIndex
	m.pendingTableIndex = -1
	m.clearStagedState()
	m.selectedTable = target
	m.resetTableContext()
	return m, m.loadViewForSelection()
}

func (m *Model) addPendingInsert() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if len(m.schema.Columns) == 0 {
		m.statusMessage = "Error: no schema loaded"
		return m, nil
	}
	row := pendingInsertRow{
		values:       make(map[int]stagedEdit, len(m.schema.Columns)),
		explicitAuto: make(map[int]bool),
	}
	for index, column := range m.schema.Columns {
		row.values[index] = stagedEdit{Value: initialInsertValue(column)}
	}
	if err := m.insertPendingRowAt(0, row); err != nil {
		m.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	m.recordOperation(stagedOperation{
		kind: opInsertAdded,
		insert: insertOperation{
			index: 0,
			row:   clonePendingInsertRow(row),
		},
	})
	m.recordSelection = 0
	m.recordColumn = m.defaultRecordColumnForRow(0)
	m.recordFieldFocus = true
	return m, nil
}

func (m *Model) toggleDeleteSelection() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if m.recordSelection < 0 || m.recordSelection >= m.totalRecordRows() {
		return m, nil
	}
	if insertIndex, isInsert := m.pendingInsertIndexForSelection(); isInsert {
		removed, err := m.removePendingInsert(insertIndex)
		if err != nil {
			m.statusMessage = "Error: " + err.Error()
			return m, nil
		}
		m.recordOperation(stagedOperation{
			kind: opInsertRemoved,
			insert: insertOperation{
				index: insertIndex,
				row:   removed,
			},
		})
		return m, nil
	}
	if !m.canEditRecords() {
		m.statusMessage = "Error: table has no primary key"
		return m, nil
	}
	key, identity, err := m.recordIdentityForVisibleRow(m.recordSelection)
	if err != nil {
		m.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	_, exists := m.pendingDeletes[key]
	nextMarked := !exists
	if err := m.setDeleteMark(key, identity, nextMarked); err != nil {
		m.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	m.recordOperation(stagedOperation{
		kind: opDeleteToggled,
		del: deleteToggleOperation{
			key:          key,
			identity:     identity,
			beforeMarked: exists,
			afterMarked:  nextMarked,
		},
	})
	return m, nil
}

func (m *Model) toggleInsertAutoFields() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	insertIndex, isInsert := m.pendingInsertIndexForSelection()
	if !isInsert {
		return m, nil
	}
	row := m.pendingInserts[insertIndex]
	row.showAuto = !row.showAuto
	m.pendingInserts[insertIndex] = row
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		m.recordColumn = 0
		m.recordFieldFocus = false
		return m, nil
	}
	if !containsInt(visibleColumns, m.recordColumn) {
		m.recordColumn = visibleColumns[0]
	}
	return m, nil
}

func (m *Model) undoStagedAction() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent || len(m.history) == 0 {
		return m, nil
	}
	lastIndex := len(m.history) - 1
	op := m.history[lastIndex]
	m.history = m.history[:lastIndex]
	if err := m.applyInverseOperation(op); err != nil {
		m.statusMessage = "Error: " + err.Error()
		m.history = append(m.history, op)
		return m, nil
	}
	m.future = append(m.future, op)
	m.normalizeRecordSelection()
	return m, nil
}

func (m *Model) redoStagedAction() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent || len(m.future) == 0 {
		return m, nil
	}
	lastIndex := len(m.future) - 1
	op := m.future[lastIndex]
	m.future = m.future[:lastIndex]
	if err := m.applyOperation(op); err != nil {
		m.statusMessage = "Error: " + err.Error()
		m.future = append(m.future, op)
		return m, nil
	}
	m.history = append(m.history, op)
	m.normalizeRecordSelection()
	return m, nil
}

func (m *Model) applyOperation(op stagedOperation) error {
	switch op.kind {
	case opInsertAdded:
		return m.insertPendingRowAt(op.insert.index, op.insert.row)
	case opInsertRemoved:
		_, err := m.removePendingInsert(op.insert.index)
		return err
	case opCellEdited:
		return m.applyCellEditState(op.cell, false)
	case opDeleteToggled:
		return m.setDeleteMark(op.del.key, op.del.identity, op.del.afterMarked)
	default:
		return fmt.Errorf("unsupported staged operation")
	}
}

func (m *Model) applyInverseOperation(op stagedOperation) error {
	switch op.kind {
	case opInsertAdded:
		_, err := m.removePendingInsert(op.insert.index)
		return err
	case opInsertRemoved:
		return m.insertPendingRowAt(op.insert.index, op.insert.row)
	case opCellEdited:
		return m.applyCellEditState(op.cell, true)
	case opDeleteToggled:
		return m.setDeleteMark(op.del.key, op.del.identity, op.del.beforeMarked)
	default:
		return fmt.Errorf("unsupported staged operation")
	}
}

func (m *Model) applyCellEditState(op cellEditOperation, useBefore bool) error {
	edit := op.after
	exists := op.afterExists
	explicitAuto := op.afterExplicitAuto
	if useBefore {
		edit = op.before
		exists = op.beforeExists
		explicitAuto = op.beforeExplicitAuto
	}
	switch op.target {
	case cellEditInsert:
		if op.insertIndex < 0 || op.insertIndex >= len(m.pendingInserts) {
			return fmt.Errorf("insert index out of range")
		}
		row := m.pendingInserts[op.insertIndex]
		if row.values == nil {
			row.values = make(map[int]stagedEdit, len(m.schema.Columns))
		}
		if row.explicitAuto == nil {
			row.explicitAuto = make(map[int]bool)
		}
		if exists {
			row.values[op.columnIndex] = edit
		} else {
			delete(row.values, op.columnIndex)
		}
		if op.columnIndex >= 0 && op.columnIndex < len(m.schema.Columns) && m.schema.Columns[op.columnIndex].AutoIncrement {
			if explicitAuto {
				row.explicitAuto[op.columnIndex] = true
			} else {
				delete(row.explicitAuto, op.columnIndex)
			}
		}
		m.pendingInserts[op.insertIndex] = row
		return nil
	case cellEditPersisted:
		if strings.TrimSpace(op.recordKey) == "" {
			return fmt.Errorf("record key missing")
		}
		if m.pendingUpdates == nil {
			m.pendingUpdates = make(map[string]recordEdits)
		}
		edits := m.pendingUpdates[op.recordKey]
		if edits.changes == nil {
			edits.changes = make(map[int]stagedEdit)
		}
		edits.identity = op.identity
		if exists {
			edits.changes[op.columnIndex] = edit
			m.pendingUpdates[op.recordKey] = edits
			return nil
		}
		delete(edits.changes, op.columnIndex)
		if len(edits.changes) == 0 {
			delete(m.pendingUpdates, op.recordKey)
			return nil
		}
		m.pendingUpdates[op.recordKey] = edits
		return nil
	default:
		return fmt.Errorf("unsupported cell edit target")
	}
}

func (m *Model) setDeleteMark(key string, identity dto.RecordIdentity, marked bool) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("record key missing")
	}
	if m.pendingDeletes == nil {
		m.pendingDeletes = make(map[string]recordDelete)
	}
	if marked {
		m.pendingDeletes[key] = recordDelete{identity: identity}
		return nil
	}
	delete(m.pendingDeletes, key)
	return nil
}

func (m *Model) recordOperation(op stagedOperation) {
	m.history = append(m.history, op)
	m.future = nil
}

func (m *Model) stageEdit(rowIndex, columnIndex int, value dto.StagedValue) error {
	if columnIndex < 0 || columnIndex >= len(m.schema.Columns) {
		return fmt.Errorf("column index out of range")
	}
	if rowIndex < 0 || rowIndex >= m.totalRecordRows() {
		return fmt.Errorf("record index out of range")
	}
	if insertIndex, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		op, changed, err := m.stageInsertEdit(insertIndex, columnIndex, value)
		if err != nil {
			return err
		}
		if changed {
			m.recordOperation(op)
		}
		return nil
	}
	op, changed, err := m.stagePersistedEdit(rowIndex, columnIndex, value)
	if err != nil {
		return err
	}
	if changed {
		m.recordOperation(op)
	}
	return nil
}

func (m *Model) stagePersistedEdit(visibleRowIndex, columnIndex int, value dto.StagedValue) (stagedOperation, bool, error) {
	key, identity, err := m.recordIdentityForVisibleRow(visibleRowIndex)
	if err != nil {
		return stagedOperation{}, false, err
	}
	if m.pendingUpdates == nil {
		m.pendingUpdates = make(map[string]recordEdits)
	}
	edits := m.pendingUpdates[key]
	if edits.changes == nil {
		edits.changes = make(map[int]stagedEdit)
	}
	edits.identity = identity
	before, beforeExists := edits.changes[columnIndex]
	after, afterExists := stagedEdit{}, false

	original := m.visibleRowValue(visibleRowIndex, columnIndex)
	if displayValue(value) == original {
		delete(edits.changes, columnIndex)
		if len(edits.changes) == 0 {
			delete(m.pendingUpdates, key)
		} else {
			m.pendingUpdates[key] = edits
		}
	} else {
		after = stagedEdit{Value: value}
		afterExists = true
		edits.changes[columnIndex] = after
		m.pendingUpdates[key] = edits
	}
	changed := beforeExists != afterExists || (beforeExists && afterExists && !stagedEditEqual(before, after))
	if !changed {
		return stagedOperation{}, false, nil
	}
	return stagedOperation{
		kind: opCellEdited,
		cell: cellEditOperation{
			target:       cellEditPersisted,
			recordKey:    key,
			identity:     identity,
			columnIndex:  columnIndex,
			before:       before,
			beforeExists: beforeExists,
			after:        after,
			afterExists:  afterExists,
		},
	}, true, nil
}

func (m *Model) stageInsertEdit(insertIndex, columnIndex int, value dto.StagedValue) (stagedOperation, bool, error) {
	if insertIndex < 0 || insertIndex >= len(m.pendingInserts) {
		return stagedOperation{}, false, fmt.Errorf("insert index out of range")
	}
	if columnIndex < 0 || columnIndex >= len(m.schema.Columns) {
		return stagedOperation{}, false, fmt.Errorf("column index out of range")
	}
	row := m.pendingInserts[insertIndex]
	if row.values == nil {
		row.values = make(map[int]stagedEdit, len(m.schema.Columns))
	}
	if row.explicitAuto == nil {
		row.explicitAuto = make(map[int]bool)
	}
	before, beforeExists := row.values[columnIndex]
	beforeExplicitAuto := row.explicitAuto[columnIndex]
	after := stagedEdit{Value: value}
	row.values[columnIndex] = after
	afterExplicitAuto := beforeExplicitAuto
	if m.schema.Columns[columnIndex].AutoIncrement {
		row.explicitAuto[columnIndex] = true
		afterExplicitAuto = true
	}
	m.pendingInserts[insertIndex] = row
	changed := !beforeExists || !stagedEditEqual(before, after) || beforeExplicitAuto != afterExplicitAuto
	if !changed {
		return stagedOperation{}, false, nil
	}
	return stagedOperation{
		kind: opCellEdited,
		cell: cellEditOperation{
			target:             cellEditInsert,
			insertIndex:        insertIndex,
			columnIndex:        columnIndex,
			before:             before,
			beforeExists:       beforeExists,
			after:              after,
			afterExists:        true,
			beforeExplicitAuto: beforeExplicitAuto,
			afterExplicitAuto:  afterExplicitAuto,
		},
	}, true, nil
}

func (m *Model) canEditRecords() bool {
	for _, column := range m.schema.Columns {
		if column.PrimaryKey {
			return true
		}
	}
	return false
}

func (m *Model) editColumn() (dto.SchemaColumn, bool) {
	index := m.editPopup.columnIndex
	if index < 0 || index >= len(m.schema.Columns) {
		return dto.SchemaColumn{}, false
	}
	return m.schema.Columns[index], true
}

func (m *Model) recordValue(rowIndex, columnIndex int) string {
	if rowIndex < 0 || rowIndex >= len(m.records) {
		return ""
	}
	values := m.records[rowIndex].Values
	if columnIndex < 0 || columnIndex >= len(values) {
		return ""
	}
	return values[columnIndex]
}

func (m *Model) stagedEditForRow(rowIndex, columnIndex int) (stagedEdit, bool) {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return stagedEdit{}, false
	}
	key, ok := m.recordKeyForPersistedRow(persistedIndex)
	if !ok {
		return stagedEdit{}, false
	}
	edits, ok := m.pendingUpdates[key]
	if !ok {
		return stagedEdit{}, false
	}
	edit, ok := edits.changes[columnIndex]
	return edit, ok
}

func (m *Model) recordKeyForPersistedRow(rowIndex int) (string, bool) {
	pkColumns := m.primaryKeyColumns()
	if len(pkColumns) == 0 || rowIndex < 0 || rowIndex >= len(m.records) {
		return "", false
	}
	values := m.records[rowIndex].Values
	parts := make([]string, 0, len(pkColumns))
	for _, pk := range pkColumns {
		if pk.index < 0 || pk.index >= len(values) {
			return "", false
		}
		parts = append(parts, fmt.Sprintf("%s=%s", pk.column.Name, values[pk.index]))
	}
	return strings.Join(parts, "|"), true
}

func (m *Model) recordIdentityForVisibleRow(rowIndex int) (string, dto.RecordIdentity, error) {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return "", dto.RecordIdentity{}, fmt.Errorf("record index out of range")
	}
	return m.recordIdentityForPersistedRow(persistedIndex)
}

func (m *Model) recordIdentityForPersistedRow(rowIndex int) (string, dto.RecordIdentity, error) {
	if rowIndex < 0 || rowIndex >= len(m.records) {
		return "", dto.RecordIdentity{}, fmt.Errorf("record index out of range")
	}
	return m.translatorUseCase().BuildRecordIdentity(m.schema, m.records[rowIndex])
}

func (m *Model) buildTableChanges() (dto.TableChanges, error) {
	return m.translatorUseCase().BuildTableChanges(
		m.schema,
		m.toPendingInsertRowsDTO(),
		m.toPendingRecordEditsDTO(),
		m.toPendingRecordDeletesDTO(),
	)
}

func (m *Model) primaryKeyColumns() []pkColumn {
	if len(m.schema.Columns) == 0 {
		return nil
	}
	var pkColumns []pkColumn
	for i, column := range m.schema.Columns {
		if column.PrimaryKey {
			pkColumns = append(pkColumns, pkColumn{index: i, column: column})
		}
	}
	return pkColumns
}

func (m *Model) dirtyEditCount() int {
	count := 0
	for _, edits := range m.pendingUpdates {
		count += len(edits.changes)
	}
	count += len(m.pendingInserts)
	count += len(m.pendingDeletes)
	return count
}

func (m *Model) hasDirtyEdits() bool {
	return m.dirtyEditCount() > 0
}

func (m *Model) clearStagedState() {
	m.pendingInserts = nil
	m.pendingUpdates = nil
	m.pendingDeletes = nil
	m.history = nil
	m.future = nil
}

func (m *Model) totalRecordRows() int {
	return len(m.pendingInserts) + len(m.records)
}

func (m *Model) pendingInsertIndex(rowIndex int) (int, bool) {
	if rowIndex < 0 || rowIndex >= len(m.pendingInserts) {
		return -1, false
	}
	return rowIndex, true
}

func (m *Model) pendingInsertIndexForSelection() (int, bool) {
	return m.pendingInsertIndex(m.recordSelection)
}

func (m *Model) persistedRowIndex(rowIndex int) int {
	persisted := rowIndex - len(m.pendingInserts)
	if persisted < 0 || persisted >= len(m.records) {
		return -1
	}
	return persisted
}

func (m *Model) visibleRowValue(rowIndex, columnIndex int) string {
	if insertIndex, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		row := m.pendingInserts[insertIndex]
		if value, ok := row.values[columnIndex]; ok {
			return displayValue(value.Value)
		}
		return ""
	}
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return ""
	}
	return m.recordValue(persistedIndex, columnIndex)
}

func (m *Model) isRowMarkedDelete(rowIndex int) bool {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return false
	}
	key, ok := m.recordKeyForPersistedRow(persistedIndex)
	if !ok {
		return false
	}
	_, marked := m.pendingDeletes[key]
	return marked
}

func (m *Model) removePendingInsert(index int) (pendingInsertRow, error) {
	if index < 0 || index >= len(m.pendingInserts) {
		return pendingInsertRow{}, fmt.Errorf("insert index out of range")
	}
	removed := clonePendingInsertRow(m.pendingInserts[index])
	m.pendingInserts = append(m.pendingInserts[:index], m.pendingInserts[index+1:]...)
	if m.recordSelection > index {
		m.recordSelection--
	}
	m.normalizeRecordSelection()
	return removed, nil
}

func (m *Model) insertPendingRowAt(index int, row pendingInsertRow) error {
	if index < 0 || index > len(m.pendingInserts) {
		return fmt.Errorf("insert index out of range")
	}
	cloned := clonePendingInsertRow(row)
	m.pendingInserts = append(m.pendingInserts, pendingInsertRow{})
	copy(m.pendingInserts[index+1:], m.pendingInserts[index:])
	m.pendingInserts[index] = cloned
	if m.recordSelection >= index {
		m.recordSelection++
	}
	m.normalizeRecordSelection()
	return nil
}

func (m *Model) normalizeRecordSelection() {
	totalRows := m.totalRecordRows()
	if totalRows == 0 {
		m.recordSelection = 0
		m.recordFieldFocus = false
		m.recordColumn = 0
		return
	}
	m.recordSelection = clamp(m.recordSelection, 0, totalRows-1)
	m.syncRecordColumnForSelection()
}

func (m *Model) syncRecordColumnForSelection() {
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		m.recordColumn = 0
		m.recordFieldFocus = false
		return
	}
	if !containsInt(visibleColumns, m.recordColumn) {
		m.recordColumn = visibleColumns[0]
	}
}

func (m *Model) visibleColumnIndicesForSelection() []int {
	if len(m.schema.Columns) == 0 {
		return nil
	}
	insertIndex, isInsert := m.pendingInsertIndexForSelection()
	columns := make([]int, 0, len(m.schema.Columns))
	for idx, column := range m.schema.Columns {
		if isInsert && !m.pendingInserts[insertIndex].showAuto && column.AutoIncrement {
			continue
		}
		columns = append(columns, idx)
	}
	return columns
}

func (m *Model) defaultRecordColumnForRow(rowIndex int) int {
	columns := m.visibleColumnIndicesForRow(rowIndex)
	if len(columns) == 0 {
		return 0
	}
	return columns[0]
}

func (m *Model) visibleColumnIndicesForRow(rowIndex int) []int {
	if len(m.schema.Columns) == 0 {
		return nil
	}
	insertIndex, isInsert := m.pendingInsertIndex(rowIndex)
	columns := make([]int, 0, len(m.schema.Columns))
	for idx, column := range m.schema.Columns {
		if isInsert && !m.pendingInserts[insertIndex].showAuto && column.AutoIncrement {
			continue
		}
		columns = append(columns, idx)
	}
	return columns
}

func initialInsertValue(column dto.SchemaColumn) dto.StagedValue {
	if column.DefaultValue != nil {
		return dto.StagedValue{Text: *column.DefaultValue, Raw: *column.DefaultValue}
	}
	if column.Nullable {
		return dto.StagedValue{IsNull: true, Text: "NULL"}
	}
	return dto.StagedValue{Text: "", Raw: ""}
}

func displayValue(value dto.StagedValue) string {
	if value.IsNull {
		return "NULL"
	}
	if strings.TrimSpace(value.Text) != "" {
		return value.Text
	}
	if value.Raw != nil {
		return fmt.Sprint(value.Raw)
	}
	return ""
}

func (m *Model) translatorUseCase() *usecase.StagedChangesTranslator {
	if m.translator != nil {
		return m.translator
	}
	return usecase.NewStagedChangesTranslator()
}

func (m *Model) toPendingInsertRowsDTO() []dto.PendingInsertRow {
	rows := make([]dto.PendingInsertRow, 0, len(m.pendingInserts))
	for _, row := range m.pendingInserts {
		dtoRow := dto.PendingInsertRow{
			Values:       make(map[int]dto.StagedEdit, len(row.values)),
			ExplicitAuto: make(map[int]bool, len(row.explicitAuto)),
		}
		for index, value := range row.values {
			dtoRow.Values[index] = dto.StagedEdit{Value: value.Value}
		}
		for index, explicit := range row.explicitAuto {
			dtoRow.ExplicitAuto[index] = explicit
		}
		rows = append(rows, dtoRow)
	}
	return rows
}

func (m *Model) toPendingRecordEditsDTO() map[string]dto.PendingRecordEdits {
	edits := make(map[string]dto.PendingRecordEdits, len(m.pendingUpdates))
	for key, update := range m.pendingUpdates {
		dtoChanges := make(map[int]dto.StagedEdit, len(update.changes))
		for columnIndex, change := range update.changes {
			dtoChanges[columnIndex] = dto.StagedEdit{Value: change.Value}
		}
		edits[key] = dto.PendingRecordEdits{
			Identity: update.identity,
			Changes:  dtoChanges,
		}
	}
	return edits
}

func (m *Model) toPendingRecordDeletesDTO() map[string]dto.PendingRecordDelete {
	deletes := make(map[string]dto.PendingRecordDelete, len(m.pendingDeletes))
	for key, deleteChange := range m.pendingDeletes {
		deletes[key] = dto.PendingRecordDelete{
			Identity: deleteChange.identity,
		}
	}
	return deletes
}

func stagedEditEqual(left, right stagedEdit) bool {
	if left.Value.IsNull != right.Value.IsNull || left.Value.Text != right.Value.Text {
		return false
	}
	return reflect.DeepEqual(left.Value.Raw, right.Value.Raw)
}

func clonePendingInsertRow(row pendingInsertRow) pendingInsertRow {
	cloned := pendingInsertRow{
		values:       make(map[int]stagedEdit, len(row.values)),
		explicitAuto: make(map[int]bool, len(row.explicitAuto)),
		showAuto:     row.showAuto,
	}
	for key, value := range row.values {
		cloned.values[key] = value
	}
	for key, value := range row.explicitAuto {
		cloned.explicitAuto[key] = value
	}
	return cloned
}
