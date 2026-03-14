package primitives

import "fmt"

func SelectorContextLinesBrowseDefault(escAction string) []string {
	return []string{
		joinShortcutSegments(
			fmt.Sprintf("%s navigate", joinKeyLabels("/", KeySelectorMoveDown, KeySelectorMoveUp)),
			fmt.Sprintf("%s select", keyLabel(KeySelectorEnter)),
			fmt.Sprintf("%s add", keyLabel(KeySelectorAdd)),
			fmt.Sprintf("%s edit", keyLabel(KeySelectorEdit)),
			fmt.Sprintf("%s delete", keyLabel(KeySelectorDelete)),
		),
		fmt.Sprintf("%s %s", keyLabel(KeySelectorCancel), escAction),
	}
}

func SelectorContextLinesBrowseFirstSetup(escAction string) []string {
	return []string{
		joinShortcutSegments(
			fmt.Sprintf("First setup: %s continue", keyLabel(KeySelectorEnter)),
			fmt.Sprintf("%s add database", keyLabel(KeySelectorAdd)),
		),
		joinShortcutSegments(
			fmt.Sprintf("%s navigate", joinKeyLabels("/", KeySelectorMoveDown, KeySelectorMoveUp)),
			fmt.Sprintf("%s %s", keyLabel(KeySelectorCancel), escAction),
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
