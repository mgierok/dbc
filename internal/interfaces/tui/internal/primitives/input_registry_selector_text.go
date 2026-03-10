package primitives

import "fmt"

func selectorContextLinesBrowseDefault() []string {
	return []string{
		joinShortcutSegments(
			fmt.Sprintf("%s navigate", joinKeyLabels("/", keySelectorMoveDown, keySelectorMoveUp)),
			fmt.Sprintf("%s select", keyLabel(keySelectorEnter)),
			fmt.Sprintf("%s add", keyLabel(keySelectorAdd)),
			fmt.Sprintf("%s edit", keyLabel(keySelectorEdit)),
			fmt.Sprintf("%s delete", keyLabel(keySelectorDelete)),
		),
		joinShortcutSegments(
			fmt.Sprintf("%s cancel", keyLabel(keyRuntimeEsc)),
			fmt.Sprintf("%s quit", keyLabel(keySelectorQuit)),
		),
	}
}

func selectorContextLinesBrowseFirstSetup() []string {
	return []string{
		joinShortcutSegments(
			fmt.Sprintf("First setup: %s continue", keyLabel(keySelectorEnter)),
			fmt.Sprintf("%s add database", keyLabel(keySelectorAdd)),
		),
		joinShortcutSegments(
			fmt.Sprintf("%s navigate", joinKeyLabels("/", keySelectorMoveDown, keySelectorMoveUp)),
			fmt.Sprintf("%s quit", keyLabel(keySelectorQuit)),
		),
	}
}

func selectorFormSwitchLine() string {
	return joinShortcutSegments(
		fmt.Sprintf("%s switch field", keyLabel(keySelectorFormSwitch)),
		fmt.Sprintf("%s clear field", keyLabel(keySelectorFormClear)),
	)
}

func selectorFormSubmitLine(escLabel string) string {
	return joinShortcutSegments(fmt.Sprintf("%s save", keyLabel(keySelectorEnter)), escLabel)
}

func selectorDeleteConfirmationLine() string {
	return joinShortcutSegments(
		fmt.Sprintf("%s confirm delete", keyLabel(keySelectorDeleteConfirm)),
		fmt.Sprintf("%s cancel", keyLabel(keySelectorDeleteCancel)),
	)
}
