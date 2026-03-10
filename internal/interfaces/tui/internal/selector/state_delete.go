package selector

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *databaseSelectorModel) handleDeleteConfirmationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeySelectorDeleteCancel, key):
		m.mode = selectorModeBrowse
		m.confirmDelete = selectorDeleteConfirm{}
		m.browse.statusMessage = "Delete canceled"
		return m, nil
	case primitives.KeyMatches(primitives.KeySelectorDeleteConfirm, key):
		optionIndex := m.confirmDelete.optionIndex
		managerIndex := m.confirmDelete.managerIndex
		if optionIndex < 0 || optionIndex >= len(m.options) || managerIndex < 0 {
			m.mode = selectorModeBrowse
			m.confirmDelete = selectorDeleteConfirm{}
			m.browse.statusMessage = "Invalid selection for delete"
			return m, nil
		}
		if err := m.manager.Delete(m.ctx, managerIndex); err != nil {
			m.mode = selectorModeBrowse
			m.confirmDelete = selectorDeleteConfirm{}
			m.browse.statusMessage = "Delete failed: " + err.Error()
			return m, nil
		}

		m.mode = selectorModeBrowse
		m.confirmDelete = selectorDeleteConfirm{}
		if err := m.refreshOptions(); err != nil {
			m.browse.statusMessage = "Delete succeeded but refresh failed: " + err.Error()
			return m, nil
		}
		m.browse.statusMessage = "Database deleted"
		return m, nil
	default:
		return m, nil
	}
}
