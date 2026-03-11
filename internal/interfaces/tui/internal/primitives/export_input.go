package primitives

import "errors"

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

var ErrUnknownRuntimeCommand = errUnknownRuntimeCommand

func KeyMatches(bindingID KeyBindingID, key string) bool {
	return keyMatches(keyBindingID(bindingID), key)
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

func RuntimeHelpPopupSummaryLine() string {
	return runtimeHelpPopupSummaryLine()
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

func RenderStatusWithRightHint(left, right string, width int) string {
	return renderStatusWithRightHint(left, right, width)
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
