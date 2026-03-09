package tui

import (
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
)

func (m *Model) renderSchema(width, height int) []string {
	if len(m.schema.Columns) == 0 {
		return padLines([]string{padRight("No schema loaded.", width)}, height, width)
	}

	items := make([]string, len(m.schema.Columns))
	for i, column := range m.schema.Columns {
		items[i] = fmt.Sprintf("%s : %s", column.Name, column.Type)
	}
	lines := renderList(items, m.schemaIndex, height, width, m.focus == FocusContent && m.viewMode == ViewSchema, m.styles)
	return padLines(lines, height, width)
}

func (m *Model) renderRecords(width, height int) []string {
	lines := make([]string, 0, height)

	columns := m.schemaColumnsForRecordsHeader()
	if len(columns) == 0 {
		lines = append(lines, padRight("No columns loaded.", width))
		return padLines(lines, height, width)
	}

	const (
		recordSelectionPrefixWidth = 2
		recordMarkerSlotWidth      = 2
	)

	rowWidth := width - recordSelectionPrefixWidth - recordMarkerSlotWidth
	if rowWidth < 1 {
		rowWidth = 1
	}
	columnWidths := allocateColumnWidths(rowWidth, len(columns))
	headerPrefix := strings.Repeat(" ", recordSelectionPrefixWidth+recordMarkerSlotWidth)
	headerRows := formatRecordsHeaderRows(columns, columnWidths, m.styles)
	for _, headerRow := range headerRows {
		lines = append(lines, padRight(headerPrefix+headerRow, width))
	}

	listHeight := height - len(headerRows)
	if listHeight < 1 {
		return padLines(lines, height, width)
	}

	if m.totalRecordRows() == 0 {
		lines = append(lines, padRight("No records.", width))
		return padLines(lines, height, width)
	}

	totalRows := m.totalRecordRows()
	start := scrollStart(m.recordSelection, listHeight, totalRows)
	end := minInt(totalRows, start+listHeight)
	for i := start; i < end; i++ {
		prefix := selectionUnselectedPrefix()
		if m.focus == FocusContent && m.viewMode == ViewRecords && i == m.recordSelection {
			prefix = selectionSelectedPrefix()
		}
		displayValues := make([]string, len(columns))
		if insertIndex, isInsert := m.pendingInsertIndex(i); isInsert {
			for colIndex := range columns {
				if value, ok := m.staging.pendingInserts[insertIndex].values[colIndex]; ok {
					displayValues[colIndex] = displayValue(value.Value)
				}
			}
		} else {
			for colIndex := range columns {
				if staged, ok := m.stagedEditForRow(i, colIndex); ok {
					displayValues[colIndex] = displayValue(staged.Value)
				} else {
					displayValues[colIndex] = m.visibleRowValue(i, colIndex)
				}
			}
		}
		focusColumn := -1
		if m.recordFieldFocus && i == m.recordSelection {
			focusColumn = m.recordColumn
		}
		row := formatRecordRow(displayValues, columnWidths, focusColumn)
		rowMarker := m.recordRowMarker(i)
		line := padRight(prefix+rowMarker+" "+row, width)
		if m.focus == FocusContent && m.viewMode == ViewRecords && i == m.recordSelection {
			line = m.styles.selected(line)
		}
		lines = append(lines, line)
	}
	return padLines(lines, height, width)
}

func (m *Model) recordRowMarker(rowIndex int) string {
	if _, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		return iconInsert
	}
	if m.isRowMarkedDelete(rowIndex) {
		return iconDelete
	}
	if m.isRowEdited(rowIndex) {
		return iconEdit
	}
	return " "
}

func (m *Model) renderRecordDetail(width, height int) []string {
	lines := make([]string, 0, height)

	contentLines := m.recordDetailContentLines(width)
	if len(contentLines) == 0 {
		lines = append(lines, padRight("No detail available.", width))
		return padLines(lines, height, width)
	}

	maxOffset := len(contentLines) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := clamp(m.recordDetail.scrollOffset, 0, maxOffset)
	end := minInt(len(contentLines), offset+height)

	for i := offset; i < end; i++ {
		lines = append(lines, padRight(contentLines[i], width))
	}

	return padLines(lines, height, width)
}

