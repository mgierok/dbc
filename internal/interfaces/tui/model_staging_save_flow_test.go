package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestConfirmSaveChanges_SubmitsBuiltTableChanges(t *testing.T) {
	// Arrange
	saveChanges := &spySaveChangesUseCase{count: 1}
	model := withTestStaging(&Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
			tables: []dto.Table{{Name: "users"}},
		},
	}, stagingState{
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: dto.StagedValue{Text: "", Raw: ""}},
					1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
	})

	// Act
	_, cmd := model.confirmSaveChanges()
	msg := cmd()

	// Assert
	result, ok := msg.(saveChangesMsg)
	if !ok {
		t.Fatalf("expected saveChangesMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("expected no error, got %v", result.err)
	}
	if result.count != 1 {
		t.Fatalf("expected save message count 1, got %d", result.count)
	}
	if len(saveChanges.lastChanges.Inserts) != 1 {
		t.Fatalf("expected one insert payload, got %d", len(saveChanges.lastChanges.Inserts))
	}
}

func TestConfirmSaveChanges_UsesAppliedRowCountFromUseCaseInsteadOfDirtyRowCount(t *testing.T) {
	// Arrange
	saveChanges := &spySaveChangesUseCase{count: 1}
	model := withTestStaging(&Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT", Nullable: false},
					{Name: "email", Type: "TEXT", Nullable: false},
				},
			},
			tables: []dto.Table{{Name: "users"}},
		},
	}, stagingState{
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
				},
				changes: map[int]stagedEdit{
					1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
					2: {Value: dto.StagedValue{Text: "alice@example.com", Raw: "alice@example.com"}},
				},
			},
		},
	})

	// Act
	_, cmd := model.confirmSaveChanges()
	msg := cmd()

	// Assert
	result, ok := msg.(saveChangesMsg)
	if !ok {
		t.Fatalf("expected saveChangesMsg, got %T", msg)
	}
	if result.count != 1 {
		t.Fatalf("expected applied row count 1 from use case, got %d", result.count)
	}
}

func TestConfirmSaveChanges_StartsBlockingSaveStateAndShowsSavingStatus(t *testing.T) {
	// Arrange
	model := newDirtyRuntimeSaveModel(ViewRecords, FocusContent)
	model.ui.statusMessage = "stale status"
	model.saveChanges = &spySaveChangesUseCase{count: 1}

	// Act
	_, cmd := model.confirmSaveChanges()

	// Assert
	if cmd == nil {
		t.Fatal("expected save command to be returned")
	}
	assertRuntimeSaveStarted(t, model, "confirmSaveChanges")
}

func TestSetTableSelection_WithDirtyStateOpensInformationalSwitchTablePopup(t *testing.T) {
	// Arrange
	model := newDirtyTableSwitchModel()

	// Act
	model.setTableSelection(1)

	// Assert
	if !model.overlay.confirmPopup.active {
		t.Fatalf("expected discard confirmation popup")
	}
	if !model.overlay.confirmPopup.modal {
		t.Fatalf("expected table switch popup to be modal")
	}
	if model.overlay.confirmPopup.title != "Switch Table" {
		t.Fatalf("expected switch table title, got %q", model.overlay.confirmPopup.title)
	}
	if !strings.Contains(model.overlay.confirmPopup.message, "Switching tables will cause loss of unsaved data (3 rows).") {
		t.Fatalf("expected message with unsaved row count, got %q", model.overlay.confirmPopup.message)
	}
	if !strings.Contains(model.overlay.confirmPopup.message, "Are you sure you want to discard unsaved data?") {
		t.Fatalf("expected discard confirmation question, got %q", model.overlay.confirmPopup.message)
	}
	if len(model.overlay.confirmPopup.options) != 2 {
		t.Fatalf("expected two explicit options, got %d", len(model.overlay.confirmPopup.options))
	}
	if model.overlay.confirmPopup.options[0].label != "Discard changes and switch table" {
		t.Fatalf("expected discard option label, got %q", model.overlay.confirmPopup.options[0].label)
	}
	if model.overlay.confirmPopup.options[1].label != "Continue editing" {
		t.Fatalf("expected continue-editing option label, got %q", model.overlay.confirmPopup.options[1].label)
	}
	if model.read.selectedTable != 0 {
		t.Fatalf("expected table selection not to change before confirmation")
	}
}

