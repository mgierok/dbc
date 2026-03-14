package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) requestSaveChanges() (tea.Model, tea.Cmd) {
	if !m.saveSupportedInCurrentContext() {
		return m, nil
	}
	if m.saveChanges == nil {
		m.ui.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	m.openConfirmPopup(confirmSave, "Save staged changes?")
	return m, nil
}

func (m *Model) requestSaveAndQuit() (tea.Model, tea.Cmd) {
	if !m.nonBlockingRuntimeCommandContextActive() {
		return m, nil
	}
	if !m.hasDirtyEdits() {
		return m, tea.Quit
	}
	if m.saveChanges == nil {
		m.ui.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	m.openConfirmPopup(confirmSaveAndQuit, "Save staged changes?")
	return m, nil
}

func (m *Model) confirmSaveChanges() (tea.Model, tea.Cmd) {
	changes, err := m.buildTableChanges()
	if err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	if len(changes.Inserts) == 0 && len(changes.Updates) == 0 && len(changes.Deletes) == 0 {
		return m, nil
	}
	return m, saveChangesCmd(m.ctx, m.saveChanges, m.currentTableName(), changes)
}

func (m *Model) confirmSaveAndQuit() (tea.Model, tea.Cmd) {
	m.ui.pendingConfigOpen = false
	m.ui.pendingQuitAfterSave = true
	updatedModel, cmd := m.confirmSaveChanges()
	if cmd == nil {
		m.ui.pendingQuitAfterSave = false
	}
	return updatedModel, cmd
}

func (m *Model) confirmConfigSaveAndOpen() (tea.Model, tea.Cmd) {
	m.ui.pendingQuitAfterSave = false
	m.ui.pendingConfigOpen = true
	updatedModel, cmd := m.confirmSaveChanges()
	if cmd == nil {
		m.ui.pendingConfigOpen = false
	}
	return updatedModel, cmd
}

func (m *Model) confirmConfigDiscardAndOpen() (tea.Model, tea.Cmd) {
	m.ui.pendingQuitAfterSave = false
	m.ui.pendingConfigOpen = false
	m.clearStagedState()
	m.ui.openConfigSelector = true
	m.ui.statusMessage = "Opening database selector"
	return m, tea.Quit
}

func (m *Model) confirmDiscardTableSwitch() (tea.Model, tea.Cmd) {
	if m.ui.pendingTableIndex < 0 || m.ui.pendingTableIndex >= len(m.read.tables) {
		m.ui.pendingTableIndex = -1
		return m, nil
	}
	target := m.ui.pendingTableIndex
	m.ui.pendingTableIndex = -1
	m.clearStagedState()
	m.read.selectedTable = target
	m.resetTableContext()
	return m, m.loadViewForSelection()
}

func (m *Model) confirmDiscardQuit() (tea.Model, tea.Cmd) {
	m.ui.pendingQuitAfterSave = false
	m.clearStagedState()
	return m, tea.Quit
}
