package tui

import (
	"fmt"
	"strings"
)

func (m *Model) View() string {
	bodyHeight := m.contentHeight()
	leftWidth, rightWidth := m.panelWidths()

	left := m.renderTables(leftWidth, bodyHeight)
	right := m.renderContent(rightWidth, bodyHeight)
	lines := mergePanels(left, right, leftWidth, rightWidth)

	if m.popup.active {
		lines = append(lines, "")
		lines = append(lines, m.renderPopup(leftWidth+rightWidth+3)...)
	}

	status := m.renderStatus(m.width)
	lines = append(lines, status)

	return strings.Join(lines, "\n")
}

func (m *Model) panelWidths() (int, int) {
	width := m.width
	if width <= 0 {
		width = 80
	}
	left := width / 3
	if left < 18 {
		left = 18
	}
	right := width - left - 3
	if right < 10 {
		right = 10
		left = width - right - 3
		if left < 10 {
			left = 10
		}
	}
	return left, right
}

func (m *Model) renderTables(width, height int) []string {
	title := "Tables"
	if m.focus == FocusTables {
		title = "Tables *"
	}
	lines := []string{padRight(title, width)}

	items := make([]string, len(m.tables))
	for i, table := range m.tables {
		items[i] = table.Name
	}

	listHeight := height - 1
	lines = append(lines, renderList(items, m.selectedTable, listHeight, width, m.focus == FocusTables)...)
	return padLines(lines, height, width)
}

func (m *Model) renderContent(width, height int) []string {
	switch m.viewMode {
	case ViewRecords:
		return m.renderRecords(width, height)
	default:
		return m.renderSchema(width, height)
	}
}

func (m *Model) renderSchema(width, height int) []string {
	title := "Schema"
	if m.focus == FocusContent && m.viewMode == ViewSchema {
		title = "Schema *"
	}
	lines := []string{padRight(title, width)}

	if len(m.schema.Columns) == 0 {
		lines = append(lines, padRight("No schema loaded.", width))
		return padLines(lines, height, width)
	}

	items := make([]string, len(m.schema.Columns))
	for i, column := range m.schema.Columns {
		items[i] = fmt.Sprintf("%s : %s", column.Name, column.Type)
	}
	listHeight := height - 1
	lines = append(lines, renderList(items, m.schemaIndex, listHeight, width, m.focus == FocusContent && m.viewMode == ViewSchema)...)
	return padLines(lines, height, width)
}

func (m *Model) renderRecords(width, height int) []string {
	title := "Records"
	if m.focus == FocusContent && m.viewMode == ViewRecords {
		title = "Records *"
	}
	lines := []string{padRight(title, width)}

	columns := m.schemaColumns()
	if len(columns) == 0 {
		lines = append(lines, padRight("No columns loaded.", width))
		return padLines(lines, height, width)
	}

	rowWidth := width - 2
	if rowWidth < 1 {
		rowWidth = 1
	}
	columnWidths := allocateColumnWidths(rowWidth, len(columns))
	header := "  " + formatRow(columns, columnWidths)
	lines = append(lines, padRight(header, width))

	listHeight := height - 2
	if listHeight < 1 {
		return padLines(lines, height, width)
	}

	if len(m.records) == 0 {
		lines = append(lines, padRight("No records.", width))
		return padLines(lines, height, width)
	}

	start := scrollStart(m.recordSelection, listHeight, len(m.records))
	end := minInt(len(m.records), start+listHeight)
	for i := start; i < end; i++ {
		prefix := "  "
		if m.focus == FocusContent && m.viewMode == ViewRecords && i == m.recordSelection {
			prefix = "> "
		}
		row := formatRow(m.records[i].Values, columnWidths)
		lines = append(lines, padRight(prefix+row, width))
	}
	return padLines(lines, height, width)
}

func (m *Model) renderPopup(totalWidth int) []string {
	width := totalWidth
	if width <= 0 {
		width = 60
	}
	if width > 60 {
		width = 60
	}
	if width < 20 {
		width = 20
	}

	border := "+" + strings.Repeat("-", width-2) + "+"
	lines := []string{border}
	lines = append(lines, "|"+padRight("Filter", width-2)+"|")

	stepLabel := ""
	switch m.popup.step {
	case filterSelectColumn:
		stepLabel = "Select column"
	case filterSelectOperator:
		stepLabel = "Select operator"
	case filterInputValue:
		stepLabel = "Enter value"
	}
	lines = append(lines, "|"+padRight(stepLabel, width-2)+"|")
	lines = append(lines, "|"+strings.Repeat("-", width-2)+"|")

	switch m.popup.step {
	case filterSelectColumn:
		items := make([]string, len(m.schema.Columns))
		for i, column := range m.schema.Columns {
			items[i] = fmt.Sprintf("%s (%s)", column.Name, column.Type)
		}
		lines = append(lines, renderPopupList(items, m.popup.columnIndex, width-2)...)
	case filterSelectOperator:
		items := make([]string, len(m.popup.operators))
		for i, operator := range m.popup.operators {
			items[i] = fmt.Sprintf("%s (%s)", operator.Name, operator.SQL)
		}
		lines = append(lines, renderPopupList(items, m.popup.operatorIndex, width-2)...)
	case filterInputValue:
		input := m.popup.input
		cursor := clamp(m.popup.cursor, 0, len(input))
		value := input[:cursor] + "|" + input[cursor:]
		lines = append(lines, "|"+padRight("Value: "+value, width-2)+"|")
	}

	lines = append(lines, border)
	return lines
}

