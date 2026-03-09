package primitives

import (
	"fmt"
	"strings"
)

type standardizedPopupWidthMode int

const (
	popupWidthClamp standardizedPopupWidthMode = iota
	popupWidthContent
)

type standardizedPopupRow struct {
	text       string
	selectable bool
	selected   bool
}

type standardizedPopupFooter struct {
	left  string
	right string
}

type standardizedPopupSpec struct {
	title               string
	summary             string
	rows                []standardizedPopupRow
	footer              standardizedPopupFooter
	scrollOffset        int
	visibleRows         int
	showScrollIndicator bool
	widthMode           standardizedPopupWidthMode
	defaultWidth        int
	minWidth            int
	maxWidth            int
	styles              renderStyles
}

const (
	popupContentSidePadding = 1
	popupMinHeightPercent   = 40
)

type popupFrameRenderer struct {
	innerWidth   int
	contentWidth int
	leftPadding  int
	rightPadding int
	styles       renderStyles
}

func renderStandardizedPopup(totalWidth, totalHeight int, spec standardizedPopupSpec) []string {
	width := resolvePopupWidth(totalWidth, spec)
	title := strings.TrimSpace(spec.title)
	if title == "" {
		title = "Popup"
	}

	contentInnerWidth := width - 2
	if contentInnerWidth < 1 {
		contentInnerWidth = 1
	}
	frame := newPopupFrameRenderer(contentInnerWidth, spec.styles)

	lines := []string{frame.topBorder(title)}
	if strings.TrimSpace(spec.summary) != "" {
		lines = append(lines, frame.contentLine(spec.styles.summary(spec.summary)))
	}
	lines = append(lines, frame.sectionDivider())

	rows := spec.rows
	visibleRows, offset, maxOffset := popupVisibleRows(len(rows), spec.scrollOffset, spec.visibleRows)
	end := minInt(len(rows), offset+visibleRows)

	for i := offset; i < end; i++ {
		lines = append(lines, frame.rowLine(rows[i]))
	}
	for i := end - offset; i < visibleRows; i++ {
		lines = append(lines, frame.blankLine())
	}

	if spec.showScrollIndicator && maxOffset > 0 {
		indicator := spec.styles.muted(fmt.Sprintf("Scroll: %d/%d", offset+1, maxOffset+1))
		lines = append(lines, frame.contentLine(indicator))
	}

	reservedBottomRows := 1
	if hasPopupFooter(spec.footer) {
		reservedBottomRows++
	}
	minHeight := popupMinHeight(totalHeight)
	for len(lines)+reservedBottomRows < minHeight {
		lines = append(lines, frame.blankLine())
	}

	if hasPopupFooter(spec.footer) {
		lines = append(lines, frame.footerLine(spec.footer))
	}
	lines = append(lines, frame.bottomBorder())
	return lines
}

func popupTextRows(rows []string) []standardizedPopupRow {
	result := make([]standardizedPopupRow, len(rows))
	for i, row := range rows {
		result[i] = standardizedPopupRow{text: row}
	}
	return result
}

func popupSelectableRows(rows []string, selected int) []standardizedPopupRow {
	result := make([]standardizedPopupRow, len(rows))
	for i, row := range rows {
		result[i] = standardizedPopupRow{
			text:       row,
			selectable: true,
			selected:   i == selected,
		}
	}
	return result
}

func newPopupFrameRenderer(innerWidth int, styles renderStyles) popupFrameRenderer {
	leftPadding := popupContentSidePadding
	rightPadding := popupContentSidePadding
	if innerWidth <= (popupContentSidePadding * 2) {
		leftPadding = 0
		rightPadding = 0
	}
	contentWidth := innerWidth - leftPadding - rightPadding
	if contentWidth < 0 {
		contentWidth = 0
	}

	return popupFrameRenderer{
		innerWidth:   innerWidth,
		contentWidth: contentWidth,
		leftPadding:  leftPadding,
		rightPadding: rightPadding,
		styles:       styles,
	}
}

func (r popupFrameRenderer) topBorder(title string) string {
	return renderTitledTopBorder(r.styles.title(title), r.innerWidth)
}

func (r popupFrameRenderer) sectionDivider() string {
	return frameJoinLeft + strings.Repeat(frameHorizontal, r.innerWidth) + frameJoinRight
}

