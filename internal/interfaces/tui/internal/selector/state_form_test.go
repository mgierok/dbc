package selector

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestDatabaseSelector_AddCreatesEntryAndRefreshesList(t *testing.T) {
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
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = typeText(model, "analytics")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	model = typeText(model, "/tmp/analytics.sqlite")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if len(manager.created) != 1 {
		t.Fatalf("expected one create call, got %d", len(manager.created))
	}
	if manager.created[0].Name != "analytics" || manager.created[0].Path != "/tmp/analytics.sqlite" {
		t.Fatalf("unexpected create payload: %#v", manager.created[0])
	}
	if len(model.options) != 2 {
		t.Fatalf("expected two options after add, got %d", len(model.options))
	}
	if model.options[1].Name != "analytics" {
		t.Fatalf("expected new option in selector list, got %q", model.options[1].Name)
	}
}

func TestDatabaseSelector_EditUpdatesEntry(t *testing.T) {
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
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyCtrlU})
	model = typeText(model, "warehouse")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyCtrlU})
	model = typeText(model, "/tmp/warehouse.sqlite")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if len(manager.updated) != 1 {
		t.Fatalf("expected one update call, got %d", len(manager.updated))
	}
	if manager.updated[0].index != 1 {
		t.Fatalf("expected update index 1, got %d", manager.updated[0].index)
	}
	if manager.updated[0].entry.Name != "warehouse" || manager.updated[0].entry.Path != "/tmp/warehouse.sqlite" {
		t.Fatalf("unexpected update payload: %#v", manager.updated[0].entry)
	}
	if model.options[1].Name != "warehouse" {
		t.Fatalf("expected updated option in selector list, got %q", model.options[1].Name)
	}
}

func TestDatabaseSelector_AddKeepsFormWhenCreateFails(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
		createErr: errors.New("cannot connect to database"),
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = typeText(model, "analytics")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	model = typeText(model, "/tmp/analytics.sqlite")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.mode != selectorModeAdd {
		t.Fatalf("expected add form to stay open, got mode %v", model.mode)
	}
	if len(model.options) != 1 {
		t.Fatalf("expected options to stay unchanged, got %d", len(model.options))
	}
	if !strings.Contains(model.form.errorMessage, "cannot connect") {
		t.Fatalf("expected connection error in form, got %q", model.form.errorMessage)
	}
	if len(manager.created) != 0 {
		t.Fatalf("expected no created entries, got %d", len(manager.created))
	}
}

func TestDatabaseSelector_EditKeepsFormWhenUpdateFails(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
		updateErr: errors.New("cannot connect to database"),
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyCtrlU})
	model = typeText(model, "prod")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyCtrlU})
	model = typeText(model, "/tmp/prod.sqlite")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.mode != selectorModeEdit {
		t.Fatalf("expected edit form to stay open, got mode %v", model.mode)
	}
	if len(model.options) != 1 {
		t.Fatalf("expected options to stay unchanged, got %d", len(model.options))
	}
	if model.options[0].Name != "local" {
		t.Fatalf("expected original option to remain unchanged, got %q", model.options[0].Name)
	}
	if !strings.Contains(model.form.errorMessage, "cannot connect") {
		t.Fatalf("expected connection error in form, got %q", model.form.errorMessage)
	}
	if len(manager.updated) != 0 {
		t.Fatalf("expected no updated entries, got %d", len(manager.updated))
	}
}

func TestDatabaseSelector_ForcedSetupEscCancelsStartup(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{},
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
		t.Fatal("expected forced setup Esc to cancel selector startup")
	}
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("expected controller Esc not to emit quit directly, got %T", cmd())
		}
	}
}

func TestDatabaseSelector_ForcedSetupAllowsContinueAfterFirstEntry(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = typeText(model, "local")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	model = typeText(model, "/tmp/local.sqlite")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if len(model.options) != 1 {
		t.Fatalf("expected one option after first setup entry, got %d", len(model.options))
	}
	if !model.chosen {
		t.Fatal("expected selector completion after first valid entry")
	}
}

func TestDatabaseSelector_ForcedSetupSupportsOptionalAdditionalEntries(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = typeText(model, "local")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	model = typeText(model, "/tmp/local.sqlite")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = typeText(model, "analytics")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	model = typeText(model, "/tmp/analytics.sqlite")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if len(manager.created) != 2 {
		t.Fatalf("expected two created entries in forced setup, got %d", len(manager.created))
	}
	if len(model.options) != 2 {
		t.Fatalf("expected two selector options after forced setup additions, got %d", len(model.options))
	}
	if model.options[0].Name != "local" || model.options[1].Name != "analytics" {
		t.Fatalf("unexpected option names after forced setup: %#v", model.options)
	}
}
