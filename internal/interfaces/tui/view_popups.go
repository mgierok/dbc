package tui

import (
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
)

func (m *Model) renderHelpPopup(totalWidth int) []string {
	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:               m.helpPopupContextTitle(),
		summary:             runtimeHelpPopupSummaryLine(),
		rows:                m.helpPopupContentLines(),
		selected:            -1,
		scrollOffset:        m.helpPopup.scrollOffset,
		visibleRows:         m.helpPopupVisibleLines(),
		showScrollIndicator: true,
		defaultWidth:        50,
		minWidth:            20,
		maxWidth:            60,
		styles:              m.styles,
	})
}

func (m *Model) renderFilterPopup(totalWidth int) []string {
	stepLabel := "Select column"
	rows := []string{}
	selected := -1
	switch m.filterPopup.step {
	case filterSelectColumn:
		stepLabel = "Select column"
		rows = make([]string, len(m.schema.Columns))
		for i, column := range m.schema.Columns {
			rows[i] = fmt.Sprintf("%s (%s)", column.Name, column.Type)
		}
		if len(rows) > 0 {
			selected = clamp(m.filterPopup.columnIndex, 0, len(rows)-1)
		}
	case filterSelectOperator:
		stepLabel = "Select operator"
		rows = make([]string, len(m.filterPopup.operators))
		for i, operator := range m.filterPopup.operators {
			rows[i] = operator.Name
		}
		if len(rows) > 0 {
			selected = clamp(m.filterPopup.operatorIndex, 0, len(rows)-1)
		}
	case filterInputValue:
		stepLabel = "Enter value"
		input := m.filterPopup.input
		cursor := clamp(m.filterPopup.cursor, 0, len(input))
		value := input[:cursor] + "|" + input[cursor:]
		rows = append(rows, "Value: "+value)
	}

	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:        "Filter",
		summary:      stepLabel,
		rows:         rows,
		selected:     selected,
		defaultWidth: 60,
		minWidth:     20,
		maxWidth:     60,
		styles:       m.styles,
	})
}

func (m *Model) renderSortPopup(totalWidth int) []string {
	stepLabel := "Select column"
	rows := []string{}
	selected := -1

	switch m.sortPopup.step {
	case sortSelectColumn:
		rows = make([]string, len(m.schema.Columns))
		for i, column := range m.schema.Columns {
			rows[i] = fmt.Sprintf("%s (%s)", column.Name, column.Type)
		}
		if len(rows) > 0 {
			selected = clamp(m.sortPopup.columnIndex, 0, len(rows)-1)
		}
	case sortSelectDirection:
		stepLabel = "Select direction"
		directions := sortDirections()
		rows = make([]string, len(directions))
		for i, direction := range directions {
			rows[i] = string(direction)
		}
		if len(rows) > 0 {
			selected = clamp(m.sortPopup.directionIndex, 0, len(rows)-1)
		}
	}

	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:        "Sort",
		summary:      stepLabel,
		rows:         rows,
		selected:     selected,
		defaultWidth: 60,
		minWidth:     20,
		maxWidth:     60,
		styles:       m.styles,
	})
}

func (m *Model) renderEditPopup(totalWidth int) []string {
	columnLabel := "Unknown column"
	nullableLabel := "NOT NULL"
	inputKind := dto.ColumnInputText
	var options []string
	if m.editPopup.columnIndex >= 0 && m.editPopup.columnIndex < len(m.schema.Columns) {
		column := m.schema.Columns[m.editPopup.columnIndex]
		columnLabel = fmt.Sprintf("%s (%s)", column.Name, column.Type)
		if column.Nullable {
			nullableLabel = "NULLABLE"
		}
		inputKind = column.Input.Kind
		options = column.Input.Options
	}
	rows := []string{}
	selected := -1

	if inputKind == dto.ColumnInputSelect {
		current := "NULL"
		if !m.editPopup.isNull {
			if len(options) > 0 {
				current = options[clamp(m.editPopup.optionIndex, 0, len(options)-1)]
			} else {
				current = m.editPopup.input
			}
		}
		rows = append(rows, "Value: "+current)
		if len(options) > 0 {
			rows = append(rows, options...)
			selected = clamp(m.editPopup.optionIndex, 0, len(options)-1) + 1
		}
	} else {
		if m.editPopup.isNull {
			rows = append(rows, "Value: NULL")
		} else {
			input := m.editPopup.input
			cursor := clamp(m.editPopup.cursor, 0, len(input))
			value := input[:cursor] + "|" + input[cursor:]
			rows = append(rows, "Value: "+value)
		}
	}

	if strings.TrimSpace(m.editPopup.errorMessage) != "" {
		rows = append(rows, m.styles.error("Error: "+m.editPopup.errorMessage))
	}

	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:        "Edit Cell",
		summary:      columnLabel + frameSegmentSeparator + nullableLabel,
		rows:         rows,
		selected:     selected,
		defaultWidth: 60,
		minWidth:     30,
		maxWidth:     60,
		styles:       m.styles,
	})
}

func (m *Model) renderConfirmPopup(totalWidth int) []string {
	title := strings.TrimSpace(m.confirmPopup.title)
	if title == "" {
		title = "Confirm"
	}
	message := m.confirmPopup.message
	if strings.TrimSpace(message) == "" {
		message = "Are you sure?"
	}
	options := make([]string, len(m.confirmPopup.options))
	for i, option := range m.confirmPopup.options {
		options[i] = option.label
	}

	selected := -1
	if len(options) > 0 {
		selected = clamp(m.confirmPopup.selected, 0, len(options)-1)
	}

	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:        title,
		summary:      message,
		rows:         options,
		selected:     selected,
		defaultWidth: 50,
		minWidth:     20,
		maxWidth:     60,
		styles:       m.styles,
	})
}