func (r popupFrameRenderer) bottomBorder() string {
	return frameBottomLeft + strings.Repeat(frameHorizontal, r.innerWidth) + frameBottomRight
}

func (r popupFrameRenderer) blankLine() string {
	return r.contentLine("")
}

func (r popupFrameRenderer) contentLine(text string) string {
	content := strings.Repeat(" ", r.leftPadding) + padRight(text, r.contentWidth) + strings.Repeat(" ", r.rightPadding)
	content = padRight(content, r.innerWidth)
	return frameVertical + content + frameVertical
}

func (r popupFrameRenderer) selectedContentLine(text string) string {
	content := strings.Repeat(" ", r.leftPadding) + padRight(text, r.contentWidth) + strings.Repeat(" ", r.rightPadding)
	content = padRight(content, r.innerWidth)
	return frameVertical + r.styles.selected(content) + frameVertical
}

func (r popupFrameRenderer) rowLine(row standardizedPopupRow) string {
	text := row.text
	if row.selectable {
		prefix := selectionUnselectedPrefix()
		if row.selected {
			prefix = selectionSelectedPrefix()
		}
		text = prefix + text
	}
	if row.selected {
		return r.selectedContentLine(text)
	}
	return r.contentLine(text)
}

func (r popupFrameRenderer) footerLine(footer standardizedPopupFooter) string {
	return r.contentLine(renderStatusWithRightHint(footer.left, footer.right, r.contentWidth))
}

func popupVisibleRows(totalRows, scrollOffset, visibleRows int) (int, int, int) {
	if visibleRows <= 0 {
		visibleRows = totalRows
	}
	if visibleRows < 0 {
		visibleRows = 0
	}

	maxOffset := totalRows - visibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := clamp(scrollOffset, 0, maxOffset)
	return visibleRows, offset, maxOffset
}

func hasPopupFooter(footer standardizedPopupFooter) bool {
	return strings.TrimSpace(footer.left) != "" || strings.TrimSpace(footer.right) != ""
}

func resolvePopupWidth(totalWidth int, spec standardizedPopupSpec) int {
	if spec.widthMode == popupWidthContent {
		return resolveContentPopupWidth(totalWidth, spec)
	}
	return clampPopupWidth(totalWidth, spec.defaultWidth, spec.minWidth, spec.maxWidth)
}

func resolveContentPopupWidth(totalWidth int, spec standardizedPopupSpec) int {
	contentWidth := popupContentWidth(spec)
	innerWidth := contentWidth + (popupContentSidePadding * 2)
	if innerWidth < 1 {
		innerWidth = 1
	}

	maxInner := totalWidth - 2
	if maxInner < 1 {
		maxInner = 1
	}
	if innerWidth > maxInner {
		innerWidth = maxInner
	}
	return innerWidth + 2
}

func popupContentWidth(spec standardizedPopupSpec) int {
	maxWidth := textWidth(strings.TrimSpace(spec.title))
	if textWidth(spec.summary) > maxWidth {
		maxWidth = textWidth(spec.summary)
	}
	for _, row := range spec.rows {
		rowWidth := textWidth(row.text)
		if row.selectable {
			rowWidth += textWidth(selectionSelectedPrefix())
		}
		if rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}
	if footerWidth := popupFooterWidth(spec.footer); footerWidth > maxWidth {
		maxWidth = footerWidth
	}
	if indicatorWidth := popupScrollIndicatorWidth(spec); indicatorWidth > maxWidth {
		maxWidth = indicatorWidth
	}
	if maxWidth < 1 {
		maxWidth = 1
	}
	return maxWidth
}

func popupFooterWidth(footer standardizedPopupFooter) int {
	left := strings.TrimSpace(footer.left)
	right := strings.TrimSpace(footer.right)
	switch {
	case left == "" && right == "":
		return 0
	case left == "":
		return textWidth(right)
	case right == "":
		return textWidth(left)
	default:
		return textWidth(left) + 1 + textWidth(right)
	}
}

func popupScrollIndicatorWidth(spec standardizedPopupSpec) int {
	if !spec.showScrollIndicator || spec.visibleRows <= 0 {
		return 0
	}
	_, offset, maxOffset := popupVisibleRows(len(spec.rows), spec.scrollOffset, spec.visibleRows)
	if maxOffset == 0 {
		return 0
	}
	indicator := fmt.Sprintf("Scroll: %d/%d", offset+1, maxOffset+1)
	return textWidth(indicator)
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
