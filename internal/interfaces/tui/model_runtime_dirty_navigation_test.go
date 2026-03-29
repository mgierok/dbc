package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKey_DirtyConfigCommandOpensRuntimeDatabaseSelectorPopup(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "full command", command: "config"},
		{name: "alias", command: "c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			current := DatabaseOption{
				Name:       "primary",
				ConnString: "/tmp/primary.sqlite",
				Source:     DatabaseOptionSourceConfig,
			}
			model := withTestStaging(&Model{
				read:                        runtimeReadState{viewMode: ViewRecords},
				runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
			}, stagingState{pendingInserts: []pendingInsertRow{{}}})

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
			for _, r := range tc.command {
				model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
			_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert
			if cmd != nil {
				if _, ok := cmd().(tea.QuitMsg); ok {
					t.Fatalf("expected dirty :%s to keep runtime active", tc.command)
				}
			}
			if !model.overlay.databaseSelector.active {
				t.Fatalf("expected dirty :%s to open runtime database selector popup", tc.command)
			}
		})
	}
}

func TestHandleKey_DirtyRuntimeDatabaseSelectionOpensReloadDecisionPrompt(t *testing.T) {
	// Arrange
	current := runtimeTestDatabaseOption()
	model := withTestStaging(&Model{
		read:                        runtimeReadState{viewMode: ViewRecords},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}, stagingState{pendingInserts: []pendingInsertRow{{}}})
	submitTypedRuntimeCommand(model, "config")

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	assertRuntimeSessionActive(t, cmd, "dirty config selection")
	if !model.overlay.confirmPopup.active {
		t.Fatal("expected dirty selection to open decision popup")
	}
	if model.overlay.confirmPopup.title != "Reload Database" {
		t.Fatalf("expected reload database prompt title, got %q", model.overlay.confirmPopup.title)
	}
	if !strings.Contains(model.overlay.confirmPopup.message, "Reloading the current database") {
		t.Fatalf("expected reload decision message, got %q", model.overlay.confirmPopup.message)
	}
}

func TestHandleConfirmPopupKey_DirtyDatabaseTransitionCancelKeepsSelectorOpenAndStagedState(t *testing.T) {
	// Arrange
	current := runtimeTestDatabaseOption()
	model := withTestStaging(&Model{
		read:                        runtimeReadState{viewMode: ViewRecords},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}, stagingState{pendingInserts: []pendingInsertRow{{}}})
	submitTypedRuntimeCommand(model, "config")
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Act
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.overlay.confirmPopup.active {
		t.Fatal("expected decision popup to close on cancel")
	}
	if !model.overlay.databaseSelector.active {
		t.Fatal("expected runtime database selector popup to remain open on cancel")
	}
	if !model.hasDirtyEdits() {
		t.Fatal("expected staged changes to stay untouched on cancel")
	}
}

func TestHandleConfirmPopupKey_DirtyDatabaseTransitionDiscardClearsStateAndStartsTransition(t *testing.T) {
	// Arrange
	current := runtimeTestDatabaseOption()
	model := withTestStaging(&Model{
		read:                        runtimeReadState{viewMode: ViewRecords},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}, stagingState{pendingInserts: []pendingInsertRow{{}}})
	submitTypedRuntimeCommand(model, "config")
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	assertQuitCommand(t, cmd, "discard database navigation decision")
	if model.hasDirtyEdits() {
		t.Fatal("expected staged changes to be cleared on discard")
	}
	if model.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected discard decision to request runtime reopen, got %v", model.exitResult.Action)
	}
}

func TestUpdate_DirtyDatabaseTransitionSaveSuccessContinuesTransition(t *testing.T) {
	// Arrange
	model := newDirtyRuntimeNavigationModel(&spySaveChangesUseCase{})
	submitTypedRuntimeCommand(model, "edit")

	// Act
	_, saveCmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if saveCmd == nil {
		t.Fatal("expected save command after selecting save decision")
	}
	msg := saveCmd()
	_, followupCmd := model.Update(msg)

	// Assert
	assertQuitCommand(t, followupCmd, "save-then-reopen flow")
	if model.hasDirtyEdits() {
		t.Fatal("expected staged changes to be cleared after successful save")
	}
	if model.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected successful save to request runtime reopen, got %v", model.exitResult.Action)
	}
}

