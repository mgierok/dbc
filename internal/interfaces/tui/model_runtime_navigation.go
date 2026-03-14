package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) moveDown() (tea.Model, tea.Cmd) {
	switch m.read.focus {
	case FocusTables:
		return m.moveTableSelection(1)
	case FocusContent:
		return m.moveContentSelection(1)
	default:
		return m, nil
	}
}

func (m *Model) moveUp() (tea.Model, tea.Cmd) {
	switch m.read.focus {
	case FocusTables:
		return m.moveTableSelection(-1)
	case FocusContent:
		return m.moveContentSelection(-1)
	default:
		return m, nil
	}
}

func (m *Model) moveLeft() (tea.Model, tea.Cmd) {
	if m.read.focus != FocusContent || m.read.viewMode != ViewRecords || !m.read.recordFieldFocus {
		return m, nil
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		return m, nil
	}
	current := indexOfInt(visibleColumns, m.read.recordColumn)
	if current == -1 {
		m.read.recordColumn = visibleColumns[0]
		return m, nil
	}
	if current > 0 {
		m.read.recordColumn = visibleColumns[current-1]
	}
	return m, nil
}

func (m *Model) moveRight() (tea.Model, tea.Cmd) {
	if m.read.focus != FocusContent || m.read.viewMode != ViewRecords || !m.read.recordFieldFocus {
		return m, nil
	}
	visibleColumns := m.visibleColumnIndicesForSelection()
	if len(visibleColumns) == 0 {
		return m, nil
	}
	current := indexOfInt(visibleColumns, m.read.recordColumn)
	if current == -1 {
		m.read.recordColumn = visibleColumns[0]
		return m, nil
	}
	if current < len(visibleColumns)-1 {
		m.read.recordColumn = visibleColumns[current+1]
	}
	return m, nil
}

func (m *Model) pageDown() (tea.Model, tea.Cmd) {
	if m.read.focus == FocusContent && m.read.viewMode == ViewRecords {
		return m.nextRecordPage()
	}
	page := m.pageSize()
	switch m.read.focus {
	case FocusTables:
		return m.moveTableSelection(page)
	case FocusContent:
		return m.moveContentSelection(page)
	default:
		return m, nil
	}
}

func (m *Model) pageUp() (tea.Model, tea.Cmd) {
	if m.read.focus == FocusContent && m.read.viewMode == ViewRecords {
		return m.prevRecordPage()
	}
	page := m.pageSize()
	switch m.read.focus {
	case FocusTables:
		return m.moveTableSelection(-page)
	case FocusContent:
		return m.moveContentSelection(-page)
	default:
		return m, nil
	}
}

func (m *Model) nextRecordPage() (tea.Model, tea.Cmd) {
	if m.read.recordTotalPages <= 1 {
		return m, nil
	}
	if m.read.recordPageIndex >= m.read.recordTotalPages-1 {
		return m, nil
	}
	m.read.recordPageIndex++
	return m, m.loadRecordsCmd(false)
}

func (m *Model) prevRecordPage() (tea.Model, tea.Cmd) {
	if m.read.recordTotalPages <= 1 {
		return m, nil
	}
	if m.read.recordPageIndex <= 0 {
		return m, nil
	}
	m.read.recordPageIndex--
	return m, m.loadRecordsCmd(false)
}

func (m *Model) jumpTop() (tea.Model, tea.Cmd) {
	switch m.read.focus {
	case FocusTables:
		return m.setTableSelection(0)
	case FocusContent:
		return m.setContentSelection(0)
	default:
		return m, nil
	}
}

func (m *Model) jumpBottom() (tea.Model, tea.Cmd) {
	switch m.read.focus {
	case FocusTables:
		return m.setTableSelection(len(m.read.tables) - 1)
	case FocusContent:
		return m.setContentSelection(m.contentMaxIndex())
	default:
		return m, nil
	}
}

func (m *Model) moveTableSelection(delta int) (tea.Model, tea.Cmd) {
	target := m.read.selectedTable + delta
	return m.setTableSelection(target)
}

func (m *Model) setTableSelection(index int) (tea.Model, tea.Cmd) {
	if len(m.read.tables) == 0 {
		return m, nil
	}
	index = clamp(index, 0, len(m.read.tables)-1)
	if index == m.read.selectedTable {
		return m, nil
	}
	if m.hasDirtyEdits() {
		prompt := m.dirtyNavigationPolicyUseCase().BuildTableSwitchPrompt(m.dirtyEditCount())
		m.ui.pendingTableIndex = index
		m.openModalConfirmPopupWithOptions(
			prompt.Title,
			prompt.Message,
			m.confirmOptionsFromDirtyPrompt(prompt, dirtyConfirmFlowTableSwitch),
			0,
		)
		return m, nil
	}
	m.read.selectedTable = index
	m.resetTableContext()
	return m, m.loadViewForSelection()
}

func (m *Model) moveContentSelection(delta int) (tea.Model, tea.Cmd) {
	target := m.contentSelection() + delta
	return m.setContentSelection(target)
}

func (m *Model) setContentSelection(index int) (tea.Model, tea.Cmd) {
	maxIndex := m.contentMaxIndex()
	if maxIndex < 0 {
		return m, nil
	}
	index = clamp(index, 0, maxIndex)
	switch m.read.viewMode {
	case ViewSchema:
		m.read.schemaIndex = index
		return m, nil
	case ViewRecords:
		m.read.recordSelection = index
		m.syncRecordColumnForSelection()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) switchToRecords() (tea.Model, tea.Cmd) {
	m.read.viewMode = ViewRecords
	m.read.focus = FocusContent
	m.read.recordFieldFocus = false
	m.closeRecordDetail()
	if m.currentTableName() == "" {
		return m, nil
	}

	var cmds []tea.Cmd
	if len(m.read.schema.Columns) == 0 {
		cmds = append(cmds, m.loadSchemaCmd())
	}
	if len(m.read.records) == 0 {
		cmds = append(cmds, m.loadRecordsCmd(true))
	}
	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
}
