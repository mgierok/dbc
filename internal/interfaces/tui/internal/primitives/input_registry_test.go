package primitives

import (
	"errors"
	"strings"
	"testing"
)

func TestResolveRuntimeCommand_ResolvesAliasesCaseInsensitive(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		action      RuntimeCommandAction
		force       bool
		connString  string
		recordLimit int
	}{
		{name: "help full", input: ":help", action: RuntimeCommandActionOpenHelp},
		{name: "help alias uppercase", input: ":H", action: RuntimeCommandActionOpenHelp},
		{name: "quit full", input: ":quit", action: RuntimeCommandActionQuit},
		{name: "quit alias", input: ":q", action: RuntimeCommandActionQuit},
		{name: "forced quit full", input: ":quit!", action: RuntimeCommandActionForcedQuit},
		{name: "forced quit alias", input: ":q!", action: RuntimeCommandActionForcedQuit},
		{name: "forced quit uppercase", input: ":Q!", action: RuntimeCommandActionForcedQuit},
		{name: "config full", input: ":config", action: RuntimeCommandActionOpenConfig},
		{name: "config alias", input: ":c", action: RuntimeCommandActionOpenConfig},
		{name: "save alias", input: ":w", action: RuntimeCommandActionSave},
		{name: "save full alias", input: ":write", action: RuntimeCommandActionSave},
		{name: "save and quit", input: ":wq", action: RuntimeCommandActionSaveAndQuit},
		{name: "save and quit uppercase", input: ":WQ", action: RuntimeCommandActionSaveAndQuit},
		{name: "edit current database", input: ":edit", action: RuntimeCommandActionEdit},
		{name: "edit alias", input: ":e", action: RuntimeCommandActionEdit},
		{name: "forced edit current database", input: ":edit!", action: RuntimeCommandActionEdit, force: true},
		{name: "forced edit alias", input: ":E!", action: RuntimeCommandActionEdit, force: true},
		{name: "edit connection string", input: ":edit /tmp/analytics.sqlite", action: RuntimeCommandActionEdit, connString: "/tmp/analytics.sqlite"},
		{name: "forced edit connection string with spaces", input: ":e!  /tmp/analytics copy.sqlite  ", action: RuntimeCommandActionEdit, force: true, connString: "/tmp/analytics copy.sqlite"},
		{name: "set limit", input: ":set limit=10", action: RuntimeCommandActionSetRecordLimit, recordLimit: 10},
		{name: "set limit keyword uppercase", input: ":SET LIMIT=25", action: RuntimeCommandActionSetRecordLimit, recordLimit: 25},
		{name: "set limit zero stays parser-valid", input: ":set limit=0", action: RuntimeCommandActionSetRecordLimit, recordLimit: 0},
		{name: "set limit negative stays parser-valid", input: ":set limit=-1", action: RuntimeCommandActionSetRecordLimit, recordLimit: -1},
		{name: "set limit oversized stays parser-valid", input: ":set limit=100000000", action: RuntimeCommandActionSetRecordLimit, recordLimit: 100000000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange

			// Act
			command, err := ParseRuntimeCommand(tc.input)

			// Assert
			if err != nil {
				t.Fatalf("expected command %q to resolve, got error %v", tc.input, err)
			}
			if command.Action != tc.action {
				t.Fatalf("expected action %v for %q, got %v", tc.action, tc.input, command.Action)
			}
			if command.Force != tc.force {
				t.Fatalf("expected force=%t for %q, got %t", tc.force, tc.input, command.Force)
			}
			if command.ConnString != tc.connString {
				t.Fatalf("expected connection string %q for %q, got %q", tc.connString, tc.input, command.ConnString)
			}
			if command.RecordLimit != tc.recordLimit {
				t.Fatalf("expected record limit %d for %q, got %d", tc.recordLimit, tc.input, command.RecordLimit)
			}
		})
	}
}

func TestParseRuntimeCommand_ClassifiesUnknownCommand(t *testing.T) {
	// Arrange
	input := ":unknown"

	// Act
	_, err := ParseRuntimeCommand(input)

	// Assert
	if err == nil {
		t.Fatalf("expected %q to be rejected", input)
	}
	if !errors.Is(err, ErrUnknownRuntimeCommand) {
		t.Fatalf("expected unknown runtime command error for %q, got %v", input, err)
	}
	if !IsUnknownRuntimeCommand(err) {
		t.Fatalf("expected IsUnknownRuntimeCommand to classify %q as unknown", input)
	}
}

