package tui

import (
	"fmt"

	"github.com/mgierok/dbc/internal/application/dto"
)

func (m *Model) recordValue(rowIndex, columnIndex int) string {
	if rowIndex < 0 || rowIndex >= len(m.read.records) {
		return ""
	}
	values := m.read.records[rowIndex].Values
	if columnIndex < 0 || columnIndex >= len(values) {
		return ""
	}
	return values[columnIndex]
}

func (m *Model) stagedEditForRow(rowIndex, columnIndex int) (dto.StagedEdit, bool) {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return dto.StagedEdit{}, false
	}
	key, ok := m.recordKeyForPersistedRow(persistedIndex)
	if !ok {
		return dto.StagedEdit{}, false
	}
	edits, ok := m.currentStagingSnapshot().PendingUpdates[key]
	if !ok {
		return dto.StagedEdit{}, false
	}
	edit, ok := edits.Changes[columnIndex]
	return edit, ok
}

func (m *Model) recordKeyForPersistedRow(rowIndex int) (string, bool) {
	recordRef, err := m.persistedRecordRefForPersistedRow(rowIndex)
	if err != nil {
		return "", false
	}
	return recordRef.RowKey, true
}

func (m *Model) persistedRecordRefForVisibleRow(rowIndex int) (dto.PersistedRecordRef, error) {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return dto.PersistedRecordRef{}, fmt.Errorf("record index out of range")
	}
	return m.persistedRecordRefForPersistedRow(persistedIndex)
}

func (m *Model) persistedRecordRefForPersistedRow(rowIndex int) (dto.PersistedRecordRef, error) {
	if rowIndex < 0 || rowIndex >= len(m.read.records) {
		return dto.PersistedRecordRef{}, fmt.Errorf("record index out of range")
	}
	return m.recordAccessResolverUseCase().ResolveForDelete(m.read.schema, m.read.records[rowIndex])
}
