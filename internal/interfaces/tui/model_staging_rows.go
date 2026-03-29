package tui

import "github.com/mgierok/dbc/internal/application/dto"

func (m *Model) totalRecordRows() int {
	return len(m.currentStagingSnapshot().PendingInserts) + len(m.read.records)
}

func (m *Model) pendingInsertIndex(rowIndex int) (int, bool) {
	pendingInserts := m.currentStagingSnapshot().PendingInserts
	if rowIndex < 0 || rowIndex >= len(pendingInserts) {
		return -1, false
	}
	return rowIndex, true
}

func (m *Model) pendingInsertForRow(rowIndex int) (dto.InsertDraftSnapshot, bool) {
	index, ok := m.pendingInsertIndex(rowIndex)
	if !ok {
		return dto.InsertDraftSnapshot{}, false
	}
	return m.currentStagingSnapshot().PendingInserts[index], true
}

func (m *Model) pendingInsertForSelection() (dto.InsertDraftSnapshot, bool) {
	return m.pendingInsertForRow(m.read.recordSelection)
}

func (m *Model) persistedRowIndex(rowIndex int) int {
	persisted := rowIndex - len(m.currentStagingSnapshot().PendingInserts)
	if persisted < 0 || persisted >= len(m.read.records) {
		return -1
	}
	return persisted
}

func (m *Model) visibleRowValue(rowIndex, columnIndex int) string {
	if row, isInsert := m.pendingInsertForRow(rowIndex); isInsert {
		if value, ok := row.Values[columnIndex]; ok {
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
	_, marked := m.currentStagingSnapshot().PendingDeletes[key]
	return marked
}

func (m *Model) normalizeRecordSelection() {
	totalRows := m.totalRecordRows()
	if totalRows == 0 {
		m.read.recordSelection = 0
		m.read.recordFieldFocus = false
		m.read.recordColumn = 0
		return
	}
	m.read.recordSelection = clamp(m.read.recordSelection, 0, totalRows-1)
	m.syncRecordColumnForSelection()
}

func (m *Model) syncRecordColumnForSelection() {
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		m.read.recordColumn = 0
		m.read.recordFieldFocus = false
		return
	}
	if !containsInt(visibleColumns, m.read.recordColumn) {
		m.read.recordColumn = visibleColumns[0]
	}
}

func (m *Model) visibleColumnIndicesForSelection() []int {
	return m.visibleColumnIndicesForRow(m.read.recordSelection)
}

func (m *Model) defaultRecordColumnForRow(rowIndex int) int {
	columns := m.visibleColumnIndicesForRow(rowIndex)
	if len(columns) == 0 {
		return 0
	}
	return columns[0]
}

func (m *Model) visibleColumnIndicesForRow(rowIndex int) []int {
	if len(m.read.schema.Columns) == 0 {
		return nil
	}
	insert, isInsert := m.pendingInsertForRow(rowIndex)
	columns := make([]int, 0, len(m.read.schema.Columns))
	for idx, column := range m.read.schema.Columns {
		if isInsert && !m.showAutoForInsert(insert.ID) && column.AutoIncrement {
			continue
		}
		columns = append(columns, idx)
	}
	return columns
}
