package tui

import (
	"context"
	"errors"
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
	if selector.selected != 1 {
		t.Fatalf("expected selection to move down, got %d", selector.selected)
	}

	// Act
	updated, _ = selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	// Assert
	selector = updated.(*databaseSelectorModel)
	if selector.selected != 0 {
		t.Fatalf("expected selection to move up, got %d", selector.selected)
	}
}

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
	selector := model
	if len(manager.created) != 1 {
		t.Fatalf("expected one create call, got %d", len(manager.created))
	}
	if manager.created[0].Name != "analytics" || manager.created[0].Path != "/tmp/analytics.sqlite" {
		t.Fatalf("unexpected create payload: %#v", manager.created[0])
	}
	if len(selector.options) != 2 {
		t.Fatalf("expected two options after add, got %d", len(selector.options))
	}
	if selector.options[1].Name != "analytics" {
		t.Fatalf("expected new option in selector list, got %q", selector.options[1].Name)
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
	selector := model
	if len(manager.updated) != 1 {
		t.Fatalf("expected one update call, got %d", len(manager.updated))
	}
	if manager.updated[0].index != 1 {
		t.Fatalf("expected update index 1, got %d", manager.updated[0].index)
	}
	if manager.updated[0].entry.Name != "warehouse" || manager.updated[0].entry.Path != "/tmp/warehouse.sqlite" {
		t.Fatalf("unexpected update payload: %#v", manager.updated[0].entry)
	}
	if selector.options[1].Name != "warehouse" {
		t.Fatalf("expected updated option in selector list, got %q", selector.options[1].Name)
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

func TestDatabaseSelector_AddFormShowsCaretInActiveField(t *testing.T) {
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
	lines := strings.Join(model.formLines(), "\n")

	// Assert
	if !strings.Contains(lines, "> Name: |") {
		t.Fatalf("expected caret in active name field, got %q", lines)
	}
	if strings.Contains(lines, "> Path: |") {
		t.Fatalf("expected path field to stay inactive before tab, got %q", lines)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	lines = strings.Join(model.formLines(), "\n")

	// Assert
	if !strings.Contains(lines, "> Path: |") {
		t.Fatalf("expected caret in active path field after tab, got %q", lines)
	}
}

func TestDatabaseSelector_EditFormShowsCaretInActiveField(t *testing.T) {
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
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	lines := strings.Join(model.formLines(), "\n")

	// Assert
	if !strings.Contains(lines, "> Name: local|") {
		t.Fatalf("expected caret in active edit name field, got %q", lines)
	}
	if strings.Contains(lines, "> Path: /tmp/local.sqlite|") {
		t.Fatalf("expected path field to stay inactive before tab, got %q", lines)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	lines = strings.Join(model.formLines(), "\n")

	// Assert
	if !strings.Contains(lines, "> Path: /tmp/local.sqlite|") {
		t.Fatalf("expected caret in active edit path field after tab, got %q", lines)
	}
}

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
	selector := model
	if !selector.confirmDelete.active {
		t.Fatal("expected delete confirmation to open")
	}
	if len(manager.deleted) != 0 {
		t.Fatalf("expected no delete before confirmation, got %d", len(manager.deleted))
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	selector = model
	if len(manager.deleted) != 1 {
		t.Fatalf("expected one delete call after confirmation, got %d", len(manager.deleted))
	}
	if len(selector.options) != 1 {
		t.Fatalf("expected one option after delete, got %d", len(selector.options))
	}
}

func TestDatabaseSelector_ViewShowsActiveConfigPath(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
		activePath: "/tmp/config.toml",
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.width = 120
	model.height = 24

	// Act
	view := model.View()

	// Assert
	if !strings.Contains(view, "/tmp/config.toml") {
		t.Fatalf("expected active config path in view, got %q", view)
	}
}

func TestDatabaseSelector_EmptyConfigStartsInForcedSetupForm(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Assert
	if !model.requiresFirstEntry {
		t.Fatal("expected forced setup mode when no configured databases exist")
	}
	if model.mode != selectorModeAdd {
		t.Fatalf("expected add form mode in forced setup, got %v", model.mode)
	}
}

func TestDatabaseSelector_ForcedSetupRequiresFirstEntryBeforeContinue(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act: attempt to leave forced setup without creating first entry.
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.mode != selectorModeAdd {
		t.Fatalf("expected add form to remain active, got mode %v", model.mode)
	}
	if !strings.Contains(model.statusMessage, "required") {
		t.Fatalf("expected required-first-entry status, got %q", model.statusMessage)
	}

	// Act: create first entry and continue.
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

	// Act: add first entry.
	model = typeText(model, "local")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	model = typeText(model, "/tmp/local.sqlite")
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Act: optionally add second entry before finishing.
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

func typeText(model *databaseSelectorModel, text string) *databaseSelectorModel {
	current := model
	for _, r := range text {
		current = sendKey(current, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return current
}

func sendKey(model *databaseSelectorModel, key tea.KeyMsg) *databaseSelectorModel {
	updated, _ := model.Update(key)
	return updated.(*databaseSelectorModel)
}

type fakeSelectorManager struct {
	entries    []dto.ConfigDatabase
	activePath string

	listErr   error
	createErr error
	updateErr error
	deleteErr error

	created []dto.ConfigDatabase
	updated []updatedEntry
	deleted []int
}

type updatedEntry struct {
	index int
	entry dto.ConfigDatabase
}

func (f *fakeSelectorManager) List(_ context.Context) ([]dto.ConfigDatabase, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	result := make([]dto.ConfigDatabase, len(f.entries))
	copy(result, f.entries)
	return result, nil
}

func (f *fakeSelectorManager) Create(_ context.Context, entry dto.ConfigDatabase) error {
	if f.createErr != nil {
		return f.createErr
	}
	if strings.TrimSpace(entry.Name) == "" || strings.TrimSpace(entry.Path) == "" {
		return errors.New("invalid entry")
	}
	f.created = append(f.created, entry)
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeSelectorManager) Update(_ context.Context, index int, entry dto.ConfigDatabase) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.updated = append(f.updated, updatedEntry{index: index, entry: entry})
	f.entries[index] = entry
	return nil
}

func (f *fakeSelectorManager) Delete(_ context.Context, index int) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.deleted = append(f.deleted, index)
	f.entries = append(f.entries[:index], f.entries[index+1:]...)
	return nil
}

func (f *fakeSelectorManager) ActivePath(_ context.Context) (string, error) {
	if f.activePath == "" {
		return "/tmp/config.toml", nil
	}
	return f.activePath, nil
}