func TestSetTableSelection_WithDirtyStatePopupDecisions(t *testing.T) {
	for _, tc := range []struct {
		name                string
		keys                []tea.KeyMsg
		wantSelectedTable   int
		wantDirty           bool
		wantPopupActive     bool
		wantPendingNav      bool
		wantPendingTable    string
		wantPendingNavReset bool
	}{
		{
			name:              "discard switches table and clears staging",
			keys:              []tea.KeyMsg{{Type: tea.KeyEnter}},
			wantSelectedTable: 1,
			wantDirty:         false,
			wantPopupActive:   false,
		},
		{
			name: "continue editing preserves staging",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune{'j'}},
				{Type: tea.KeyEnter},
			},
			wantSelectedTable:   0,
			wantDirty:           true,
			wantPopupActive:     false,
			wantPendingNavReset: true,
		},
		{
			name: "n keeps popup open and navigation pending",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune{'n'}},
			},
			wantSelectedTable: 0,
			wantDirty:         true,
			wantPopupActive:   true,
			wantPendingNav:    true,
			wantPendingTable:  "orders",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := newDirtyTableSwitchModel()
			model.setTableSelection(1)

			// Act
			for _, key := range tc.keys {
				model.handleConfirmPopupKey(key)
			}

			// Assert
			if model.read.selectedTable != tc.wantSelectedTable {
				t.Fatalf("expected selected table %d, got %d", tc.wantSelectedTable, model.read.selectedTable)
			}
			if model.hasDirtyEdits() != tc.wantDirty {
				t.Fatalf("expected dirty state %t, got %t", tc.wantDirty, model.hasDirtyEdits())
			}
			if model.overlay.confirmPopup.active != tc.wantPopupActive {
				t.Fatalf("expected popup active=%t, got %t", tc.wantPopupActive, model.overlay.confirmPopup.active)
			}
			if tc.wantPendingNav && model.ui.pendingNavigation == nil {
				t.Fatal("expected pending navigation to remain set")
			}
			if !tc.wantPendingNav && tc.wantPendingNavReset && model.ui.pendingNavigation != nil {
				t.Fatalf("expected pending navigation reset, got %+v", model.ui.pendingNavigation)
			}
			if tc.wantPendingTable != "" && model.ui.pendingNavigation.Action.TargetTableName != tc.wantPendingTable {
				t.Fatalf("expected pending table target %q, got %q", tc.wantPendingTable, model.ui.pendingNavigation.Action.TargetTableName)
			}
		})
	}
}

func TestSetTableSelection_ClearsSortOnTableSwitch(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 0,
			currentSort: &dto.Sort{
				Column:    "name",
				Direction: dto.SortDirectionAsc,
			},
		},
	}

	// Act
	model.setTableSelection(1)

	// Assert
	if model.read.selectedTable != 1 {
		t.Fatalf("expected selected table to switch, got %d", model.read.selectedTable)
	}
	if model.read.currentSort != nil {
		t.Fatalf("expected sort to reset on table switch, got %+v", model.read.currentSort)
	}
}

func TestHandleConfirmPopupKey_SaveDecisionForPendingDatabaseNavigationStartsSaveAndKeepsPendingNavigation(t *testing.T) {
	// Arrange
	model := newPendingDatabaseNavigationConfirmModel(&spySaveChangesUseCase{})

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected save command to be returned")
	}
	if model.ui.pendingNavigation == nil {
		t.Fatal("expected pending navigation to stay set for save-start flow")
	}
}

func TestHandleConfirmPopupKey_SaveDecisionForPendingDatabaseNavigationCleansPendingStateWhenSaveFailsBeforeStart(t *testing.T) {
	// Arrange
	model := withTestStaging(&Model{
		saveChanges: &spySaveChangesUseCase{},
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			tables:   []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
		},
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active: true,
				options: []confirmOption{
					{label: "Save and reload database", decisionID: usecase.DirtyDecisionSave},
				},
			},
		},
		ui: runtimeUIState{
			pendingNavigation: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind: usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: usecase.RuntimeDatabaseTarget{
						Option:         runtimeDatabaseOptionFromSelectorOption(runtimeTestDatabaseOption()),
						TransitionKind: usecase.RuntimeDatabaseTransitionReloadCurrent,
					},
				},
			},
			pendingCommandInput: "edit",
		},
	}, stagingState{
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: dto.StagedValue{Text: "", Raw: ""}},
				},
				explicitAuto: map[int]bool{},
			},
		},
	})

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected synchronous save-start failure not to return async command")
	}
	if model.ui.saveInFlight {
		t.Fatal("expected failed pre-save path not to enter save-in-flight state")
	}
	if model.ui.pendingNavigation != nil {
		t.Fatalf("expected pending navigation cleanup after pre-save failure, got %+v", model.ui.pendingNavigation)
	}
	if model.ui.pendingCommandInput != "" {
		t.Fatalf("expected pending command input cleanup after pre-save failure, got %q", model.ui.pendingCommandInput)
	}
	if !model.overlay.commandInput.active {
		t.Fatal("expected :edit spotlight to reopen after pre-save failure")
	}
	if model.overlay.commandInput.value != "edit" {
		t.Fatalf("expected restored :edit input, got %q", model.overlay.commandInput.value)
	}
	if model.ui.statusMessage != `Error: value for column "name" is required` {
		t.Fatalf("expected validation error status, got %q", model.ui.statusMessage)
	}
}

