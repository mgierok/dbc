package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	runtimecontract "github.com/mgierok/dbc/internal/interfaces/tui/internal"
)

func (m *Model) applyRecordLimit(recordLimit int) (tea.Model, tea.Cmd) {
	if !runtimecontract.IsValidRecordPageLimit(recordLimit) {
		m.ui.statusMessage = fmt.Sprintf("Error: %s", runtimecontract.InvalidSetRecordLimitHint())
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
