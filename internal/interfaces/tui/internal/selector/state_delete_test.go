package selector

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestDatabaseSelector_DeleteRequiresConfirmation(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
			{Name: "analytics", Path: "/tmp/analytics.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if !model.confirmDelete.active {
		t.Fatal("expected delete confirmation to open")
	}
	if len(manager.deleted) != 0 {
		t.Fatalf("expected no delete before confirmation, got %d", len(manager.deleted))
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if len(manager.deleted) != 1 {
		t.Fatalf("expected one delete call after confirmation, got %d", len(manager.deleted))
	}
	if len(model.options) != 1 {
		t.Fatalf("expected one option after delete, got %d", len(model.options))
	}
}

func TestDatabaseSelector_DeleteLastEntryLeavesEmptyListInBrowseMode(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if len(manager.deleted) != 1 {
		t.Fatalf("expected one delete call after confirmation, got %d", len(manager.deleted))
	}
	if len(model.options) != 0 {
		t.Fatalf("expected zero options after deleting last entry, got %d", len(model.options))
	}
	if model.mode != selectorModeBrowse {
		t.Fatalf("expected browse mode after deleting last entry, got %v", model.mode)
	}
	if !model.requiresFirstEntry {
		t.Fatal("expected first-setup mode to become required after deleting last entry")
	}
}