func TestStartSaveForPendingNavigation_CleansPendingStateWhenSaveIsImmediateNoOp(t *testing.T) {
	// Arrange
	model := &Model{
		ui: runtimeUIState{
			pendingNavigation: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind: usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: usecase.RuntimeDatabaseTarget{
						Option:         runtimeDatabaseOptionFromSelectorOption(runtimeTestDatabaseOption()),
						TransitionKind: usecase.RuntimeDatabaseTransitionReloadCurrent,
					},
				},
			},
			pendingCommandInput: "edit",
		},
	}

	// Act
	_, cmd := model.startSaveForPendingNavigation()

	// Assert
	if cmd != nil {
		t.Fatal("expected immediate no-op save path not to return async command")
	}
	if model.ui.saveInFlight {
		t.Fatal("expected immediate no-op save path not to enter save-in-flight state")
	}
	if model.ui.pendingNavigation != nil {
		t.Fatalf("expected pending navigation cleanup after no-op save path, got %+v", model.ui.pendingNavigation)
	}
	if model.ui.pendingCommandInput != "" {
		t.Fatalf("expected pending command input cleanup after no-op save path, got %q", model.ui.pendingCommandInput)
	}
	if !model.overlay.commandInput.active {
		t.Fatal("expected :edit spotlight to reopen after no-op save path")
	}
	if model.overlay.commandInput.value != "edit" {
		t.Fatalf("expected restored :edit input, got %q", model.overlay.commandInput.value)
	}
	if model.ui.statusMessage != "No changes to save" {
		t.Fatalf("expected no-op save status, got %q", model.ui.statusMessage)
	}
}

func TestRequestSaveAndQuit_BlocksRuntimeInputUntilSaveResponse(t *testing.T) {
	// Arrange
	model := newDirtyRuntimeSaveModel(ViewRecords, FocusContent)
	model.saveChanges = &spySaveChangesUseCase{count: 1}
	model.read.tables = []dto.Table{{Name: "users"}, {Name: "orders"}}

	// Act
	_, saveCmd := model.requestSaveAndQuit()
	if saveCmd == nil {
		t.Fatal("expected save command to be returned")
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	msg := saveCmd()
	_, quitCmd := model.Update(msg)

	// Assert
	if model.overlay.commandInput.active {
		t.Fatal("expected command input to stay blocked while save is in flight")
	}
	if model.read.selectedTable != 0 {
		t.Fatalf("expected table selection to stay unchanged while save is in flight, got %d", model.read.selectedTable)
	}
	if model.ui.pendingSaveSuccessAction != usecase.RuntimeSaveSuccessActionNone {
		t.Fatal("expected pending save action to clear after save response")
	}
	if quitCmd == nil {
		t.Fatal("expected save-and-quit flow to quit after save response")
	}
	if _, ok := quitCmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after save response, got %T", quitCmd())
	}
}

func TestRequestSaveChanges_StartsSaveImmediatelyFromNonRecordsDirtyContexts(t *testing.T) {
	for _, tc := range []struct {
		name     string
		viewMode ViewMode
		focus    PanelFocus
	}{
		{name: "schema", viewMode: ViewSchema, focus: FocusContent},
		{name: "tables", viewMode: ViewSchema, focus: FocusTables},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := newDirtyRuntimeSaveModel(tc.viewMode, tc.focus)
			model.saveChanges = &spySaveChangesUseCase{count: 1}

			// Act
			_, cmd := model.requestSaveChanges()

			// Assert
			if cmd == nil {
				t.Fatalf("expected %s save request to start save immediately", tc.name)
			}
			assertRuntimeSaveStarted(t, model, "requestSaveChanges/"+tc.name)
		})
	}
}

func TestRequestSaveChanges_WithNoDirtyStateShowsNoChangesStatus(t *testing.T) {
	// Arrange
	model := &Model{
		saveChanges: &spySaveChangesUseCase{},
		read: runtimeReadState{
			viewMode: ViewSchema,
			focus:    FocusTables,
		},
	}

	// Act
	_, cmd := model.requestSaveChanges()

	// Assert
	if cmd != nil {
		t.Fatal("expected clean save request not to start save")
	}
	if model.overlay.confirmPopup.active {
		t.Fatal("expected clean save request not to open confirm popup")
	}
	if model.ui.saveInFlight {
		t.Fatal("expected clean save request to keep save out of flight")
	}
	if model.ui.statusMessage != "No changes to save" {
		t.Fatalf("expected clean save request to show no-op status, got %q", model.ui.statusMessage)
	}
}

func TestRequestSaveAndQuit_StartsSaveImmediatelyWithoutPopup(t *testing.T) {
	// Arrange
	model := newDirtyRuntimeSaveModel(ViewRecords, FocusContent)
	model.saveChanges = &spySaveChangesUseCase{count: 1}

	// Act
	_, cmd := model.requestSaveAndQuit()

	// Assert
	if cmd == nil {
		t.Fatal("expected save-and-quit request to start save flow")
	}
	assertRuntimeSaveStarted(t, model, "requestSaveAndQuit")
	if model.ui.pendingSaveSuccessAction != usecase.RuntimeSaveSuccessActionQuitRuntime {
		t.Fatalf("expected save-and-quit flow to set quit action, got %v", model.ui.pendingSaveSuccessAction)
	}
}
