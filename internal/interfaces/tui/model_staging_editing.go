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
	key, identity, err := m.recordIdentityForVisibleRow(m.read.recordSelection)
	if err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	_, exists := m.staging.pendingDeletes[key]
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
	row := m.staging.pendingInserts[insertIndex]
	row.showAuto = !row.showAuto
	m.staging.pendingInserts[insertIndex] = row
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
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent || len(m.staging.history) == 0 {
		return m, nil
	}
	lastIndex := len(m.staging.history) - 1
	op := m.staging.history[lastIndex]
	m.staging.history = m.staging.history[:lastIndex]
	if err := m.applyInverseOperation(op); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		m.staging.history = append(m.staging.history, op)
		return m, nil
	}
	m.staging.future = append(m.staging.future, op)
	m.normalizeRecordSelection()
	return m, nil
}

func (m *Model) redoStagedAction() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent || len(m.staging.future) == 0 {
		return m, nil
	}
	lastIndex := len(m.staging.future) - 1
	op := m.staging.future[lastIndex]
	m.staging.future = m.staging.future[:lastIndex]
	if err := m.applyOperation(op); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		m.staging.future = append(m.staging.future, op)
		return m, nil
	}
	m.staging.history = append(m.staging.history, op)
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
	key, identity, err := m.recordIdentityForVisibleRow(visibleRowIndex)
	if err != nil {
		return stagedOperation{}, false, err
	}
	if m.staging.pendingUpdates == nil {
		m.staging.pendingUpdates = make(map[string]recordEdits)
	}
	edits := m.staging.pendingUpdates[key]
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
			delete(m.staging.pendingUpdates, key)
		} else {
			m.staging.pendingUpdates[key] = edits
		}
	} else {
		after = stagedEdit{Value: value}
		afterExists = true
		edits.changes[columnIndex] = after
		m.staging.pendingUpdates[key] = edits
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
	if insertIndex < 0 || insertIndex >= len(m.staging.pendingInserts) {
		return stagedOperation{}, false, fmt.Errorf("insert index out of range")
	}
	if columnIndex < 0 || columnIndex >= len(m.read.schema.Columns) {
		return stagedOperation{}, false, fmt.Errorf("column index out of range")
	}
	row := m.staging.pendingInserts[insertIndex]
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
	m.staging.pendingInserts[insertIndex] = row
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
