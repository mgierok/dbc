package primitives

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func RenderList(items []string, selected, height, width int, focused bool, styles RenderStyles) []string {
	if height < 1 {
		return nil
	}
	if len(items) == 0 {
		return []string{PadRight("No items.", width)}
	}

	start := ScrollStart(selected, height, len(items))
	end := MinInt(len(items), start+height)
	lines := make([]string, 0, height)
	for i := start; i < end; i++ {
		prefix := SelectionUnselectedPrefix()
		if focused && i == selected {
			prefix = SelectionSelectedPrefix()
		}
		line := PadRight(prefix+items[i], width)
		if focused && i == selected {
			line = styles.Selected(line)
		}
		lines = append(lines, line)
	}
	return PadLines(lines, height, width)
}

func SelectionSelectedPrefix() string {
	return IconSelection + " "
}

func SelectionUnselectedPrefix() string {
	return strings.Repeat(" ", TextWidth(SelectionSelectedPrefix()))
}

func RenderPanelBox(title string, content []string, contentWidth int, styles RenderStyles) []string {
	if contentWidth < 1 {
		contentWidth = 1
	}
	top := renderTitledTopBorder(styles.Title(title), contentWidth)
	bottom := FrameBottomLeft + strings.Repeat(FrameHorizontal, contentWidth) + FrameBottomRight

	lines := make([]string, 0, len(content)+2)
	lines = append(lines, top)
	for _, line := range content {
		lines = append(lines, FrameVertical+PadRight(line, contentWidth)+FrameVertical)
	}
	lines = append(lines, bottom)
	return lines
}

func renderTitledTopBorder(title string, contentWidth int) string {
	if contentWidth < 1 {
		contentWidth = 1
	}
	title = Truncate(strings.TrimSpace(title), contentWidth)
	fillWidth := contentWidth - TextWidth(title)
	if fillWidth < 0 {
		fillWidth = 0
	}
	return FrameTopLeft + title + strings.Repeat(FrameHorizontal, fillWidth) + FrameTopRight
}

func MergePanelBoxes(left, right []string, leftWidth, rightWidth, gapWidth int) []string {
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := strings.Repeat(" ", gapWidth)

	maxLines := MaxInt(len(left), len(right))
	lines := make([]string, 0, maxLines)
	for i := 0; i < maxLines; i++ {
		leftLine := strings.Repeat(" ", leftWidth)
		if i < len(left) {
			leftLine = PadRight(left[i], leftWidth)
		}
		rightLine := strings.Repeat(" ", rightWidth)
		if i < len(right) {
			rightLine = PadRight(right[i], rightWidth)
		}
		combined := leftLine + gap + rightLine
		lines = append(lines, combined)
	}
	return lines
}

func FitLinesToHeight(lines []string, height, width int) []string {
	if height < 1 {
		return nil
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for i := range lines {
		lines[i] = PadRight(lines[i], width)
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return lines
}

func CenterBoxLines(lines []string, width, height int) string {
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
		boxWidth = MaxInt(boxWidth, TextWidth(line))
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
		full = append(full, PadRight(strings.Repeat(" ", leftPad)+line, width))
	}
	for len(full) < height {
		full = append(full, strings.Repeat(" ", width))
	}
	return strings.Join(full, "\n")
}

func RenderFrameEdge(width int, leftCorner, horizontal, rightCorner string) string {
	switch {
	case width <= 0:
		return ""
	case width == 1:
		return Truncate(leftCorner, 1)
	case width == 2:
		return Truncate(leftCorner+rightCorner, 2)
	default:
		return leftCorner + strings.Repeat(horizontal, width-2) + rightCorner
	}
}

func RenderFrameContent(value string, width int) string {
	switch {
	case width <= 0:
		return ""
	case width == 1:
		return Truncate(FrameVertical, 1)
	case width == 2:
		return FrameVertical + FrameVertical
	default:
		innerWidth := width - 2
		return FrameVertical + centerText(value, innerWidth) + FrameVertical
	}
}

func centerText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = Truncate(value, width)
	valueWidth := TextWidth(value)
	if valueWidth >= width {
		return value
	}
	leftPadding := (width - valueWidth) / 2
	rightPadding := width - valueWidth - leftPadding
	return strings.Repeat(" ", leftPadding) + value + strings.Repeat(" ", rightPadding)
}

func ScrollStart(selection, height, total int) int {
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

func PadLines(lines []string, height, width int) []string {
	for len(lines) < height {
		lines = append(lines, PadRight("", width))
	}
	return lines
}

func PadRight(text string, width int) string {
	if width <= 0 {
		return ""
	}
	text = Truncate(text, width)
	textLength := TextWidth(text)
	if textLength >= width {
		return text
	}
	return text + strings.Repeat(" ", width-textLength)
}

func Truncate(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if TextWidth(text) <= width {
		return text
	}
	if width <= 3 {
		return ansi.Truncate(text, width, "")
	}
	return ansi.Truncate(text, width, "...")
}

func WrapTextToWidth(text string, width int) []string {
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

func TextWidth(text string) int {
	return ansi.StringWidth(text)
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