func (m *Model) recordDetailContentLines(width int) []string {
	if m.totalRecordRows() == 0 {
		return []string{"No records."}
	}
	if len(m.schema.Columns) == 0 {
		return []string{"No columns loaded."}
	}

	rowIndex := clamp(m.recordSelection, 0, m.totalRecordRows()-1)
	lines := make([]string, 0, len(m.schema.Columns)*4)
	rowLine := iconInfo + " Persisted record"
	if _, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		rowLine = iconInfo + " Pending insert"
	} else if m.isRowMarkedDelete(rowIndex) {
		rowLine = iconInfo + " Marked for delete"
	} else if m.isRowEdited(rowIndex) {
		rowLine = iconInfo + " Edited record"
	}
	lines = append(lines, wrapTextToWidth(m.styles.summary(rowLine), width)...)
	lines = append(lines, "")

	valueWidth := width - 2
	if valueWidth < 1 {
		valueWidth = 1
	}

	for columnIndex, column := range m.schema.Columns {
		value, edited := m.effectiveRecordDetailValue(rowIndex, columnIndex)
		header := fmt.Sprintf("%s (%s)", m.styles.title(column.Name), column.Type)
		if edited {
			header += " " + iconEdit
		}
		lines = append(lines, wrapTextToWidth(header, width)...)

		for _, wrappedLine := range wrapTextToWidth(value, valueWidth) {
			lines = append(lines, "  "+wrappedLine)
		}
		if columnIndex < len(m.schema.Columns)-1 {
			lines = append(lines, "")
		}
	}
	return lines
}

func (m *Model) effectiveRecordDetailValue(rowIndex, columnIndex int) (string, bool) {
	if insertIndex, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		if value, ok := m.staging.pendingInserts[insertIndex].values[columnIndex]; ok {
			return displayValue(value.Value), false
		}
		return "", false
	}
	if staged, ok := m.stagedEditForRow(rowIndex, columnIndex); ok {
		return displayValue(staged.Value), true
	}
	return m.visibleRowValue(rowIndex, columnIndex), false
}

func (m *Model) isRowEdited(rowIndex int) bool {
	persistedIndex := m.persistedRowIndex(rowIndex)
	if persistedIndex < 0 {
		return false
	}
	key, ok := m.recordKeyForPersistedRow(persistedIndex)
	if !ok {
		return false
	}
	edits, ok := m.staging.pendingUpdates[key]
	if !ok {
		return false
	}
	return len(edits.changes) > 0
}

func (m *Model) schemaColumnsForRecordsHeader() []string {
	if len(m.schema.Columns) == 0 {
		return nil
	}
	columns := make([]string, len(m.schema.Columns))
	for i, column := range m.schema.Columns {
		label := column.Name
		if m.currentSort != nil && column.Name == m.currentSort.Column {
			switch m.currentSort.Direction {
			case dto.SortDirectionAsc:
				label += " " + iconSortAsc
			case dto.SortDirectionDesc:
				label += " " + iconSortDesc
			}
		}
		columns[i] = label
	}
	return columns
}

func allocateColumnWidths(totalWidth, columns int) []int {
	if columns <= 0 {
		return nil
	}
	separatorWidth := (columns - 1) * textWidth(recordsColumnSeparator)
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

func formatRecordsHeaderRows(values []string, widths []int, styles renderStyles) []string {
	if len(widths) == 0 {
		return nil
	}

	topParts := make([]string, len(widths))
	middleParts := make([]string, len(widths))
	bottomParts := make([]string, len(widths))
	for i, width := range widths {
		value := ""
		if i < len(values) {
			value = values[i]
		}
		value = styles.title(value)
		top, middle, bottom := formatRecordsHeaderCell(value, width)
		topParts[i] = top
		middleParts[i] = middle
		bottomParts[i] = bottom
	}

	return []string{
		strings.Join(topParts, recordsColumnSeparator),
		strings.Join(middleParts, recordsColumnSeparator),
		strings.Join(bottomParts, recordsColumnSeparator),
	}
}

func formatRecordsHeaderCell(value string, columnWidth int) (string, string, string) {
	if columnWidth <= 0 {
		return "", "", ""
	}

	boxWidth := boxWidthForRecordHeaderColumn(columnWidth)
	leftPadding := (columnWidth - boxWidth) / 2
	rightPadding := columnWidth - boxWidth - leftPadding

	top := strings.Repeat(" ", leftPadding) +
		renderFrameEdge(boxWidth, frameTopLeft, frameHorizontal, frameTopRight) +
		strings.Repeat(" ", rightPadding)
	middle := strings.Repeat(" ", leftPadding) +
		renderFrameContent(value, boxWidth) +
		strings.Repeat(" ", rightPadding)
	bottom := strings.Repeat(" ", leftPadding) +
		renderFrameEdge(boxWidth, frameBottomLeft, frameHorizontal, frameBottomRight) +
		strings.Repeat(" ", rightPadding)

	return top, middle, bottom
}

func boxWidthForRecordHeaderColumn(columnWidth int) int {
	if columnWidth <= 0 {
		return 0
	}
	return columnWidth
}

func formatRecordRow(values []string, widths []int, focusColumn int) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		value := ""
		if i < len(values) {
			value = values[i]
		}
		focused := i == focusColumn
		parts[i] = formatRecordCell(value, width, focused)
	}
	return strings.Join(parts, recordsColumnSeparator)
}

func formatRecordCell(value string, width int, focused bool) string {
	if focused {
		if width <= 1 {
			return padRight(">", width)
		}
		innerWidth := width - 2
		value = truncate(value, innerWidth)
		value = padRight(value, innerWidth)
		return "[" + value + "]"
	}
	return padRight(value, width)
}
