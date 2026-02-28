package tui

import (
	"fmt"
	"strings"
)

type keyBindingID string

type keyBinding struct {
	keys  []string
	label string
}

type runtimeCommandAction int

const (
	runtimeCommandActionNone runtimeCommandAction = iota
	runtimeCommandActionOpenHelp
	runtimeCommandActionQuit
	runtimeCommandActionOpenConfig
)

type runtimeCommandSpec struct {
	aliases     []string
	description string
	action      runtimeCommandAction
}

const (
	keyRuntimeOpenCommandInput keyBindingID = "runtime.open_command_input"
	keyRuntimeOpenContextHelp  keyBindingID = "runtime.open_context_help"
	keyRuntimeJumpTopPending   keyBindingID = "runtime.jump_top_pending"
	keyRuntimeJumpTopDisplay   keyBindingID = "runtime.jump_top_display"
	keyRuntimeJumpBottom       keyBindingID = "runtime.jump_bottom"
	keyRuntimeEnter            keyBindingID = "runtime.enter"
	keyRuntimeEdit             keyBindingID = "runtime.edit"
	keyRuntimeEsc              keyBindingID = "runtime.esc"
	keyRuntimeFilter           keyBindingID = "runtime.filter"
	keyRuntimeSort             keyBindingID = "runtime.sort"
	keyRuntimeRecordDetail     keyBindingID = "runtime.record_detail"
	keyRuntimeSave             keyBindingID = "runtime.save"
	keyRuntimeInsert           keyBindingID = "runtime.insert"
	keyRuntimeDelete           keyBindingID = "runtime.delete"
	keyRuntimeUndo             keyBindingID = "runtime.undo"
	keyRuntimeRedo             keyBindingID = "runtime.redo"
	keyRuntimeToggleAutoFields keyBindingID = "runtime.toggle_auto_fields"
	keyRuntimeMoveDown         keyBindingID = "runtime.move_down"
	keyRuntimeMoveUp           keyBindingID = "runtime.move_up"
	keyRuntimeMoveLeft         keyBindingID = "runtime.move_left"
	keyRuntimeMoveRight        keyBindingID = "runtime.move_right"
	keyRuntimePageDown         keyBindingID = "runtime.page_down"
	keyRuntimePageUp           keyBindingID = "runtime.page_up"

	keyPopupMoveDown   keyBindingID = "popup.move_down"
	keyPopupMoveUp     keyBindingID = "popup.move_up"
	keyPopupJumpTop    keyBindingID = "popup.jump_top"
	keyPopupJumpBottom keyBindingID = "popup.jump_bottom"

	keyInputMoveLeft  keyBindingID = "input.move_left"
	keyInputMoveRight keyBindingID = "input.move_right"
	keyInputBackspace keyBindingID = "input.backspace"
	keyEditSetNull    keyBindingID = "edit.set_null"

	keyConfirmCancel keyBindingID = "confirm.cancel"
	keyConfirmAccept keyBindingID = "confirm.accept"

	keySelectorCancel        keyBindingID = "selector.cancel"
	keySelectorQuit          keyBindingID = "selector.quit"
	keySelectorEnter         keyBindingID = "selector.enter"
	keySelectorMoveDown      keyBindingID = "selector.move_down"
	keySelectorMoveUp        keyBindingID = "selector.move_up"
	keySelectorJumpTop       keyBindingID = "selector.jump_top"
	keySelectorJumpBottom    keyBindingID = "selector.jump_bottom"
	keySelectorPageDown      keyBindingID = "selector.page_down"
	keySelectorPageUp        keyBindingID = "selector.page_up"
	keySelectorAdd           keyBindingID = "selector.add"
	keySelectorEdit          keyBindingID = "selector.edit"
	keySelectorDelete        keyBindingID = "selector.delete"
	keySelectorFormEsc       keyBindingID = "selector.form_esc"
	keySelectorFormSwitch    keyBindingID = "selector.form_switch"
	keySelectorFormClear     keyBindingID = "selector.form_clear"
	keySelectorFormBackspace keyBindingID = "selector.form_backspace"
	keySelectorDeleteCancel  keyBindingID = "selector.delete_cancel"
	keySelectorDeleteConfirm keyBindingID = "selector.delete_confirm"
)

