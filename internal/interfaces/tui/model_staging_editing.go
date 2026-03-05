package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

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
		row.values[index] = stagedEdit{Value: m.stagingPolicyUseCase().InitialInsertValue(column)}
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
