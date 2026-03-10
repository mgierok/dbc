package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) applyRecordLimit(recordLimit int) (tea.Model, tea.Cmd) {
	if recordLimit <= 0 || recordLimit > primitives.RuntimeMaxRecordPageLimit {
		m.statusMessage = fmt.Sprintf("Error: expected :set limit=<1-%d>", primitives.RuntimeMaxRecordPageLimit)
		return m, nil
	}
	if m.runtimeSession == nil {
		m.runtimeSession = &RuntimeSessionState{}
	}

	m.runtimeSession.RecordsPageLimit = recordLimit
	m.statusMessage = fmt.Sprintf("Record limit set to %d", recordLimit)
	m.recordPageIndex = 0
	m.recordSelection = 0
	m.recordColumn = 0
	m.recordTotalPages = 1
	m.recordTotalCount = 0
	m.recordRequestID++
	m.recordLoading = false
	m.recordFieldFocus = false
	m.closeRecordDetail()

	if m.viewMode == ViewRecords {
		return m, m.loadRecordsCmd(true)
	}

	m.records = nil
	return m, nil
}
