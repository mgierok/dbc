package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) applyRecordLimit(recordLimit int) (tea.Model, tea.Cmd) {
	if err := m.recordLimitPolicyUseCase().Validate(recordLimit); err != nil {
		m.ui.statusMessage = fmt.Sprintf("Error: %s", err.Error())
		return m, nil
	}
	if m.runtimeSession == nil {
		m.runtimeSession = &RuntimeSessionState{}
	}

	m.runtimeSession.RecordsPageLimit = recordLimit
	m.ui.statusMessage = fmt.Sprintf("Record limit set to %d", recordLimit)
	m.resetReadRecordBrowsingState()
	m.read.recordRequestID++

	if m.read.viewMode == ViewRecords {
		return m, m.loadRecordsCmd(true)
	}

	return m, nil
}
