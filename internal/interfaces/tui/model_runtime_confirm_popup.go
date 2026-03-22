package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) openModalConfirmPopupWithOptions(title, message string, options []confirmOption, selected int) {
	if len(options) == 0 {
		m.overlay.confirmPopup = confirmPopup{
			active:  true,
			title:   title,
			action:  confirmDatabaseTransitionCancel,
			message: message,
			modal:   true,
		}
		return
	}
	m.overlay.confirmPopup = confirmPopup{
		active:   true,
		title:    title,
		message:  message,
		options:  options,
		selected: clamp(selected, 0, len(options)-1),
		modal:    true,
	}
}

func (m *Model) closeConfirmPopup() {
	m.overlay.confirmPopup = confirmPopup{}
}

func (m *Model) handleConfirmPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyPopupMoveDown, key):
		if len(m.overlay.confirmPopup.options) > 0 {
			m.overlay.confirmPopup.selected = clamp(m.overlay.confirmPopup.selected+1, 0, len(m.overlay.confirmPopup.options)-1)
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveUp, key):
		if len(m.overlay.confirmPopup.options) > 0 {
			m.overlay.confirmPopup.selected = clamp(m.overlay.confirmPopup.selected-1, 0, len(m.overlay.confirmPopup.options)-1)
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyConfirmCancel, key):
		m.closeConfirmPopup()
		m.ui.pendingTableIndex = -1
		m.ui.pendingDatabaseTransition = nil
		return m, nil
	case primitives.KeyMatches(primitives.KeyConfirmAccept, key):
		action := m.overlay.confirmPopup.action
		if len(m.overlay.confirmPopup.options) > 0 {
			action = m.overlay.confirmPopup.options[clamp(m.overlay.confirmPopup.selected, 0, len(m.overlay.confirmPopup.options)-1)].action
		}
		m.closeConfirmPopup()
		switch action {
		case confirmDiscardTable:
			return m.confirmDiscardTableSwitch()
		case confirmCancelTableSwitch:
			m.ui.pendingTableIndex = -1
			return m, nil
		case confirmDatabaseTransitionSave:
			return m.confirmDatabaseTransitionSave()
		case confirmDatabaseTransitionDiscard:
			return m.confirmDatabaseTransitionDiscard()
		case confirmDatabaseTransitionCancel:
			m.ui.pendingDatabaseTransition = nil
			return m, nil
		case confirmDiscardQuit:
			return m.confirmDiscardQuit()
		case confirmCancelQuit:
			return m, nil
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}
