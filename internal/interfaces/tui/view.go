package tui

import "strings"

const (
	recordsColumnSeparator = "  "
	panelBoxGapWidth       = 0
	panelBoxBorderWidth    = 2
	statusBoxSidePadding   = 1
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

	left := renderPanelBox("Tables", m.renderTables(leftWidth, bodyHeight), leftWidth)
	right := renderPanelBox(m.contentPanelTitle(), m.renderContent(rightWidth, bodyHeight), rightWidth)
	lines := mergePanelBoxes(left, right, leftWidth+panelBoxBorderWidth, rightWidth+panelBoxBorderWidth, panelBoxGapWidth)
	lines = append(lines, m.renderStatusBox(width)...)
	lines = fitLinesToHeight(lines, height, width)
	return strings.Join(lines, "\n")
}

func (m *Model) panelWidths() (int, int) {
	width := m.width
	if width <= 0 {
		width = 80
	}
	available := width - (panelBoxBorderWidth * 2) - panelBoxGapWidth
	if available < 2 {
		available = 2
	}

	left := available / 3
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

	right := available - left
	if right < 10 {
		right = 10
		left = available - right
		if left < 10 {
			left = 10
			right = available - left
		}
	}
	if right < 1 {
		right = 1
		left = available - right
	}
	if left < 1 {
		left = 1
		right = available - left
	}
	if right < 1 {
		right = 1
	}
	return left, right
}

func (m *Model) contentPanelTitle() string {
	if m.viewMode == ViewRecords {
		if m.recordDetail.active {
			return "Record Detail"
		}
		return "Records"
	}
	return "Schema"
}

func (m *Model) maxTablePanelWidth() int {
	const (
		tablePrefixWidth = 2
		nameMargin       = 1
	)

	maxWidth := maxInt(textWidth("Tables"), textWidth("No items."))
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
	items := make([]string, len(m.tables))
	for i, table := range m.tables {
		items[i] = table.Name
	}

	listLines := renderList(items, m.selectedTable, height, width, true)
	return padLines(listLines, height, width)
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
