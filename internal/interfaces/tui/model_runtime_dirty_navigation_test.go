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
			// Arrange
			model := &Model{
				read:    runtimeReadState{viewMode: ViewRecords},
				staging: stagingState{pendingInserts: []pendingInsertRow{{}}},
			}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
			for _, r := range tc.command {
				model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
			_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert
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
	// Arrange
	model := &Model{
		read:    runtimeReadState{viewMode: ViewRecords},
		staging: stagingState{pendingInserts: []pendingInsertRow{{}}},
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
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
	// Arrange
	model := &Model{
		read:    runtimeReadState{viewMode: ViewRecords},
		staging: stagingState{pendingInserts: []pendingInsertRow{{}}},
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
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
	// Arrange
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
		staging: stagingState{
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
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	_, saveCmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("expected save command after selecting save decision")
	}
	msg := saveCmd()
	_, quitCmd := model.Update(msg)

	// Assert
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
	// Arrange
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
		staging: stagingState{
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
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	_, saveCmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("expected save command after selecting save decision")
	}
	msg := saveCmd()
	_, quitCmd := model.Update(msg)

	// Assert
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

func TestHandleKey_DirtyQuitCommandOpensDecisionPrompt(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "full command", command: "quit"},
		{name: "alias", command: "q"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{
				read:    runtimeReadState{viewMode: ViewRecords},
				staging: stagingState{pendingInserts: []pendingInsertRow{{}}},
			}

			// Act
			_, cmd := submitTypedRuntimeCommand(model, tc.command)

			// Assert
			assertRuntimeSessionActive(t, cmd, "dirty :"+tc.command)
			if !model.overlay.confirmPopup.active {
				t.Fatalf("expected dirty :%s decision popup to open", tc.command)
			}
			if !model.overlay.confirmPopup.modal {
				t.Fatalf("expected dirty :%s decision popup to be modal", tc.command)
			}
			if model.overlay.confirmPopup.title != "Quit" {
				t.Fatalf("expected dirty :%s popup title Quit, got %q", tc.command, model.overlay.confirmPopup.title)
			}
			if model.overlay.confirmPopup.message != "Quitting will cause loss of unsaved data (1 rows). Are you sure you want to discard unsaved data and quit?" {
				t.Fatalf("unexpected dirty :%s popup message: %q", tc.command, model.overlay.confirmPopup.message)
			}
			if len(model.overlay.confirmPopup.options) != 2 {
				t.Fatalf("expected dirty :%s popup to show two explicit options, got %d", tc.command, len(model.overlay.confirmPopup.options))
			}
		})
	}
}

func TestHandleConfirmPopupKey_DirtyQuitEscapeKeepsStagedState(t *testing.T) {
	// Arrange
	model := &Model{
		read:    runtimeReadState{viewMode: ViewRecords},
		staging: stagingState{pendingInserts: []pendingInsertRow{{}}},
	}
	submitTypedRuntimeCommand(model, "quit")

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	assertRuntimeSessionActive(t, cmd, "dirty quit cancel")
	if model.overlay.confirmPopup.active {
		t.Fatal("expected decision popup to close on escape")
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged changes to stay untouched on escape")
	}
}

func TestHandleConfirmPopupKey_DirtyQuitNKeyIsIgnored(t *testing.T) {
	// Arrange
	model := &Model{
		read:    runtimeReadState{viewMode: ViewRecords},
		staging: stagingState{pendingInserts: []pendingInsertRow{{}}},
	}
	submitTypedRuntimeCommand(model, "quit")
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Assert
	assertRuntimeSessionActive(t, cmd, "dirty quit n ignored")
	if !model.overlay.confirmPopup.active {
		t.Fatal("expected decision popup to remain open on n key")
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged changes to stay untouched on n key")
	}
}

func TestHandleConfirmPopupKey_DirtyQuitYKeyIsIgnored(t *testing.T) {
	// Arrange
	model := &Model{
		read:    runtimeReadState{viewMode: ViewRecords},
		staging: stagingState{pendingInserts: []pendingInsertRow{{}}},
	}
	submitTypedRuntimeCommand(model, "quit")

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	// Assert
	assertRuntimeSessionActive(t, cmd, "dirty quit y ignored")
	if !model.overlay.confirmPopup.active {
		t.Fatal("expected decision popup to remain open on y key")
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged changes to stay untouched on y key")
	}
}
