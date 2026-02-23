package tui

import "testing"

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
	if lines[3] != ":q / :quit - Quit the application." {
		t.Fatalf("unexpected third command line: %q", lines[3])
	}
}
