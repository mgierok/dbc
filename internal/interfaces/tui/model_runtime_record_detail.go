package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) openRecordDetail() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	if m.totalRecordRows() == 0 {
		return m, nil
	}
	m.overlay.recordDetail = recordDetailState{
		active:       true,
		scrollOffset: 0,
	}
	return m, nil
}

func (m *Model) closeRecordDetail() {
	m.overlay.recordDetail = recordDetailState{}
}

func (m *Model) moveRecordDetailScroll(delta int) {
	maxOffset := m.recordDetailMaxOffset()
	m.overlay.recordDetail.scrollOffset = clamp(m.overlay.recordDetail.scrollOffset+delta, 0, maxOffset)
}

func (m *Model) recordDetailVisibleLines() int {
	visible := m.contentHeight() - 1
	if visible < 1 {
		return 1
	}
	return visible
}

func (m *Model) recordDetailMaxOffset() int {
	_, rightWidth := m.panelWidths()
	maxOffset := len(m.recordDetailContentLines(rightWidth)) - m.recordDetailVisibleLines()
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (m *Model) handleRecordDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeRecordDetail()
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveDown, key):
		m.moveRecordDetailScroll(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveUp, key):
		m.moveRecordDetailScroll(-1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageDown, key):
		m.moveRecordDetailScroll(m.recordDetailVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageUp, key):
		m.moveRecordDetailScroll(-m.recordDetailVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpTop, key):
		m.overlay.recordDetail.scrollOffset = 0
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpBottom, key):
		m.overlay.recordDetail.scrollOffset = m.recordDetailMaxOffset()
		return m, nil
	default:
		return m, nil
	}
}