func TestUpdate_DirtyDatabaseTransitionSaveFailureKeepsStateAndBlocksTransition(t *testing.T) {
	// Arrange
	model := newDirtyRuntimeNavigationModel(&spySaveChangesUseCase{err: errors.New("boom")})
	submitTypedRuntimeCommand(model, "edit")

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
	if !strings.Contains(model.ui.statusMessage, "boom") {
		t.Fatalf("expected save error status to be surfaced, got %q", model.ui.statusMessage)
	}
	if !model.overlay.commandInput.active {
		t.Fatal("expected :edit spotlight to reopen after save failure")
	}
	if model.overlay.commandInput.value != "edit" {
		t.Fatalf("expected :edit spotlight to preserve submitted command after save failure, got %q", model.overlay.commandInput.value)
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
			model := withTestStaging(&Model{
				read: runtimeReadState{viewMode: ViewRecords},
			}, stagingState{pendingInserts: []pendingInsertRow{{}}})

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

func TestHandleKey_DirtyForcedQuitCommandDiscardsStateWithoutPrompt(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "full command", command: "quit!"},
		{name: "alias", command: "q!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := withTestStaging(&Model{
				read: runtimeReadState{viewMode: ViewRecords},
			}, stagingState{pendingInserts: []pendingInsertRow{{}}})

			// Act
			_, cmd := submitTypedRuntimeCommand(model, tc.command)

			// Assert
			if cmd == nil {
				t.Fatalf("expected quit command for dirty :%s", tc.command)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatalf("expected tea.QuitMsg for dirty :%s, got %T", tc.command, cmd())
			}
			if model.overlay.confirmPopup.active {
				t.Fatalf("expected dirty :%s to bypass confirmation popup", tc.command)
			}
			if model.hasDirtyEdits() {
				t.Fatalf("expected dirty :%s to clear staged changes before quit", tc.command)
			}
		})
	}
}

func TestHandleConfirmPopupKey_DirtyQuitNonAcceptKeysPreserveState(t *testing.T) {
	for _, tc := range []struct {
		name               string
		key                tea.KeyMsg
		prepareSelection   bool
		expectPopupOpen    bool
		assertActiveCtxMsg string
	}{
		{
			name:               "escape closes prompt",
			key:                tea.KeyMsg{Type: tea.KeyEsc},
			expectPopupOpen:    false,
			assertActiveCtxMsg: "dirty quit cancel",
		},
		{
			name:               "n key is ignored",
			key:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
			prepareSelection:   true,
			expectPopupOpen:    true,
			assertActiveCtxMsg: "dirty quit n ignored",
		},
		{
			name:               "y key is ignored",
			key:                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			expectPopupOpen:    true,
			assertActiveCtxMsg: "dirty quit y ignored",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := withTestStaging(&Model{
				read: runtimeReadState{viewMode: ViewRecords},
			}, stagingState{pendingInserts: []pendingInsertRow{{}}})
			submitTypedRuntimeCommand(model, "quit")
			if tc.prepareSelection {
				model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
			}

			// Act
			_, cmd := model.handleConfirmPopupKey(tc.key)

			// Assert
			assertRuntimeSessionActive(t, cmd, tc.assertActiveCtxMsg)
			if model.overlay.confirmPopup.active != tc.expectPopupOpen {
				t.Fatalf("expected prompt active=%t, got %t", tc.expectPopupOpen, model.overlay.confirmPopup.active)
			}
			if !model.hasDirtyEdits() {
				t.Fatal("expected staged changes to stay untouched")
			}
		})
	}
}
