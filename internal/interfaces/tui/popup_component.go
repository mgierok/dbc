package tui

import (
	"fmt"
	"strings"
)

type standardizedPopupSpec struct {
	title               string
	summary             string
	rows                []string
	selected            int
	scrollOffset        int
	visibleRows         int
	showScrollIndicator bool
	defaultWidth        int
	minWidth            int
	maxWidth            int
}

const (
	popupContentSidePadding = 1
	popupMinHeightPercent   = 40
)

func renderStandardizedPopup(totalWidth, totalHeight int, spec standardizedPopupSpec) []string {
	width := clampPopupWidth(totalWidth, spec.defaultWidth, spec.minWidth, spec.maxWidth)
	title := strings.TrimSpace(spec.title)
	if title == "" {
		title = "Popup"
	}

	contentInnerWidth := width - 2
	if contentInnerWidth < 1 {
		contentInnerWidth = 1
	}

	leftPadding := popupContentSidePadding
	rightPadding := popupContentSidePadding
	if contentInnerWidth <= (popupContentSidePadding * 2) {
		leftPadding = 0
		rightPadding = 0
	}
	contentWidth := contentInnerWidth - leftPadding - rightPadding
	if contentWidth < 0 {
		contentWidth = 0
	}

	buildContentLine := func(text string) string {
		content := strings.Repeat(" ", leftPadding) + padRight(text, contentWidth) + strings.Repeat(" ", rightPadding)
		content = padRight(content, contentInnerWidth)
		return frameVertical + content + frameVertical
	}

	topBorder := renderTitledTopBorder(title, width-2)
	bottomBorder := frameBottomLeft + strings.Repeat(frameHorizontal, width-2) + frameBottomRight
	sectionDivider := frameJoinLeft + strings.Repeat(frameHorizontal, width-2) + frameJoinRight

	lines := []string{topBorder}
	if strings.TrimSpace(spec.summary) != "" {
		lines = append(lines, buildContentLine(spec.summary))
	}
	lines = append(lines, sectionDivider)

	rows := spec.rows
	selected := spec.selected
	if len(rows) == 0 && selected >= 0 {
		rows = []string{"No options."}
		selected = -1
	}

	visibleRows := spec.visibleRows
	if visibleRows <= 0 {
		visibleRows = len(rows)
	}
	if visibleRows < 0 {
		visibleRows = 0
	}

	maxOffset := len(rows) - visibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := clamp(spec.scrollOffset, 0, maxOffset)
	end := minInt(len(rows), offset+visibleRows)

	for i := offset; i < end; i++ {
		row := rows[i]
		if selected >= 0 {
			prefix := selectionUnselectedPrefix()
			if i == selected {
				prefix = selectionSelectedPrefix()
			}
			row = prefix + row
		}
		lines = append(lines, buildContentLine(row))
	}
	for i := end - offset; i < visibleRows; i++ {
		lines = append(lines, buildContentLine(""))
	}

	if spec.showScrollIndicator && maxOffset > 0 {
		indicator := fmt.Sprintf("Scroll: %d/%d", offset+1, maxOffset+1)
		lines = append(lines, buildContentLine(indicator))
	}

	minHeight := popupMinHeight(totalHeight)
	for len(lines)+1 < minHeight {
		lines = append(lines, buildContentLine(""))
	}

	lines = append(lines, bottomBorder)
	return lines
}

func popupMinHeight(totalHeight int) int {
	if totalHeight <= 0 {
		totalHeight = 24
	}
	minHeight := (totalHeight*popupMinHeightPercent + 99) / 100
	if minHeight < 1 {
		minHeight = 1
	}
	return minHeight
}

func clampPopupWidth(totalWidth, defaultWidth, minWidth, maxWidth int) int {
	if defaultWidth <= 0 {
		defaultWidth = 50
	}
	if minWidth <= 0 {
		minWidth = 20
	}
	if maxWidth <= 0 {
		maxWidth = 60
	}

	width := totalWidth
	if width <= 0 {
		width = defaultWidth
	}
	if width > maxWidth {
		width = maxWidth
	}
	if width < minWidth {
		width = minWidth
	}
	if width < 2 {
		width = 2
	}
	return width
}
