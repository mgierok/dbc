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

func renderStandardizedPopup(totalWidth int, spec standardizedPopupSpec) []string {
	width := clampPopupWidth(totalWidth, spec.defaultWidth, spec.minWidth, spec.maxWidth)
	title := strings.TrimSpace(spec.title)
	if title == "" {
		title = "Popup"
	}

	topBorder := frameTopLeft + strings.Repeat(frameHorizontal, width-2) + frameTopRight
	bottomBorder := frameBottomLeft + strings.Repeat(frameHorizontal, width-2) + frameBottomRight
	sectionDivider := frameJoinLeft + strings.Repeat(frameHorizontal, width-2) + frameJoinRight

	lines := []string{topBorder}
	lines = append(lines, frameVertical+padRight(title, width-2)+frameVertical)
	if strings.TrimSpace(spec.summary) != "" {
		lines = append(lines, frameVertical+padRight(spec.summary, width-2)+frameVertical)
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
			prefix := "  "
			if i == selected {
				prefix = "> "
			}
			row = prefix + row
		}
		lines = append(lines, frameVertical+padRight(row, width-2)+frameVertical)
	}
	for i := end - offset; i < visibleRows; i++ {
		lines = append(lines, frameVertical+padRight("", width-2)+frameVertical)
	}

	if spec.showScrollIndicator && maxOffset > 0 {
		indicator := fmt.Sprintf("Scroll: %d/%d", offset+1, maxOffset+1)
		lines = append(lines, frameVertical+padRight(indicator, width-2)+frameVertical)
	}

	lines = append(lines, bottomBorder)
	return lines
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
