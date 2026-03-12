package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func (m *Model) addPendingInsert() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	if len(m.read.schema.Columns) == 0 {
		m.ui.statusMessage = "Error: no schema loaded"
		return m, nil
	}
	m.syncActiveTableSchema()
	row := pendingInsertRow{
		values:       make(map[int]stagedEdit, len(m.read.schema.Columns)),
		explicitAuto: make(map[int]bool),
	}
	for index, column := range m.read.schema.Columns {
		row.values[index] = stagedEdit{Value: m.stagingPolicyUseCase().InitialInsertValue(column)}
	}
	if err := m.insertPendingRowAt(0, row); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	m.recordOperation(stagedOperation{
		kind: opInsertAdded,
		insert: insertOperation{
			index: 0,
			row:   clonePendingInsertRow(row),
		},
	})
	m.read.recordSelection = 0
	m.read.recordColumn = m.defaultRecordColumnForRow(0)
	m.read.recordFieldFocus = true
	return m, nil
}

func (m *Model) toggleDeleteSelection() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	if m.read.recordSelection < 0 || m.read.recordSelection >= m.totalRecordRows() {
		return m, nil
	}
	if insertIndex, isInsert := m.pendingInsertIndexForSelection(); isInsert {
		removed, err := m.removePendingInsert(insertIndex)
		if err != nil {
			m.ui.statusMessage = "Error: " + err.Error()
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
		m.ui.statusMessage = "Error: table has no primary key"
		return m, nil
	}
	m.syncActiveTableSchema()
	key, identity, err := m.recordIdentityForVisibleRow(m.read.recordSelection)
	if err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	_, exists := m.activeTableStaging().pendingDeletes[key]
	nextMarked := !exists
	if err := m.setDeleteMark(key, identity, nextMarked); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
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
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	insertIndex, isInsert := m.pendingInsertIndexForSelection()
	if !isInsert {
		return m, nil
	}
	staging := m.activeTableStagingPtr()
	row := staging.pendingInserts[insertIndex]
	row.showAuto = !row.showAuto
	staging.pendingInserts[insertIndex] = row
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		m.read.recordColumn = 0
		m.read.recordFieldFocus = false
		return m, nil
	}
	if !containsInt(visibleColumns, m.read.recordColumn) {
		m.read.recordColumn = visibleColumns[0]
	}
	return m, nil
}

func (m *Model) undoStagedAction() (tea.Model, tea.Cmd) {
	staging := m.activeTableStagingPtr()
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent || len(staging.history) == 0 {
		return m, nil
	}
	lastIndex := len(staging.history) - 1
	op := staging.history[lastIndex]
	staging.history = staging.history[:lastIndex]
	if err := m.applyInverseOperation(op); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		staging.history = append(staging.history, op)
		return m, nil
	}
	staging.future = append(staging.future, op)
	m.normalizeRecordSelection()
	return m, nil
}

func (m *Model) redoStagedAction() (tea.Model, tea.Cmd) {
	staging := m.activeTableStagingPtr()
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent || len(staging.future) == 0 {
		return m, nil
	}
	lastIndex := len(staging.future) - 1
	op := staging.future[lastIndex]
	staging.future = staging.future[:lastIndex]
	if err := m.applyOperation(op); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		staging.future = append(staging.future, op)
		return m, nil
	}
	staging.history = append(staging.history, op)
	m.normalizeRecordSelection()
	return m, nil
}

func (m *Model) stageEdit(rowIndex, columnIndex int, value dto.StagedValue) error {
	if columnIndex < 0 || columnIndex >= len(m.read.schema.Columns) {
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
	m.syncActiveTableSchema()
	key, identity, err := m.recordIdentityForVisibleRow(visibleRowIndex)
	if err != nil {
		return stagedOperation{}, false, err
	}
	staging := m.activeTableStagingPtr()
	if staging.pendingUpdates == nil {
		staging.pendingUpdates = make(map[string]recordEdits)
	}
	edits := staging.pendingUpdates[key]
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
			delete(staging.pendingUpdates, key)
		} else {
			staging.pendingUpdates[key] = edits
		}
	} else {
		after = stagedEdit{Value: value}
		afterExists = true
		edits.changes[columnIndex] = after
		staging.pendingUpdates[key] = edits
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
	m.syncActiveTableSchema()
	staging := m.activeTableStagingPtr()
	if insertIndex < 0 || insertIndex >= len(staging.pendingInserts) {
		return stagedOperation{}, false, fmt.Errorf("insert index out of range")
	}
	if columnIndex < 0 || columnIndex >= len(m.read.schema.Columns) {
		return stagedOperation{}, false, fmt.Errorf("column index out of range")
	}
	row := staging.pendingInserts[insertIndex]
	if row.values == nil {
		row.values = make(map[int]stagedEdit, len(m.read.schema.Columns))
	}
	if row.explicitAuto == nil {
		row.explicitAuto = make(map[int]bool)
	}
	before, beforeExists := row.values[columnIndex]
	beforeExplicitAuto := row.explicitAuto[columnIndex]
	after := stagedEdit{Value: value}
	row.values[columnIndex] = after
	afterExplicitAuto := beforeExplicitAuto
	if m.read.schema.Columns[columnIndex].AutoIncrement {
		row.explicitAuto[columnIndex] = true
		afterExplicitAuto = true
	}
	staging.pendingInserts[insertIndex] = row
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
	for _, column := range m.read.schema.Columns {
		if column.PrimaryKey {
			return true
		}
	}
	return false
}

func (m *Model) editColumn() (dto.SchemaColumn, bool) {
	index := m.overlay.editPopup.columnIndex
	if index < 0 || index >= len(m.read.schema.Columns) {
		return dto.SchemaColumn{}, false
	}
	return m.read.schema.Columns[index], true
}
