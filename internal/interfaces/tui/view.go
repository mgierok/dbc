package tui

import (
	"strings"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

const (
	recordsColumnSeparator = "  "
	panelBoxGapWidth       = 0
	panelBoxBorderWidth    = 2
	statusBoxSidePadding   = 1
)

func (m *Model) View() string {
	width := m.ui.width
	if width <= 0 {
		width = 80
	}
	height := m.ui.height
	if height <= 0 {
		height = 24
	}

	if m.overlay.helpPopup.active {
		return primitives.CenterBoxLines(m.renderHelpPopup(width), width, height)
	}
	if m.overlay.confirmPopup.active {
		return primitives.CenterBoxLines(m.renderConfirmPopup(width), width, height)
	}
	if m.overlay.editPopup.active {
		return primitives.CenterBoxLines(m.renderEditPopup(width), width, height)
	}
	if m.overlay.filterPopup.active {
		return primitives.CenterBoxLines(m.renderFilterPopup(width), width, height)
	}
	if m.overlay.sortPopup.active {
		return primitives.CenterBoxLines(m.renderSortPopup(width), width, height)
	}

	bodyHeight := m.contentHeight()
	leftWidth, rightWidth := m.panelWidths()

	left := primitives.RenderPanelBox("Tables", m.renderTables(leftWidth, bodyHeight), leftWidth, m.styles)
	right := primitives.RenderPanelBox(m.contentPanelTitle(), m.renderContent(rightWidth, bodyHeight), rightWidth, m.styles)
	lines := primitives.MergePanelBoxes(left, right, leftWidth+panelBoxBorderWidth, rightWidth+panelBoxBorderWidth, panelBoxGapWidth)
	lines = append(lines, m.renderStatusBox(width)...)
	lines = primitives.FitLinesToHeight(lines, height, width)
	return strings.Join(lines, "\n")
}

func (m *Model) panelWidths() (int, int) {
	width := m.ui.width
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
	if m.read.viewMode == ViewRecords {
		if m.overlay.recordDetail.active {
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

	maxWidth := primitives.MaxInt(primitives.TextWidth("Tables"), primitives.TextWidth("No items."))
	longestNameWidth := 0
	for _, table := range m.read.tables {
		longestNameWidth = primitives.MaxInt(longestNameWidth, primitives.TextWidth(table.Name))
	}
	if longestNameWidth == 0 {
		return maxWidth
	}

	tableListWidth := tablePrefixWidth + longestNameWidth + nameMargin
	return primitives.MaxInt(maxWidth, tableListWidth)
}

func (m *Model) renderTables(width, height int) []string {
	items := make([]string, len(m.read.tables))
	for i, table := range m.read.tables {
		items[i] = table.Name
	}

	listLines := primitives.RenderList(items, m.read.selectedTable, height, width, true, m.styles)
	return primitives.PadLines(listLines, height, width)
}

func (m *Model) renderContent(width, height int) []string {
	switch m.read.viewMode {
	case ViewRecords:
		if m.overlay.recordDetail.active {
			return m.renderRecordDetail(width, height)
		}
		return m.renderRecords(width, height)
	default:
		return m.renderSchema(width, height)
	}
}
