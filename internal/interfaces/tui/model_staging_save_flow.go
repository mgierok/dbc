package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func saveConfirmOptions(primaryLabel string, primaryAction confirmAction) []confirmOption {
	return []confirmOption{
		{label: primaryLabel, action: primaryAction},
		{label: "Cancel", action: confirmDatabaseTransitionCancel},
	}
}

func (m *Model) requestSaveChanges() (tea.Model, tea.Cmd) {
	if !m.saveSupportedInCurrentContext() {
		return m, nil
	}
	if m.saveChanges == nil {
		m.ui.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	m.openModalConfirmPopupWithOptions(
		"Save",
		"Choose whether to save staged changes.",
		saveConfirmOptions("Save changes", confirmSave),
		0,
	)
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
	m.openModalConfirmPopupWithOptions(
		"Save",
		"Choose whether to save staged changes before quitting.",
		saveConfirmOptions("Save changes and quit", confirmSaveAndQuit),
		0,
	)
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
	m.ui.saveInFlight = true
	m.ui.statusMessage = "Saving changes..."
	return m, saveChangesCmd(m.ctx, m.saveChanges, m.currentTableName(), changes)
}

func (m *Model) confirmSaveAndQuit() (tea.Model, tea.Cmd) {
	m.ui.pendingDatabaseTransition = nil
	m.ui.pendingQuitAfterSave = true
	updatedModel, cmd := m.confirmSaveChanges()
	if cmd == nil {
		m.ui.pendingQuitAfterSave = false
	}
	return updatedModel, cmd
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
	m.ui.pendingDatabaseTransition = nil
	m.clearStagedState()
	return m, tea.Quit
}
