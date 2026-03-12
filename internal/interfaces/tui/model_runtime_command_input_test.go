package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKey_CommandInputAcceptsLiteralSpaceForSetLimitCommand(t *testing.T) {
	// Arrange
	runtimeSession := &RuntimeSessionState{}
	model := &Model{
		runtimeSession: runtimeSession,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	typeCommandInputText(model, "set")
	model.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	typeCommandInputText(model, "limit=10")
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no immediate records reload outside records view")
	}
	if runtimeSession.RecordsPageLimit != 10 {
		t.Fatalf("expected record limit 10 after typing :set limit=10, got %d", runtimeSession.RecordsPageLimit)
	}
	if model.ui.statusMessage != "Record limit set to 10" {
		t.Fatalf("expected success status, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_StartingNewCommandClearsStaleUnknownCommandStatus(t *testing.T) {
	// Arrange
	model := &Model{}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	typeCommandInputText(model, "setlimit")
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	status := stripANSI(model.renderStatus(200))

	// Assert
	if !model.overlay.commandInput.active {
		t.Fatal("expected command input to be active after pressing :")
	}
	if model.ui.statusMessage != "" {
		t.Fatalf("expected stale status message to be cleared, got %q", model.ui.statusMessage)
	}
	if strings.Contains(status, "Unknown command") {
		t.Fatalf("expected rendered status to hide stale unknown-command message, got %q", status)
	}
	if !strings.Contains(status, "Command: :|") {
		t.Fatalf("expected fresh command prompt in status, got %q", status)
	}
}

func TestHandleKey_CommandInputSecondColonPreservesTypedCommand(t *testing.T) {
	// Arrange
	model := &Model{}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	typeCommandInputText(model, "set")
	model.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	typeCommandInputText(model, "limit=10")

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})

	// Assert
	if !model.overlay.commandInput.active {
		t.Fatal("expected command input to remain active")
	}
	if model.overlay.commandInput.value != "set limit=10:" {
		t.Fatalf("expected repeated : to append to command input, got %q", model.overlay.commandInput.value)
	}
	if model.overlay.commandInput.cursor != len("set limit=10:") {
		t.Fatalf("expected cursor at end of appended command, got %d", model.overlay.commandInput.cursor)
	}
}

func typeCommandInputText(model *Model, value string) {
	for _, r := range value {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}
