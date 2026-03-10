package primitives

const (
	IconInsert            = iconInsert
	IconEdit              = iconEdit
	IconDelete            = iconDelete
	IconSortAsc           = iconSortAsc
	IconSortDesc          = iconSortDesc
	IconConfigSource      = iconConfigSource
	IconCLISource         = iconCLISource
	IconInfo              = iconInfo
	IconSelection         = iconSelection
	FrameVertical         = frameVertical
	FrameHorizontal       = frameHorizontal
	FrameTopLeft          = frameTopLeft
	FrameTopRight         = frameTopRight
	FrameBottomLeft       = frameBottomLeft
	FrameBottomRight      = frameBottomRight
	FrameJoinLeft         = frameJoinLeft
	FrameJoinRight        = frameJoinRight
	FrameJoinTop          = frameJoinTop
	FrameJoinBottom       = frameJoinBottom
	FrameJoinCenter       = frameJoinCenter
	FrameSegmentSeparator = frameSegmentSeparator
)

func RenderList(items []string, selected, height, width int, focused bool, styles RenderStyles) []string {
	return renderList(items, selected, height, width, focused, styles.inner)
}

func SelectionSelectedPrefix() string {
	return selectionSelectedPrefix()
}

func SelectionUnselectedPrefix() string {
	return selectionUnselectedPrefix()
}

func RenderPanelBox(title string, content []string, contentWidth int, styles RenderStyles) []string {
	return renderPanelBox(title, content, contentWidth, styles.inner)
}

func MergePanelBoxes(left, right []string, leftWidth, rightWidth, gapWidth int) []string {
	return mergePanelBoxes(left, right, leftWidth, rightWidth, gapWidth)
}

func FitLinesToHeight(lines []string, height, width int) []string {
	return fitLinesToHeight(lines, height, width)
}

func CenterBoxLines(lines []string, width, height int) string {
	return centerBoxLines(lines, width, height)
}

func RenderFrameEdge(width int, leftCorner, horizontal, rightCorner string) string {
	return renderFrameEdge(width, leftCorner, horizontal, rightCorner)
}

func RenderFrameContent(value string, width int) string {
	return renderFrameContent(value, width)
}

func ScrollStart(selection, height, total int) int {
	return scrollStart(selection, height, total)
}

func PadLines(lines []string, height, width int) []string {
	return padLines(lines, height, width)
}

func PadRight(text string, width int) string {
	return padRight(text, width)
}

func Truncate(text string, width int) string {
	return truncate(text, width)
}

func WrapTextToWidth(text string, width int) []string {
	return wrapTextToWidth(text, width)
}

func TextWidth(text string) int {
	return textWidth(text)
}

func MinInt(a, b int) int {
	return minInt(a, b)
}

func MaxInt(a, b int) int {
	return maxInt(a, b)
}
