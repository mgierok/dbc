package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) openModalConfirmPopupWithOptions(title, message string, options []confirmOption, selected int) {
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
		return m, nil
	case primitives.KeyMatches(primitives.KeyConfirmAccept, key):
		if len(m.overlay.confirmPopup.options) == 0 {
			m.closeConfirmPopup()
			return m, nil
		}
		decisionID := m.overlay.confirmPopup.options[clamp(m.overlay.confirmPopup.selected, 0, len(m.overlay.confirmPopup.options)-1)].decisionID
		m.closeConfirmPopup()
		pending := m.ui.pendingNavigation
		m.ui.pendingNavigation = nil
		pendingCommandInput := m.ui.pendingCommandInput
		m.ui.pendingCommandInput = ""
		if pending == nil {
			return m, nil
		}
		resolution := m.navigationWorkflowUseCase().ResolveDecision(decisionID, *pending)
		m.ui.pendingNavigation = clonePendingRuntimeNavigation(resolution.Pending)
		if resolution.Pending != nil {
			m.ui.pendingCommandInput = pendingCommandInput
		}
		if resolution.NextAction.Kind == usecase.RuntimeNavigationNextActionStartSave {
			return m.startSaveForPendingNavigation()
		}
		return m.executeRuntimeNavigationNextAction(resolution.NextAction)
	default:
		return m, nil
	}
}
