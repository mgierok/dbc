package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestConfirmSaveChanges_SubmitsBuiltTableChanges(t *testing.T) {
	// Arrange
	saveChanges := &spySaveChangesUseCase{count: 1}
	model := &Model{
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
	model := &Model{
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
		staging: stagingState{
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
		},
	}

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
	model := &Model{
		ctx:         context.Background(),
		saveChanges: &spySaveChangesUseCase{count: 1},
		read: runtimeReadState{
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT", Nullable: false},
				},
			},
			tables: []dto.Table{{Name: "users"}},
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
		ui: runtimeUIState{statusMessage: "stale status"},
	}

	// Act
	_, cmd := model.confirmSaveChanges()

	// Assert
	if cmd == nil {
		t.Fatal("expected save command to be returned")
	}
	if !model.ui.saveInFlight {
		t.Fatal("expected save-in-flight flag to be enabled when save starts")
	}
	if model.ui.statusMessage != "Saving changes..." {
		t.Fatalf("expected saving status message, got %q", model.ui.statusMessage)
	}
}

func TestSetTableSelection_WithDirtyStateOpensInformationalSwitchTablePopup(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 0,
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
			pendingUpdates: map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
		ui: runtimeUIState{pendingTableIndex: -1},
	}

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

func TestSetTableSelection_WithDirtyStateDiscardOptionClearsStagingAndSwitches(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 0,
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
			pendingUpdates: map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
		ui: runtimeUIState{pendingTableIndex: -1},
	}

	// Act
	model.setTableSelection(1)
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.read.selectedTable != 1 {
		t.Fatalf("expected table switch after selecting discard")
	}
	if model.hasDirtyEdits() {
		t.Fatalf("expected staged state to be cleared after discard")
	}
}

func TestSetTableSelection_WithDirtyStateContinueEditingOptionPreservesStagingAndSelection(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 0,
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
			pendingUpdates: map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
		ui: runtimeUIState{pendingTableIndex: -1},
	}

	// Act
	model.setTableSelection(1)
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.read.selectedTable != 0 {
		t.Fatalf("expected table selection to stay unchanged after selecting continue editing")
	}
	if !model.hasDirtyEdits() {
		t.Fatalf("expected staged state to remain after selecting continue editing")
	}
	if model.ui.pendingTableIndex != -1 {
		t.Fatalf("expected pending table index reset after selecting continue editing, got %d", model.ui.pendingTableIndex)
	}
}

func TestSetTableSelection_WithDirtyStateNKeyIsIgnored(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 0,
		},
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
			pendingUpdates: map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
		ui: runtimeUIState{pendingTableIndex: -1},
	}

	// Act
	model.setTableSelection(1)
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Assert
	if !model.overlay.confirmPopup.active {
		t.Fatal("expected n key to leave table-switch popup open")
	}
	if model.read.selectedTable != 0 {
		t.Fatalf("expected table selection to stay unchanged after n key")
	}
	if !model.hasDirtyEdits() {
		t.Fatalf("expected staged state to remain after n key")
	}
	if model.ui.pendingTableIndex != 1 {
		t.Fatalf("expected pending table index to remain pending after n key, got %d", model.ui.pendingTableIndex)
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

func TestHandleConfirmPopupKey_DirtyDatabaseTransitionSaveStartsSaveFlow(t *testing.T) {
	// Arrange
	saveChanges := &spySaveChangesUseCase{}
	model := &Model{
		saveChanges: saveChanges,
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
						1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active: true,
				action: confirmDatabaseTransitionSave,
			},
		},
		ui: runtimeUIState{
			pendingDatabaseTransition: &runtimeDatabaseTransitionRequest{
				Target: runtimeDatabaseTransitionTarget{
					Option: DatabaseOption{
						Name:       "primary",
						ConnString: "/tmp/primary.sqlite",
						Source:     DatabaseOptionSourceConfig,
					},
					Kind: reloadCurrentDatabase,
				},
				Origin: runtimeDatabaseTransitionOriginEditCommand,
			},
		},
	}

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected save command to be returned")
	}
	if model.ui.pendingDatabaseTransition == nil {
		t.Fatal("expected pending database transition to stay set for save-and-transition flow")
	}
}

