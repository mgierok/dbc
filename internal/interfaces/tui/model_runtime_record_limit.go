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
	m.read.recordPageIndex = 0
	m.read.recordSelection = 0
	m.read.recordColumn = 0
	m.read.recordTotalPages = 1
	m.read.recordTotalCount = 0
	m.read.recordRequestID++
	m.read.recordLoading = false
	m.read.recordFieldFocus = false
	m.closeRecordDetail()

	if m.read.viewMode == ViewRecords {
		return m, m.loadRecordsCmd(true)
	}

	m.read.records = nil
	return m, nil
}
