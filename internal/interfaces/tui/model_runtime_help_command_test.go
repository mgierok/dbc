package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKey_CommandConfigQuitsToOpenSelector(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "full command", command: "config"},
		{name: "alias", command: "c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

			// Act
			_, cmd := submitTypedRuntimeCommand(model, tc.command)

			// Assert
			if cmd == nil {
				t.Fatalf("expected quit command for :%s", tc.command)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatalf("expected tea.QuitMsg for :%s, got %T", tc.command, cmd())
			}
		})
	}
}

func TestHandleKey_CommandQuitQuitsRuntime(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "short command", command: "q"},
		{name: "full command", command: "quit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

			// Act
			_, cmd := submitTypedRuntimeCommand(model, tc.command)

			// Assert
			if cmd == nil {
				t.Fatalf("expected quit command for :%s", tc.command)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatalf("expected tea.QuitMsg for :%s, got %T", tc.command, cmd())
			}
		})
	}
}

func TestHandleKey_CommandSaveAliasesOpenSaveConfirmation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
		model   *Model
	}{
		{
			name:    "short command in records",
			command: "w",
			model: &Model{
				ctx:         context.Background(),
				saveChanges: &spySaveChangesUseCase{},
				read: runtimeReadState{
					viewMode: ViewRecords,
					focus:    FocusContent,
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{{}},
				},
			},
		},
		{
			name:    "full command in records",
			command: "write",
			model: &Model{
				ctx:         context.Background(),
				saveChanges: &spySaveChangesUseCase{},
				read: runtimeReadState{
					viewMode: ViewRecords,
					focus:    FocusContent,
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{{}},
				},
			},
		},
		{
			name:    "save from schema",
			command: "w",
			model: &Model{
				ctx:         context.Background(),
				saveChanges: &spySaveChangesUseCase{},
				read: runtimeReadState{
					viewMode: ViewSchema,
					focus:    FocusContent,
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{{}},
				},
			},
		},
		{
			name:    "save from tables",
			command: "w",
			model: &Model{
				ctx:         context.Background(),
				saveChanges: &spySaveChangesUseCase{},
				read: runtimeReadState{
					viewMode: ViewSchema,
					focus:    FocusTables,
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{{}},
				},
			},
		},
		{
			name:    "save from record detail",
			command: "w",
			model: &Model{
				ctx:         context.Background(),
				saveChanges: &spySaveChangesUseCase{},
				read: runtimeReadState{
					viewMode: ViewRecords,
					focus:    FocusContent,
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{{}},
				},
				overlay: runtimeOverlayState{
					recordDetail: recordDetailState{active: true},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := tc.model

			// Act
			_, cmd := submitTypedRuntimeCommand(model, tc.command)

			// Assert
			assertRuntimeSessionActive(t, cmd, ":"+tc.command)
			if !model.overlay.confirmPopup.active {
				t.Fatalf("expected :%s to open save confirmation", tc.command)
			}
			if model.overlay.confirmPopup.action != confirmSave {
				t.Fatalf("expected :%s to open save action, got %v", tc.command, model.overlay.confirmPopup.action)
			}
		})
	}
}

func TestHandleKey_ContextHelpQuestionMarkOpensRecordsHelpPopup(t *testing.T) {
	// Arrange
	model := newRuntimeHelpCommandModel()

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Assert
	if !model.overlay.helpPopup.active {
		t.Fatal("expected ? to open help popup")
	}
	if model.overlay.helpPopup.context != helpPopupContextRecords {
		t.Fatalf("expected records context, got %v", model.overlay.helpPopup.context)
	}
}

func TestHandleKey_CommandHelpAliasOpensRecordsHelpPopup(t *testing.T) {
	// Arrange
	model := newRuntimeHelpCommandModel()

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "help")

	// Assert
	assertRuntimeSessionActive(t, cmd, ":help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected :help to open help popup")
	}
	if model.overlay.helpPopup.context != helpPopupContextRecords {
		t.Fatalf("expected records context for :help, got %v", model.overlay.helpPopup.context)
	}
	if strings.Contains(strings.ToLower(model.ui.statusMessage), "unknown command") {
		t.Fatalf("expected no unknown-command status for :help, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_ContextHelpPopupShowsCurrentContextBindings(t *testing.T) {
	// Arrange
	model := newRuntimeHelpCommandModel()
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Act
	popup := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Records: Esc tables") {
		t.Fatalf("expected records shortcuts in context help popup, got %q", popup)
	}
	if strings.Contains(popup, "Supported Commands") || strings.Contains(popup, "Supported Keywords") {
		t.Fatalf("expected context-only help content, got %q", popup)
	}
}

func TestHandleKey_ContextHelpPopupShowsSaveInSchemaTablesAndRecordDetail(t *testing.T) {
	for _, tc := range []struct {
		name           string
		model          *Model
		expectedHeader string
		expectedRow    string
	}{
		{
			name: "schema",
			model: &Model{
				ui: runtimeUIState{height: 40},
				read: runtimeReadState{
					viewMode: ViewSchema,
					focus:    FocusContent,
				},
			},
			expectedHeader: "Context Help: Schema",
			expectedRow:    "Schema: Esc tables",
		},
		{
			name: "tables",
			model: &Model{
				ui: runtimeUIState{height: 40},
				read: runtimeReadState{
					viewMode: ViewSchema,
					focus:    FocusTables,
				},
			},
			expectedHeader: "Context Help: Tables",
			expectedRow:    "Tables: Enter records",
		},
		{
			name: "record detail",
			model: &Model{
				ui: runtimeUIState{height: 40},
				read: runtimeReadState{
					viewMode: ViewRecords,
					focus:    FocusContent,
				},
				overlay: runtimeOverlayState{
					recordDetail: recordDetailState{active: true},
				},
			},
			expectedHeader: "Context Help: Record Detail",
			expectedRow:    "Detail: Esc back",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange

			// Act
			tc.model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
			popup := strings.Join(tc.model.renderHelpPopup(60), "\n")

			// Assert
			if !tc.model.overlay.helpPopup.active {
				t.Fatal("expected help popup to open")
			}
			if !strings.Contains(popup, tc.expectedHeader) {
				t.Fatalf("expected help popup header %q, got %q", tc.expectedHeader, popup)
			}
			if !strings.Contains(popup, tc.expectedRow) {
				t.Fatalf("expected context row %q, got %q", tc.expectedRow, popup)
			}
			if !strings.Contains(popup, ":w / :write save") {
				t.Fatalf("expected save shortcut in %s help popup, got %q", tc.name, popup)
			}
		})
	}
}

