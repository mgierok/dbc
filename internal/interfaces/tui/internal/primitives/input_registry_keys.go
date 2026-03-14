package primitives

import "strings"

type KeyBindingID string

type keyBinding struct {
	keys  []string
	label string
}

const (
	KeyRuntimeOpenCommandInput KeyBindingID = "runtime.open_command_input"
	KeyRuntimeOpenContextHelp  KeyBindingID = "runtime.open_context_help"
	KeyRuntimeJumpTopPending   KeyBindingID = "runtime.jump_top_pending"
	KeyRuntimeJumpTopDisplay   KeyBindingID = "runtime.jump_top_display"
	KeyRuntimeJumpBottom       KeyBindingID = "runtime.jump_bottom"
	KeyRuntimeEnter            KeyBindingID = "runtime.enter"
	KeyRuntimeEdit             KeyBindingID = "runtime.edit"
	KeyRuntimeEsc              KeyBindingID = "runtime.esc"
	KeyRuntimeFilter           KeyBindingID = "runtime.filter"
	KeyRuntimeSort             KeyBindingID = "runtime.sort"
	KeyRuntimeRecordDetail     KeyBindingID = "runtime.record_detail"
	KeyRuntimeInsert           KeyBindingID = "runtime.insert"
	KeyRuntimeDelete           KeyBindingID = "runtime.delete"
	KeyRuntimeUndo             KeyBindingID = "runtime.undo"
	KeyRuntimeRedo             KeyBindingID = "runtime.redo"
	KeyRuntimeToggleAutoFields KeyBindingID = "runtime.toggle_auto_fields"
	KeyRuntimeMoveDown         KeyBindingID = "runtime.move_down"
	KeyRuntimeMoveUp           KeyBindingID = "runtime.move_up"
	KeyRuntimeMoveLeft         KeyBindingID = "runtime.move_left"
	KeyRuntimeMoveRight        KeyBindingID = "runtime.move_right"
	KeyRuntimePageDown         KeyBindingID = "runtime.page_down"
	KeyRuntimePageUp           KeyBindingID = "runtime.page_up"

	KeyPopupMoveDown   KeyBindingID = "popup.move_down"
	KeyPopupMoveUp     KeyBindingID = "popup.move_up"
	KeyPopupJumpTop    KeyBindingID = "popup.jump_top"
	KeyPopupJumpBottom KeyBindingID = "popup.jump_bottom"

	KeyInputMoveLeft  KeyBindingID = "input.move_left"
	KeyInputMoveRight KeyBindingID = "input.move_right"
	KeyInputBackspace KeyBindingID = "input.backspace"
	KeyEditSetNull    KeyBindingID = "edit.set_null"

	KeyConfirmCancel KeyBindingID = "confirm.cancel"
	KeyConfirmAccept KeyBindingID = "confirm.accept"

	KeySelectorCancel          KeyBindingID = "selector.cancel"
	KeySelectorEnter           KeyBindingID = "selector.enter"
	KeySelectorMoveDown        KeyBindingID = "selector.move_down"
	KeySelectorMoveUp          KeyBindingID = "selector.move_up"
	KeySelectorJumpTop         KeyBindingID = "selector.jump_top"
	KeySelectorJumpBottom      KeyBindingID = "selector.jump_bottom"
	KeySelectorPageDown        KeyBindingID = "selector.page_down"
	KeySelectorPageUp          KeyBindingID = "selector.page_up"
	KeySelectorOpenContextHelp KeyBindingID = "selector.open_context_help"
	KeySelectorAdd             KeyBindingID = "selector.add"
	KeySelectorEdit            KeyBindingID = "selector.edit"
	KeySelectorDelete          KeyBindingID = "selector.delete"
	KeySelectorFormEsc         KeyBindingID = "selector.form_esc"
	KeySelectorFormSwitch      KeyBindingID = "selector.form_switch"
	KeySelectorFormClear       KeyBindingID = "selector.form_clear"
	KeySelectorFormBackspace   KeyBindingID = "selector.form_backspace"
	KeySelectorDeleteCancel    KeyBindingID = "selector.delete_cancel"
	KeySelectorDeleteConfirm   KeyBindingID = "selector.delete_confirm"
)

