package primitives

import "fmt"

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
		if joinWith == "" {
			joinWith = " / "
		}
		lines = append(lines, fmt.Sprintf("%s - %s", joinKeyLabels(joinWith, keyword.bindings...), keyword.description))
	}
	return lines
}

func runtimeStatusEditShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Edit: %s confirm", keyLabel(keyRuntimeEnter)),
		fmt.Sprintf("%s cancel", keyLabel(keyRuntimeEsc)),
		fmt.Sprintf("%s null", keyLabel(keyEditSetNull)),
	)
}

func runtimeStatusConfirmShortcuts(withOptions bool) string {
	if withOptions {
		return joinShortcutSegments(
			fmt.Sprintf("Confirm: %s choose", joinKeyLabels("/", keyPopupMoveDown, keyPopupMoveUp)),
			fmt.Sprintf("%s select", keyLabel(keyRuntimeEnter)),
			fmt.Sprintf("%s cancel", keyLabel(keyRuntimeEsc)),
		)
	}
	return joinShortcutSegments(
		fmt.Sprintf("Confirm: %s yes", keyLabel(keyRuntimeEnter)),
		fmt.Sprintf("%s no", keyLabel(keyRuntimeEsc)),
	)
}

func runtimeStatusFilterPopupShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Popup: %s apply", keyLabel(keyRuntimeEnter)),
		fmt.Sprintf("%s close", keyLabel(keyRuntimeEsc)),
	)
}

func runtimeStatusSortPopupShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Popup: %s apply", keyLabel(keyRuntimeEnter)),
		fmt.Sprintf("%s close", keyLabel(keyRuntimeEsc)),
	)
}

func runtimeStatusHelpPopupShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Help: %s scroll", joinKeyLabels("/", keyPopupMoveDown, keyPopupMoveUp)),
		fmt.Sprintf("%s close", keyLabel(keyRuntimeEsc)),
	)
}

func runtimeStatusCommandInputShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Command: %s run", keyLabel(keyRuntimeEnter)),
		fmt.Sprintf("%s cancel", keyLabel(keyRuntimeEsc)),
	)
}

func runtimeStatusTablesShortcuts() string {
	return fmt.Sprintf("Tables: %s records", keyLabel(keyRuntimeEnter))
}

func runtimeStatusSchemaShortcuts() string {
	return fmt.Sprintf("Schema: %s tables", keyLabel(keyRuntimeEsc))
}

func runtimeStatusRecordsShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Records: %s tables", keyLabel(keyRuntimeEsc)),
		fmt.Sprintf("%s edit", keyLabel(keyRuntimeEdit)),
		fmt.Sprintf("%s detail", keyLabel(keyRuntimeRecordDetail)),
		fmt.Sprintf("%s insert", keyLabel(keyRuntimeInsert)),
		fmt.Sprintf("%s delete", keyLabel(keyRuntimeDelete)),
		fmt.Sprintf("%s undo", keyLabel(keyRuntimeUndo)),
		fmt.Sprintf("%s redo", keyLabel(keyRuntimeRedo)),
		fmt.Sprintf("%s save", keyLabel(keyRuntimeSave)),
		fmt.Sprintf("%s next page", keyLabel(keyRuntimePageDown)),
		fmt.Sprintf("%s prev page", keyLabel(keyRuntimePageUp)),
		fmt.Sprintf("%s filter", keyLabel(keyRuntimeFilter)),
		fmt.Sprintf("%s sort", keyLabel(keyRuntimeSort)),
	)
}

func runtimeStatusRecordDetailShortcuts() string {
	return joinShortcutSegments(
		fmt.Sprintf("Detail: %s back", keyLabel(keyRuntimeEsc)),
		fmt.Sprintf("%s scroll", joinKeyLabels("/", keyPopupMoveDown, keyPopupMoveUp)),
		fmt.Sprintf("%s page", joinKeyLabels("/", keyRuntimePageDown, keyRuntimePageUp)),
	)
}

func runtimeStatusContextHelpHint() string {
	return fmt.Sprintf("Context help: %s", keyLabel(keyRuntimeOpenContextHelp))
}