func TestHandleConfirmPopupKey_DirtyDatabaseTransitionSaveKeepsTransitionPendingWhenSaveStarts(t *testing.T) {
	// Arrange
	saveChanges := &spySaveChangesUseCase{}
	model := &Model{
		saveChanges: saveChanges,
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
						1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
		overlay: runtimeOverlayState{
			confirmPopup: confirmPopup{
				active: true,
				action: confirmDatabaseTransitionSave,
			},
		},
		ui: runtimeUIState{
			pendingDatabaseTransition: &runtimeDatabaseTransitionRequest{
				Target: runtimeDatabaseTransitionTarget{
					Option: DatabaseOption{
						Name:       "primary",
						ConnString: "/tmp/primary.sqlite",
						Source:     DatabaseOptionSourceConfig,
					},
					Kind: reloadCurrentDatabase,
				},
				Origin: runtimeDatabaseTransitionOriginEditCommand,
			},
		},
	}

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected save command to be returned")
	}
	if model.ui.pendingDatabaseTransition == nil {
		t.Fatal("expected pending database transition to stay set for save-and-transition flow")
	}
}

func TestRequestSaveChanges_StartsSaveImmediatelyFromSchemaWithDirtyStateStartedInRecords(t *testing.T) {
	// Arrange
	model := &Model{
		ctx:         context.Background(),
		saveChanges: &spySaveChangesUseCase{count: 1},
		read: runtimeReadState{
			focus:         FocusContent,
			viewMode:      ViewSchema,
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
	_, cmd := model.requestSaveChanges()

	// Assert
	if cmd == nil {
		t.Fatal("expected schema save request to start save immediately")
	}
	if model.overlay.confirmPopup.active {
		t.Fatal("expected schema save request not to open confirm popup")
	}
	if !model.ui.saveInFlight {
		t.Fatal("expected schema save request to enter save-in-flight state")
	}
	if model.ui.statusMessage != "Saving changes..." {
		t.Fatalf("expected saving status message, got %q", model.ui.statusMessage)
	}
}

func TestRequestSaveAndQuit_BlocksRuntimeInputUntilSaveResponse(t *testing.T) {
	// Arrange
	saveChanges := &spySaveChangesUseCase{count: 1}
	model := &Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		read: runtimeReadState{
			focus:         FocusContent,
			viewMode:      ViewRecords,
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
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
						1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
	}

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
	if model.ui.pendingQuitAfterSave {
		t.Fatal("expected pending quit flag to clear after save response")
	}
	if quitCmd == nil {
		t.Fatal("expected save-and-quit flow to quit after save response")
	}
	if _, ok := quitCmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after save response, got %T", quitCmd())
	}
}

func TestRequestSaveChanges_StartsSaveImmediatelyFromTablesWithDirtyStateStartedInRecords(t *testing.T) {
	// Arrange
	model := &Model{
		ctx:         context.Background(),
		saveChanges: &spySaveChangesUseCase{count: 1},
		read: runtimeReadState{
			focus:         FocusTables,
			viewMode:      ViewSchema,
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
	_, cmd := model.requestSaveChanges()

	// Assert
	if cmd == nil {
		t.Fatal("expected tables save request to start save immediately")
	}
	if model.overlay.confirmPopup.active {
		t.Fatal("expected tables save request not to open confirm popup")
	}
	if !model.ui.saveInFlight {
		t.Fatal("expected tables save request to enter save-in-flight state")
	}
	if model.ui.statusMessage != "Saving changes..." {
		t.Fatalf("expected saving status message, got %q", model.ui.statusMessage)
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
	saveChanges := &spySaveChangesUseCase{count: 1}
	model := &Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
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
	_, cmd := model.requestSaveAndQuit()

	// Assert
	if cmd == nil {
		t.Fatal("expected save-and-quit request to start save flow")
	}
	if model.overlay.confirmPopup.active {
		t.Fatal("expected save-and-quit request not to open confirm popup")
	}
	if !model.ui.pendingQuitAfterSave {
		t.Fatal("expected save-and-quit flow to set pending quit flag")
	}
	if !model.ui.saveInFlight {
		t.Fatal("expected save-and-quit flow to enter save-in-flight state")
	}
}
