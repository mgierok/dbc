package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_DirtyConfigCommandOpensDecisionPrompt(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "full command", command: "config"},
		{name: "alias", command: "c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := &Model{
				read: runtimeReadState{
					viewMode:      ViewRecords,
					tables:        []dto.Table{{Name: "users"}},
					selectedTable: 0,
				},
				staging: testDatabaseStaging(map[string]stagingState{
					"users": {pendingInserts: []pendingInsertRow{{}}},
				}),
			}

			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
			for _, r := range tc.command {
				model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
			_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

			if cmd != nil {
				if _, ok := cmd().(tea.QuitMsg); ok {
					t.Fatalf("expected dirty :%s to wait for explicit decision", tc.command)
				}
			}
			if !model.overlay.confirmPopup.active {
				t.Fatalf("expected dirty :%s decision popup to open", tc.command)
			}
			if !model.overlay.confirmPopup.modal {
				t.Fatalf("expected dirty :%s decision popup to be modal", tc.command)
			}
			if model.overlay.confirmPopup.title != "Config" {
				t.Fatalf("expected dirty :%s popup title Config, got %q", tc.command, model.overlay.confirmPopup.title)
			}
			if model.ui.openConfigSelector {
				t.Fatalf("expected :%s navigation to remain blocked until explicit decision", tc.command)
			}
		})
	}
}

func TestHandleConfirmPopupKey_DirtyConfigCancelKeepsStagedState(t *testing.T) {
	model := &Model{
		read: runtimeReadState{
			viewMode:      ViewRecords,
			tables:        []dto.Table{{Name: "users"}},
			selectedTable: 0,
		},
		staging: testDatabaseStaging(map[string]stagingState{
			"users": {pendingInserts: []pendingInsertRow{{}}},
		}),
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEsc})

	if model.overlay.confirmPopup.active {
		t.Fatal("expected decision popup to close on cancel")
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged changes to stay untouched on cancel")
	}
	if model.ui.openConfigSelector {
		t.Fatal("expected no navigation on cancel")
	}
}

func TestHandleConfirmPopupKey_DirtyConfigDiscardClearsStateAndNavigates(t *testing.T) {
	model := &Model{
		read: runtimeReadState{
			viewMode:      ViewRecords,
			tables:        []dto.Table{{Name: "users"}},
			selectedTable: 0,
		},
		staging: testDatabaseStaging(map[string]stagingState{
			"users": {pendingInserts: []pendingInsertRow{{}}},
		}),
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected quit command after discard decision")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after discard decision, got %T", cmd())
	}
	if model.hasDirtyEdits() {
		t.Fatal("expected staged changes to be cleared on discard")
	}
	if !model.ui.openConfigSelector {
		t.Fatal("expected selector navigation after discard")
	}
}

func TestUpdate_DirtyConfigSaveSuccessNavigatesAfterSave(t *testing.T) {
	saveChanges := &spySaveChangesUseCase{}
	model := &Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		read: runtimeReadState{
			viewMode: ViewRecords,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
			tables: []dto.Table{{Name: "users"}},
		},
		staging: testDatabaseStaging(map[string]stagingState{
			"users": {
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
						{Name: "name", Type: "TEXT", Nullable: false},
					},
				},
				pendingInserts: []pendingInsertRow{
					{
						values: map[int]stagedEdit{
							0: {Value: dto.StagedValue{Text: "", Raw: ""}},
							1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
						},
						explicitAuto: map[int]bool{},
					},
				},
			},
		}),
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	_, saveCmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("expected save command after selecting save decision")
	}
	msg := saveCmd()
	_, quitCmd := model.Update(msg)

	if quitCmd == nil {
		t.Fatal("expected quit command after successful save decision")
	}
	if _, ok := quitCmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after successful save decision, got %T", quitCmd())
	}
	if model.hasDirtyEdits() {
		t.Fatal("expected staged changes to be cleared after successful save")
	}
	if !model.ui.openConfigSelector {
		t.Fatal("expected selector navigation after successful save")
	}
}

func TestUpdate_DirtyConfigSaveFailureKeepsStateAndBlocksNavigation(t *testing.T) {
	saveChanges := &spySaveChangesUseCase{err: errors.New("boom")}
	model := &Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		read: runtimeReadState{
			viewMode: ViewRecords,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
			tables: []dto.Table{{Name: "users"}},
		},
		staging: testDatabaseStaging(map[string]stagingState{
			"users": {
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
						{Name: "name", Type: "TEXT", Nullable: false},
					},
				},
				pendingInserts: []pendingInsertRow{
					{
						values: map[int]stagedEdit{
							0: {Value: dto.StagedValue{Text: "", Raw: ""}},
							1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
						},
						explicitAuto: map[int]bool{},
					},
				},
			},
		}),
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	_, saveCmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("expected save command after selecting save decision")
	}
	msg := saveCmd()
	_, quitCmd := model.Update(msg)

	if quitCmd != nil {
		if _, ok := quitCmd().(tea.QuitMsg); ok {
			t.Fatal("expected no navigation when save fails")
		}
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged changes to be preserved on save error")
	}
	if model.ui.openConfigSelector {
		t.Fatal("expected selector navigation to remain blocked on save error")
	}
	if !strings.Contains(model.ui.statusMessage, "boom") {
		t.Fatalf("expected save error status to be surfaced, got %q", model.ui.statusMessage)
	}
}
