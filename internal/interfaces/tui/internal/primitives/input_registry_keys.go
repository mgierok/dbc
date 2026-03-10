package primitives

import "strings"

type keyBindingID string

type keyBinding struct {
	keys  []string
	label string
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

	keySelectorCancel          keyBindingID = "selector.cancel"
	keySelectorQuit            keyBindingID = "selector.quit"
	keySelectorEnter           keyBindingID = "selector.enter"
	keySelectorMoveDown        keyBindingID = "selector.move_down"
	keySelectorMoveUp          keyBindingID = "selector.move_up"
	keySelectorJumpTop         keyBindingID = "selector.jump_top"
	keySelectorJumpBottom      keyBindingID = "selector.jump_bottom"
	keySelectorPageDown        keyBindingID = "selector.page_down"
	keySelectorPageUp          keyBindingID = "selector.page_up"
	keySelectorOpenContextHelp keyBindingID = "selector.open_context_help"
	keySelectorAdd             keyBindingID = "selector.add"
	keySelectorEdit            keyBindingID = "selector.edit"
	keySelectorDelete          keyBindingID = "selector.delete"
	keySelectorFormEsc         keyBindingID = "selector.form_esc"
	keySelectorFormSwitch      keyBindingID = "selector.form_switch"
	keySelectorFormClear       keyBindingID = "selector.form_clear"
	keySelectorFormBackspace   keyBindingID = "selector.form_backspace"
	keySelectorDeleteCancel    keyBindingID = "selector.delete_cancel"
	keySelectorDeleteConfirm   keyBindingID = "selector.delete_confirm"
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

	keySelectorCancel:          {keys: []string{"ctrl+c", "q", "esc"}, label: "Esc"},
	keySelectorQuit:            {keys: []string{"q"}, label: "q"},
	keySelectorEnter:           {keys: []string{"enter"}, label: "Enter"},
	keySelectorMoveDown:        {keys: []string{"j", "down"}, label: "j"},
	keySelectorMoveUp:          {keys: []string{"k", "up"}, label: "k"},
	keySelectorJumpTop:         {keys: []string{"g", "home"}, label: "g"},
	keySelectorJumpBottom:      {keys: []string{"G", "end"}, label: "G"},
	keySelectorPageDown:        {keys: []string{"ctrl+f", "pgdown"}, label: "Ctrl+f"},
	keySelectorPageUp:          {keys: []string{"ctrl+b", "pgup"}, label: "Ctrl+b"},
	keySelectorOpenContextHelp: {keys: []string{"?"}, label: "?"},
	keySelectorAdd:             {keys: []string{"a"}, label: "a"},
	keySelectorEdit:            {keys: []string{"e"}, label: "e"},
	keySelectorDelete:          {keys: []string{"d"}, label: "d"},
	keySelectorFormEsc:         {keys: []string{"esc"}, label: "Esc"},
	keySelectorFormSwitch:      {keys: []string{"tab", "shift+tab"}, label: "Tab"},
	keySelectorFormClear:       {keys: []string{"ctrl+u"}, label: "Ctrl+u"},
	keySelectorFormBackspace:   {keys: []string{"backspace", "ctrl+h"}, label: "Backspace"},
	keySelectorDeleteCancel:    {keys: []string{"esc"}, label: "Esc"},
	keySelectorDeleteConfirm:   {keys: []string{"enter"}, label: "Enter"},
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

func joinShortcutSegments(parts ...string) string {
	return strings.Join(parts, frameSegmentSeparator)
}
