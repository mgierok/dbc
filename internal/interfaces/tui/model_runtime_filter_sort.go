package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) startFilterPopup() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	if m.currentTableName() == "" {
		return m, nil
	}
	if len(m.read.schema.Columns) == 0 {
		m.overlay.pendingFilterOpen = true
		return m, m.loadSchemaCmd()
	}
	m.openFilterPopup()
	return m, nil
}

func (m *Model) startSortPopup() (tea.Model, tea.Cmd) {
	if m.read.viewMode != ViewRecords || m.read.focus != FocusContent {
		return m, nil
	}
	if m.currentTableName() == "" {
		return m, nil
	}
	if len(m.read.schema.Columns) == 0 {
		m.overlay.pendingSortOpen = true
		return m, m.loadSchemaCmd()
	}
	m.openSortPopup()
	return m, nil
}

func (m *Model) openFilterPopup() {
	m.overlay.filterPopup = filterPopup{
		active:        true,
		step:          filterSelectColumn,
		columnIndex:   0,
		operatorIndex: 0,
		input:         "",
		operators:     nil,
		cursor:        0,
	}
}

func (m *Model) closeFilterPopup() {
	m.overlay.filterPopup = filterPopup{}
}

func (m *Model) openSortPopup() {
	directionIndex := 0
	columnIndex := 0
	if m.read.currentSort != nil {
		for i, column := range m.read.schema.Columns {
			if column.Name == m.read.currentSort.Column {
				columnIndex = i
				break
			}
		}
		if m.read.currentSort.Direction == dto.SortDirectionDesc {
			directionIndex = 1
		}
	}
	m.overlay.sortPopup = sortPopup{
		active:         true,
		step:           sortSelectColumn,
		columnIndex:    columnIndex,
		directionIndex: directionIndex,
	}
}

func (m *Model) closeSortPopup() {
	m.overlay.sortPopup = sortPopup{}
}

func (m *Model) handleFilterPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeFilterPopup()
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		return m.confirmPopupSelection()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveDown, key):
		m.movePopupSelection(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeMoveUp, key):
		m.movePopupSelection(-1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveLeft, key):
		if m.overlay.filterPopup.step == filterInputValue {
			m.overlay.filterPopup.cursor = clamp(m.overlay.filterPopup.cursor-1, 0, len(m.overlay.filterPopup.input))
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveRight, key):
		if m.overlay.filterPopup.step == filterInputValue {
			m.overlay.filterPopup.cursor = clamp(m.overlay.filterPopup.cursor+1, 0, len(m.overlay.filterPopup.input))
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputBackspace, key):
		if m.overlay.filterPopup.step == filterInputValue && m.overlay.filterPopup.input != "" {
			m.overlay.filterPopup.input, m.overlay.filterPopup.cursor = deleteAtCursor(m.overlay.filterPopup.input, m.overlay.filterPopup.cursor)
		}
		return m, nil
	}

	if m.overlay.filterPopup.step == filterInputValue && msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.overlay.filterPopup.input, m.overlay.filterPopup.cursor = insertAtCursor(m.overlay.filterPopup.input, insert, m.overlay.filterPopup.cursor)
	}
	return m, nil
}

