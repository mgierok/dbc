package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) handleFilterPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeFilterPopup()
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		return m.confirmPopupSelection()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveDown, key):
		m.movePopupSelection(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeMoveUp, key):
		m.movePopupSelection(-1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveLeft, key):
		if m.filterPopup.step == filterInputValue {
			m.filterPopup.cursor = clamp(m.filterPopup.cursor-1, 0, len(m.filterPopup.input))
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveRight, key):
		if m.filterPopup.step == filterInputValue {
			m.filterPopup.cursor = clamp(m.filterPopup.cursor+1, 0, len(m.filterPopup.input))
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputBackspace, key):
		if m.filterPopup.step == filterInputValue && m.filterPopup.input != "" {
			m.filterPopup.input, m.filterPopup.cursor = deleteAtCursor(m.filterPopup.input, m.filterPopup.cursor)
		}
		return m, nil
	}

	if m.filterPopup.step == filterInputValue && msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.filterPopup.input, m.filterPopup.cursor = insertAtCursor(m.filterPopup.input, insert, m.filterPopup.cursor)
	}
	return m, nil
}

func (m *Model) handleSortPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeSortPopup()
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		return m.confirmSortPopupSelection()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveDown, key):
		m.moveSortPopupSelection(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeMoveUp, key):
		m.moveSortPopupSelection(-1)
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) handleHelpPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.commandInput.active {
		return m.handleCommandInputKey(msg)
	}

	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeHelpPopup()
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeOpenCommandInput, key):
		return m.startCommandInput()
	case primitives.KeyMatches(primitives.KeyPopupMoveDown, key):
		m.moveHelpPopupScroll(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveUp, key):
		m.moveHelpPopupScroll(-1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageDown, key):
		m.moveHelpPopupScroll(m.helpPopupVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageUp, key):
		m.moveHelpPopupScroll(-m.helpPopupVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpTop, key):
		m.helpPopup.scrollOffset = 0
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpBottom, key):
		m.helpPopup.scrollOffset = m.helpPopupMaxOffset()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) handleCommandInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.commandInput = commandInput{}
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		return m.submitCommandInput()
	case primitives.KeyMatches(primitives.KeyInputMoveLeft, key):
		m.commandInput.cursor = clamp(m.commandInput.cursor-1, 0, len(m.commandInput.value))
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveRight, key):
		m.commandInput.cursor = clamp(m.commandInput.cursor+1, 0, len(m.commandInput.value))
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputBackspace, key):
		if m.commandInput.value != "" {
			m.commandInput.value, m.commandInput.cursor = deleteAtCursor(m.commandInput.value, m.commandInput.cursor)
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
		insert := string(msg.Runes)
		if msg.Type == tea.KeySpace {
			insert = " "
		}
		m.commandInput.value, m.commandInput.cursor = insertAtCursor(m.commandInput.value, insert, m.commandInput.cursor)
	}
	return m, nil
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
			m.editPopup.errorMessage = "Column is not nullable"
			return m, nil
		}
		m.editPopup.isNull = true
		m.editPopup.errorMessage = ""
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		return m.confirmEditPopup()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveDown, key):
		if column.Input.Kind == dto.ColumnInputSelect {
			if len(column.Input.Options) > 0 {
				m.editPopup.optionIndex = clamp(m.editPopup.optionIndex+1, 0, len(column.Input.Options)-1)
				m.editPopup.isNull = false
				m.editPopup.errorMessage = ""
			}
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeMoveUp, key):
		if column.Input.Kind == dto.ColumnInputSelect {
			if len(column.Input.Options) > 0 {
				m.editPopup.optionIndex = clamp(m.editPopup.optionIndex-1, 0, len(column.Input.Options)-1)
				m.editPopup.isNull = false
				m.editPopup.errorMessage = ""
			}
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveLeft, key):
		if column.Input.Kind == dto.ColumnInputText {
			m.editPopup.cursor = clamp(m.editPopup.cursor-1, 0, len(m.editPopup.input))
			m.editPopup.errorMessage = ""
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveRight, key):
		if column.Input.Kind == dto.ColumnInputText {
			m.editPopup.cursor = clamp(m.editPopup.cursor+1, 0, len(m.editPopup.input))
			m.editPopup.errorMessage = ""
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputBackspace, key):
		if column.Input.Kind == dto.ColumnInputText && m.editPopup.input != "" {
			m.editPopup.input, m.editPopup.cursor = deleteAtCursor(m.editPopup.input, m.editPopup.cursor)
			m.editPopup.isNull = false
			m.editPopup.errorMessage = ""
		}
		return m, nil
	}

	if column.Input.Kind == dto.ColumnInputText && msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.editPopup.input, m.editPopup.cursor = insertAtCursor(m.editPopup.input, insert, m.editPopup.cursor)
		m.editPopup.isNull = false
		m.editPopup.errorMessage = ""
	}
	return m, nil
}

func (m *Model) handleConfirmPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyPopupMoveDown, key):
		if len(m.confirmPopup.options) > 0 {
			m.confirmPopup.selected = clamp(m.confirmPopup.selected+1, 0, len(m.confirmPopup.options)-1)
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveUp, key):
		if len(m.confirmPopup.options) > 0 {
			m.confirmPopup.selected = clamp(m.confirmPopup.selected-1, 0, len(m.confirmPopup.options)-1)
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyConfirmCancel, key):
		m.closeConfirmPopup()
		m.pendingTableIndex = -1
		m.pendingConfigOpen = false
		return m, nil
	case primitives.KeyMatches(primitives.KeyConfirmAccept, key):
		action := m.confirmPopup.action
		if len(m.confirmPopup.options) > 0 {
			action = m.confirmPopup.options[clamp(m.confirmPopup.selected, 0, len(m.confirmPopup.options)-1)].action
		}
		m.closeConfirmPopup()
		switch action {
		case confirmSave:
			return m.confirmSaveChanges()
		case confirmDiscardTable:
			return m.confirmDiscardTableSwitch()
		case confirmCancelTableSwitch:
			m.pendingTableIndex = -1
			return m, nil
		case confirmConfigSaveAndOpen:
			return m.confirmConfigSaveAndOpen()
		case confirmConfigDiscardAndOpen:
			return m.confirmConfigDiscardAndOpen()
		case confirmConfigCancel:
			m.pendingConfigOpen = false
			return m, nil
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}

func (m *Model) handleRecordDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeRecordDetail()
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveDown, key):
		m.moveRecordDetailScroll(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveUp, key):
		m.moveRecordDetailScroll(-1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageDown, key):
		m.moveRecordDetailScroll(m.recordDetailVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageUp, key):
		m.moveRecordDetailScroll(-m.recordDetailVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpTop, key):
		m.recordDetail.scrollOffset = 0
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpBottom, key):
		m.recordDetail.scrollOffset = m.recordDetailMaxOffset()
		return m, nil
	default:
		return m, nil
	}
}