func TestParseRuntimeCommand_RejectsInvalidSetLimitFormsWithValidationError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "missing value", input: ":set limit="},
		{name: "missing equals", input: ":set limit"},
		{name: "non integer", input: ":set limit=abc"},
		{name: "space after equals", input: ":set limit= 10"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange

			// Act
			_, err := ParseRuntimeCommand(tc.input)

			// Assert
			if err == nil {
				t.Fatalf("expected validation error for %q", tc.input)
			}
			if !errors.Is(err, errInvalidRuntimeCommand) {
				t.Fatalf("expected invalid runtime command error for %q, got %v", tc.input, err)
			}
			if !strings.Contains(err.Error(), ":set limit=<n>") {
				t.Fatalf("expected deterministic syntax hint for %q, got %v", tc.input, err)
			}
		})
	}
}

func TestRuntimeHelpPopupSummaryLine_IsDeterministic(t *testing.T) {
	// Arrange

	// Act
	summary := RuntimeHelpPopupSummaryLine()

	// Assert
	if summary != "Use j/k, Ctrl+f/Ctrl+b to scroll. Esc closes." {
		t.Fatalf("expected deterministic summary, got %q", summary)
	}
}

func TestRuntimeStatusContextHelpHint_IsDeterministic(t *testing.T) {
	// Arrange

	// Act
	hint := RuntimeStatusContextHelpHint()

	// Assert
	if hint != "Context help: ?" {
		t.Fatalf("expected deterministic context-help hint, got %q", hint)
	}
}

