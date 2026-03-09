package tui

import "github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"

const (
	iconInsert            = primitives.IconInsert
	iconEdit              = primitives.IconEdit
	iconDelete            = primitives.IconDelete
	iconSortAsc           = primitives.IconSortAsc
	iconSortDesc          = primitives.IconSortDesc
	iconConfigSource      = primitives.IconConfigSource
	iconCLISource         = primitives.IconCLISource
	iconInfo              = primitives.IconInfo
	iconSelection         = primitives.IconSelection
	frameVertical         = primitives.FrameVertical
	frameHorizontal       = primitives.FrameHorizontal
	frameTopLeft          = primitives.FrameTopLeft
	frameTopRight         = primitives.FrameTopRight
	frameBottomLeft       = primitives.FrameBottomLeft
	frameBottomRight      = primitives.FrameBottomRight
	frameJoinLeft         = primitives.FrameJoinLeft
	frameJoinRight        = primitives.FrameJoinRight
	frameJoinTop          = primitives.FrameJoinTop
	frameJoinBottom       = primitives.FrameJoinBottom
	frameJoinCenter       = primitives.FrameJoinCenter
	frameSegmentSeparator = primitives.FrameSegmentSeparator
)

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

func (s renderStyles) title(text string) string {
	return s.toPrimitives().Title(text)
}

func (s renderStyles) selected(text string) string {
	return s.toPrimitives().Selected(text)
}

func (s renderStyles) muted(text string) string {
	return s.toPrimitives().Muted(text)
}

func (s renderStyles) error(text string) string {
	return s.toPrimitives().Error(text)
}

func (s renderStyles) dirty(text string) string {
	return s.toPrimitives().Dirty(text)
}

func (s renderStyles) label(text string) string {
	return s.toPrimitives().Label(text)
}

func (s renderStyles) summary(text string) string {
	return s.toPrimitives().Summary(text)
}

func isErrorLikeMessage(message string) bool {
	return primitives.IsErrorLikeMessage(message)
}

type standardizedPopupWidthMode int

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

type keyBindingID = primitives.KeyBindingID

const (
	keyRuntimeOpenCommandInput = primitives.KeyRuntimeOpenCommandInput
	keyRuntimeOpenContextHelp  = primitives.KeyRuntimeOpenContextHelp
	keyRuntimeJumpTopPending   = primitives.KeyRuntimeJumpTopPending
	keyRuntimeJumpTopDisplay   = primitives.KeyRuntimeJumpTopDisplay
	keyRuntimeJumpBottom       = primitives.KeyRuntimeJumpBottom
	keyRuntimeEnter            = primitives.KeyRuntimeEnter
	keyRuntimeEdit             = primitives.KeyRuntimeEdit
	keyRuntimeEsc              = primitives.KeyRuntimeEsc
	keyRuntimeFilter           = primitives.KeyRuntimeFilter
	keyRuntimeSort             = primitives.KeyRuntimeSort
	keyRuntimeRecordDetail     = primitives.KeyRuntimeRecordDetail
	keyRuntimeSave             = primitives.KeyRuntimeSave
	keyRuntimeInsert           = primitives.KeyRuntimeInsert
	keyRuntimeDelete           = primitives.KeyRuntimeDelete
	keyRuntimeUndo             = primitives.KeyRuntimeUndo
	keyRuntimeRedo             = primitives.KeyRuntimeRedo
	keyRuntimeToggleAutoFields = primitives.KeyRuntimeToggleAutoFields
	keyRuntimeMoveDown         = primitives.KeyRuntimeMoveDown
	keyRuntimeMoveUp           = primitives.KeyRuntimeMoveUp
	keyRuntimeMoveLeft         = primitives.KeyRuntimeMoveLeft
	keyRuntimeMoveRight        = primitives.KeyRuntimeMoveRight
	keyRuntimePageDown         = primitives.KeyRuntimePageDown
	keyRuntimePageUp           = primitives.KeyRuntimePageUp
	keyPopupMoveDown           = primitives.KeyPopupMoveDown
	keyPopupMoveUp             = primitives.KeyPopupMoveUp
	keyPopupJumpTop            = primitives.KeyPopupJumpTop
	keyPopupJumpBottom         = primitives.KeyPopupJumpBottom
	keyInputMoveLeft           = primitives.KeyInputMoveLeft
	keyInputMoveRight          = primitives.KeyInputMoveRight
	keyInputBackspace          = primitives.KeyInputBackspace
	keyEditSetNull             = primitives.KeyEditSetNull
	keyConfirmCancel           = primitives.KeyConfirmCancel
	keyConfirmAccept           = primitives.KeyConfirmAccept
	keySelectorOpenContextHelp = primitives.KeySelectorOpenContextHelp
)