var keyBindings = map[keyBindingID]keyBinding{
	keyRuntimeOpenCommandInput: {keys: []string{":"}, label: ":"},
	keyRuntimeOpenContextHelp:  {keys: []string{"?"}, label: "?"},
	keyRuntimeJumpTopPending:   {keys: []string{"g"}, label: "g"},
	keyRuntimeJumpTopDisplay:   {label: "gg"},
	keyRuntimeJumpBottom:       {keys: []string{"G"}, label: "G"},
	keyRuntimeEnter:            {keys: []string{"enter"}, label: "Enter"},
	keyRuntimeEdit:             {keys: []string{"e"}, label: "e"},
	keyRuntimeEsc:              {keys: []string{"esc"}, label: "Esc"},
	keyRuntimeFilter:           {keys: []string{"F"}, label: "Shift+F"},
	keyRuntimeSort:             {keys: []string{"S"}, label: "Shift+S"},
	keyRuntimeRecordDetail:     {keys: []string{"enter"}, label: "Enter"},
	keyRuntimeSave:             {keys: []string{"w"}, label: "w"},
	keyRuntimeInsert:           {keys: []string{"i"}, label: "i"},
	keyRuntimeDelete:           {keys: []string{"d"}, label: "d"},
	keyRuntimeUndo:             {keys: []string{"u"}, label: "u"},
	keyRuntimeRedo:             {keys: []string{"ctrl+r"}, label: "Ctrl+r"},
	keyRuntimeToggleAutoFields: {keys: []string{"ctrl+a"}, label: "Ctrl+a"},
	keyRuntimeMoveDown:         {keys: []string{"j"}, label: "j"},
	keyRuntimeMoveUp:           {keys: []string{"k"}, label: "k"},
	keyRuntimeMoveLeft:         {keys: []string{"h"}, label: "h"},
	keyRuntimeMoveRight:        {keys: []string{"l"}, label: "l"},
	keyRuntimePageDown:         {keys: []string{"ctrl+f"}, label: "Ctrl+f"},
	keyRuntimePageUp:           {keys: []string{"ctrl+b"}, label: "Ctrl+b"},

	keyPopupMoveDown:   {keys: []string{"j", "down"}, label: "j"},
	keyPopupMoveUp:     {keys: []string{"k", "up"}, label: "k"},
	keyPopupJumpTop:    {keys: []string{"g", "home"}, label: "g"},
	keyPopupJumpBottom: {keys: []string{"G", "end"}, label: "G"},

	keyInputMoveLeft:  {keys: []string{"left"}, label: "left"},
	keyInputMoveRight: {keys: []string{"right"}, label: "right"},
	keyInputBackspace: {keys: []string{"backspace"}, label: "backspace"},
	keyEditSetNull:    {keys: []string{"ctrl+n"}, label: "Ctrl+n"},

	keyConfirmCancel: {keys: []string{"esc", "n"}, label: "Esc"},
	keyConfirmAccept: {keys: []string{"enter", "y"}, label: "Enter"},

	keySelectorCancel:        {keys: []string{"ctrl+c", "q", "esc"}, label: "Esc"},
	keySelectorQuit:          {keys: []string{"q"}, label: "q"},
	keySelectorEnter:         {keys: []string{"enter"}, label: "Enter"},
	keySelectorMoveDown:      {keys: []string{"j", "down"}, label: "j"},
	keySelectorMoveUp:        {keys: []string{"k", "up"}, label: "k"},
	keySelectorJumpTop:       {keys: []string{"g", "home"}, label: "g"},
	keySelectorJumpBottom:    {keys: []string{"G", "end"}, label: "G"},
	keySelectorPageDown:      {keys: []string{"ctrl+f", "pgdown"}, label: "Ctrl+f"},
	keySelectorPageUp:        {keys: []string{"ctrl+b", "pgup"}, label: "Ctrl+b"},
	keySelectorAdd:           {keys: []string{"a"}, label: "a"},
	keySelectorEdit:          {keys: []string{"e"}, label: "e"},
	keySelectorDelete:        {keys: []string{"d"}, label: "d"},
	keySelectorFormEsc:       {keys: []string{"esc"}, label: "Esc"},
	keySelectorFormSwitch:    {keys: []string{"tab", "shift+tab"}, label: "Tab"},
	keySelectorFormClear:     {keys: []string{"ctrl+u"}, label: "Ctrl+u"},
	keySelectorFormBackspace: {keys: []string{"backspace", "ctrl+h"}, label: "Backspace"},
	keySelectorDeleteCancel:  {keys: []string{"esc"}, label: "Esc"},
	keySelectorDeleteConfirm: {keys: []string{"enter"}, label: "Enter"},
}

var runtimeCommandSpecs = []runtimeCommandSpec{
	{
		aliases:     []string{"config", "c"},
		description: "Open database selector and config manager.",
		action:      runtimeCommandActionOpenConfig,
	},
	{
		aliases:     []string{"help", "h"},
		description: "Open runtime help popup reference.",
		action:      runtimeCommandActionOpenHelp,
	},
	{
		aliases:     []string{"quit", "q"},
		description: "Quit the application.",
		action:      runtimeCommandActionQuit,
	},
}

