package selector

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *databaseSelectorModel) openHelpPopup() {
	m.helpPopup = selectorHelpPopup{
		active:       true,
		scrollOffset: 0,
		context:      m.mode,
	}
}

func (m *databaseSelectorModel) closeHelpPopup() {
	m.helpPopup = selectorHelpPopup{}
}

func (m *databaseSelectorModel) handleHelpPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeHelpPopup()
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveDown, key):
		m.moveHelpPopupScroll(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveUp, key):
		m.moveHelpPopupScroll(-1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageDown, key):
		m.moveHelpPopupScroll(m.helpPopupVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageUp, key):
		m.moveHelpPopupScroll(-m.helpPopupVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpTop, key):
		m.helpPopup.scrollOffset = 0
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpBottom, key):
		m.helpPopup.scrollOffset = m.helpPopupMaxOffset()
		return m, nil
	default:
		return m, nil
	}
}

func (m *databaseSelectorModel) helpPopupVisibleLines() int {
	height := m.height
	if height <= 0 {
		height = 24
	}
	visible := height - 8
	if visible < 1 {
		visible = 1
	}
	return visible
}

func (m *databaseSelectorModel) helpPopupMaxOffset() int {
	rows := m.helpPopupContentLines()
	maxOffset := len(rows) - m.helpPopupVisibleLines()
	if maxOffset < 0 {
		maxOffset = 0
	}
	return maxOffset
}

func (m *databaseSelectorModel) moveHelpPopupScroll(delta int) {
	maxOffset := m.helpPopupMaxOffset()
	m.helpPopup.scrollOffset = clamp(m.helpPopup.scrollOffset+delta, 0, maxOffset)
}
