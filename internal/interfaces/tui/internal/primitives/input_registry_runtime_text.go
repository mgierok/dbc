package primitives

import "fmt"

type runtimeHelpKeywordSpec struct {
	bindings    []KeyBindingID
	joinWith    string
	command     RuntimeCommandAction
	description string
}

var runtimeHelpKeywordSpecs = []runtimeHelpKeywordSpec{
	{
		bindings:    []KeyBindingID{KeyRuntimeMoveDown, KeyRuntimeMoveUp},
		joinWith:    " / ",
		description: "Move selection down or up.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeMoveLeft, KeyRuntimeMoveRight},
		joinWith:    " / ",
		description: "Move field focus left or right.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeJumpTopDisplay, KeyRuntimeJumpBottom},
		joinWith:    " / ",
		description: "Jump to first or last item.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimePageDown, KeyRuntimePageUp},
		joinWith:    " / ",
		description: "Page down or up.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeEnter},
		description: "Open records, detail, or confirm action.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeEdit},
		description: "Enter field focus or open edit popup.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeEsc},
		description: "Close active popup/context.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeFilter},
		description: "Open filter flow for current table.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeSort},
		description: "Open sort flow for current table.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeRecordDetail},
		description: "Open selected record detail view.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeInsert},
		description: "Stage a new insert row.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeDelete},
		description: "Toggle delete marker/remove insert.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeUndo, KeyRuntimeRedo},
		joinWith:    " / ",
		description: "Undo or redo staged action.",
	},
	{
		command:     RuntimeCommandActionSave,
		description: "Save staged changes.",
	},
	{
		bindings:    []KeyBindingID{KeyRuntimeToggleAutoFields},
		description: "Toggle auto field visibility for inserts.",
	},
}

func RuntimeHelpPopupSummaryLine() string {
	return fmt.Sprintf(
		"Use %s, %s to scroll. %s closes.",
		joinKeyLabels("/", KeyPopupMoveDown, KeyPopupMoveUp),
		joinKeyLabels("/", KeyRuntimePageDown, KeyRuntimePageUp),
		keyLabel(KeyRuntimeEsc),
	)
}

func runtimeHelpPopupContentLines() []string {
	lines := []string{"Supported Commands"}
	for _, command := range runtimeCommandSpecs {
		lines = append(lines, fmt.Sprintf("%s - %s", runtimeCommandLabel(command), command.Description))
	}

	lines = append(lines, "")
	lines = append(lines, "Supported Keywords")
	for _, keyword := range runtimeHelpKeywordSpecs {
		lines = append(lines, fmt.Sprintf("%s - %s", runtimeHelpKeywordLabel(keyword), keyword.description))
	}
	return lines
}

func runtimeHelpKeywordLabel(keyword runtimeHelpKeywordSpec) string {
	if keyword.command != RuntimeCommandActionNone {
		return runtimeCommandLabelForAction(keyword.command)
	}

	joinWith := keyword.joinWith
	if joinWith == "" {
		joinWith = " / "
	}
	return joinKeyLabels(joinWith, keyword.bindings...)
}

func RuntimeStatusEditShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Edit: %s confirm", keyLabel(KeyRuntimeEnter)),
		fmt.Sprintf("%s cancel", keyLabel(KeyRuntimeEsc)),
		fmt.Sprintf("%s null", keyLabel(KeyEditSetNull)),
	)
}

func RuntimeStatusConfirmShortcuts(withOptions bool) string {
	if withOptions {
		return joinShortcutSegments(
			fmt.Sprintf("Confirm: %s choose", joinKeyLabels("/", KeyPopupMoveDown, KeyPopupMoveUp)),
			fmt.Sprintf("%s select", keyLabel(KeyRuntimeEnter)),
			fmt.Sprintf("%s cancel", keyLabel(KeyRuntimeEsc)),
		)
	}
	return joinShortcutSegments(
		fmt.Sprintf("Confirm: %s yes", keyLabel(KeyRuntimeEnter)),
		fmt.Sprintf("%s no", keyLabel(KeyRuntimeEsc)),
	)
}

func RuntimeStatusFilterPopupShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Popup: %s apply", keyLabel(KeyRuntimeEnter)),
		fmt.Sprintf("%s close", keyLabel(KeyRuntimeEsc)),
	)
}

func RuntimeStatusSortPopupShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Popup: %s apply", keyLabel(KeyRuntimeEnter)),
		fmt.Sprintf("%s close", keyLabel(KeyRuntimeEsc)),
	)
}

func RuntimeStatusHelpPopupShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Help: %s scroll", joinKeyLabels("/", KeyPopupMoveDown, KeyPopupMoveUp)),
		fmt.Sprintf("%s close", keyLabel(KeyRuntimeEsc)),
	)
}

func RuntimeStatusCommandInputShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Command: %s run", keyLabel(KeyRuntimeEnter)),
		fmt.Sprintf("%s cancel", keyLabel(KeyRuntimeEsc)),
	)
}

func runtimeSaveShortcutSegment() string {
	return fmt.Sprintf("%s save", runtimeCommandLabelForAction(RuntimeCommandActionSave))
}

func RuntimeStatusTablesShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Tables: %s records", keyLabel(KeyRuntimeEnter)),
		runtimeSaveShortcutSegment(),
	)
}

func RuntimeStatusSchemaShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Schema: %s tables", keyLabel(KeyRuntimeEsc)),
		runtimeSaveShortcutSegment(),
	)
}

func RuntimeStatusRecordsShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Records: %s tables", keyLabel(KeyRuntimeEsc)),
		fmt.Sprintf("%s edit", keyLabel(KeyRuntimeEdit)),
		fmt.Sprintf("%s detail", keyLabel(KeyRuntimeRecordDetail)),
		fmt.Sprintf("%s insert", keyLabel(KeyRuntimeInsert)),
		fmt.Sprintf("%s delete", keyLabel(KeyRuntimeDelete)),
		fmt.Sprintf("%s undo", keyLabel(KeyRuntimeUndo)),
		fmt.Sprintf("%s redo", keyLabel(KeyRuntimeRedo)),
		runtimeSaveShortcutSegment(),
		fmt.Sprintf("%s next page", keyLabel(KeyRuntimePageDown)),
		fmt.Sprintf("%s prev page", keyLabel(KeyRuntimePageUp)),
		fmt.Sprintf("%s filter", keyLabel(KeyRuntimeFilter)),
		fmt.Sprintf("%s sort", keyLabel(KeyRuntimeSort)),
	)
}

func RuntimeStatusRecordDetailShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Detail: %s back", keyLabel(KeyRuntimeEsc)),
		fmt.Sprintf("%s scroll", joinKeyLabels("/", KeyPopupMoveDown, KeyPopupMoveUp)),
		fmt.Sprintf("%s page", joinKeyLabels("/", KeyRuntimePageDown, KeyRuntimePageUp)),
		runtimeSaveShortcutSegment(),
	)
}

func RuntimeStatusContextHelpHint() string {
	return fmt.Sprintf("Context help: %s", keyLabel(KeyRuntimeOpenContextHelp))
}