func TestRuntimeHelpPopupContentLines_UsesRegistryDefinitions(t *testing.T) {
	// Arrange

	// Act
	lines := runtimeHelpPopupContentLines()

	// Assert
	if len(lines) < 6 {
		t.Fatalf("expected multi-line help content, got %v", lines)
	}
	if lines[0] != "Supported Commands" {
		t.Fatalf("expected commands section header, got %q", lines[0])
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, ":config / :c - Open database selector.") {
		t.Fatalf("expected config command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":help / :h - Open runtime help popup reference.") {
		t.Fatalf("expected help command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":w / :write - Save staged changes immediately.") {
		t.Fatalf("expected save command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":edit[!] / :e[!] [<connection-string>] - Reload current database or open another database.") {
		t.Fatalf("expected edit command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":wq - Save staged changes immediately and quit on success.") {
		t.Fatalf("expected save-and-quit command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":set limit=<n> - Set records page limit for the current app session.") {
		t.Fatalf("expected set-limit command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":quit / :q - Quit the application.") {
		t.Fatalf("expected quit command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":quit! / :q! - Discard staged changes and quit the application immediately.") {
		t.Fatalf("expected forced-quit command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, "Shift+F - Open filter flow for current table.") {
		t.Fatalf("expected filter keyword in help content, got %q", joined)
	}
	if !strings.Contains(joined, "Shift+S - Open sort flow for current table.") {
		t.Fatalf("expected sort keyword in help content, got %q", joined)
	}
}

func TestRuntimeStatusRecordsShortcuts_IncludesPaginationBindings(t *testing.T) {
	// Arrange

	// Act
	shortcuts := RuntimeStatusRecordsShortcuts()

	// Assert
	if !strings.Contains(shortcuts, ":w / :write save") {
		t.Fatalf("expected save command hint in records shortcuts, got %q", shortcuts)
	}
	if !strings.Contains(shortcuts, "Ctrl+f next page") {
		t.Fatalf("expected next-page shortcut in records shortcuts, got %q", shortcuts)
	}
	if !strings.Contains(shortcuts, "Ctrl+b prev page") {
		t.Fatalf("expected prev-page shortcut in records shortcuts, got %q", shortcuts)
	}
}

func TestRuntimeStatusContextShortcuts_IncludeSharedSaveCommandWhereSupported(t *testing.T) {
	for _, tc := range []struct {
		name      string
		shortcuts string
	}{
		{name: "tables", shortcuts: RuntimeStatusTablesShortcuts()},
		{name: "schema", shortcuts: RuntimeStatusSchemaShortcuts()},
		{name: "records", shortcuts: RuntimeStatusRecordsShortcuts()},
		{name: "record detail", shortcuts: RuntimeStatusRecordDetailShortcuts()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange

			// Act

			// Assert
			if !strings.Contains(tc.shortcuts, ":w / :write save") {
				t.Fatalf("expected shared save shortcut in %s shortcuts, got %q", tc.name, tc.shortcuts)
			}
		})
	}
}

func TestJoinShortcutSegments_UsesASCIIPipeSeparator(t *testing.T) {
	// Arrange

	// Act
	joined := joinShortcutSegments("A", "B")

	// Assert
	if joined != "A | B" {
		t.Fatalf("expected ASCII separator, got %q", joined)
	}
}

func TestSelectorContextHelpBinding_IsQuestionMark(t *testing.T) {
	// Arrange

	// Act + Assert
	if !KeyMatches(KeySelectorOpenContextHelp, "?") {
		t.Fatalf("expected selector context-help binding to match question mark")
	}
}

func TestSelectorCancelBinding_IsEscOnly(t *testing.T) {
	// Arrange

	// Act + Assert
	if !KeyMatches(KeySelectorCancel, "esc") {
		t.Fatal("expected selector cancel binding to match Esc")
	}
	if KeyMatches(KeySelectorCancel, "q") {
		t.Fatal("expected selector cancel binding not to match q")
	}
	if KeyMatches(KeySelectorCancel, "ctrl+c") {
		t.Fatal("expected selector cancel binding not to match Ctrl+C")
	}
}

func TestConfirmBindings_UseEnterAndEscOnly(t *testing.T) {
	// Arrange

	// Act + Assert
	if !KeyMatches(KeyConfirmAccept, "enter") {
		t.Fatal("expected confirm accept binding to match Enter")
	}
	if KeyMatches(KeyConfirmAccept, "y") {
		t.Fatal("expected confirm accept binding not to match y")
	}
	if !KeyMatches(KeyConfirmCancel, "esc") {
		t.Fatal("expected confirm cancel binding to match Esc")
	}
	if KeyMatches(KeyConfirmCancel, "n") {
		t.Fatal("expected confirm cancel binding not to match n")
	}
}

func TestSelectorContextLinesBrowseDefault_AreDeterministic(t *testing.T) {
	// Arrange

	// Act
	lines := SelectorContextLinesBrowseDefault("quit")

	// Assert
	if len(lines) != 2 {
		t.Fatalf("expected two selector browse help lines, got %v", lines)
	}
	if lines[0] != "j/k navigate | Enter select | a add | e edit | d delete" {
		t.Fatalf("expected deterministic browse shortcut line, got %q", lines[0])
	}
	if lines[1] != "Esc quit" {
		t.Fatalf("expected deterministic browse exit line, got %q", lines[1])
	}
}

func TestSelectorContextLinesBrowseFirstSetup_AreDeterministic(t *testing.T) {
	// Arrange

	// Act
	lines := SelectorContextLinesBrowseFirstSetup("exit app")

	// Assert
	if len(lines) != 2 {
		t.Fatalf("expected two selector first-setup help lines, got %v", lines)
	}
	if lines[0] != "First setup: Enter continue | a add database" {
		t.Fatalf("expected deterministic first-setup action line, got %q", lines[0])
	}
	if lines[1] != "j/k navigate | Esc exit app" {
		t.Fatalf("expected deterministic first-setup exit line, got %q", lines[1])
	}
}

func TestSelectorContextLinesBrowseDefault_RuntimeResumeAreDeterministic(t *testing.T) {
	// Arrange

	// Act
	lines := SelectorContextLinesBrowseDefault("close")

	// Assert
	if len(lines) != 2 {
		t.Fatalf("expected two selector browse help lines, got %v", lines)
	}
	if lines[1] != "Esc close" {
		t.Fatalf("expected runtime-resume browse exit line, got %q", lines[1])
	}
}

func TestSelectorFormAndDeleteLines_AreDeterministic(t *testing.T) {
	// Arrange

	// Act
	switchLine := SelectorFormSwitchLine()
	submitLine := SelectorFormSubmitLine("Esc cancel")
	deleteLine := SelectorDeleteConfirmationLine()

	// Assert
	if switchLine != "Tab switch field | Ctrl+u clear field" {
		t.Fatalf("expected deterministic form-switch line, got %q", switchLine)
	}
	if submitLine != "Enter save | Esc cancel" {
		t.Fatalf("expected deterministic form-submit line, got %q", submitLine)
	}
	if deleteLine != "Enter confirm delete | Esc cancel" {
		t.Fatalf("expected deterministic delete-confirmation line, got %q", deleteLine)
	}
}
