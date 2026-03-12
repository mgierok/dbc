package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) requestSaveChanges() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || !m.hasDirtyEdits() {
		return m, nil
	}
	if m.saveChanges == nil {
		m.ui.statusMessage = "Error: save use case unavailable"
		return m, nil
	}
	message := "Save staged changes?"
	if dirtyTables := m.dirtyTableCount(); dirtyTables > 1 {
		message = fmt.Sprintf("Save staged changes for %d tables?", dirtyTables)
	}
	m.openConfirmPopup(confirmSave, message)
	return m, nil
}

func (m *Model) confirmSaveChanges() (tea.Model, tea.Cmd) {
	changes, err := m.buildDatabaseChanges()
	if err != nil {
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}
	if len(changes) == 0 {
		return m, nil
	}
	count := m.dirtyEditCount()
	return m, saveChangesCmd(m.ctx, m.saveChanges, changes, count)
}

func (m *Model) confirmLeaveSave() (tea.Model, tea.Cmd) {
	updatedModel, cmd := m.confirmSaveChanges()
	if cmd == nil {
		m.ui.pendingLeaveTarget = leaveRuntimeNone
	}
	return updatedModel, cmd
}

func (m *Model) confirmLeaveDiscard() (tea.Model, tea.Cmd) {
	target := m.ui.pendingLeaveTarget
	m.ui.pendingLeaveTarget = leaveRuntimeNone
	m.clearStagedState()
	switch target {
	case leaveRuntimeConfig:
		m.ui.openConfigSelector = true
		m.ui.statusMessage = "Opening database selector"
		return m, tea.Quit
	case leaveRuntimeQuit:
		return m, tea.Quit
	default:
		return m, nil
	}
}