var keyBindings = map[KeyBindingID]keyBinding{
	KeyRuntimeOpenCommandInput: {keys: []string{":"}, label: ":"},
	KeyRuntimeOpenContextHelp:  {keys: []string{"?"}, label: "?"},
	KeyRuntimeJumpTopPending:   {keys: []string{"g"}, label: "g"},
	KeyRuntimeJumpTopDisplay:   {label: "gg"},
	KeyRuntimeJumpBottom:       {keys: []string{"G"}, label: "G"},
	KeyRuntimeEnter:            {keys: []string{"enter"}, label: "Enter"},
	KeyRuntimeEdit:             {keys: []string{"e"}, label: "e"},
	KeyRuntimeEsc:              {keys: []string{"esc"}, label: "Esc"},
	KeyRuntimeFilter:           {keys: []string{"F"}, label: "Shift+F"},
	KeyRuntimeSort:             {keys: []string{"S"}, label: "Shift+S"},
	KeyRuntimeRecordDetail:     {keys: []string{"enter"}, label: "Enter"},
	KeyRuntimeInsert:           {keys: []string{"i"}, label: "i"},
	KeyRuntimeDelete:           {keys: []string{"d"}, label: "d"},
	KeyRuntimeUndo:             {keys: []string{"u"}, label: "u"},
	KeyRuntimeRedo:             {keys: []string{"ctrl+r"}, label: "Ctrl+r"},
	KeyRuntimeToggleAutoFields: {keys: []string{"ctrl+a"}, label: "Ctrl+a"},
	KeyRuntimeMoveDown:         {keys: []string{"j"}, label: "j"},
	KeyRuntimeMoveUp:           {keys: []string{"k"}, label: "k"},
	KeyRuntimeMoveLeft:         {keys: []string{"h"}, label: "h"},
	KeyRuntimeMoveRight:        {keys: []string{"l"}, label: "l"},
	KeyRuntimePageDown:         {keys: []string{"ctrl+f"}, label: "Ctrl+f"},
	KeyRuntimePageUp:           {keys: []string{"ctrl+b"}, label: "Ctrl+b"},

	KeyPopupMoveDown:   {keys: []string{"j", "down"}, label: "j"},
	KeyPopupMoveUp:     {keys: []string{"k", "up"}, label: "k"},
	KeyPopupJumpTop:    {keys: []string{"g", "home"}, label: "g"},
	KeyPopupJumpBottom: {keys: []string{"G", "end"}, label: "G"},

	KeyInputMoveLeft:  {keys: []string{"left"}, label: "left"},
	KeyInputMoveRight: {keys: []string{"right"}, label: "right"},
	KeyInputBackspace: {keys: []string{"backspace"}, label: "backspace"},
	KeyEditSetNull:    {keys: []string{"ctrl+n"}, label: "Ctrl+n"},

	KeyConfirmCancel: {keys: []string{"esc", "n"}, label: "Esc"},
	KeyConfirmAccept: {keys: []string{"enter", "y"}, label: "Enter"},

	KeySelectorCancel:          {keys: []string{"esc"}, label: "Esc"},
	KeySelectorEnter:           {keys: []string{"enter"}, label: "Enter"},
	KeySelectorMoveDown:        {keys: []string{"j", "down"}, label: "j"},
	KeySelectorMoveUp:          {keys: []string{"k", "up"}, label: "k"},
	KeySelectorJumpTop:         {keys: []string{"g", "home"}, label: "g"},
	KeySelectorJumpBottom:      {keys: []string{"G", "end"}, label: "G"},
	KeySelectorPageDown:        {keys: []string{"ctrl+f", "pgdown"}, label: "Ctrl+f"},
	KeySelectorPageUp:          {keys: []string{"ctrl+b", "pgup"}, label: "Ctrl+b"},
	KeySelectorOpenContextHelp: {keys: []string{"?"}, label: "?"},
	KeySelectorAdd:             {keys: []string{"a"}, label: "a"},
	KeySelectorEdit:            {keys: []string{"e"}, label: "e"},
	KeySelectorDelete:          {keys: []string{"d"}, label: "d"},
	KeySelectorFormEsc:         {keys: []string{"esc"}, label: "Esc"},
	KeySelectorFormSwitch:      {keys: []string{"tab", "shift+tab"}, label: "Tab"},
	KeySelectorFormClear:       {keys: []string{"ctrl+u"}, label: "Ctrl+u"},
	KeySelectorFormBackspace:   {keys: []string{"backspace", "ctrl+h"}, label: "Backspace"},
	KeySelectorDeleteCancel:    {keys: []string{"esc"}, label: "Esc"},
	KeySelectorDeleteConfirm:   {keys: []string{"enter"}, label: "Enter"},
}

func KeyMatches(bindingID KeyBindingID, key string) bool {
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

func keyLabel(bindingID KeyBindingID) string {
	binding, ok := keyBindings[bindingID]
	if !ok {
		return ""
	}
	return binding.label
}

func joinKeyLabels(joinWith string, bindingIDs ...KeyBindingID) string {
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
	return strings.Join(parts, FrameSegmentSeparator)
}