func (m *Model) renderStatus(width int) string {
	if width <= 0 {
		width = 80
	}
	parts := []string{
		"READ-ONLY",
		fmt.Sprintf("View: %s", m.viewModeLabel()),
		fmt.Sprintf("Table: %s", m.currentTableName()),
		m.filterSummary(),
	}
	shortcuts := m.statusShortcuts()
	if strings.TrimSpace(shortcuts) != "" {
		parts = append(parts, shortcuts)
	}
	if strings.TrimSpace(m.statusMessage) != "" {
		parts = append(parts, m.statusMessage)
	}
	status := strings.Join(parts, " | ")
	return padRight(status, width)
}

func (m *Model) viewModeLabel() string {
	if m.viewMode == ViewRecords {
		return "Records"
	}
	return "Schema"
}

func (m *Model) filterSummary() string {
	if m.currentFilter == nil {
		return "Filter: none"
	}
	if m.currentFilter.Operator.RequiresValue {
		return fmt.Sprintf("Filter: %s %s %s", m.currentFilter.Column, m.currentFilter.Operator.SQL, m.currentFilter.Value)
	}
	return fmt.Sprintf("Filter: %s %s", m.currentFilter.Column, m.currentFilter.Operator.SQL)
}

func (m *Model) statusShortcuts() string {
	switch {
	case m.popup.active:
		return "Popup: Enter apply | Esc close"
	case m.focus == FocusTables:
		return "Tables: F filter"
	case m.focus == FocusContent && m.viewMode == ViewSchema:
		return "Schema: F filter"
	case m.focus == FocusContent && m.viewMode == ViewRecords:
		return "Records: F filter"
	default:
		return ""
	}
}

func renderList(items []string, selected, height, width int, focused bool) []string {
	if height < 1 {
		return nil
	}
	if len(items) == 0 {
		return []string{padRight("No items.", width)}
	}

	start := scrollStart(selected, height, len(items))
	end := minInt(len(items), start+height)
	lines := make([]string, 0, height)
	for i := start; i < end; i++ {
		prefix := "  "
		if focused && i == selected {
			prefix = "> "
		}
		line := prefix + items[i]
		lines = append(lines, padRight(line, width))
	}
	return padLines(lines, height, width)
}

func renderPopupList(items []string, selected, width int) []string {
	if len(items) == 0 {
		return []string{"|" + padRight("No options.", width) + "|"}
	}
	lines := make([]string, 0, len(items))
	for i, item := range items {
		prefix := "  "
		if i == selected {
			prefix = "> "
		}
		lines = append(lines, "|"+padRight(prefix+item, width)+"|")
	}
	return lines
}

func mergePanels(left, right []string, leftWidth, rightWidth int) []string {
	maxLines := maxInt(len(left), len(right))
	lines := make([]string, 0, maxLines)
	for i := 0; i < maxLines; i++ {
		leftLine := ""
		if i < len(left) {
			leftLine = left[i]
		}
		rightLine := ""
		if i < len(right) {
			rightLine = right[i]
		}
		combined := padRight(leftLine, leftWidth) + " | " + padRight(rightLine, rightWidth)
		lines = append(lines, combined)
	}
	return lines
}

func allocateColumnWidths(totalWidth, columns int) []int {
	if columns <= 0 {
		return nil
	}
	separatorWidth := (columns - 1) * 3
	available := totalWidth - separatorWidth
	if available < columns {
		available = columns
	}
	base := available / columns
	remainder := available % columns
	widths := make([]int, columns)
	for i := 0; i < columns; i++ {
		widths[i] = base
		if i < remainder {
			widths[i]++
		}
	}
	return widths
}

func formatRow(values []string, widths []int) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		value := ""
		if i < len(values) {
			value = values[i]
		}
		parts[i] = padRight(value, width)
	}
	return strings.Join(parts, " | ")
}

func scrollStart(selection, height, total int) int {
	if height <= 0 || total <= height {
		return 0
	}
	start := selection - height + 1
	if start < 0 {
		start = 0
	}
	maxStart := total - height
	if start > maxStart {
		start = maxStart
	}
	return start
}

func padLines(lines []string, height, width int) []string {
	for len(lines) < height {
		lines = append(lines, padRight("", width))
	}
	return lines
}

func padRight(text string, width int) string {
	if width <= 0 {
		return ""
	}
	text = truncate(text, width)
	if len(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}

func truncate(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}
	if width <= 3 {
		return text[:width]
	}
	return text[:width-3] + "..."
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
