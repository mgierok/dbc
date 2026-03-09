package primitives

import "errors"

type RenderStyles struct {
	inner renderStyles
}

func NewRenderStyles(enabled bool) RenderStyles {
	return RenderStyles{inner: renderStyles{enabled: enabled}}
}

func ResolveRenderStylesFromEnv() RenderStyles {
	return RenderStyles{inner: resolveRenderStylesFromEnv()}
}

func (s RenderStyles) Enabled() bool {
	return s.inner.enabled
}

func (s RenderStyles) Title(text string) string {
	return s.inner.title(text)
}

func (s RenderStyles) Selected(text string) string {
	return s.inner.selected(text)
}

func (s RenderStyles) Muted(text string) string {
	return s.inner.muted(text)
}

func (s RenderStyles) Error(text string) string {
	return s.inner.error(text)
}

func (s RenderStyles) Dirty(text string) string {
	return s.inner.dirty(text)
}

func (s RenderStyles) Label(text string) string {
	return s.inner.label(text)
}

func (s RenderStyles) Summary(text string) string {
	return s.inner.summary(text)
}

func IsErrorLikeMessage(message string) bool {
	return isErrorLikeMessage(message)
}

type StandardizedPopupWidthMode int

const (
	RuntimeMaxRecordPageLimit                            = maxRuntimeRecordLimit
	PopupWidthClamp           StandardizedPopupWidthMode = StandardizedPopupWidthMode(popupWidthClamp)
	PopupWidthContent         StandardizedPopupWidthMode = StandardizedPopupWidthMode(popupWidthContent)
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

func RenderStandardizedPopup(totalWidth, totalHeight int, spec StandardizedPopupSpec) []string {
	rows := make([]standardizedPopupRow, len(spec.Rows))
	for i, row := range spec.Rows {
		rows[i] = standardizedPopupRow{
			text:       row.Text,
			selectable: row.Selectable,
			selected:   row.Selected,
		}
	}

	return renderStandardizedPopup(totalWidth, totalHeight, standardizedPopupSpec{
		title:               spec.Title,
		summary:             spec.Summary,
		rows:                rows,
		footer:              standardizedPopupFooter{left: spec.Footer.Left, right: spec.Footer.Right},
		scrollOffset:        spec.ScrollOffset,
		visibleRows:         spec.VisibleRows,
		showScrollIndicator: spec.ShowScrollIndicator,
		widthMode:           standardizedPopupWidthMode(spec.WidthMode),
		defaultWidth:        spec.DefaultWidth,
		minWidth:            spec.MinWidth,
		maxWidth:            spec.MaxWidth,
		styles:              spec.Styles.inner,
	})
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

type KeyBindingID string

const (
	KeyRuntimeOpenCommandInput KeyBindingID = KeyBindingID(keyRuntimeOpenCommandInput)
	KeyRuntimeOpenContextHelp  KeyBindingID = KeyBindingID(keyRuntimeOpenContextHelp)
	KeyRuntimeJumpTopPending   KeyBindingID = KeyBindingID(keyRuntimeJumpTopPending)
	KeyRuntimeJumpTopDisplay   KeyBindingID = KeyBindingID(keyRuntimeJumpTopDisplay)
	KeyRuntimeJumpBottom       KeyBindingID = KeyBindingID(keyRuntimeJumpBottom)
	KeyRuntimeEnter            KeyBindingID = KeyBindingID(keyRuntimeEnter)
	KeyRuntimeEdit             KeyBindingID = KeyBindingID(keyRuntimeEdit)
	KeyRuntimeEsc              KeyBindingID = KeyBindingID(keyRuntimeEsc)
	KeyRuntimeFilter           KeyBindingID = KeyBindingID(keyRuntimeFilter)
	KeyRuntimeSort             KeyBindingID = KeyBindingID(keyRuntimeSort)
	KeyRuntimeRecordDetail     KeyBindingID = KeyBindingID(keyRuntimeRecordDetail)
	KeyRuntimeSave             KeyBindingID = KeyBindingID(keyRuntimeSave)
	KeyRuntimeInsert           KeyBindingID = KeyBindingID(keyRuntimeInsert)
	KeyRuntimeDelete           KeyBindingID = KeyBindingID(keyRuntimeDelete)
	KeyRuntimeUndo             KeyBindingID = KeyBindingID(keyRuntimeUndo)
	KeyRuntimeRedo             KeyBindingID = KeyBindingID(keyRuntimeRedo)
	KeyRuntimeToggleAutoFields KeyBindingID = KeyBindingID(keyRuntimeToggleAutoFields)
	KeyRuntimeMoveDown         KeyBindingID = KeyBindingID(keyRuntimeMoveDown)
	KeyRuntimeMoveUp           KeyBindingID = KeyBindingID(keyRuntimeMoveUp)
	KeyRuntimeMoveLeft         KeyBindingID = KeyBindingID(keyRuntimeMoveLeft)
	KeyRuntimeMoveRight        KeyBindingID = KeyBindingID(keyRuntimeMoveRight)
	KeyRuntimePageDown         KeyBindingID = KeyBindingID(keyRuntimePageDown)
	KeyRuntimePageUp           KeyBindingID = KeyBindingID(keyRuntimePageUp)
	KeyPopupMoveDown           KeyBindingID = KeyBindingID(keyPopupMoveDown)
	KeyPopupMoveUp             KeyBindingID = KeyBindingID(keyPopupMoveUp)
	KeyPopupJumpTop            KeyBindingID = KeyBindingID(keyPopupJumpTop)
	KeyPopupJumpBottom         KeyBindingID = KeyBindingID(keyPopupJumpBottom)
	KeyInputMoveLeft           KeyBindingID = KeyBindingID(keyInputMoveLeft)
	KeyInputMoveRight          KeyBindingID = KeyBindingID(keyInputMoveRight)
	KeyInputBackspace          KeyBindingID = KeyBindingID(keyInputBackspace)
	KeyEditSetNull             KeyBindingID = KeyBindingID(keyEditSetNull)
	KeyConfirmCancel           KeyBindingID = KeyBindingID(keyConfirmCancel)
	KeyConfirmAccept           KeyBindingID = KeyBindingID(keyConfirmAccept)
	KeySelectorCancel          KeyBindingID = KeyBindingID(keySelectorCancel)
	KeySelectorQuit            KeyBindingID = KeyBindingID(keySelectorQuit)
	KeySelectorEnter           KeyBindingID = KeyBindingID(keySelectorEnter)
	KeySelectorMoveDown        KeyBindingID = KeyBindingID(keySelectorMoveDown)
	KeySelectorMoveUp          KeyBindingID = KeyBindingID(keySelectorMoveUp)
	KeySelectorJumpTop         KeyBindingID = KeyBindingID(keySelectorJumpTop)
	KeySelectorJumpBottom      KeyBindingID = KeyBindingID(keySelectorJumpBottom)
	KeySelectorPageDown        KeyBindingID = KeyBindingID(keySelectorPageDown)
	KeySelectorPageUp          KeyBindingID = KeyBindingID(keySelectorPageUp)
	KeySelectorOpenContextHelp KeyBindingID = KeyBindingID(keySelectorOpenContextHelp)
	KeySelectorAdd             KeyBindingID = KeyBindingID(keySelectorAdd)
	KeySelectorEdit            KeyBindingID = KeyBindingID(keySelectorEdit)
	KeySelectorDelete          KeyBindingID = KeyBindingID(keySelectorDelete)
	KeySelectorFormEsc         KeyBindingID = KeyBindingID(keySelectorFormEsc)
	KeySelectorFormSwitch      KeyBindingID = KeyBindingID(keySelectorFormSwitch)
	KeySelectorFormClear       KeyBindingID = KeyBindingID(keySelectorFormClear)
	KeySelectorFormBackspace   KeyBindingID = KeyBindingID(keySelectorFormBackspace)
	KeySelectorDeleteCancel    KeyBindingID = KeyBindingID(keySelectorDeleteCancel)
	KeySelectorDeleteConfirm   KeyBindingID = KeyBindingID(keySelectorDeleteConfirm)
)

type RuntimeCommandAction int

const (
	RuntimeCommandActionNone           RuntimeCommandAction = RuntimeCommandAction(runtimeCommandActionNone)
	RuntimeCommandActionOpenHelp       RuntimeCommandAction = RuntimeCommandAction(runtimeCommandActionOpenHelp)
	RuntimeCommandActionQuit           RuntimeCommandAction = RuntimeCommandAction(runtimeCommandActionQuit)
	RuntimeCommandActionOpenConfig     RuntimeCommandAction = RuntimeCommandAction(runtimeCommandActionOpenConfig)
	RuntimeCommandActionSetRecordLimit RuntimeCommandAction = RuntimeCommandAction(runtimeCommandActionSetRecordLimit)
)

type RuntimeCommandSpec struct {
	Aliases     []string
	Usage       string
	Description string
	Action      RuntimeCommandAction
	RecordLimit int
}

var (
	ErrUnknownRuntimeCommand = errUnknownRuntimeCommand
	ErrInvalidRuntimeCommand = errInvalidRuntimeCommand
)

func KeyMatches(bindingID KeyBindingID, key string) bool {
	return keyMatches(keyBindingID(bindingID), key)
}

func KeyLabel(bindingID KeyBindingID) string {
	return keyLabel(keyBindingID(bindingID))
}

func JoinKeyLabels(joinWith string, bindingIDs ...KeyBindingID) string {
	converted := make([]keyBindingID, len(bindingIDs))
	for i, bindingID := range bindingIDs {
		converted[i] = keyBindingID(bindingID)
	}
	return joinKeyLabels(joinWith, converted...)
}

func JoinShortcutSegments(parts ...string) string {
	return joinShortcutSegments(parts...)
}

func RuntimeHelpPopupSummaryLine() string {
	return runtimeHelpPopupSummaryLine()
}

func RuntimeHelpPopupContentLines() []string {
	return runtimeHelpPopupContentLines()
}

func ResolveRuntimeCommand(input string) (RuntimeCommandSpec, bool) {
	spec, ok := resolveRuntimeCommand(input)
	if !ok {
		return RuntimeCommandSpec{}, false
	}
	return RuntimeCommandSpec{
		Aliases:     append([]string(nil), spec.aliases...),
		Usage:       spec.usage,
		Description: spec.description,
		Action:      RuntimeCommandAction(spec.action),
		RecordLimit: spec.recordLimit,
	}, true
}

func ParseRuntimeCommand(input string) (RuntimeCommandSpec, error) {
	spec, err := parseRuntimeCommand(input)
	if err != nil {
		return RuntimeCommandSpec{}, err
	}
	return RuntimeCommandSpec{
		Aliases:     append([]string(nil), spec.aliases...),
		Usage:       spec.usage,
		Description: spec.description,
		Action:      RuntimeCommandAction(spec.action),
		RecordLimit: spec.recordLimit,
	}, nil
}

func IsUnknownRuntimeCommand(err error) bool {
	return errors.Is(err, ErrUnknownRuntimeCommand)
}

func IsInvalidRuntimeCommand(err error) bool {
	return errors.Is(err, ErrInvalidRuntimeCommand)
}

func RuntimeStatusEditShortcuts() string {
	return runtimeStatusEditShortcuts()
}

func RuntimeStatusConfirmShortcuts(withOptions bool) string {
	return runtimeStatusConfirmShortcuts(withOptions)
}

func RuntimeStatusFilterPopupShortcuts() string {
	return runtimeStatusFilterPopupShortcuts()
}

func RuntimeStatusSortPopupShortcuts() string {
	return runtimeStatusSortPopupShortcuts()
}

func RuntimeStatusHelpPopupShortcuts() string {
	return runtimeStatusHelpPopupShortcuts()
}

func RuntimeStatusCommandInputShortcuts() string {
	return runtimeStatusCommandInputShortcuts()
}

func RuntimeStatusTablesShortcuts() string {
	return runtimeStatusTablesShortcuts()
}

func RuntimeStatusSchemaShortcuts() string {
	return runtimeStatusSchemaShortcuts()
}

func RuntimeStatusRecordsShortcuts() string {
	return runtimeStatusRecordsShortcuts()
}

func RuntimeStatusRecordDetailShortcuts() string {
	return runtimeStatusRecordDetailShortcuts()
}

func RuntimeStatusContextHelpHint() string {
	return runtimeStatusContextHelpHint()
}

func SelectorContextLinesBrowseDefault() []string {
	return selectorContextLinesBrowseDefault()
}

func SelectorContextLinesBrowseFirstSetup() []string {
	return selectorContextLinesBrowseFirstSetup()
}

func SelectorFormSwitchLine() string {
	return selectorFormSwitchLine()
}

func SelectorFormSubmitLine(escLabel string) string {
	return selectorFormSubmitLine(escLabel)
}

func SelectorDeleteConfirmationLine() string {
	return selectorDeleteConfirmationLine()
}

const (
	IconInsert              = iconInsert
	IconEdit                = iconEdit
	IconDelete              = iconDelete
	IconSortAsc             = iconSortAsc
	IconSortDesc            = iconSortDesc
	IconConfigSource        = iconConfigSource
	IconCLISource           = iconCLISource
	IconInfo                = iconInfo
	IconSelection           = iconSelection
	FrameVertical           = frameVertical
	FrameHorizontal         = frameHorizontal
	FrameTopLeft            = frameTopLeft
	FrameTopRight           = frameTopRight
	FrameBottomLeft         = frameBottomLeft
	FrameBottomRight        = frameBottomRight
	FrameJoinLeft           = frameJoinLeft
	FrameJoinRight          = frameJoinRight
	FrameJoinTop            = frameJoinTop
	FrameJoinBottom         = frameJoinBottom
	FrameJoinCenter         = frameJoinCenter
	FrameSegmentSeparator   = frameSegmentSeparator
	PopupContentSidePadding = popupContentSidePadding
	PopupMinHeightPercent   = popupMinHeightPercent
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

func RenderTitledTopBorder(title string, contentWidth int) string {
	return renderTitledTopBorder(title, contentWidth)
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

func CenterText(value string, width int) string {
	return centerText(value, width)
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
