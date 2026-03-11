package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) applyRecordLimit(recordLimit int) (tea.Model, tea.Cmd) {
	if recordLimit <= 0 || recordLimit > primitives.RuntimeMaxRecordPageLimit {
		m.ui.statusMessage = fmt.Sprintf("Error: expected :set limit=<1-%d>", primitives.RuntimeMaxRecordPageLimit)
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
