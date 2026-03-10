package selector

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestDatabaseSelector_EnterSelects(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/example.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if !selector.chosen {
		t.Fatal("expected selection to be confirmed")
	}
}

func TestDatabaseSelector_EscCancels(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/example.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if !selector.canceled {
		t.Fatal("expected selection to be canceled")
	}
}

func TestDatabaseSelector_MoveSelection(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/example.sqlite"},
			{Name: "analytics", Path: "/tmp/analytics.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if selector.browse.selected != 1 {
		t.Fatalf("expected selection to move down, got %d", selector.browse.selected)
	}

	// Act
	updated, _ = selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	// Assert
	selector = updated.(*databaseSelectorModel)
	if selector.browse.selected != 0 {
		t.Fatalf("expected selection to move up, got %d", selector.browse.selected)
	}
}

func TestDatabaseSelector_EditIsUnavailableForSessionScopedCLIEntry(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	})
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.browse.selected = 1

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Assert
	if model.mode != selectorModeBrowse {
		t.Fatalf("expected browse mode to stay active, got %v", model.mode)
	}
	if !strings.Contains(strings.ToLower(model.browse.statusMessage), "cannot be edited") {
		t.Fatalf("expected edit-block status message, got %q", model.browse.statusMessage)
	}
	if len(manager.updated) != 0 {
		t.Fatalf("expected no update calls for CLI session entry, got %d", len(manager.updated))
	}
}

func TestDatabaseSelector_DeleteIsUnavailableForSessionScopedCLIEntry(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	})
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.browse.selected = 1

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if model.mode != selectorModeBrowse {
		t.Fatalf("expected browse mode to stay active, got %v", model.mode)
	}
	if !strings.Contains(strings.ToLower(model.browse.statusMessage), "cannot be deleted") {
		t.Fatalf("expected delete-block status message, got %q", model.browse.statusMessage)
	}
	if len(manager.deleted) != 0 {
		t.Fatalf("expected no delete calls for CLI session entry, got %d", len(manager.deleted))
	}
}
