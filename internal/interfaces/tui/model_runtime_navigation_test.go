package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_EnterFromTablesSwitchesToRecordsAndContentFocus(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewSchema,
			focus:    FocusTables,
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.read.viewMode != ViewRecords {
		t.Fatalf("expected view mode to switch to records, got %v", model.read.viewMode)
	}
	if model.read.focus != FocusContent {
		t.Fatalf("expected focus to switch to content, got %v", model.read.focus)
	}
}

func TestHandleKey_EscClearsFieldFocus(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			recordFieldFocus: true,
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.read.recordFieldFocus {
		t.Fatalf("expected record field focus to be disabled")
	}
	if model.read.focus != FocusContent {
		t.Fatalf("expected focus to remain on content in nested context, got %v", model.read.focus)
	}
}

func TestHandleKey_EscInRightPanelNeutralReturnsToTables(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.read.focus != FocusTables {
		t.Fatalf("expected focus to return to tables, got %v", model.read.focus)
	}
	if model.read.viewMode != ViewSchema {
		t.Fatalf("expected schema view to be active, got %v", model.read.viewMode)
	}
}

func TestHandleKey_EscFromFieldFocusThenNeutralSwitchesToTablesAndSchema(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode:         ViewRecords,
			focus:            FocusContent,
			recordFieldFocus: true,
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.read.recordFieldFocus {
		t.Fatalf("expected record field focus to be disabled")
	}
	if model.read.focus != FocusTables {
		t.Fatalf("expected focus to return to tables, got %v", model.read.focus)
	}
	if model.read.viewMode != ViewSchema {
		t.Fatalf("expected schema view to be active, got %v", model.read.viewMode)
	}
}

func TestHandleKey_CommandInputClearsPendingGPrefix(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			focus:         FocusTables,
			viewMode:      ViewSchema,
			tables:        []dto.Table{{Name: "a"}, {Name: "b"}, {Name: "c"}},
			selectedTable: 2,
		},
	}

	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	if model.overlay.pendingG {
		t.Fatal("expected starting command input to clear pending runtime key state")
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

	// Assert
	if model.read.selectedTable != 2 {
		t.Fatalf("expected first g after closing command input to keep selection in place, got %d", model.read.selectedTable)
	}
	if !model.overlay.pendingG {
		t.Fatal("expected first g after closing command input to start a fresh gg sequence")
	}
}
