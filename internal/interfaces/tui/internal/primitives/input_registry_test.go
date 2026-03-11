package primitives

import (
	"errors"
	"strings"
	"testing"

	runtimecontract "github.com/mgierok/dbc/internal/interfaces/tui/internal"
)

func TestResolveRuntimeCommand_ResolvesAliasesCaseInsensitive(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		action      runtimeCommandAction
		recordLimit int
	}{
		{name: "help full", input: ":help", action: runtimeCommandActionOpenHelp},
		{name: "help alias uppercase", input: ":H", action: runtimeCommandActionOpenHelp},
		{name: "quit full", input: ":quit", action: runtimeCommandActionQuit},
		{name: "quit alias", input: ":q", action: runtimeCommandActionQuit},
		{name: "config full", input: ":config", action: runtimeCommandActionOpenConfig},
		{name: "config alias", input: ":c", action: runtimeCommandActionOpenConfig},
		{name: "set limit", input: ":set limit=10", action: runtimeCommandActionSetRecordLimit, recordLimit: 10},
		{name: "set limit keyword uppercase", input: ":SET LIMIT=25", action: runtimeCommandActionSetRecordLimit, recordLimit: 25},
		{name: "set limit max boundary", input: ":set limit=1000", action: runtimeCommandActionSetRecordLimit, recordLimit: runtimecontract.MaxRecordPageLimit},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange

			// Act
			command, ok := resolveRuntimeCommand(tc.input)

			// Assert
			if !ok {
				t.Fatalf("expected command %q to resolve", tc.input)
			}
			if command.action != tc.action {
				t.Fatalf("expected action %v for %q, got %v", tc.action, tc.input, command.action)
			}
			if command.recordLimit != tc.recordLimit {
				t.Fatalf("expected record limit %d for %q, got %d", tc.recordLimit, tc.input, command.recordLimit)
			}
		})
	}
}

func TestResolveRuntimeCommand_RejectsUnknownCommand(t *testing.T) {
	// Arrange
	input := ":unknown"

	// Act
	_, ok := resolveRuntimeCommand(input)

	// Assert
	if ok {
		t.Fatalf("expected %q to be rejected", input)
	}
}

func TestParseRuntimeCommand_RejectsInvalidSetLimitFormsWithValidationError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "missing value", input: ":set limit="},
		{name: "missing equals", input: ":set limit"},
		{name: "zero", input: ":set limit=0"},
		{name: "negative", input: ":set limit=-1"},
		{name: "too large", input: ":set limit=100000000"},
		{name: "non integer", input: ":set limit=abc"},
		{name: "space after equals", input: ":set limit= 10"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange

			// Act
			_, err := parseRuntimeCommand(tc.input)

			// Assert
			if err == nil {
				t.Fatalf("expected validation error for %q", tc.input)
			}
			if !errors.Is(err, errInvalidRuntimeCommand) {
				t.Fatalf("expected invalid runtime command error for %q, got %v", tc.input, err)
			}
			if !strings.Contains(err.Error(), ":set limit=<1-1000>") {
				t.Fatalf("expected deterministic validation hint for %q, got %v", tc.input, err)
			}
		})
	}
}

func TestRuntimeHelpPopupSummaryLine_IsDeterministic(t *testing.T) {
	// Arrange

	// Act
	summary := runtimeHelpPopupSummaryLine()

	// Assert
	if summary != "Use j/k, Ctrl+f/Ctrl+b to scroll. Esc closes." {
		t.Fatalf("expected deterministic summary, got %q", summary)
	}
}

func TestRuntimeStatusContextHelpHint_IsDeterministic(t *testing.T) {
	// Arrange

	// Act
	hint := runtimeStatusContextHelpHint()

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
	if !strings.Contains(joined, ":config / :c - Open database selector and config manager.") {
		t.Fatalf("expected config command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":help / :h - Open runtime help popup reference.") {
		t.Fatalf("expected help command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":set limit=<n> - Set records page limit for the current app session.") {
		t.Fatalf("expected set-limit command line in help content, got %q", joined)
	}
	if !strings.Contains(joined, ":quit / :q - Quit the application.") {
		t.Fatalf("expected quit command line in help content, got %q", joined)
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
	shortcuts := runtimeStatusRecordsShortcuts()

	// Assert
	if !strings.Contains(shortcuts, "Ctrl+f next page") {
		t.Fatalf("expected next-page shortcut in records shortcuts, got %q", shortcuts)
	}
	if !strings.Contains(shortcuts, "Ctrl+b prev page") {
		t.Fatalf("expected prev-page shortcut in records shortcuts, got %q", shortcuts)
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
	if !keyMatches(keySelectorOpenContextHelp, "?") {
		t.Fatalf("expected selector context-help binding to match question mark")
	}
	if keyLabel(keySelectorOpenContextHelp) != "?" {
		t.Fatalf("expected selector context-help label ?, got %q", keyLabel(keySelectorOpenContextHelp))
	}
}

func TestSelectorContextLinesBrowseDefault_AreDeterministic(t *testing.T) {
	// Arrange

	// Act
	lines := selectorContextLinesBrowseDefault()

	// Assert
	if len(lines) != 2 {
		t.Fatalf("expected two selector browse help lines, got %v", lines)
	}
	if lines[0] != "j/k navigate | Enter select | a add | e edit | d delete" {
		t.Fatalf("expected deterministic browse shortcut line, got %q", lines[0])
	}
	if lines[1] != "Esc cancel | q quit" {
		t.Fatalf("expected deterministic browse exit line, got %q", lines[1])
	}
}

func TestSelectorContextLinesBrowseFirstSetup_AreDeterministic(t *testing.T) {
	// Arrange

	// Act
	lines := selectorContextLinesBrowseFirstSetup()

	// Assert
	if len(lines) != 2 {
		t.Fatalf("expected two selector first-setup help lines, got %v", lines)
	}
	if lines[0] != "First setup: Enter continue | a add database" {
		t.Fatalf("expected deterministic first-setup action line, got %q", lines[0])
	}
	if lines[1] != "j/k navigate | q quit" {
		t.Fatalf("expected deterministic first-setup exit line, got %q", lines[1])
	}
}

func TestSelectorFormAndDeleteLines_AreDeterministic(t *testing.T) {
	// Arrange

	// Act
	switchLine := selectorFormSwitchLine()
	submitLine := selectorFormSubmitLine("Esc cancel")
	deleteLine := selectorDeleteConfirmationLine()

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
