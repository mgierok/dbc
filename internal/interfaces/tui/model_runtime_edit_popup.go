package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) enableRecordFieldFocus() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.totalRecordRows() == 0 {
		return m, nil
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		return m, nil
	}
	if !containsInt(visibleColumns, m.read.recordColumn) {
		m.read.recordColumn = visibleColumns[0]
	}
	m.read.recordFieldFocus = true
	return m, nil
}

func (m *Model) openEditPopup() (tea.Model, tea.Cmd) {
	if m.read.recordSelection < 0 || m.read.recordSelection >= m.totalRecordRows() {
		return m, nil
	}
	if insertIndex, isInsert := m.pendingInsertIndexForSelection(); !isInsert && !m.canEditRecords() {
		m.ui.statusMessage = "Error: table has no primary key"
		return m, nil
	} else if isInsert && (insertIndex < 0 || insertIndex >= len(m.staging.pendingInserts)) {
		return m, nil
	} else if !isInsert {
		if _, _, err := m.recordIdentityForVisibleRow(m.read.recordSelection); err != nil {
			m.ui.statusMessage = "Error: " + err.Error()
			return m, nil
		}
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		return m, nil
	}
	if !containsInt(visibleColumns, m.read.recordColumn) {
		m.read.recordColumn = visibleColumns[0]
	}
	if m.read.recordColumn < 0 || m.read.recordColumn >= len(m.read.schema.Columns) {
		return m, nil
	}
	column := m.read.schema.Columns[m.read.recordColumn]
	currentValue := m.visibleRowValue(m.read.recordSelection, m.read.recordColumn)
	if insertIndex, isInsert := m.pendingInsertIndexForSelection(); isInsert {
		if value, ok := m.staging.pendingInserts[insertIndex].values[m.read.recordColumn]; ok {
			currentValue = displayValue(value.Value)
		}
	} else if staged, ok := m.stagedEditForRow(m.read.recordSelection, m.read.recordColumn); ok {
		currentValue = displayValue(staged.Value)
	} else if persistedIndex := m.persistedRowIndex(m.read.recordSelection); persistedIndex >= 0 {
		if !m.recordCellEditableFromBrowse(persistedIndex, m.read.recordColumn) {
			m.ui.statusMessage = "Error: selected cell has no safe editable source"
			return m, nil
		}
	}

	popup := editPopup{
		active:      true,
		rowIndex:    m.read.recordSelection,
		columnIndex: m.read.recordColumn,
		input:       currentValue,
		cursor:      len(currentValue),
	}
	if strings.EqualFold(currentValue, "NULL") {
		popup.isNull = true
		popup.input = ""
		popup.cursor = 0
	}
	if column.Input.Kind == dto.ColumnInputSelect {
		popup.optionIndex = optionIndex(column.Input.Options, currentValue)
	}
	m.overlay.editPopup = popup
	return m, nil
}

func (m *Model) confirmEditPopup() (tea.Model, tea.Cmd) {
	column, ok := m.editColumn()
	if !ok {
		m.closeEditPopup()
		return m, nil
	}
	input := m.overlay.editPopup.input
	if column.Input.Kind == dto.ColumnInputSelect && len(column.Input.Options) > 0 {
		input = column.Input.Options[clamp(m.overlay.editPopup.optionIndex, 0, len(column.Input.Options)-1)]
	}

	value, err := m.translatorUseCase().ParseStagedValue(column, input, m.overlay.editPopup.isNull)
	if err != nil {
		m.overlay.editPopup.errorMessage = err.Error()
		return m, nil
	}
	if err := m.stageEdit(m.overlay.editPopup.rowIndex, m.overlay.editPopup.columnIndex, value); err != nil {
		m.overlay.editPopup.errorMessage = err.Error()
		return m, nil
	}
	m.closeEditPopup()
	return m, nil
}

func (m *Model) closeEditPopup() {
	m.overlay.editPopup = editPopup{}
}

func (m *Model) handleEditPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	column, ok := m.editColumn()
	if !ok {
		m.closeEditPopup()
		return m, nil
	}

	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeEditPopup()
		return m, nil
	case primitives.KeyMatches(primitives.KeyEditSetNull, key):
		if !column.Nullable {
			m.overlay.editPopup.errorMessage = "Column is not nullable"
			return m, nil
		}
		m.overlay.editPopup.isNull = true
		m.overlay.editPopup.errorMessage = ""
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		return m.confirmEditPopup()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveDown, key):
		if column.Input.Kind == dto.ColumnInputSelect {
			if len(column.Input.Options) > 0 {
				m.overlay.editPopup.optionIndex = clamp(m.overlay.editPopup.optionIndex+1, 0, len(column.Input.Options)-1)
				m.overlay.editPopup.isNull = false
				m.overlay.editPopup.errorMessage = ""
			}
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeMoveUp, key):
		if column.Input.Kind == dto.ColumnInputSelect {
			if len(column.Input.Options) > 0 {
				m.overlay.editPopup.optionIndex = clamp(m.overlay.editPopup.optionIndex-1, 0, len(column.Input.Options)-1)
				m.overlay.editPopup.isNull = false
				m.overlay.editPopup.errorMessage = ""
			}
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveLeft, key):
		if column.Input.Kind == dto.ColumnInputText {
			m.overlay.editPopup.cursor = clamp(m.overlay.editPopup.cursor-1, 0, len(m.overlay.editPopup.input))
			m.overlay.editPopup.errorMessage = ""
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveRight, key):
		if column.Input.Kind == dto.ColumnInputText {
			m.overlay.editPopup.cursor = clamp(m.overlay.editPopup.cursor+1, 0, len(m.overlay.editPopup.input))
			m.overlay.editPopup.errorMessage = ""
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputBackspace, key):
		if column.Input.Kind == dto.ColumnInputText && m.overlay.editPopup.input != "" {
			m.overlay.editPopup.input, m.overlay.editPopup.cursor = deleteAtCursor(m.overlay.editPopup.input, m.overlay.editPopup.cursor)
			m.overlay.editPopup.isNull = false
			m.overlay.editPopup.errorMessage = ""
		}
		return m, nil
	}

	if column.Input.Kind == dto.ColumnInputText && msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.overlay.editPopup.input, m.overlay.editPopup.cursor = insertAtCursor(m.overlay.editPopup.input, insert, m.overlay.editPopup.cursor)
		m.overlay.editPopup.isNull = false
		m.overlay.editPopup.errorMessage = ""
	}
	return m, nil
}
