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

func (m *Model) recordCellEditableFromBrowse(rowIndex, columnIndex int) bool {
	if rowIndex < 0 || rowIndex >= len(m.read.records) {
		return false
	}
	editable := m.read.records[rowIndex].EditableFromBrowse
	if len(editable) == 0 {
		return true
	}
	if columnIndex < 0 || columnIndex >= len(editable) {
		return false
	}
	return editable[columnIndex]
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
	key, _, err := m.recordIdentityForPersistedRow(rowIndex)
	if err != nil {
		return "", false
	}
	return key, true
}

func (m *Model) recordIdentityForVisibleRow(rowIndex int) (string, dto.RecordIdentity, error) {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return "", dto.RecordIdentity{}, fmt.Errorf("record index out of range")
	}
	return m.recordIdentityForPersistedRow(persistedIndex)
}

func (m *Model) recordIdentityForPersistedRow(rowIndex int) (string, dto.RecordIdentity, error) {
	if rowIndex < 0 || rowIndex >= len(m.read.records) {
		return "", dto.RecordIdentity{}, fmt.Errorf("record index out of range")
	}
	return m.translatorUseCase().BuildRecordIdentity(m.read.schema, m.read.records[rowIndex])
}