func TestHandleKey_HelpPopupScrollCanReachFinalContextShortcut(t *testing.T) {
	// Arrange
	model := newRuntimeHelpCommandModel()
	model.ui.height = 12
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	initial := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
	for range 30 {
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	scrolled := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if strings.Contains(initial, "Shift+S sort") {
		t.Fatalf("expected final help item to be hidden before scrolling, got %q", initial)
	}
	if !strings.Contains(scrolled, "Shift+S sort") {
		t.Fatalf("expected final help item to be reachable after scrolling, got %q", scrolled)
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
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

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
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	assertRuntimeSessionActive(t, cmd, "command without prefix")
}

func TestHandleKey_QKeyWithoutCommandPrefixKeepsSessionActive(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Assert
	assertRuntimeSessionActive(t, cmd, "q without prefix")
}

func TestHandleKey_WKeyWithoutCommandPrefixDoesNotOpenSaveConfirmation(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
		},
	}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})

	// Assert
	assertRuntimeSessionActive(t, cmd, "w without prefix")
	if model.overlay.confirmPopup.active {
		t.Fatal("expected raw w to leave save confirmation closed")
	}
}

func TestHandleKey_CtrlCWithoutCommandPrefixKeepsSessionActive(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Assert
	assertRuntimeSessionActive(t, cmd, "ctrl+c without prefix")
}

func TestHandleKey_CommandHelpReenterKeepsPopupOpen(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}
	submitTypedRuntimeCommand(model, "help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to open before idempotence check")
	}

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "help")

	// Assert
	assertRuntimeSessionActive(t, cmd, "repeated :help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to remain open when :help is re-entered")
	}
	if strings.Contains(strings.ToLower(model.ui.statusMessage), "unknown command") {
		t.Fatalf("expected no unknown-command status for repeated :help, got %q", model.ui.statusMessage)
	}
}

func TestHandleKey_CommandHelpReenterClearsStaleStatusBeforeNewCommand(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{viewMode: ViewRecords},
		ui:   runtimeUIState{statusMessage: "existing status"},
	}
	submitTypedRuntimeCommand(model, "help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to open before re-entering :help")
	}

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "help")

	// Assert
	assertRuntimeSessionActive(t, cmd, "repeated :help after stale status")
	if model.ui.statusMessage != "" {
		t.Fatalf("expected stale status message to clear before new command, got %q", model.ui.statusMessage)
	}
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to remain open")
	}
}

func TestHandleKey_HelpPopupEscClosesPopup(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}
	submitTypedRuntimeCommand(model, "help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to open before close check")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.overlay.helpPopup.active {
		t.Fatal("expected Esc to close help popup")
	}
}

func TestHandleKey_HelpPopupColonDoesNotOpenCommandInput(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		overlay: runtimeOverlayState{
			helpPopup: helpPopup{
				active:  true,
				context: helpPopupContextRecords,
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})

	// Assert
	if model.overlay.commandInput.active {
		t.Fatal("expected help popup to block command input")
	}
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to stay open after :")
	}
}

func TestHandleKey_HelpPopupUnrelatedKeysDoNotClosePopup(t *testing.T) {
	// Arrange
	model := &Model{read: runtimeReadState{viewMode: ViewRecords}}
	submitTypedRuntimeCommand(model, "help")
	if !model.overlay.helpPopup.active {
		t.Fatal("expected help popup to open before unrelated-key check")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.overlay.helpPopup.active {
		t.Fatal("expected unrelated keys to keep help popup open")
	}
}

func TestHandleKey_ContextHelpFromFilterPopupUsesFilterContext(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active: true,
				step:   filterSelectColumn,
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Assert
	if !model.overlay.helpPopup.active {
		t.Fatal("expected ? to open help popup from filter context")
	}
	if model.overlay.helpPopup.context != helpPopupContextFilterPopup {
		t.Fatalf("expected filter-popup context help, got %v", model.overlay.helpPopup.context)
	}
	if !model.overlay.filterPopup.active {
		t.Fatal("expected filter popup state to stay preserved under help overlay")
	}
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

func TestHandleKey_PopupPriority_HelpPopupConsumesEscBeforeOtherPopups(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			helpPopup:   helpPopup{active: true},
			filterPopup: filterPopup{active: true},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.overlay.helpPopup.active {
		t.Fatal("expected help popup to close first")
	}
	if !model.overlay.filterPopup.active {
		t.Fatal("expected filter popup to remain active when help popup handled Esc")
	}
}

func newRuntimeHelpCommandModel() *Model {
	return &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
		ui: runtimeUIState{
			height: 40,
		},
	}
}

func submitTypedRuntimeCommand(model *Model, value string) (tea.Model, tea.Cmd) {
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range value {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
}

func assertRuntimeSessionActive(t *testing.T, cmd tea.Cmd, context string) {
	t.Helper()
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("expected %s to keep session active", context)
		}
	}
}
