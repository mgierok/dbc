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
	if _, err := m.stagingSessionUseCase().AddInsert(m.read.schema); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	m.syncStagingSnapshot()
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
	if insert, isInsert := m.pendingInsertForSelection(); isInsert {
		if err := m.stagingSessionUseCase().RemoveInsert(insert.ID); err != nil {
			m.ui.statusMessage = "Error: " + err.Error()
			return m, nil
		}
		m.syncStagingSnapshot()
		m.normalizeRecordSelection()
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
	_, exists := m.currentStagingSnapshot().PendingDeletes[key]
	if err := m.stagingSessionUseCase().SetDeleteMark(key, identity, !exists); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	m.syncStagingSnapshot()
	return m, nil
}

func (m *Model) toggleInsertAutoFields() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	insert, isInsert := m.pendingInsertForSelection()
	if !isInsert {
		return m, nil
	}
	m.setShowAutoForInsert(insert.ID, !m.showAutoForInsert(insert.ID))
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
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	if err := m.stagingSessionUseCase().Undo(); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	m.syncStagingSnapshot()
	m.normalizeRecordSelection()
	return m, nil
}

func (m *Model) redoStagedAction() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	if err := m.stagingSessionUseCase().Redo(); err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	m.syncStagingSnapshot()
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
	if insert, isInsert := m.pendingInsertForRow(rowIndex); isInsert {
		if err := m.stagingSessionUseCase().StageInsertEdit(insert.ID, columnIndex, value); err != nil {
			return err
		}
		m.syncStagingSnapshot()
		return nil
	}
	key, identity, err := m.recordIdentityForVisibleRow(rowIndex)
	if err != nil {
		return err
	}
	if err := m.stagingSessionUseCase().StagePersistedEdit(
		key,
		identity,
		columnIndex,
		m.visibleRowValue(rowIndex, columnIndex),
		value,
	); err != nil {
		return err
	}
	m.syncStagingSnapshot()
	return nil
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
