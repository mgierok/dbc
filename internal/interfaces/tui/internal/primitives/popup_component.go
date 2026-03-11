package primitives

import (
	"fmt"
	"strings"
)

type StandardizedPopupWidthMode int

const (
	PopupWidthClamp StandardizedPopupWidthMode = iota
	PopupWidthContent
)

type StandardizedPopupRow struct {
	Text       string
	Selectable bool
	Selected   bool
}

type StandardizedPopupFooter struct {
	Left  string
	Right string
}

type StandardizedPopupSpec struct {
	Title               string
	Summary             string
	Rows                []StandardizedPopupRow
	Footer              StandardizedPopupFooter
	ScrollOffset        int
	VisibleRows         int
	ShowScrollIndicator bool
	WidthMode           StandardizedPopupWidthMode
	DefaultWidth        int
	MinWidth            int
	MaxWidth            int
	Styles              RenderStyles
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
	styles       RenderStyles
}

func RenderStandardizedPopup(totalWidth, totalHeight int, spec StandardizedPopupSpec) []string {
	width := resolvePopupWidth(totalWidth, spec)
	title := strings.TrimSpace(spec.Title)
	if title == "" {
		title = "Popup"
	}

	contentInnerWidth := width - 2
	if contentInnerWidth < 1 {
		contentInnerWidth = 1
	}
	frame := newPopupFrameRenderer(contentInnerWidth, spec.Styles)

	lines := []string{frame.topBorder(title)}
	if strings.TrimSpace(spec.Summary) != "" {
		lines = append(lines, frame.contentLine(spec.Styles.Summary(spec.Summary)))
	}
	lines = append(lines, frame.sectionDivider())

	rows := spec.Rows
	visibleRows, offset, maxOffset := popupVisibleRows(len(rows), spec.ScrollOffset, spec.VisibleRows)
	end := MinInt(len(rows), offset+visibleRows)

	for i := offset; i < end; i++ {
		lines = append(lines, frame.rowLine(rows[i]))
	}
	for i := end - offset; i < visibleRows; i++ {
		lines = append(lines, frame.blankLine())
	}

	if spec.ShowScrollIndicator && maxOffset > 0 {
		indicator := spec.Styles.Muted(fmt.Sprintf("Scroll: %d/%d", offset+1, maxOffset+1))
		lines = append(lines, frame.contentLine(indicator))
	}

	reservedBottomRows := 1
	if hasPopupFooter(spec.Footer) {
		reservedBottomRows++
	}
	minHeight := popupMinHeight(totalHeight)
	for len(lines)+reservedBottomRows < minHeight {
		lines = append(lines, frame.blankLine())
	}

	if hasPopupFooter(spec.Footer) {
		lines = append(lines, frame.footerLine(spec.Footer))
	}
	lines = append(lines, frame.bottomBorder())
	return lines
}

func PopupTextRows(rows []string) []StandardizedPopupRow {
	result := make([]StandardizedPopupRow, len(rows))
	for i, row := range rows {
		result[i] = StandardizedPopupRow{Text: row}
	}
	return result
}

func PopupSelectableRows(rows []string, selected int) []StandardizedPopupRow {
	result := make([]StandardizedPopupRow, len(rows))
	for i, row := range rows {
		result[i] = StandardizedPopupRow{
			Text:       row,
			Selectable: true,
			Selected:   i == selected,
		}
	}
	return result
}

func newPopupFrameRenderer(innerWidth int, styles RenderStyles) popupFrameRenderer {
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
	return renderTitledTopBorder(r.styles.Title(title), r.innerWidth)
}

func (r popupFrameRenderer) sectionDivider() string {
	return FrameJoinLeft + strings.Repeat(FrameHorizontal, r.innerWidth) + FrameJoinRight
}

func (r popupFrameRenderer) bottomBorder() string {
	return FrameBottomLeft + strings.Repeat(FrameHorizontal, r.innerWidth) + FrameBottomRight
}

func (r popupFrameRenderer) blankLine() string {
	return r.contentLine("")
}

func (r popupFrameRenderer) contentLine(text string) string {
	content := strings.Repeat(" ", r.leftPadding) + PadRight(text, r.contentWidth) + strings.Repeat(" ", r.rightPadding)
	content = PadRight(content, r.innerWidth)
	return FrameVertical + content + FrameVertical
}

func (r popupFrameRenderer) selectedContentLine(text string) string {
	content := strings.Repeat(" ", r.leftPadding) + PadRight(text, r.contentWidth) + strings.Repeat(" ", r.rightPadding)
	content = PadRight(content, r.innerWidth)
	return FrameVertical + r.styles.Selected(content) + FrameVertical
}

func (r popupFrameRenderer) rowLine(row StandardizedPopupRow) string {
	text := row.Text
	if row.Selectable {
		prefix := SelectionUnselectedPrefix()
		if row.Selected {
			prefix = SelectionSelectedPrefix()
		}
		text = prefix + text
	}
	if row.Selected {
		return r.selectedContentLine(text)
	}
	return r.contentLine(text)
}

func (r popupFrameRenderer) footerLine(footer StandardizedPopupFooter) string {
	return r.contentLine(RenderStatusWithRightHint(footer.Left, footer.Right, r.contentWidth))
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

func hasPopupFooter(footer StandardizedPopupFooter) bool {
	return strings.TrimSpace(footer.Left) != "" || strings.TrimSpace(footer.Right) != ""
}

func resolvePopupWidth(totalWidth int, spec StandardizedPopupSpec) int {
	if spec.WidthMode == PopupWidthContent {
		return resolveContentPopupWidth(totalWidth, spec)
	}
	return clampPopupWidth(totalWidth, spec.DefaultWidth, spec.MinWidth, spec.MaxWidth)
}

func resolveContentPopupWidth(totalWidth int, spec StandardizedPopupSpec) int {
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

func popupContentWidth(spec StandardizedPopupSpec) int {
	maxWidth := TextWidth(strings.TrimSpace(spec.Title))
	if TextWidth(spec.Summary) > maxWidth {
		maxWidth = TextWidth(spec.Summary)
	}
	for _, row := range spec.Rows {
		rowWidth := TextWidth(row.Text)
		if row.Selectable {
			rowWidth += TextWidth(SelectionSelectedPrefix())
		}
		if rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}
	if footerWidth := popupFooterWidth(spec.Footer); footerWidth > maxWidth {
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

func popupFooterWidth(footer StandardizedPopupFooter) int {
	left := strings.TrimSpace(footer.Left)
	right := strings.TrimSpace(footer.Right)
	switch {
	case left == "" && right == "":
		return 0
	case left == "":
		return TextWidth(right)
	case right == "":
		return TextWidth(left)
	default:
		return TextWidth(left) + 1 + TextWidth(right)
	}
}

func popupScrollIndicatorWidth(spec StandardizedPopupSpec) int {
	if !spec.ShowScrollIndicator || spec.VisibleRows <= 0 {
		return 0
	}
	_, offset, maxOffset := popupVisibleRows(len(spec.Rows), spec.ScrollOffset, spec.VisibleRows)
	if maxOffset == 0 {
		return 0
	}
	indicator := fmt.Sprintf("Scroll: %d/%d", offset+1, maxOffset+1)
	return TextWidth(indicator)
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
