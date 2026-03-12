package tui

import (
	"fmt"
)

func (m *Model) totalRecordRows() int {
	return len(m.activeTableStaging().pendingInserts) + len(m.read.records)
}

func (m *Model) pendingInsertIndex(rowIndex int) (int, bool) {
	staging := m.activeTableStaging()
	if rowIndex < 0 || rowIndex >= len(staging.pendingInserts) {
		return -1, false
	}
	return rowIndex, true
}

func (m *Model) pendingInsertIndexForSelection() (int, bool) {
	return m.pendingInsertIndex(m.read.recordSelection)
}

func (m *Model) persistedRowIndex(rowIndex int) int {
	persisted := rowIndex - len(m.activeTableStaging().pendingInserts)
	if persisted < 0 || persisted >= len(m.read.records) {
		return -1
	}
	return persisted
}

func (m *Model) visibleRowValue(rowIndex, columnIndex int) string {
	if insertIndex, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		row := m.activeTableStaging().pendingInserts[insertIndex]
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
	_, marked := m.activeTableStaging().pendingDeletes[key]
	return marked
}

func (m *Model) removePendingInsert(index int) (pendingInsertRow, error) {
	staging := m.activeTableStagingPtr()
	if index < 0 || index >= len(staging.pendingInserts) {
		return pendingInsertRow{}, fmt.Errorf("insert index out of range")
	}
	removed := clonePendingInsertRow(staging.pendingInserts[index])
	staging.pendingInserts = append(staging.pendingInserts[:index], staging.pendingInserts[index+1:]...)
	if m.read.recordSelection > index {
		m.read.recordSelection--
	}
	m.normalizeRecordSelection()
	return removed, nil
}

func (m *Model) insertPendingRowAt(index int, row pendingInsertRow) error {
	staging := m.activeTableStagingPtr()
	if index < 0 || index > len(staging.pendingInserts) {
		return fmt.Errorf("insert index out of range")
	}
	cloned := clonePendingInsertRow(row)
	staging.pendingInserts = append(staging.pendingInserts, pendingInsertRow{})
	copy(staging.pendingInserts[index+1:], staging.pendingInserts[index:])
	staging.pendingInserts[index] = cloned
	if m.read.recordSelection >= index {
		m.read.recordSelection++
	}
	m.normalizeRecordSelection()
	return nil
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
	if len(m.read.schema.Columns) == 0 {
		return nil
	}
	insertIndex, isInsert := m.pendingInsertIndexForSelection()
	columns := make([]int, 0, len(m.read.schema.Columns))
	for idx, column := range m.read.schema.Columns {
		if isInsert && !m.activeTableStaging().pendingInserts[insertIndex].showAuto && column.AutoIncrement {
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
	if len(m.read.schema.Columns) == 0 {
		return nil
	}
	insertIndex, isInsert := m.pendingInsertIndex(rowIndex)
	columns := make([]int, 0, len(m.read.schema.Columns))
	for idx, column := range m.read.schema.Columns {
		if isInsert && !m.activeTableStaging().pendingInserts[insertIndex].showAuto && column.AutoIncrement {
			continue
		}
		columns = append(columns, idx)
	}
	return columns
}
