package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) openConfirmPopup(action confirmAction, message string) {
	m.confirmPopup = confirmPopup{
		active:  true,
		title:   "Confirm",
		action:  action,
		message: message,
	}
}

func (m *Model) openModalConfirmPopupWithOptions(title, message string, options []confirmOption, selected int) {
	if len(options) == 0 {
		m.confirmPopup = confirmPopup{
			active:  true,
			title:   title,
			action:  confirmConfigCancel,
			message: message,
			modal:   true,
		}
		return
	}
	m.confirmPopup = confirmPopup{
		active:   true,
		title:    title,
		message:  message,
		options:  options,
		selected: clamp(selected, 0, len(options)-1),
		modal:    true,
	}
}

func (m *Model) closeConfirmPopup() {
	m.confirmPopup = confirmPopup{}
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
