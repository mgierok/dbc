package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/mgierok/dbc/internal/application/dto"
)

func (m *Model) View() string {
	width := m.width
	if width <= 0 {
		width = 80
	}
	height := m.height
	if height <= 0 {
		height = 24
	}

	if m.helpPopup.active {
		return centerBoxLines(m.renderHelpPopup(width), width, height)
	}
	if m.confirmPopup.active {
		return centerBoxLines(m.renderConfirmPopup(width), width, height)
	}
	if m.editPopup.active {
		return centerBoxLines(m.renderEditPopup(width), width, height)
	}
	if m.filterPopup.active {
		return centerBoxLines(m.renderFilterPopup(width), width, height)
	}
	if m.sortPopup.active {
		return centerBoxLines(m.renderSortPopup(width), width, height)
	}

	bodyHeight := m.contentHeight()
	leftWidth, rightWidth := m.panelWidths()

	left := m.renderTables(leftWidth, bodyHeight)
	right := m.renderContent(rightWidth, bodyHeight)
	lines := mergePanels(left, right, leftWidth, rightWidth)

	status := m.renderStatus(width)
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

	maxLeft := m.maxTablePanelWidth()
	if maxLeft < 18 {
		maxLeft = 18
	}
	if left > maxLeft {
		left = maxLeft
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

func (m *Model) maxTablePanelWidth() int {
	const (
		tablePrefixWidth = 2
		nameMargin       = 1
	)

	maxWidth := maxInt(textWidth("Tables *"), textWidth("No items."))
	longestNameWidth := 0
	for _, table := range m.tables {
		longestNameWidth = maxInt(longestNameWidth, textWidth(table.Name))
	}
	if longestNameWidth == 0 {
		return maxWidth
	}

	tableListWidth := tablePrefixWidth + longestNameWidth + nameMargin
	return maxInt(maxWidth, tableListWidth)
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
	listLines := renderList(items, m.selectedTable, listHeight, width, m.focus == FocusTables)
	if len(items) > 0 && len(listLines) > 0 && listHeight > 0 {
		selected := clamp(m.selectedTable, 0, len(items)-1)
		start := scrollStart(selected, listHeight, len(items))
		selectedLine := selected - start
		if selectedLine >= 0 && selectedLine < len(listLines) {
			listLines[selectedLine] = bold(listLines[selectedLine])
		}
	}
	lines = append(lines, listLines...)
	return padLines(lines, height, width)
}

func (m *Model) renderContent(width, height int) []string {
	switch m.viewMode {
	case ViewRecords:
		if m.recordDetail.active {
			return m.renderRecordDetail(width, height)
		}
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

	columns := m.schemaColumnsForRecordsHeader()
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

	if m.totalRecordRows() == 0 {
		lines = append(lines, padRight("No records.", width))
		return padLines(lines, height, width)
	}

	totalRows := m.totalRecordRows()
	start := scrollStart(m.recordSelection, listHeight, totalRows)
	end := minInt(totalRows, start+listHeight)
	for i := start; i < end; i++ {
		prefix := "  "
		if m.focus == FocusContent && m.viewMode == ViewRecords && i == m.recordSelection {
			prefix = "> "
		}
		displayValues := make([]string, len(columns))
		edited := make([]bool, len(columns))
		if insertIndex, isInsert := m.pendingInsertIndex(i); isInsert {
			for colIndex := range columns {
				if value, ok := m.pendingInserts[insertIndex].values[colIndex]; ok {
					displayValues[colIndex] = displayValue(value.Value)
				}
			}
		} else {
			for colIndex := range columns {
				if staged, ok := m.stagedEditForRow(i, colIndex); ok {
					displayValues[colIndex] = displayValue(staged.Value)
					edited[colIndex] = true
				} else {
					displayValues[colIndex] = m.visibleRowValue(i, colIndex)
				}
			}
		}
		focusColumn := -1
		if m.recordFieldFocus && i == m.recordSelection {
			focusColumn = m.recordColumn
		}
		row := formatRecordRow(displayValues, columnWidths, focusColumn, edited)
		rowTag := ""
		if _, isInsert := m.pendingInsertIndex(i); isInsert {
			rowTag = "[INS] "
		} else if m.isRowMarkedDelete(i) {
			rowTag = "[DEL] "
		}
		lines = append(lines, padRight(prefix+rowTag+row, width))
	}
	return padLines(lines, height, width)
}

func (m *Model) renderRecordDetail(width, height int) []string {
	title := "Record Detail"
	if m.focus == FocusContent && m.viewMode == ViewRecords {
		title = "Record Detail *"
	}
	lines := []string{padRight(title, width)}

	listHeight := height - 1
	if listHeight < 1 {
		return padLines(lines, height, width)
	}

	contentLines := m.recordDetailContentLines(width)
	if len(contentLines) == 0 {
		lines = append(lines, padRight("No detail available.", width))
		return padLines(lines, height, width)
	}

	maxOffset := len(contentLines) - listHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := clamp(m.recordDetail.scrollOffset, 0, maxOffset)
	end := minInt(len(contentLines), offset+listHeight)

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
	rowLine := "ⓘ Persisted record"
	if _, isInsert := m.pendingInsertIndex(rowIndex); isInsert {
		rowLine = "ⓘ [INS] Pending insert"
	} else if m.isRowMarkedDelete(rowIndex) {
		rowLine = "ⓘ [DEL] Marked for delete"
	}
	lines = append(lines, wrapTextToWidth(rowLine, width)...)
	lines = append(lines, "")

	valueWidth := width - 2
	if valueWidth < 1 {
		valueWidth = 1
	}

	for columnIndex, column := range m.schema.Columns {
		value, edited := m.effectiveRecordDetailValue(rowIndex, columnIndex)
		header := fmt.Sprintf("%s (%s)", bold(column.Name), column.Type)
		if edited {
			header += " *"
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
		if value, ok := m.pendingInserts[insertIndex].values[columnIndex]; ok {
			return displayValue(value.Value), false
		}
		return "", false
	}
	if staged, ok := m.stagedEditForRow(rowIndex, columnIndex); ok {
		return displayValue(staged.Value), true
	}
	return m.visibleRowValue(rowIndex, columnIndex), false
}

func (m *Model) renderHelpPopup(totalWidth int) []string {
	return renderStandardizedPopup(totalWidth, standardizedPopupSpec{
		title:               "Help",
		summary:             runtimeHelpPopupSummaryLine(),
		rows:                helpPopupContentLines(),
		selected:            -1,
		scrollOffset:        m.helpPopup.scrollOffset,
		visibleRows:         m.helpPopupVisibleLines(),
		showScrollIndicator: true,
		defaultWidth:        50,
		minWidth:            20,
		maxWidth:            60,
	})
}

func helpPopupContentLines() []string {
	return runtimeHelpPopupContentLines()
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
			rows[i] = fmt.Sprintf("%s (%s)", operator.Name, operator.SQL)
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

	return renderStandardizedPopup(totalWidth, standardizedPopupSpec{
		title:        "Filter",
		summary:      stepLabel,
		rows:         rows,
		selected:     selected,
		defaultWidth: 60,
		minWidth:     20,
		maxWidth:     60,
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

	return renderStandardizedPopup(totalWidth, standardizedPopupSpec{
		title:        "Sort",
		summary:      stepLabel,
		rows:         rows,
		selected:     selected,
		defaultWidth: 60,
		minWidth:     20,
		maxWidth:     60,
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
		rows = append(rows, "Error: "+m.editPopup.errorMessage)
	}

	return renderStandardizedPopup(totalWidth, standardizedPopupSpec{
		title:        "Edit Cell",
		summary:      fmt.Sprintf("%s | %s", columnLabel, nullableLabel),
		rows:         rows,
		selected:     selected,
		defaultWidth: 60,
		minWidth:     30,
		maxWidth:     60,
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

	return renderStandardizedPopup(totalWidth, standardizedPopupSpec{
		title:        title,
		summary:      message,
		rows:         options,
		selected:     selected,
		defaultWidth: 50,
		minWidth:     20,
		maxWidth:     60,
	})
}

func (m *Model) renderStatus(width int) string {
	if width <= 0 {
		width = 80
	}
	mode := "READ-ONLY"
	if m.hasDirtyEdits() {
		mode = fmt.Sprintf("WRITE (dirty: %d)", m.dirtyEditCount())
	}
	parts := []string{
		mode,
		fmt.Sprintf("View: %s", m.viewModeLabel()),
		fmt.Sprintf("Table: %s", m.currentTableName()),
		m.filterSummary(),
		m.sortSummary(),
	}
	if m.commandInput.active {
		parts = append(parts, "Command: "+m.commandPrompt())
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

func (m *Model) sortSummary() string {
	if m.currentSort == nil {
		return "Sort: none"
	}
	return fmt.Sprintf("Sort: %s %s", m.currentSort.Column, m.currentSort.Direction)
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
				label += " ↑"
			case dto.SortDirectionDesc:
				label += " ↓"
			}
		}
		columns[i] = label
	}
	return columns
}

func (m *Model) statusShortcuts() string {
	switch {
	case m.editPopup.active:
		return runtimeStatusEditShortcuts()
	case m.confirmPopup.active:
		return runtimeStatusConfirmShortcuts(len(m.confirmPopup.options) > 0)
	case m.filterPopup.active:
		return runtimeStatusFilterPopupShortcuts()
	case m.sortPopup.active:
		return runtimeStatusSortPopupShortcuts()
	case m.helpPopup.active:
		return runtimeStatusHelpPopupShortcuts()
	case m.commandInput.active:
		return runtimeStatusCommandInputShortcuts()
	case m.recordDetail.active:
		return runtimeStatusRecordDetailShortcuts()
	case m.focus == FocusTables:
		return runtimeStatusTablesShortcuts()
	case m.focus == FocusContent && m.viewMode == ViewSchema:
		return runtimeStatusSchemaShortcuts()
	case m.focus == FocusContent && m.viewMode == ViewRecords:
		return runtimeStatusRecordsShortcuts()
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

func centerBoxLines(lines []string, width, height int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if len(lines) == 0 {
		lines = []string{""}
	}

	boxHeight := len(lines)
	if boxHeight > height {
		boxHeight = height
		lines = lines[:boxHeight]
	}

	boxWidth := len(lines[0])
	leftPad := 0
	if width > boxWidth {
		leftPad = (width - boxWidth) / 2
	}
	topPad := 0
	if height > boxHeight {
		topPad = (height - boxHeight) / 2
	}

	full := make([]string, 0, height)
	for i := 0; i < topPad; i++ {
		full = append(full, strings.Repeat(" ", width))
	}
	for _, line := range lines {
		full = append(full, padRight(strings.Repeat(" ", leftPad)+line, width))
	}
	for len(full) < height {
		full = append(full, strings.Repeat(" ", width))
	}
	return strings.Join(full, "\n")
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

func formatRecordRow(values []string, widths []int, focusColumn int, edited []bool) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		value := ""
		if i < len(values) {
			value = values[i]
		}
		editedCell := false
		if i < len(edited) {
			editedCell = edited[i]
		}
		focused := i == focusColumn
		parts[i] = formatRecordCell(value, width, focused, editedCell)
	}
	return strings.Join(parts, " | ")
}

func formatRecordCell(value string, width int, focused, edited bool) string {
	if edited && width > 0 {
		value += "*"
	}
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
	textLength := textWidth(text)
	if textLength >= width {
		return text
	}
	return text + strings.Repeat(" ", width-textLength)
}

func truncate(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if textWidth(text) <= width {
		return text
	}
	if width <= 3 {
		return ansi.Truncate(text, width, "")
	}
	return ansi.Truncate(text, width, "...")
}

func wrapTextToWidth(text string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if text == "" {
		return []string{""}
	}

	segments := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "" {
			lines = append(lines, "")
			continue
		}
		remaining := segment
		for ansi.StringWidth(remaining) > width {
			line := ansi.Truncate(remaining, width, "")
			if line == "" {
				break
			}
			lines = append(lines, line)
			remaining = ansi.Cut(remaining, width, ansi.StringWidth(remaining))
		}
		lines = append(lines, remaining)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func textWidth(text string) int {
	return ansi.StringWidth(text)
}

func bold(text string) string {
	return "\x1b[1m" + text + "\x1b[0m"
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
