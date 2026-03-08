package tui

import (
	"fmt"
	"strings"
)

func (m *Model) renderStatus(width int) string {
	if width <= 0 {
		width = 80
	}
	mode := "READ-ONLY"
	if m.hasDirtyEdits() {
		mode = m.styles.dirty(fmt.Sprintf("WRITE (dirty: %d)", m.dirtyEditCount()))
	}
	parts := []string{
		mode,
		m.statusSegment("Table", m.currentTableName()),
	}
	if m.viewMode == ViewRecords {
		parts = append(parts, m.recordsSummary(), m.pageSummary())
	}
	parts = append(parts, m.filterSummary(), m.sortSummary())
	if m.commandInput.active {
		parts = append(parts, m.statusSegment("Command", m.commandPrompt()))
	}
	if strings.TrimSpace(m.statusMessage) != "" {
		parts = append(parts, m.styleStatusMessage(m.statusMessage))
	}
	left := strings.Join(parts, frameSegmentSeparator)
	return renderStatusWithRightHint(left, m.styles.muted(runtimeStatusContextHelpHint()), width)
}

func (m *Model) filterSummary() string {
	if m.currentFilter == nil {
		return m.statusSegment("Filter", "none")
	}
	if m.currentFilter.Operator.RequiresValue {
		return m.statusSegment("Filter", fmt.Sprintf("%s %s %s", m.currentFilter.Column, m.currentFilter.Operator.Name, m.currentFilter.Value))
	}
	return m.statusSegment("Filter", fmt.Sprintf("%s %s", m.currentFilter.Column, m.currentFilter.Operator.Name))
}

func (m *Model) sortSummary() string {
	if m.currentSort == nil {
		return m.statusSegment("Sort", "none")
	}
	return m.statusSegment("Sort", fmt.Sprintf("%s %s", m.currentSort.Column, m.currentSort.Direction))
}

func (m *Model) recordsSummary() string {
	return m.statusSegment("Records", fmt.Sprintf("%d/%d", len(m.records), m.recordTotalCount))
}

func (m *Model) pageSummary() string {
	currentPage := clamp(m.recordPageIndex+1, 1, maxInt(1, m.recordTotalPages))
	return m.statusSegment("Page", fmt.Sprintf("%d/%d", currentPage, maxInt(1, m.recordTotalPages)))
}

func (m *Model) statusSegment(label, value string) string {
	return m.styles.label(label+":") + " " + value
}

func (m *Model) styleStatusMessage(message string) string {
	if isErrorLikeMessage(message) {
		return m.styles.error(message)
	}
	return message
}

func renderStatusWithRightHint(left, right string, width int) string {
	if width <= 0 {
		width = 80
	}

	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if right == "" {
		return padRight(left, width)
	}

	right = truncate(right, width)
	rightWidth := textWidth(right)
	if rightWidth >= width {
		return padRight(right, width)
	}

	leftWidth := width - rightWidth - 1
	if leftWidth <= 0 {
		return padRight(right, width)
	}

	if left == "" {
		return padRight(strings.Repeat(" ", leftWidth+1)+right, width)
	}

	return padRight(truncate(left, leftWidth), leftWidth) + " " + right
}

func (m *Model) renderStatusBox(width int) []string {
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

	status := m.renderStatus(statusWidth)
	content := strings.Repeat(" ", leftPadding) + padRight(status, statusWidth) + strings.Repeat(" ", rightPadding)
	content = padRight(content, innerWidth)
	return []string{
		frameTopLeft + strings.Repeat(frameHorizontal, innerWidth) + frameTopRight,
		frameVertical + content + frameVertical,
		frameBottomLeft + strings.Repeat(frameHorizontal, innerWidth) + frameBottomRight,
	}
}
