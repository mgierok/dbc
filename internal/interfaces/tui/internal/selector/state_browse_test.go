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
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if !selector.canceled {
		t.Fatal("expected selection to be canceled")
	}
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("expected controller Esc not to emit quit directly, got %T", cmd())
		}
	}
}

func TestDatabaseSelector_EscDismissesWhenBrowseEscBehaviorIsRuntimeResume(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/example.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		BrowseEscBehavior: SelectorBrowseEscBehaviorRuntimeResume,
	})
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if selector.canceled {
		t.Fatal("expected runtime-resume Esc not to cancel startup")
	}
	if !selector.dismissed {
		t.Fatal("expected runtime-resume Esc to dismiss selector")
	}
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("expected controller Esc not to emit quit directly, got %T", cmd())
		}
	}
}

func TestDatabaseSelector_QDoesNotCancelOrQuit(t *testing.T) {
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
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if selector.canceled {
		t.Fatal("expected q to keep selector active")
	}
	if selector.chosen {
		t.Fatal("expected q not to confirm selection")
	}
	if selector.mode != selectorModeBrowse {
		t.Fatalf("expected browse mode to stay active, got %v", selector.mode)
	}
	if cmd != nil {
		t.Fatalf("expected q not to quit selector, got %T", cmd())
	}
}

func TestDatabaseSelector_CtrlCDoesNotCancelOrQuit(t *testing.T) {
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
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if selector.canceled {
		t.Fatal("expected Ctrl+C to keep selector active")
	}
	if selector.chosen {
		t.Fatal("expected Ctrl+C not to confirm selection")
	}
	if selector.mode != selectorModeBrowse {
		t.Fatalf("expected browse mode to stay active, got %v", selector.mode)
	}
	if cmd != nil {
		t.Fatalf("expected Ctrl+C not to quit selector, got %T", cmd())
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

func TestDatabaseSelector_EditIsUnavailableWhenLoadedStateDisablesEditForConfigSource(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		loadState: &dto.DatabaseSelectorState{
			ActiveConfigPath: "/tmp/config.json",
			Options: []dto.DatabaseSelectorOption{
				{
					Name:        "local",
					ConnString:  "/tmp/local.sqlite",
					Source:      dto.DatabaseSelectorOptionSourceConfig,
					ConfigIndex: 0,
					CanEdit:     false,
					CanDelete:   true,
				},
			},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

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
		t.Fatalf("expected no update calls when edit permission is disabled, got %d", len(manager.updated))
	}
}

func TestDatabaseSelector_DeleteIsUnavailableWhenLoadedStateDisablesDeleteForConfigSource(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		loadState: &dto.DatabaseSelectorState{
			ActiveConfigPath: "/tmp/config.json",
			Options: []dto.DatabaseSelectorOption{
				{
					Name:        "local",
					ConnString:  "/tmp/local.sqlite",
					Source:      dto.DatabaseSelectorOptionSourceConfig,
					ConfigIndex: 0,
					CanEdit:     true,
					CanDelete:   false,
				},
			},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

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
		t.Fatalf("expected no delete calls when delete permission is disabled, got %d", len(manager.deleted))
	}
}
