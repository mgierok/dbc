package tui

import (
	"fmt"
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

	overlayLines := m.activeRuntimeOverlay(width, height)
	styles := m.styles
	if len(overlayLines) > 0 {
		styles = styles.Backdrop()
	}

	lines := m.renderRuntimeLayout(width, height, styles)
	if len(overlayLines) > 0 {
		lines = primitives.OverlayCenteredBoxLines(lines, overlayLines, width, height)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) activeRuntimeOverlay(width, height int) []string {
	switch {
	case m.overlay.helpPopup.active:
		return m.renderHelpPopup(width)
	case m.overlay.confirmPopup.active:
		return m.renderConfirmPopup(width)
	case m.overlay.editPopup.active:
		return m.renderEditPopup(width)
	case m.overlay.filterPopup.active:
		return m.renderFilterPopup(width)
	case m.overlay.sortPopup.active:
		return m.renderSortPopup(width)
	case m.overlay.databaseSelector.active && m.overlay.databaseSelector.controller != nil:
		return m.overlay.databaseSelector.controller.PopupLines(width, height)
	case m.overlay.commandInput.active:
		return m.renderCommandSpotlight(width)
	default:
		return nil
	}
}

func (m *Model) renderRuntimeLayout(width, height int, styles primitives.RenderStyles) []string {
	bodyHeight := m.contentHeight()
	leftWidth, rightWidth := m.panelWidths()

	left := primitives.RenderPanelBox(primitives.SemanticText(primitives.SemanticRoleTitle, "Tables"), m.renderTablesWithStyles(leftWidth, bodyHeight, styles), leftWidth, styles)
	right := primitives.RenderPanelBox(primitives.SemanticText(primitives.SemanticRoleTitle, m.contentPanelTitle()), m.renderContentWithStyles(rightWidth, bodyHeight, styles), rightWidth, styles)
	lines := primitives.MergePanelBoxes(left, right, leftWidth+panelBoxBorderWidth, rightWidth+panelBoxBorderWidth, panelBoxGapWidth)
	lines = append(lines, m.renderStatusBox(width, styles)...)
	return primitives.FitLinesToHeight(lines, height, width)
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
		if m.hasDirtyEdits() {
			return fmt.Sprintf("Records [staged rows: %d]", m.dirtyEditCount())
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
		sanitizedName := primitives.SanitizeDisplayText(table.Name, primitives.DisplaySanitizeSingleLine)
		longestNameWidth = primitives.MaxInt(longestNameWidth, primitives.TextWidth(sanitizedName))
	}
	if longestNameWidth == 0 {
		return maxWidth
	}

	tableListWidth := tablePrefixWidth + longestNameWidth + nameMargin
	return primitives.MaxInt(maxWidth, tableListWidth)
}

func (m *Model) renderTables(width, height int) []string {
	return m.renderTablesWithStyles(width, height, m.styles)
}

func (m *Model) renderTablesWithStyles(width, height int, styles primitives.RenderStyles) []string {
	semanticItems := make([]primitives.SemanticLine, len(m.read.tables))
	for i, table := range m.read.tables {
		semanticItems[i] = primitives.SemanticText(primitives.SemanticRoleBody, table.Name)
	}

	listLines := primitives.RenderList(semanticItems, m.read.selectedTable, height, width, true, styles)
	return primitives.PadLines(listLines, height, width)
}

func (m *Model) renderContent(width, height int) []string {
	return m.renderContentWithStyles(width, height, m.styles)
}

func (m *Model) renderContentWithStyles(width, height int, styles primitives.RenderStyles) []string {
	switch m.read.viewMode {
	case ViewRecords:
		if m.overlay.recordDetail.active {
			if styles == m.styles {
				return m.renderRecordDetail(width, height)
			}
			return m.renderRecordDetailWithStyles(width, height, styles)
		}
		if styles == m.styles {
			return m.renderRecords(width, height)
		}
		return m.renderRecordsWithStyles(width, height, styles)
	default:
		if styles == m.styles {
			return m.renderSchema(width, height)
		}
		return m.renderSchemaWithStyles(width, height, styles)
	}
}
