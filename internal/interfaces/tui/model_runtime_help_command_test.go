package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_CommandConfigOpensRuntimeDatabaseSelectorPopup(t *testing.T) {
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
			model := &Model{
				read:                        runtimeReadState{viewMode: ViewRecords},
				runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
			}

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

func TestHandleKey_CommandEditReloadsCurrentDatabaseAndQuits(t *testing.T) {
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		read:                        runtimeReadState{viewMode: ViewRecords},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}

	updated, cmd := submitTypedRuntimeCommand(model, "edit")

	if cmd == nil {
		t.Fatal("expected :edit to quit runtime")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg for :edit, got %T", cmd())
	}
	runtimeModel := updated.(*Model)
	if runtimeModel.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected :edit to request runtime reopen, got %v", runtimeModel.exitResult.Action)
	}
	if runtimeModel.exitResult.NextDatabase.ConnString != current.ConnString {
		t.Fatalf("expected :edit to reload current database %q, got %q", current.ConnString, runtimeModel.exitResult.NextDatabase.ConnString)
	}
	if runtimeModel.overlay.commandInput.active {
		t.Fatal("expected :edit spotlight to close before runtime exit")
	}
}

func TestHandleKey_CommandEditWithConnectionStringRequestsReopenAndQuits(t *testing.T) {
	current := DatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	model := &Model{
		read:                        runtimeReadState{viewMode: ViewRecords},
		runtimeDatabaseSelectorDeps: runtimeDatabaseSelectorDepsForTest(current),
	}

	updated, cmd := submitTypedRuntimeCommand(model, "edit /tmp/analytics.sqlite")

	if cmd == nil {
		t.Fatal("expected :edit <conn> to quit runtime")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg for :edit <conn>, got %T", cmd())
	}
	runtimeModel := updated.(*Model)
	if runtimeModel.exitResult.Action != RuntimeExitActionOpenDatabaseNext {
		t.Fatalf("expected :edit <conn> to request runtime reopen, got %v", runtimeModel.exitResult.Action)
	}
	if runtimeModel.exitResult.NextDatabase.ConnString != "/tmp/analytics.sqlite" {
		t.Fatalf("expected :edit <conn> to target /tmp/analytics.sqlite, got %q", runtimeModel.exitResult.NextDatabase.ConnString)
	}
	if runtimeModel.overlay.commandInput.active {
		t.Fatal("expected :edit <conn> spotlight to close before runtime exit")
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

func TestHandleKey_CommandForcedQuitQuitsRuntime(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "short command", command: "q!"},
		{name: "full command", command: "quit!"},
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

func TestHandleKey_CommandSaveAndQuitQuitsRuntimeWhenClean(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
	}{
		{name: "short command", command: "wq"},
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

func TestHandleKey_CommandSaveAliasesStartSaveImmediately(t *testing.T) {
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
					viewMode:      ViewRecords,
					focus:         FocusContent,
					tables:        []dto.Table{{Name: "users"}},
					selectedTable: 0,
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER", PrimaryKey: true},
							{Name: "name", Type: "TEXT", Nullable: false},
						},
					},
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{
						{
							values: map[int]stagedEdit{
								0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
								1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
							},
							explicitAuto: map[int]bool{},
						},
					},
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
					viewMode:      ViewRecords,
					focus:         FocusContent,
					tables:        []dto.Table{{Name: "users"}},
					selectedTable: 0,
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER", PrimaryKey: true},
							{Name: "name", Type: "TEXT", Nullable: false},
						},
					},
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{
						{
							values: map[int]stagedEdit{
								0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
								1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
							},
							explicitAuto: map[int]bool{},
						},
					},
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
					viewMode:      ViewSchema,
					focus:         FocusContent,
					tables:        []dto.Table{{Name: "users"}},
					selectedTable: 0,
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER", PrimaryKey: true},
							{Name: "name", Type: "TEXT", Nullable: false},
						},
					},
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{
						{
							values: map[int]stagedEdit{
								0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
								1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
							},
							explicitAuto: map[int]bool{},
						},
					},
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
					viewMode:      ViewSchema,
					focus:         FocusTables,
					tables:        []dto.Table{{Name: "users"}},
					selectedTable: 0,
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER", PrimaryKey: true},
							{Name: "name", Type: "TEXT", Nullable: false},
						},
					},
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{
						{
							values: map[int]stagedEdit{
								0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
								1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
							},
							explicitAuto: map[int]bool{},
						},
					},
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
					viewMode:      ViewRecords,
					focus:         FocusContent,
					tables:        []dto.Table{{Name: "users"}},
					selectedTable: 0,
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER", PrimaryKey: true},
							{Name: "name", Type: "TEXT", Nullable: false},
						},
					},
				},
				staging: stagingState{
					pendingInserts: []pendingInsertRow{
						{
							values: map[int]stagedEdit{
								0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
								1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
							},
							explicitAuto: map[int]bool{},
						},
					},
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
			if model.overlay.confirmPopup.active {
				t.Fatalf("expected :%s not to open a save popup", tc.command)
			}
			if !model.ui.saveInFlight {
				t.Fatalf("expected :%s to start save immediately", tc.command)
			}
			if model.ui.statusMessage != "Saving changes..." {
				t.Fatalf("expected :%s to show saving status, got %q", tc.command, model.ui.statusMessage)
			}
		})
	}
}

func TestHandleKey_CommandSaveAndQuitStartsSaveImmediatelyWhenDirty(t *testing.T) {
	// Arrange
	model := &Model{
		ctx:         context.Background(),
		saveChanges: &spySaveChangesUseCase{},
		read: runtimeReadState{
			viewMode:      ViewRecords,
			focus:         FocusContent,
			tables:        []dto.Table{{Name: "users"}},
			selectedTable: 0,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
						1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
	}

	// Act
	_, cmd := submitTypedRuntimeCommand(model, "wq")

	// Assert
	assertRuntimeSessionActive(t, cmd, ":wq")
	if model.overlay.confirmPopup.active {
		t.Fatal("expected :wq not to open a save popup")
	}
	if !model.ui.saveInFlight {
		t.Fatal("expected :wq to start save immediately")
	}
	if !model.ui.pendingQuitAfterSave {
		t.Fatal("expected :wq to set pending quit flag")
	}
	if model.ui.statusMessage != "Saving changes..." {
		t.Fatalf("expected :wq to show saving status, got %q", model.ui.statusMessage)
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

func TestHandleKey_WKeyWithoutCommandPrefixDoesNotOpenSavePopup(t *testing.T) {
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
		t.Fatal("expected raw w to leave save popup closed")
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

func TestHandleKey_CommandHelpReenterPreservesExistingStatusWhileOpeningSpotlight(t *testing.T) {
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
	if model.ui.statusMessage != "existing status" {
		t.Fatalf("expected existing status message to remain while opening new command, got %q", model.ui.statusMessage)
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
