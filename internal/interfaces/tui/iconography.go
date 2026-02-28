package tui

const (
	iconInsert       = "✚"
	iconEdit         = "✱"
	iconDelete       = "✖"
	iconSortAsc      = "↑"
	iconSortDesc     = "↓"
	iconConfigSource = "⚙"
	iconCLISource    = "⌨"
	iconInfo         = "ℹ"
	iconActivePrefix = "➤"

	// Outer frame uses the existing heavy style.
	outerFrameVertical   = "┃"
	outerFrameHorizontal = "━"

	outerFrameTopLeft     = "┏"
	outerFrameTopRight    = "┓"
	outerFrameBottomLeft  = "┗"
	outerFrameBottomRight = "┛"
	outerFrameJoinLeft    = "┣"
	outerFrameJoinRight   = "┫"
	outerFrameJoinTop     = "┳"
	outerFrameJoinBottom  = "┻"
	outerFrameJoinCenter  = "╋"

	outerFrameSegmentSeparator = " ┃ "
)

// Inner frame provides a separate light-style set.
const (
	innerFrameVertical   = "│"
	innerFrameHorizontal = "─"

	innerFrameTopLeft     = "┌"
	innerFrameTopRight    = "┐"
	innerFrameBottomLeft  = "└"
	innerFrameBottomRight = "┘"
	innerFrameJoinLeft    = "├"
	innerFrameJoinRight   = "┤"
	innerFrameJoinTop     = "┬"
	innerFrameJoinBottom  = "┴"
	innerFrameJoinCenter  = "┼"
)
