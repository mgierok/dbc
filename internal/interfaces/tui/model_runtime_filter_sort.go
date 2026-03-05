package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

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