type runtimeCommandAction int

const (
	runtimeCommandActionNone       runtimeCommandAction = runtimeCommandAction(primitives.RuntimeCommandActionNone)
	runtimeCommandActionOpenHelp   runtimeCommandAction = runtimeCommandAction(primitives.RuntimeCommandActionOpenHelp)
	runtimeCommandActionQuit       runtimeCommandAction = runtimeCommandAction(primitives.RuntimeCommandActionQuit)
	runtimeCommandActionOpenConfig runtimeCommandAction = runtimeCommandAction(primitives.RuntimeCommandActionOpenConfig)
)

type runtimeCommandSpec struct {
	aliases     []string
	description string
	action      runtimeCommandAction
}

func keyMatches(bindingID keyBindingID, key string) bool {
	return primitives.KeyMatches(bindingID, key)
}

func runtimeHelpPopupSummaryLine() string {
	return primitives.RuntimeHelpPopupSummaryLine()
}

func resolveRuntimeCommand(input string) (runtimeCommandSpec, bool) {
	spec, ok := primitives.ResolveRuntimeCommand(input)
	if !ok {
		return runtimeCommandSpec{}, false
	}
	return runtimeCommandSpec{
		aliases:     append([]string(nil), spec.Aliases...),
		description: spec.Description,
		action:      runtimeCommandAction(spec.Action),
	}, true
}

func runtimeStatusEditShortcuts() string {
	return primitives.RuntimeStatusEditShortcuts()
}

func runtimeStatusConfirmShortcuts(withOptions bool) string {
	return primitives.RuntimeStatusConfirmShortcuts(withOptions)
}

func runtimeStatusFilterPopupShortcuts() string {
	return primitives.RuntimeStatusFilterPopupShortcuts()
}

func runtimeStatusSortPopupShortcuts() string {
	return primitives.RuntimeStatusSortPopupShortcuts()
}

func runtimeStatusHelpPopupShortcuts() string {
	return primitives.RuntimeStatusHelpPopupShortcuts()
}

func runtimeStatusCommandInputShortcuts() string {
	return primitives.RuntimeStatusCommandInputShortcuts()
}

func runtimeStatusTablesShortcuts() string {
	return primitives.RuntimeStatusTablesShortcuts()
}

func runtimeStatusSchemaShortcuts() string {
	return primitives.RuntimeStatusSchemaShortcuts()
}

func runtimeStatusRecordsShortcuts() string {
	return primitives.RuntimeStatusRecordsShortcuts()
}

func runtimeStatusRecordDetailShortcuts() string {
	return primitives.RuntimeStatusRecordDetailShortcuts()
}

func runtimeStatusContextHelpHint() string {
	return primitives.RuntimeStatusContextHelpHint()
}

func renderList(items []string, selected, height, width int, focused bool, styles renderStyles) []string {
	return primitives.RenderList(items, selected, height, width, focused, styles.toPrimitives())
}

func selectionSelectedPrefix() string {
	return primitives.SelectionSelectedPrefix()
}

func selectionUnselectedPrefix() string {
	return primitives.SelectionUnselectedPrefix()
}

func renderPanelBox(title string, content []string, contentWidth int, styles renderStyles) []string {
	return primitives.RenderPanelBox(title, content, contentWidth, styles.toPrimitives())
}

func mergePanelBoxes(left, right []string, leftWidth, rightWidth, gapWidth int) []string {
	return primitives.MergePanelBoxes(left, right, leftWidth, rightWidth, gapWidth)
}

func fitLinesToHeight(lines []string, height, width int) []string {
	return primitives.FitLinesToHeight(lines, height, width)
}

func centerBoxLines(lines []string, width, height int) string {
	return primitives.CenterBoxLines(lines, width, height)
}

func renderFrameEdge(width int, leftCorner, horizontal, rightCorner string) string {
	return primitives.RenderFrameEdge(width, leftCorner, horizontal, rightCorner)
}

func renderFrameContent(value string, width int) string {
	return primitives.RenderFrameContent(value, width)
}

func scrollStart(selection, height, total int) int {
	return primitives.ScrollStart(selection, height, total)
}

func padLines(lines []string, height, width int) []string {
	return primitives.PadLines(lines, height, width)
}

func padRight(text string, width int) string {
	return primitives.PadRight(text, width)
}

func truncate(text string, width int) string {
	return primitives.Truncate(text, width)
}

func wrapTextToWidth(text string, width int) []string {
	return primitives.WrapTextToWidth(text, width)
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
