package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestHandleKey_CommandConfigOpensRuntimeDatabaseSelectorPopup(t *testing.T) {
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}

	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "full command", command: "config"},
		{name: "alias", command: "c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := newRuntimeCommandModelWithCurrentDatabase(current)

			// Act
			updated, cmd := submitTypedRuntimeCommand(model, tc.command)

			// Assert
			assertRuntimeSessionActive(t, cmd, ":"+tc.command)
			runtimeModel := updated.(*Model)
			if !runtimeModel.overlay.databaseSelector.active {
				t.Fatalf("expected runtime database selector popup for :%s", tc.command)
			}
		})
	}
}

func TestHandleKey_CommandEditRequestsRuntimeReopenAndQuits(t *testing.T) {
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}

	for _, tc := range []struct {
		name           string
		command        string
		expectedTarget string
	}{
		{
			name:           "reload current database",
			command:        "edit",
			expectedTarget: current.ConnString,
		},
		{
			name:           "open explicit connection string",
			command:        "edit /tmp/analytics.sqlite",
			expectedTarget: "/tmp/analytics.sqlite",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := newRuntimeCommandModelWithCurrentDatabase(current)

			// Act
			updated, cmd := submitTypedRuntimeCommand(model, tc.command)

			// Assert
			assertQuitCommand(t, cmd, ":"+tc.command)
			runtimeModel := updated.(*Model)
			if runtimeModel.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
				t.Fatalf("expected :%s to request runtime reopen, got %v", tc.command, runtimeModel.exitResult.Action)
			}
			if runtimeModel.exitResult.NextDatabase.ConnString != tc.expectedTarget {
				t.Fatalf(
					"expected :%s to target %q, got %q",
					tc.command,
					tc.expectedTarget,
					runtimeModel.exitResult.NextDatabase.ConnString,
				)
			}
			if runtimeModel.overlay.commandInput.active {
				t.Fatalf("expected :%s spotlight to close before runtime exit", tc.command)
			}
		})
	}
}

func TestHandleKey_RuntimeExitCommandsQuitRuntime(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "quit alias", command: "q"},
		{name: "quit command", command: "quit"},
		{name: "forced quit alias", command: "q!"},
		{name: "forced quit command", command: "quit!"},
		{name: "save and quit when clean", command: "wq"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

			// Act
			_, cmd := submitTypedRuntimeCommand(model, tc.command)

			// Assert
			assertQuitCommand(t, cmd, ":"+tc.command)
		})
	}
}

func TestHandleKey_CommandSaveAliasesStartSaveImmediately(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
		model   *Model
	}{
		{
			name:    "short command in records",
			command: "w",
			model:   newDirtyRuntimeSaveModel(ViewRecords, FocusContent),
		},
		{
			name:    "full command in records",
			command: "write",
			model:   newDirtyRuntimeSaveModel(ViewRecords, FocusContent),
		},
		{
			name:    "save from schema",
			command: "w",
			model:   newDirtyRuntimeSaveModel(ViewSchema, FocusContent),
		},
		{
			name:    "save from tables",
			command: "w",
			model:   newDirtyRuntimeSaveModel(ViewSchema, FocusTables),
		},
		{
			name:    "save from record detail",
			command: "w",
			model: func() *Model {
				model := newDirtyRuntimeSaveModel(ViewRecords, FocusContent)
				model.overlay.recordDetail.active = true
				return model
			}(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			_, cmd := submitTypedRuntimeCommand(tc.model, tc.command)

			// Assert
			assertRuntimeSessionActive(t, cmd, ":"+tc.command)
			assertRuntimeSaveStarted(t, tc.model, ":"+tc.command)
		})
	}
}

func TestHandleKey_CommandSaveAndQuitStartsSaveImmediatelyWhenDirty(t *testing.T) {
	// Arrange
	model := newDirtyRuntimeSaveModel(ViewRecords, FocusContent)

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "wq")

	// Assert
	assertRuntimeSessionActive(t, cmd, ":wq")
	assertRuntimeSaveStarted(t, model, ":wq")
	if model.ui.pendingSaveSuccessAction != usecase.RuntimeSaveSuccessActionQuitRuntime {
		t.Fatalf("expected :wq to set quit action, got %v", model.ui.pendingSaveSuccessAction)
	}
}

func TestHandleKey_InvalidCommandShowsErrorAndKeepsSessionActive(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "unknown")

	// Assert
	assertRuntimeSessionActive(t, cmd, "invalid command")
	if !strings.Contains(strings.ToLower(model.ui.statusMessage), "unknown command") {
		t.Fatalf("expected unknown command status message, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_HelpCommandRequiresExplicitPrefix(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

	// Act
	for _, r := range "help" {
		model.handleKey(keyRunesMsg(r))
	}
	_, cmd := model.handleKey(keyEnterMsg())

	// Assert
	assertRuntimeSessionActive(t, cmd, "help without prefix")
	if model.overlay.helpPopup.active {
		t.Fatal("expected help popup to stay closed without ':' prefix")
	}
}

func TestHandleKey_CommandRequiresExplicitPrefix(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

	// Act
	for _, r := range "config" {
		model.handleKey(keyRunesMsg(r))
	}
	_, cmd := model.handleKey(keyEnterMsg())

	// Assert
	assertRuntimeSessionActive(t, cmd, "command without prefix")
}

func TestHandleKey_QKeyWithoutCommandPrefixKeepsSessionActive(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

	// Act
	_, cmd := model.handleKey(keyRunesMsg('q'))

	// Assert
	assertRuntimeSessionActive(t, cmd, "q without prefix")
}

func TestHandleKey_WKeyWithoutCommandPrefixDoesNotOpenSavePopup(t *testing.T) {
	// Arrange
	model := withTestStaging(&Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
	}, stagingState{
		pendingInserts: []pendingInsertRow{{}},
	})

	// Act
	_, cmd := model.handleKey(keyRunesMsg('w'))

	// Assert
	assertRuntimeSessionActive(t, cmd, "w without prefix")
	if model.overlay.confirmPopup.active {
		t.Fatal("expected raw w to leave save popup closed")
	}
}

func TestHandleKey_CtrlCWithoutCommandPrefixKeepsSessionActive(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

	// Act
	_, cmd := model.handleKey(keyCtrlCMsg())

	// Assert
	assertRuntimeSessionActive(t, cmd, "ctrl+c without prefix")
}

func TestHandleKey_MisspelledHelpCommandUsesUnknownCommandFallback(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "helpp")

	// Assert
	assertRuntimeSessionActive(t, cmd, "misspelled :help")
	if model.overlay.helpPopup.active {
		t.Fatal("expected misspelled :help to keep help popup closed")
	}
	if !strings.Contains(strings.ToLower(model.ui.statusMessage), "unknown command") {
		t.Fatalf("expected unknown-command status for misspelled :help, got %q", model.ui.statusMessage)
	}
}

func keyRunesMsg(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func keyEnterMsg() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyEnter}
}

func keyCtrlCMsg() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyCtrlC}
}
