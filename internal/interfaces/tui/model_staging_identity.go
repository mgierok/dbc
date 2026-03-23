package tui

import (
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
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

func (m *Model) stagedEditForRow(rowIndex, columnIndex int) (stagedEdit, bool) {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return stagedEdit{}, false
	}
	key, ok := m.recordKeyForPersistedRow(persistedIndex)
	if !ok {
		return stagedEdit{}, false
	}
	edits, ok := m.staging.pendingUpdates[key]
	if !ok {
		return stagedEdit{}, false
	}
	edit, ok := edits.changes[columnIndex]
	return edit, ok
}

func (m *Model) recordKeyForPersistedRow(rowIndex int) (string, bool) {
	if rowIndex < 0 || rowIndex >= len(m.read.records) {
		return "", false
	}
	row := m.read.records[rowIndex]
	if row.IdentityUnavailable {
		return "", false
	}
	if row.RowKey != "" {
		return row.RowKey, true
	}

	pkColumns := m.primaryKeyColumns()
	if len(pkColumns) == 0 {
		return "", false
	}
	values := row.Values
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
	if rowIndex < 0 || rowIndex >= len(m.read.records) {
		return "", dto.RecordIdentity{}, fmt.Errorf("record index out of range")
	}
	row := m.read.records[rowIndex]
	if row.IdentityUnavailable {
		return "", dto.RecordIdentity{}, usecase.ErrSelectedRecordIdentityExceedsSafeBrowseLimit
	}
	if row.RowKey != "" || len(row.Identity.Keys) > 0 {
		if row.RowKey == "" || len(row.Identity.Keys) == 0 {
			return "", dto.RecordIdentity{}, fmt.Errorf("record identity missing")
		}
		return row.RowKey, row.Identity, nil
	}
	return m.translatorUseCase().BuildRecordIdentity(m.read.schema, row)
}

func (m *Model) primaryKeyColumns() []pkColumn {
	if len(m.read.schema.Columns) == 0 {
		return nil
	}
	var pkColumns []pkColumn
	for i, column := range m.read.schema.Columns {
		if column.PrimaryKey {
			pkColumns = append(pkColumns, pkColumn{index: i, column: column})
		}
	}
	return pkColumns
}
