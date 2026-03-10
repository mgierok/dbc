package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) startFilterPopup() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if m.currentTableName() == "" {
		return m, nil
	}
	if len(m.schema.Columns) == 0 {
		m.pendingFilterOpen = true
		return m, m.loadSchemaCmd()
	}
	m.openFilterPopup()
	return m, nil
}

func (m *Model) startSortPopup() (tea.Model, tea.Cmd) {
	if m.viewMode != ViewRecords || m.focus != FocusContent {
		return m, nil
	}
	if m.currentTableName() == "" {
		return m, nil
	}
	if len(m.schema.Columns) == 0 {
		m.pendingSortOpen = true
		return m, m.loadSchemaCmd()
	}
	m.openSortPopup()
	return m, nil
}

func (m *Model) openFilterPopup() {
	m.filterPopup = filterPopup{
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
	m.filterPopup = filterPopup{}
}

func (m *Model) openSortPopup() {
	directionIndex := 0
	columnIndex := 0
	if m.currentSort != nil {
		for i, column := range m.schema.Columns {
			if column.Name == m.currentSort.Column {
				columnIndex = i
				break
			}
		}
		if m.currentSort.Direction == dto.SortDirectionDesc {
			directionIndex = 1
		}
	}
	m.sortPopup = sortPopup{
		active:         true,
		step:           sortSelectColumn,
		columnIndex:    columnIndex,
		directionIndex: directionIndex,
	}
}

func (m *Model) closeSortPopup() {
	m.sortPopup = sortPopup{}
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
		if m.filterPopup.step == filterInputValue {
			m.filterPopup.cursor = clamp(m.filterPopup.cursor-1, 0, len(m.filterPopup.input))
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveRight, key):
		if m.filterPopup.step == filterInputValue {
			m.filterPopup.cursor = clamp(m.filterPopup.cursor+1, 0, len(m.filterPopup.input))
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputBackspace, key):
		if m.filterPopup.step == filterInputValue && m.filterPopup.input != "" {
			m.filterPopup.input, m.filterPopup.cursor = deleteAtCursor(m.filterPopup.input, m.filterPopup.cursor)
		}
		return m, nil
	}

	if m.filterPopup.step == filterInputValue && msg.Type == tea.KeyRunes {
		insert := string(msg.Runes)
		m.filterPopup.input, m.filterPopup.cursor = insertAtCursor(m.filterPopup.input, insert, m.filterPopup.cursor)
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
	switch m.filterPopup.step {
	case filterSelectColumn:
		if len(m.schema.Columns) == 0 {
			return m, nil
		}
		column := m.schema.Columns[m.filterPopup.columnIndex]
		operators, err := m.listOperators.Execute(m.ctx, column.Type)
		if err != nil {
			m.statusMessage = "Error: " + err.Error()
			return m, nil
		}
		m.filterPopup.operators = operators
		m.filterPopup.operatorIndex = 0
		m.filterPopup.step = filterSelectOperator
		return m, nil
	case filterSelectOperator:
		if len(m.filterPopup.operators) == 0 {
			return m, nil
		}
		operator := m.filterPopup.operators[m.filterPopup.operatorIndex]
		if operator.RequiresValue {
			m.filterPopup.input = ""
			m.filterPopup.cursor = 0
			m.filterPopup.step = filterInputValue
			return m, nil
		}
		return m.applyFilter(operator, "")
	case filterInputValue:
		operator := m.filterPopup.operators[m.filterPopup.operatorIndex]
		return m.applyFilter(operator, m.filterPopup.input)
	default:
		return m, nil
	}
}

func (m *Model) confirmSortPopupSelection() (tea.Model, tea.Cmd) {
	switch m.sortPopup.step {
	case sortSelectColumn:
		if len(m.schema.Columns) == 0 {
			return m, nil
		}
		m.sortPopup.step = sortSelectDirection
		return m, nil
	case sortSelectDirection:
		if len(m.schema.Columns) == 0 {
			return m, nil
		}
		directions := sortDirections()
		if len(directions) == 0 {
			return m, nil
		}
		column := m.schema.Columns[m.sortPopup.columnIndex]
		direction := directions[clamp(m.sortPopup.directionIndex, 0, len(directions)-1)]
		return m.applySort(column.Name, direction)
	default:
		return m, nil
	}
}

func (m *Model) movePopupSelection(delta int) {
	switch m.filterPopup.step {
	case filterSelectColumn:
		if len(m.schema.Columns) == 0 {
			return
		}
		m.filterPopup.columnIndex = clamp(m.filterPopup.columnIndex+delta, 0, len(m.schema.Columns)-1)
	case filterSelectOperator:
		if len(m.filterPopup.operators) == 0 {
			return
		}
		m.filterPopup.operatorIndex = clamp(m.filterPopup.operatorIndex+delta, 0, len(m.filterPopup.operators)-1)
	}
}

func (m *Model) moveSortPopupSelection(delta int) {
	switch m.sortPopup.step {
	case sortSelectColumn:
		if len(m.schema.Columns) == 0 {
			return
		}
		m.sortPopup.columnIndex = clamp(m.sortPopup.columnIndex+delta, 0, len(m.schema.Columns)-1)
	case sortSelectDirection:
		directions := sortDirections()
		if len(directions) == 0 {
			return
		}
		m.sortPopup.directionIndex = clamp(m.sortPopup.directionIndex+delta, 0, len(directions)-1)
	}
}

func (m *Model) applyFilter(operator dto.Operator, value string) (tea.Model, tea.Cmd) {
	if len(m.schema.Columns) == 0 {
		return m, nil
	}
	column := m.schema.Columns[m.filterPopup.columnIndex]
	filter := &dto.Filter{
		Column: column.Name,
		Operator: dto.Operator{
			Name:          operator.Name,
			Kind:          operator.Kind,
			RequiresValue: operator.RequiresValue,
		},
		Value: value,
	}
	m.currentFilter = filter
	m.closeFilterPopup()

	if m.viewMode == ViewRecords {
		return m, m.loadRecordsCmd(true)
	}
	return m, nil
}

func (m *Model) applySort(column string, direction dto.SortDirection) (tea.Model, tea.Cmd) {
	m.currentSort = &dto.Sort{
		Column:    column,
		Direction: direction,
	}
	m.closeSortPopup()
	return m, m.loadRecordsCmd(true)
}
