package selector

import "github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"

type renderStyles struct {
	enabled bool
}

var detectRenderStyles = resolveRenderStylesFromEnv

func resolveRenderStylesFromEnv() renderStyles {
	styles := primitives.ResolveRenderStylesFromEnv()
	return renderStyles{enabled: styles.Enabled()}
}

func (s renderStyles) toPrimitives() primitives.RenderStyles {
	return primitives.NewRenderStyles(s.enabled)
}

func (s renderStyles) muted(text string) string {
	return s.toPrimitives().Muted(text)
}

func (s renderStyles) error(text string) string {
	return s.toPrimitives().Error(text)
}

const (
	iconConfigSource      = primitives.IconConfigSource
	iconCLISource         = primitives.IconCLISource
	iconSelection         = primitives.IconSelection
	frameVertical         = primitives.FrameVertical
	frameTopLeft          = primitives.FrameTopLeft
	frameTopRight         = primitives.FrameTopRight
	frameBottomLeft       = primitives.FrameBottomLeft
	frameBottomRight      = primitives.FrameBottomRight
	frameJoinLeft         = primitives.FrameJoinLeft
	frameJoinRight        = primitives.FrameJoinRight
	frameSegmentSeparator = primitives.FrameSegmentSeparator
)

type standardizedPopupWidthMode int

const (
	popupWidthClamp   standardizedPopupWidthMode = standardizedPopupWidthMode(primitives.PopupWidthClamp)
	popupWidthContent standardizedPopupWidthMode = standardizedPopupWidthMode(primitives.PopupWidthContent)
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

func renderStandardizedPopup(totalWidth, totalHeight int, spec standardizedPopupSpec) []string {
	rows := make([]primitives.StandardizedPopupRow, len(spec.rows))
	for i, row := range spec.rows {
		rows[i] = primitives.StandardizedPopupRow{
			Text:       row.text,
			Selectable: row.selectable,
			Selected:   row.selected,
		}
	}

	return primitives.RenderStandardizedPopup(totalWidth, totalHeight, primitives.StandardizedPopupSpec{
		Title:               spec.title,
		Summary:             spec.summary,
		Rows:                rows,
		Footer:              primitives.StandardizedPopupFooter{Left: spec.footer.left, Right: spec.footer.right},
		ScrollOffset:        spec.scrollOffset,
		VisibleRows:         spec.visibleRows,
		ShowScrollIndicator: spec.showScrollIndicator,
		WidthMode:           primitives.StandardizedPopupWidthMode(spec.widthMode),
		DefaultWidth:        spec.defaultWidth,
		MinWidth:            spec.minWidth,
		MaxWidth:            spec.maxWidth,
		Styles:              spec.styles.toPrimitives(),
	})
}

func popupTextRows(rows []string) []standardizedPopupRow {
	result := make([]standardizedPopupRow, len(rows))
	for i, row := range rows {
		result[i] = standardizedPopupRow{text: row}
	}
	return result
}

func selectionSelectedPrefix() string {
	return primitives.SelectionSelectedPrefix()
}

func selectionUnselectedPrefix() string {
	return primitives.SelectionUnselectedPrefix()
}

func centerBoxLines(lines []string, width, height int) string {
	return primitives.CenterBoxLines(lines, width, height)
}

func scrollStart(selection, height, total int) int {
	return primitives.ScrollStart(selection, height, total)
}

func textWidth(text string) int {
	return primitives.TextWidth(text)
}

func minInt(a, b int) int {
	return primitives.MinInt(a, b)
}

func maxInt(a, b int) int {
	return primitives.MaxInt(a, b)
}

type keyBindingID = primitives.KeyBindingID

const (
	keyRuntimeEsc              = primitives.KeyRuntimeEsc
	keyRuntimePageDown         = primitives.KeyRuntimePageDown
	keyRuntimePageUp           = primitives.KeyRuntimePageUp
	keyPopupMoveDown           = primitives.KeyPopupMoveDown
	keyPopupMoveUp             = primitives.KeyPopupMoveUp
	keyPopupJumpTop            = primitives.KeyPopupJumpTop
	keyPopupJumpBottom         = primitives.KeyPopupJumpBottom
	keySelectorCancel          = primitives.KeySelectorCancel
	keySelectorQuit            = primitives.KeySelectorQuit
	keySelectorEnter           = primitives.KeySelectorEnter
	keySelectorMoveDown        = primitives.KeySelectorMoveDown
	keySelectorMoveUp          = primitives.KeySelectorMoveUp
	keySelectorJumpTop         = primitives.KeySelectorJumpTop
	keySelectorJumpBottom      = primitives.KeySelectorJumpBottom
	keySelectorPageDown        = primitives.KeySelectorPageDown
	keySelectorPageUp          = primitives.KeySelectorPageUp
	keySelectorOpenContextHelp = primitives.KeySelectorOpenContextHelp
	keySelectorAdd             = primitives.KeySelectorAdd
	keySelectorEdit            = primitives.KeySelectorEdit
	keySelectorDelete          = primitives.KeySelectorDelete
	keySelectorFormEsc         = primitives.KeySelectorFormEsc
	keySelectorFormSwitch      = primitives.KeySelectorFormSwitch
	keySelectorFormClear       = primitives.KeySelectorFormClear
	keySelectorFormBackspace   = primitives.KeySelectorFormBackspace
	keySelectorDeleteCancel    = primitives.KeySelectorDeleteCancel
	keySelectorDeleteConfirm   = primitives.KeySelectorDeleteConfirm
)

func keyMatches(bindingID keyBindingID, key string) bool {
	return primitives.KeyMatches(bindingID, key)
}

func runtimeHelpPopupSummaryLine() string {
	return primitives.RuntimeHelpPopupSummaryLine()
}

func runtimeStatusContextHelpHint() string {
	return primitives.RuntimeStatusContextHelpHint()
}

func selectorContextLinesBrowseDefault() []string {
	return primitives.SelectorContextLinesBrowseDefault()
}

func selectorContextLinesBrowseFirstSetup() []string {
	return primitives.SelectorContextLinesBrowseFirstSetup()
}

func selectorFormSwitchLine() string {
	return primitives.SelectorFormSwitchLine()
}

func selectorFormSubmitLine(escLabel string) string {
	return primitives.SelectorFormSubmitLine(escLabel)
}

func selectorDeleteConfirmationLine() string {
	return primitives.SelectorDeleteConfirmationLine()
}

func isErrorLikeMessage(message string) bool {
	return primitives.IsErrorLikeMessage(message)
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