type runtimeHelpKeywordSpec struct {
	bindings    []keyBindingID
	joinWith    string
	description string
}

var runtimeHelpKeywordSpecs = []runtimeHelpKeywordSpec{
	{
		bindings:    []keyBindingID{keyRuntimeMoveDown, keyRuntimeMoveUp},
		joinWith:    " / ",
		description: "Move selection down or up.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeMoveLeft, keyRuntimeMoveRight},
		joinWith:    " / ",
		description: "Move field focus left or right.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeJumpTopDisplay, keyRuntimeJumpBottom},
		joinWith:    " / ",
		description: "Jump to first or last item.",
	},
	{
		bindings:    []keyBindingID{keyRuntimePageDown, keyRuntimePageUp},
		joinWith:    " / ",
		description: "Page down or up.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeEnter},
		description: "Open records, detail, or confirm action.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeEdit},
		description: "Enter field focus or open edit popup.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeEsc},
		description: "Close active popup/context.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeFilter},
		description: "Open filter flow for current table.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeSort},
		description: "Open sort flow for current table.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeRecordDetail},
		description: "Open selected record detail view.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeInsert},
		description: "Stage a new insert row.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeDelete},
		description: "Toggle delete marker/remove insert.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeUndo, keyRuntimeRedo},
		joinWith:    " / ",
		description: "Undo or redo staged action.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeSave},
		description: "Save staged changes.",
	},
	{
		bindings:    []keyBindingID{keyRuntimeToggleAutoFields},
		description: "Toggle auto field visibility for inserts.",
	},
}

func keyMatches(bindingID keyBindingID, key string) bool {
	binding, ok := keyBindings[bindingID]
	if !ok {
		return false
	}
	for _, candidate := range binding.keys {
		if key == candidate {
			return true
		}
	}
	return false
}

func keyLabel(bindingID keyBindingID) string {
	binding, ok := keyBindings[bindingID]
	if !ok {
		return ""
	}
	return binding.label
}

