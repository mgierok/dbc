package primitives

import "fmt"

func SelectorContextLinesBrowseDefault() []string {
	return []string{
		joinShortcutSegments(
			fmt.Sprintf("%s navigate", joinKeyLabels("/", KeySelectorMoveDown, KeySelectorMoveUp)),
			fmt.Sprintf("%s select", keyLabel(KeySelectorEnter)),
			fmt.Sprintf("%s add", keyLabel(KeySelectorAdd)),
			fmt.Sprintf("%s edit", keyLabel(KeySelectorEdit)),
			fmt.Sprintf("%s delete", keyLabel(KeySelectorDelete)),
		),
		joinShortcutSegments(
			fmt.Sprintf("%s cancel", keyLabel(KeyRuntimeEsc)),
			fmt.Sprintf("%s quit", keyLabel(KeySelectorQuit)),
		),
	}
}

func SelectorContextLinesBrowseFirstSetup() []string {
	return []string{
		joinShortcutSegments(
			fmt.Sprintf("First setup: %s continue", keyLabel(KeySelectorEnter)),
			fmt.Sprintf("%s add database", keyLabel(KeySelectorAdd)),
		),
		joinShortcutSegments(
			fmt.Sprintf("%s navigate", joinKeyLabels("/", KeySelectorMoveDown, KeySelectorMoveUp)),
			fmt.Sprintf("%s quit", keyLabel(KeySelectorQuit)),
		),
	}
}

func SelectorFormSwitchLine() string {
	return joinShortcutSegments(
		fmt.Sprintf("%s switch field", keyLabel(KeySelectorFormSwitch)),
		fmt.Sprintf("%s clear field", keyLabel(KeySelectorFormClear)),
	)
}

func SelectorFormSubmitLine(escLabel string) string {
	return joinShortcutSegments(fmt.Sprintf("%s save", keyLabel(KeySelectorEnter)), escLabel)
}

func SelectorDeleteConfirmationLine() string {
	return joinShortcutSegments(
		fmt.Sprintf("%s confirm delete", keyLabel(KeySelectorDeleteConfirm)),
		fmt.Sprintf("%s cancel", keyLabel(KeySelectorDeleteCancel)),
	)
}
