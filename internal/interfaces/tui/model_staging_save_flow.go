package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) requestSaveChanges() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || !m.hasDirtyEdits() {
		return m, nil
	}
	if m.saveChanges == nil {
		m.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	m.openConfirmPopup(confirmSave, "Save staged changes?")
	return m, nil
}

func (m *Model) confirmSaveChanges() (tea.Model, tea.Cmd) {
	changes, err := m.buildTableChanges()
	if err != nil {
		m.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	if len(changes.Inserts) == 0 && len(changes.Updates) == 0 && len(changes.Deletes) == 0 {
		return m, nil
	}
	count := m.dirtyEditCount()
	return m, saveChangesCmd(m.ctx, m.saveChanges, m.currentTableName(), changes, count)
}

func (m *Model) confirmConfigSaveAndOpen() (tea.Model, tea.Cmd) {
	m.pendingConfigOpen = true
	updatedModel, cmd := m.confirmSaveChanges()
	if cmd == nil {
		m.pendingConfigOpen = false
	}
	return updatedModel, cmd
}

func (m *Model) confirmConfigDiscardAndOpen() (tea.Model, tea.Cmd) {
	m.pendingConfigOpen = false
	m.clearStagedState()
	m.openConfigSelector = true
	m.statusMessage = "Opening config manager"
	return m, tea.Quit
}

func (m *Model) confirmDiscardTableSwitch() (tea.Model, tea.Cmd) {
	if m.pendingTableIndex < 0 || m.pendingTableIndex >= len(m.tables) {
		m.pendingTableIndex = -1
		return m, nil
	}
	target := m.pendingTableIndex
	m.pendingTableIndex = -1
	m.clearStagedState()
	m.selectedTable = target
	m.resetTableContext()
	return m, m.loadViewForSelection()
}
