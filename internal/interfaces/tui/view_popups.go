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
		rows:                popupTextRows(m.helpPopupContentLines()),
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
	rows := []standardizedPopupRow{}
	switch m.filterPopup.step {
	case filterSelectColumn:
		columnRows := make([]string, len(m.schema.Columns))
		for i, column := range m.schema.Columns {
			columnRows[i] = fmt.Sprintf("%s (%s)", column.Name, column.Type)
		}
		selected := -1
		if len(columnRows) > 0 {
			selected = clamp(m.filterPopup.columnIndex, 0, len(columnRows)-1)
		}
		rows = popupSelectableRows(columnRows, selected)
	case filterSelectOperator:
		stepLabel = "Select operator"
		operatorRows := make([]string, len(m.filterPopup.operators))
		for i, operator := range m.filterPopup.operators {
			operatorRows[i] = operator.Name
		}
		selected := -1
		if len(operatorRows) > 0 {
			selected = clamp(m.filterPopup.operatorIndex, 0, len(operatorRows)-1)
		}
		rows = popupSelectableRows(operatorRows, selected)
	case filterInputValue:
		stepLabel = "Enter value"
		input := m.filterPopup.input
		cursor := clamp(m.filterPopup.cursor, 0, len(input))
		value := input[:cursor] + "|" + input[cursor:]
		rows = popupTextRows([]string{"Value: " + value})
	}

	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:        "Filter",
		summary:      stepLabel,
		rows:         rows,
		defaultWidth: 60,
		minWidth:     20,
		maxWidth:     60,
		styles:       m.styles,
	})
}

func (m *Model) renderSortPopup(totalWidth int) []string {
	stepLabel := "Select column"
	rows := []standardizedPopupRow{}

	switch m.sortPopup.step {
	case sortSelectColumn:
		columnRows := make([]string, len(m.schema.Columns))
		for i, column := range m.schema.Columns {
			columnRows[i] = fmt.Sprintf("%s (%s)", column.Name, column.Type)
		}
		selected := -1
		if len(columnRows) > 0 {
			selected = clamp(m.sortPopup.columnIndex, 0, len(columnRows)-1)
		}
		rows = popupSelectableRows(columnRows, selected)
	case sortSelectDirection:
		stepLabel = "Select direction"
		directions := sortDirections()
		directionRows := make([]string, len(directions))
		for i, direction := range directions {
			directionRows[i] = string(direction)
		}
		selected := -1
		if len(directionRows) > 0 {
			selected = clamp(m.sortPopup.directionIndex, 0, len(directionRows)-1)
		}
		rows = popupSelectableRows(directionRows, selected)
	}

	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:        "Sort",
		summary:      stepLabel,
		rows:         rows,
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
	rows := []standardizedPopupRow{}

	if inputKind == dto.ColumnInputSelect {
		current := "NULL"
		if !m.editPopup.isNull {
			if len(options) > 0 {
				current = options[clamp(m.editPopup.optionIndex, 0, len(options)-1)]
			} else {
				current = m.editPopup.input
			}
		}
		rows = append(rows, standardizedPopupRow{text: "Value: " + current})
		if len(options) > 0 {
			selected := clamp(m.editPopup.optionIndex, 0, len(options)-1)
			rows = append(rows, popupSelectableRows(options, selected)...)
		}
	} else {
		if m.editPopup.isNull {
			rows = append(rows, standardizedPopupRow{text: "Value: NULL"})
		} else {
			input := m.editPopup.input
			cursor := clamp(m.editPopup.cursor, 0, len(input))
			value := input[:cursor] + "|" + input[cursor:]
			rows = append(rows, standardizedPopupRow{text: "Value: " + value})
		}
	}

	if strings.TrimSpace(m.editPopup.errorMessage) != "" {
		rows = append(rows, standardizedPopupRow{text: m.styles.error("Error: " + m.editPopup.errorMessage)})
	}

	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:        "Edit Cell",
		summary:      columnLabel + frameSegmentSeparator + nullableLabel,
		rows:         rows,
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
		rows:         popupSelectableRows(options, selected),
		defaultWidth: 50,
		minWidth:     20,
		maxWidth:     60,
		styles:       m.styles,
	})
}
