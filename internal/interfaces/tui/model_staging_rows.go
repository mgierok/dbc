package tui

import (
	"fmt"
)

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
