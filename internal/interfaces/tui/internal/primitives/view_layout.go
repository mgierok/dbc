package primitives

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func renderList(items []string, selected, height, width int, focused bool, styles renderStyles) []string {
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
		prefix := selectionUnselectedPrefix()
		if focused && i == selected {
			prefix = selectionSelectedPrefix()
		}
		line := padRight(prefix+items[i], width)
		if focused && i == selected {
			line = styles.selected(line)
		}
		lines = append(lines, line)
	}
	return padLines(lines, height, width)
}

func selectionSelectedPrefix() string {
	return iconSelection + " "
}

func selectionUnselectedPrefix() string {
	return strings.Repeat(" ", textWidth(selectionSelectedPrefix()))
}

func renderPanelBox(title string, content []string, contentWidth int, styles renderStyles) []string {
	if contentWidth < 1 {
		contentWidth = 1
	}
	top := renderTitledTopBorder(styles.title(title), contentWidth)
	bottom := frameBottomLeft + strings.Repeat(frameHorizontal, contentWidth) + frameBottomRight

	lines := make([]string, 0, len(content)+2)
	lines = append(lines, top)
	for _, line := range content {
		lines = append(lines, frameVertical+padRight(line, contentWidth)+frameVertical)
	}
	lines = append(lines, bottom)
	return lines
}

func renderTitledTopBorder(title string, contentWidth int) string {
	if contentWidth < 1 {
		contentWidth = 1
	}
	title = truncate(strings.TrimSpace(title), contentWidth)
	fillWidth := contentWidth - textWidth(title)
	if fillWidth < 0 {
		fillWidth = 0
	}
	return frameTopLeft + title + strings.Repeat(frameHorizontal, fillWidth) + frameTopRight
}

func mergePanelBoxes(left, right []string, leftWidth, rightWidth, gapWidth int) []string {
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := strings.Repeat(" ", gapWidth)

	maxLines := maxInt(len(left), len(right))
	lines := make([]string, 0, maxLines)
	for i := 0; i < maxLines; i++ {
		leftLine := strings.Repeat(" ", leftWidth)
		if i < len(left) {
			leftLine = padRight(left[i], leftWidth)
		}
		rightLine := strings.Repeat(" ", rightWidth)
		if i < len(right) {
			rightLine = padRight(right[i], rightWidth)
		}
		combined := leftLine + gap + rightLine
		lines = append(lines, combined)
	}
	return lines
}

func fitLinesToHeight(lines []string, height, width int) []string {
	if height < 1 {
		return nil
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for i := range lines {
		lines[i] = padRight(lines[i], width)
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
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

	boxWidth := 0
	for _, line := range lines {
		boxWidth = maxInt(boxWidth, textWidth(line))
	}
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

func renderFrameEdge(width int, leftCorner, horizontal, rightCorner string) string {
	switch {
	case width <= 0:
		return ""
	case width == 1:
		return truncate(leftCorner, 1)
	case width == 2:
		return truncate(leftCorner+rightCorner, 2)
	default:
		return leftCorner + strings.Repeat(horizontal, width-2) + rightCorner
	}
}

func renderFrameContent(value string, width int) string {
	switch {
	case width <= 0:
		return ""
	case width == 1:
		return truncate(frameVertical, 1)
	case width == 2:
		return frameVertical + frameVertical
	default:
		innerWidth := width - 2
		return frameVertical + centerText(value, innerWidth) + frameVertical
	}
}

func centerText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = truncate(value, width)
	valueWidth := textWidth(value)
	if valueWidth >= width {
		return value
	}
	leftPadding := (width - valueWidth) / 2
	rightPadding := width - valueWidth - leftPadding
	return strings.Repeat(" ", leftPadding) + value + strings.Repeat(" ", rightPadding)
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