func (m *Model) handleSortPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeSortPopup()
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		return m.confirmSortPopupSelection()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveDown, key):
		m.moveSortPopupSelection(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeMoveUp, key):
		m.moveSortPopupSelection(-1)
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) confirmPopupSelection() (tea.Model, tea.Cmd) {
	switch m.overlay.filterPopup.step {
	case filterSelectColumn:
		if len(m.read.schema.Columns) == 0 {
			return m, nil
		}
		column := m.read.schema.Columns[m.overlay.filterPopup.columnIndex]
		operators, err := m.listOperators.Execute(m.ctx, column.Type)
		if err != nil {
			m.ui.statusMessage = "Error: " + err.Error()
			return m, nil
		}
		m.overlay.filterPopup.operators = operators
		m.overlay.filterPopup.operatorIndex = 0
		m.overlay.filterPopup.step = filterSelectOperator
		return m, nil
	case filterSelectOperator:
		if len(m.overlay.filterPopup.operators) == 0 {
			return m, nil
		}
		operator := m.overlay.filterPopup.operators[m.overlay.filterPopup.operatorIndex]
		if operator.RequiresValue {
			m.overlay.filterPopup.input = ""
			m.overlay.filterPopup.cursor = 0
			m.overlay.filterPopup.step = filterInputValue
			return m, nil
		}
		return m.applyFilter(operator, "")
	case filterInputValue:
		operator := m.overlay.filterPopup.operators[m.overlay.filterPopup.operatorIndex]
		return m.applyFilter(operator, m.overlay.filterPopup.input)
	default:
		return m, nil
	}
}

func (m *Model) confirmSortPopupSelection() (tea.Model, tea.Cmd) {
	switch m.overlay.sortPopup.step {
	case sortSelectColumn:
		if len(m.read.schema.Columns) == 0 {
			return m, nil
		}
		m.overlay.sortPopup.step = sortSelectDirection
		return m, nil
	case sortSelectDirection:
		if len(m.read.schema.Columns) == 0 {
			return m, nil
		}
		directions := sortDirections()
		if len(directions) == 0 {
			return m, nil
		}
		column := m.read.schema.Columns[m.overlay.sortPopup.columnIndex]
		direction := directions[clamp(m.overlay.sortPopup.directionIndex, 0, len(directions)-1)]
		return m.applySort(column.Name, direction)
	default:
		return m, nil
	}
}

func (m *Model) movePopupSelection(delta int) {
	switch m.overlay.filterPopup.step {
	case filterSelectColumn:
		if len(m.read.schema.Columns) == 0 {
			return
		}
		m.overlay.filterPopup.columnIndex = clamp(m.overlay.filterPopup.columnIndex+delta, 0, len(m.read.schema.Columns)-1)
	case filterSelectOperator:
		if len(m.overlay.filterPopup.operators) == 0 {
			return
		}
		m.overlay.filterPopup.operatorIndex = clamp(m.overlay.filterPopup.operatorIndex+delta, 0, len(m.overlay.filterPopup.operators)-1)
	}
}

func (m *Model) moveSortPopupSelection(delta int) {
	switch m.overlay.sortPopup.step {
	case sortSelectColumn:
		if len(m.read.schema.Columns) == 0 {
			return
		}
		m.overlay.sortPopup.columnIndex = clamp(m.overlay.sortPopup.columnIndex+delta, 0, len(m.read.schema.Columns)-1)
	case sortSelectDirection:
		directions := sortDirections()
		if len(directions) == 0 {
			return
		}
		m.overlay.sortPopup.directionIndex = clamp(m.overlay.sortPopup.directionIndex+delta, 0, len(directions)-1)
	}
}

func (m *Model) applyFilter(operator dto.Operator, value string) (tea.Model, tea.Cmd) {
	if len(m.read.schema.Columns) == 0 {
		return m, nil
	}
	column := m.read.schema.Columns[m.overlay.filterPopup.columnIndex]
	filter := &dto.Filter{
		Column: column.Name,
		Operator: dto.Operator{
			Name:          operator.Name,
			Kind:          operator.Kind,
			RequiresValue: operator.RequiresValue,
		},
		Value: value,
	}
	m.read.currentFilter = filter
	m.closeFilterPopup()

	if m.read.viewMode == ViewRecords {
		return m, m.loadRecordsCmd(true)
	}
	return m, nil
}

func (m *Model) applySort(column string, direction dto.SortDirection) (tea.Model, tea.Cmd) {
	m.read.currentSort = &dto.Sort{
		Column:    column,
		Direction: direction,
	}
	m.closeSortPopup()
	return m, m.loadRecordsCmd(true)
}
