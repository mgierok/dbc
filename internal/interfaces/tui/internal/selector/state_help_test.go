package selector

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestDatabaseSelector_ContextHelpPopupOpensWithQuestionMarkInBrowseMode(t *testing.T) {
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
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Assert
	if !model.helpPopup.active {
		t.Fatal("expected context help popup to open with ? in browse mode")
	}
	if model.helpPopup.context != selectorModeBrowse {
		t.Fatalf("expected browse help context, got %v", model.helpPopup.context)
	}
}

func TestDatabaseSelector_ContextHelpPopupCapturesFormContext(t *testing.T) {
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
	model.openAddForm()

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Assert
	if !model.helpPopup.active {
		t.Fatal("expected context help popup to open in add form mode")
	}
	if model.helpPopup.context != selectorModeAdd {
		t.Fatalf("expected add-form help context, got %v", model.helpPopup.context)
	}
}

func TestDatabaseSelector_ContextHelpPopupEscClosesAndRestoresSelectorMode(t *testing.T) {
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
	model.mode = selectorModeConfirmDelete
	model.confirmDelete = selectorDeleteConfirm{active: true, optionIndex: 0, managerIndex: 0}
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !model.helpPopup.active {
		t.Fatal("expected context help popup to open before close check")
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.helpPopup.active {
		t.Fatal("expected Esc to close selector context help popup")
	}
	if model.mode != selectorModeConfirmDelete {
		t.Fatalf("expected selector mode to stay unchanged after closing help popup, got %v", model.mode)
	}
}