func joinKeyLabels(joinWith string, bindingIDs ...keyBindingID) string {
	labels := make([]string, 0, len(bindingIDs))
	for _, bindingID := range bindingIDs {
		label := keyLabel(bindingID)
		if strings.TrimSpace(label) == "" {
			continue
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, joinWith)
}

func runtimeHelpPopupSummaryLine() string {
	return fmt.Sprintf(
		"Use %s, %s to scroll. %s closes.",
		joinKeyLabels("/", keyPopupMoveDown, keyPopupMoveUp),
		joinKeyLabels("/", keyRuntimePageDown, keyRuntimePageUp),
		keyLabel(keyRuntimeEsc),
	)
}

func runtimeHelpPopupContentLines() []string {
	lines := []string{"Supported Commands"}
	for _, command := range runtimeCommandSpecs {
		lines = append(lines, fmt.Sprintf("%s - %s", runtimeCommandLabel(command), command.description))
	}

	lines = append(lines, "")
	lines = append(lines, "Supported Keywords")
	for _, keyword := range runtimeHelpKeywordSpecs {
		joinWith := keyword.joinWith
		if strings.TrimSpace(joinWith) == "" {
			joinWith = " / "
		}
		lines = append(lines, fmt.Sprintf("%s - %s", joinKeyLabels(joinWith, keyword.bindings...), keyword.description))
	}
	return lines
}

func resolveRuntimeCommand(input string) (runtimeCommandSpec, bool) {
	command := strings.TrimSpace(input)
	command = strings.TrimPrefix(command, ":")
	command = strings.ToLower(command)
	if command == "" {
		return runtimeCommandSpec{}, false
	}

	for _, candidate := range runtimeCommandSpecs {
		for _, alias := range candidate.aliases {
			if command == strings.ToLower(alias) {
				return candidate, true
			}
		}
	}
	return runtimeCommandSpec{}, false
}

func runtimeCommandLabel(command runtimeCommandSpec) string {
	aliases := make([]string, 0, len(command.aliases))
	for _, alias := range command.aliases {
		aliases = append(aliases, ":"+alias)
	}
	return strings.Join(aliases, " / ")
}

func runtimeStatusEditShortcuts() string {
	return fmt.Sprintf(
		"Edit: %s confirm | %s cancel | %s null",
		keyLabel(keyRuntimeEnter),
		keyLabel(keyRuntimeEsc),
		keyLabel(keyEditSetNull),
	)
}

func runtimeStatusConfirmShortcuts(withOptions bool) string {
	if withOptions {
		return fmt.Sprintf(
			"Confirm: %s choose | %s select | %s cancel",
			joinKeyLabels("/", keyPopupMoveDown, keyPopupMoveUp),
			keyLabel(keyRuntimeEnter),
			keyLabel(keyRuntimeEsc),
		)
	}
	return fmt.Sprintf(
		"Confirm: %s yes | %s no",
		keyLabel(keyRuntimeEnter),
		keyLabel(keyRuntimeEsc),
	)
}

func runtimeStatusFilterPopupShortcuts() string {
	return fmt.Sprintf(
		"Popup: %s apply | %s close",
		keyLabel(keyRuntimeEnter),
		keyLabel(keyRuntimeEsc),
	)
}

func runtimeStatusSortPopupShortcuts() string {
	return fmt.Sprintf(
		"Popup: %s apply | %s close",
		keyLabel(keyRuntimeEnter),
		keyLabel(keyRuntimeEsc),
	)
}

func runtimeStatusHelpPopupShortcuts() string {
	return fmt.Sprintf(
		"Help: %s scroll | %s close",
		joinKeyLabels("/", keyPopupMoveDown, keyPopupMoveUp),
		keyLabel(keyRuntimeEsc),
	)
}

func runtimeStatusCommandInputShortcuts() string {
	return fmt.Sprintf(
		"Command: %s run | %s cancel",
		keyLabel(keyRuntimeEnter),
		keyLabel(keyRuntimeEsc),
	)
}

func runtimeStatusTablesShortcuts() string {
	return fmt.Sprintf(
		"Tables: %s records",
		keyLabel(keyRuntimeEnter),
	)
}

func runtimeStatusSchemaShortcuts() string {
	return fmt.Sprintf(
		"Schema: %s tables",
		keyLabel(keyRuntimeEsc),
	)
}

func runtimeStatusRecordsShortcuts() string {
	return fmt.Sprintf(
		"Records: %s tables | %s edit | %s detail | %s insert | %s delete | %s undo | %s redo | %s save | %s next page | %s prev page | %s filter | %s sort",
		keyLabel(keyRuntimeEsc),
		keyLabel(keyRuntimeEdit),
		keyLabel(keyRuntimeRecordDetail),
		keyLabel(keyRuntimeInsert),
		keyLabel(keyRuntimeDelete),
		keyLabel(keyRuntimeUndo),
		keyLabel(keyRuntimeRedo),
		keyLabel(keyRuntimeSave),
		keyLabel(keyRuntimePageDown),
		keyLabel(keyRuntimePageUp),
		keyLabel(keyRuntimeFilter),
		keyLabel(keyRuntimeSort),
	)
}

func runtimeStatusRecordDetailShortcuts() string {
	return fmt.Sprintf(
		"Detail: %s back | %s scroll | %s page",
		keyLabel(keyRuntimeEsc),
		joinKeyLabels("/", keyPopupMoveDown, keyPopupMoveUp),
		joinKeyLabels("/", keyRuntimePageDown, keyRuntimePageUp),
	)
}

func runtimeStatusContextHelpHint() string {
	return fmt.Sprintf("Context help: %s", keyLabel(keyRuntimeOpenContextHelp))
}

func selectorContextLinesBrowseDefault() []string {
	return []string{
		fmt.Sprintf(
			"%s navigate | %s select | %s add | %s edit | %s delete",
			joinKeyLabels("/", keySelectorMoveDown, keySelectorMoveUp),
			keyLabel(keySelectorEnter),
			keyLabel(keySelectorAdd),
			keyLabel(keySelectorEdit),
			keyLabel(keySelectorDelete),
		),
		fmt.Sprintf(
			"%s cancel | %s quit",
			keyLabel(keyRuntimeEsc),
			keyLabel(keySelectorQuit),
		),
	}
}

func selectorContextLinesBrowseFirstSetup() []string {
	return []string{
		fmt.Sprintf(
			"First setup: %s continue | %s add database",
			keyLabel(keySelectorEnter),
			keyLabel(keySelectorAdd),
		),
		fmt.Sprintf(
			"%s navigate | %s quit",
			joinKeyLabels("/", keySelectorMoveDown, keySelectorMoveUp),
			keyLabel(keySelectorQuit),
		),
	}
}

func selectorFormSwitchLine() string {
	return fmt.Sprintf(
		"%s switch field | %s clear field",
		keyLabel(keySelectorFormSwitch),
		keyLabel(keySelectorFormClear),
	)
}

func selectorFormSubmitLine(escLabel string) string {
	return fmt.Sprintf("%s save | %s", keyLabel(keySelectorEnter), escLabel)
}

func selectorDeleteConfirmationLine() string {
	return fmt.Sprintf(
		"%s confirm delete | %s cancel",
		keyLabel(keySelectorDeleteConfirm),
		keyLabel(keySelectorDeleteCancel),
	)
}
