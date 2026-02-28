package tui

import (
	"strings"
	"testing"
)

func TestResolveRuntimeCommand_ResolvesAliasesCaseInsensitive(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		action runtimeCommandAction
	}{
		{name: "help full", input: ":help", action: runtimeCommandActionOpenHelp},
		{name: "help alias uppercase", input: ":H", action: runtimeCommandActionOpenHelp},
		{name: "quit full", input: ":quit", action: runtimeCommandActionQuit},
		{name: "quit alias", input: ":q", action: runtimeCommandActionQuit},
		{name: "config full", input: ":config", action: runtimeCommandActionOpenConfig},
		{name: "config alias", input: ":c", action: runtimeCommandActionOpenConfig},
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
	if lines[1] != ":config / :c - Open database selector and config manager." {
		t.Fatalf("unexpected first command line: %q", lines[1])
	}
	if lines[2] != ":help / :h - Open runtime help popup reference." {
		t.Fatalf("unexpected second command line: %q", lines[2])
	}
	if lines[3] != ":quit / :q - Quit the application." {
		t.Fatalf("unexpected third command line: %q", lines[3])
	}
	joined := strings.Join(lines, "\n")
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
