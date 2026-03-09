package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKey_EnterFromTablesSwitchesToRecordsAndContentFocus(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewSchema,
		focus:    FocusTables,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.viewMode != ViewRecords {
		t.Fatalf("expected view mode to switch to records, got %v", model.viewMode)
	}
	if model.focus != FocusContent {
		t.Fatalf("expected focus to switch to content, got %v", model.focus)
	}
}

func TestHandleKey_EscClearsFieldFocus(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.recordFieldFocus {
		t.Fatalf("expected record field focus to be disabled")
	}
	if model.focus != FocusContent {
		t.Fatalf("expected focus to remain on content in nested context, got %v", model.focus)
	}
}

func TestHandleKey_EscInRightPanelNeutralReturnsToTables(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.focus != FocusTables {
		t.Fatalf("expected focus to return to tables, got %v", model.focus)
	}
	if model.viewMode != ViewSchema {
		t.Fatalf("expected schema view to be active, got %v", model.viewMode)
	}
}

func TestHandleKey_EscFromFieldFocusThenNeutralSwitchesToTablesAndSchema(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.recordFieldFocus {
		t.Fatalf("expected record field focus to be disabled")
	}
	if model.focus != FocusTables {
		t.Fatalf("expected focus to return to tables, got %v", model.focus)
	}
	if model.viewMode != ViewSchema {
		t.Fatalf("expected schema view to be active, got %v", model.viewMode)
	}
}
