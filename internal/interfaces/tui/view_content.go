package tui

import (
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) renderSchema(width, height int) []string {
	return m.renderSchemaWithStyles(width, height, m.styles)
}

func (m *Model) renderSchemaWithStyles(width, height int, styles primitives.RenderStyles) []string {
	if len(m.read.schema.Columns) == 0 {
		return primitives.PadLines([]string{primitives.PadRight(styles.Render(primitives.SemanticRoleBody, "No schema loaded."), width)}, height, width)
	}

	items := make([]primitives.SemanticLine, len(m.read.schema.Columns))
	for i, column := range m.read.schema.Columns {
		items[i] = primitives.SemanticLine{
			primitives.Span(primitives.SemanticRoleHeader, column.Name),
			primitives.Span(primitives.SemanticRoleBody, " : "+column.Type),
		}
	}
	lines := primitives.RenderList(items, m.read.schemaIndex, height, width, m.read.focus == FocusContent && m.read.viewMode == ViewSchema, styles)
	return primitives.PadLines(lines, height, width)
}

func (m *Model) renderRecords(width, height int) []string {
	return m.renderRecordsWithStyles(width, height, m.styles)
}

func (m *Model) renderRecordsWithStyles(width, height int, styles primitives.RenderStyles) []string {
	lines := make([]string, 0, height)

	columns := m.schemaColumnsForRecordsHeader()
	if len(columns) == 0 {
		lines = append(lines, primitives.PadRight(styles.Render(primitives.SemanticRoleBody, "No columns loaded."), width))
		return primitives.PadLines(lines, height, width)
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
	headerRows := formatRecordsHeaderRows(columns, columnWidths, styles)
	for _, headerRow := range headerRows {
		lines = append(lines, primitives.PadRight(headerPrefix+headerRow, width))
	}

	listHeight := height - len(headerRows)
	if listHeight < 1 {
		return primitives.PadLines(lines, height, width)
	}

	if m.totalRecordRows() == 0 {
		lines = append(lines, primitives.PadRight(styles.Render(primitives.SemanticRoleBody, "No records."), width))
		return primitives.PadLines(lines, height, width)
	}

	totalRows := m.totalRecordRows()
	start := primitives.ScrollStart(m.read.recordSelection, listHeight, totalRows)
	end := primitives.MinInt(totalRows, start+listHeight)
	for i := start; i < end; i++ {
		prefix := primitives.SelectionUnselectedPrefix()
		selected := m.read.focus == FocusContent && m.read.viewMode == ViewRecords && i == m.read.recordSelection
		if selected {
			prefix = primitives.SelectionSelectedPrefix()
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
		if m.read.recordFieldFocus && i == m.read.recordSelection {
			focusColumn = m.read.recordColumn
		}
		row := formatRecordRow(displayValues, columnWidths, focusColumn)
		rowMarker := m.recordRowMarker(i)
		lineRole := primitives.SemanticRoleBody
		if m.isRowMarkedDelete(i) {
			lineRole = primitives.SemanticRoleDeleted
			if selected {
				lineRole = primitives.SemanticRoleSelectedDeleted
			}
		} else if selected {
			lineRole = primitives.SemanticRoleSelected
		}
		line := styles.Render(lineRole, primitives.PadRight(prefix+rowMarker+" "+row, width))
		lines = append(lines, line)
	}
	return primitives.PadLines(lines, height, width)
}

func (m *Model) recordRowMarker(rowIndex int) string {
	if _, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		return primitives.IconInsert
	}
	if m.isRowMarkedDelete(rowIndex) {
		return primitives.IconDelete
	}
	if m.isRowEdited(rowIndex) {
		return primitives.IconEdit
	}
	return " "
}

func (m *Model) renderRecordDetail(width, height int) []string {
	return m.renderRecordDetailWithStyles(width, height, m.styles)
}

func (m *Model) renderRecordDetailWithStyles(width, height int, styles primitives.RenderStyles) []string {
	lines := make([]string, 0, height)

	contentLines := m.recordDetailContentLinesWithStyles(width, styles)
	if len(contentLines) == 0 {
		lines = append(lines, primitives.PadRight(styles.Render(primitives.SemanticRoleBody, "No detail available."), width))
		return primitives.PadLines(lines, height, width)
	}

	maxOffset := len(contentLines) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := clamp(m.overlay.recordDetail.scrollOffset, 0, maxOffset)
	end := primitives.MinInt(len(contentLines), offset+height)

	for i := offset; i < end; i++ {
		lines = append(lines, primitives.PadRight(contentLines[i], width))
	}

	return primitives.PadLines(lines, height, width)
}

func (m *Model) recordDetailContentLines(width int) []string {
	return m.recordDetailContentLinesWithStyles(width, m.styles)
}

func (m *Model) recordDetailContentLinesWithStyles(width int, styles primitives.RenderStyles) []string {
	if m.totalRecordRows() == 0 {
		return []string{styles.Render(primitives.SemanticRoleBody, "No records.")}
	}
	if len(m.read.schema.Columns) == 0 {
		return []string{styles.Render(primitives.SemanticRoleBody, "No columns loaded.")}
	}

	rowIndex := clamp(m.read.recordSelection, 0, m.totalRecordRows()-1)
	lines := make([]string, 0, len(m.read.schema.Columns)*4)
	deleted := m.isRowMarkedDelete(rowIndex)
	rowLine := primitives.IconInfo + " Persisted record"
	if _, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		rowLine = primitives.IconInfo + " Pending insert"
	} else if deleted {
		rowLine = primitives.IconInfo + " Marked for delete"
	} else if m.isRowEdited(rowIndex) {
		rowLine = primitives.IconInfo + " Edited record"
	}
	lines = append(lines, primitives.WrapTextToWidth(styles.Render(primitives.SemanticRoleSummary, rowLine), width)...)
	lines = append(lines, "")

	valueWidth := width - 2
	if valueWidth < 1 {
		valueWidth = 1
	}

	for columnIndex, column := range m.read.schema.Columns {
		value, edited := m.effectiveRecordDetailValue(rowIndex, columnIndex)
		header := styles.RenderLine(primitives.SemanticLine{
			primitives.Span(primitives.SemanticRoleHeader, column.Name),
			primitives.Span(primitives.SemanticRoleBody, " ("+column.Type+")"),
		})
		if edited {
			header += " " + styles.Render(primitives.SemanticRoleBody, primitives.IconEdit)
		}
		lines = append(lines, primitives.WrapTextToWidth(header, width)...)

		valueRole := primitives.SemanticRoleBody
		if deleted {
			valueRole = primitives.SemanticRoleDeleted
		}
		for _, wrappedLine := range primitives.WrapTextToWidth(styles.Render(valueRole, value), valueWidth) {
			lines = append(lines, "  "+wrappedLine)
		}
		if columnIndex < len(m.read.schema.Columns)-1 {
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
	if len(m.read.schema.Columns) == 0 {
		return nil
	}
	columns := make([]string, len(m.read.schema.Columns))
	for i, column := range m.read.schema.Columns {
		label := column.Name
		if m.read.currentSort != nil && column.Name == m.read.currentSort.Column {
			switch m.read.currentSort.Direction {
			case dto.SortDirectionAsc:
				label += " " + primitives.IconSortAsc
			case dto.SortDirectionDesc:
				label += " " + primitives.IconSortDesc
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
	separatorWidth := (columns - 1) * primitives.TextWidth(recordsColumnSeparator)
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

func formatRecordsHeaderRows(values []string, widths []int, styles primitives.RenderStyles) []string {
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
		value = styles.Render(primitives.SemanticRoleHeader, value)
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
		primitives.RenderFrameEdge(boxWidth, primitives.FrameTopLeft, primitives.FrameHorizontal, primitives.FrameTopRight) +
		strings.Repeat(" ", rightPadding)
	middle := strings.Repeat(" ", leftPadding) +
		primitives.RenderFrameContent(value, boxWidth) +
		strings.Repeat(" ", rightPadding)
	bottom := strings.Repeat(" ", leftPadding) +
		primitives.RenderFrameEdge(boxWidth, primitives.FrameBottomLeft, primitives.FrameHorizontal, primitives.FrameBottomRight) +
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
			return primitives.PadRight(">", width)
		}
		innerWidth := width - 2
		value = primitives.Truncate(value, innerWidth)
		value = primitives.PadRight(value, innerWidth)
		return "[" + value + "]"
	}
	return primitives.PadRight(value, width)
}
