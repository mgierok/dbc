package tui

import (
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) renderStatus(width int) string {
	return m.renderStatusWithStyles(width, m.styles)
}

func (m *Model) renderStatusWithStyles(width int, styles primitives.RenderStyles) string {
	if width <= 0 {
		width = 80
	}
	mode := primitives.SemanticText(primitives.SemanticRoleBody, primitives.IconClean)
	if m.hasDirtyEdits() {
		mode = primitives.SemanticText(primitives.SemanticRoleDirty, primitives.IconEdit)
	}
	parts := []primitives.SemanticLine{
		mode,
		m.statusSegment("Table", m.currentTableName()),
	}
	if m.read.viewMode == ViewRecords {
		parts = append(parts, m.recordsSummary(), m.pageSummary())
	}
	parts = append(parts, m.filterSummary(), m.sortSummary())
	if strings.TrimSpace(m.ui.statusMessage) != "" {
		parts = append(parts, m.styleStatusMessage(m.ui.statusMessage))
	}
	left := primitives.JoinSemanticLines(parts, primitives.FrameSegmentSeparator, primitives.SemanticRoleBody)
	return primitives.RenderSemanticStatusWithRightHint(left, primitives.SemanticText(primitives.SemanticRoleMuted, primitives.RuntimeStatusContextHelpHint()), styles, width)
}

func (m *Model) filterSummary() primitives.SemanticLine {
	if m.read.currentFilter == nil {
		return m.statusSegment("Filter", "none")
	}
	if m.read.currentFilter.Operator.RequiresValue {
		return m.statusSegment("Filter", fmt.Sprintf("%s %s %s", m.read.currentFilter.Column, m.read.currentFilter.Operator.Name, m.read.currentFilter.Value))
	}
	return m.statusSegment("Filter", fmt.Sprintf("%s %s", m.read.currentFilter.Column, m.read.currentFilter.Operator.Name))
}

func (m *Model) sortSummary() primitives.SemanticLine {
	if m.read.currentSort == nil {
		return m.statusSegment("Sort", "none")
	}
	return m.statusSegment("Sort", fmt.Sprintf("%s %s", m.read.currentSort.Column, m.read.currentSort.Direction))
}

func (m *Model) recordsSummary() primitives.SemanticLine {
	return m.statusSegment("Records", fmt.Sprintf("%d/%d", len(m.read.records), m.read.recordTotalCount))
}

func (m *Model) pageSummary() primitives.SemanticLine {
	currentPage := clamp(m.read.recordPageIndex+1, 1, primitives.MaxInt(1, m.read.recordTotalPages))
	return m.statusSegment("Page", fmt.Sprintf("%d/%d", currentPage, primitives.MaxInt(1, m.read.recordTotalPages)))
}

func (m *Model) statusSegment(label, value string) primitives.SemanticLine {
	return primitives.SemanticLine{
		primitives.Span(primitives.SemanticRoleLabel, label+":"),
		primitives.Span(primitives.SemanticRoleBody, " "+value),
	}
}

func (m *Model) styleStatusMessage(message string) primitives.SemanticLine {
	sanitizedMessage := primitives.SanitizeDisplayText(message, primitives.DisplaySanitizeSingleLine)
	if primitives.IsErrorLikeMessage(sanitizedMessage) {
		return primitives.SemanticText(primitives.SemanticRoleError, sanitizedMessage)
	}
	return primitives.SemanticText(primitives.SemanticRoleBody, sanitizedMessage)
}

func (m *Model) renderStatusBox(width int, styles primitives.RenderStyles) []string {
	if width <= 0 {
		width = 80
	}
	innerWidth := width - panelBoxBorderWidth
	if innerWidth < 1 {
		innerWidth = 1
	}
	leftPadding := statusBoxSidePadding
	rightPadding := statusBoxSidePadding
	if innerWidth == 1 {
		leftPadding = 0
		rightPadding = 0
	}
	statusWidth := innerWidth - leftPadding - rightPadding
	if statusWidth < 0 {
		statusWidth = 0
	}

	status := m.renderStatusWithStyles(statusWidth, styles)
	content := strings.Repeat(" ", leftPadding) + primitives.PadRight(status, statusWidth) + strings.Repeat(" ", rightPadding)
	content = primitives.PadRight(content, innerWidth)
	return []string{
		primitives.FrameTopLeft + strings.Repeat(primitives.FrameHorizontal, innerWidth) + primitives.FrameTopRight,
		primitives.FrameVertical + content + primitives.FrameVertical,
		primitives.FrameBottomLeft + strings.Repeat(primitives.FrameHorizontal, innerWidth) + primitives.FrameBottomRight,
	}
}
