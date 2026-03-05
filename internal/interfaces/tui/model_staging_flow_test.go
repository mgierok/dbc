package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_InsertCreatesPendingRowAtTop(t *testing.T) {
	// Arrange
	defaultName := "guest"
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: false, DefaultValue: &defaultName},
				{Name: "note", Type: "TEXT", Nullable: true},
				{Name: "age", Type: "INTEGER", Nullable: false},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	if len(model.pendingInserts) != 1 {
		t.Fatalf("expected one pending insert, got %d", len(model.pendingInserts))
	}
	if model.recordSelection != 0 {
		t.Fatalf("expected selection at top pending row, got %d", model.recordSelection)
	}
	if model.recordColumn != 1 {
		t.Fatalf("expected first editable column to skip auto field, got %d", model.recordColumn)
	}
	row := model.pendingInserts[0]
	if got := displayValue(row.values[1].Value); got != "guest" {
		t.Fatalf("expected default value guest, got %q", got)
	}
	if !row.values[2].Value.IsNull {
		t.Fatalf("expected nullable column to default to NULL")
	}
	if got := displayValue(row.values[3].Value); got != "" {
		t.Fatalf("expected required no-default column to be empty, got %q", got)
	}
}

func TestHandleKey_DeleteTogglesPersistedRow(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.pendingDeletes) != 1 {
		t.Fatalf("expected one pending delete, got %d", len(model.pendingDeletes))
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.pendingDeletes) != 0 {
		t.Fatalf("expected pending delete to toggle off")
	}
}

func TestHandleKey_DeleteRemovesPendingInsert(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:       ViewRecords,
		focus:          FocusContent,
		pendingInserts: []pendingInsertRow{{values: map[int]stagedEdit{}}},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.pendingInserts) != 0 {
		t.Fatalf("expected pending insert to be removed")
	}
}

func TestBuildTableChanges_IgnoresUpdatesForDeletedRows(t *testing.T) {
	// Arrange
	model := &Model{
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
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
				},
				changes: map[int]stagedEdit{
					1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
				},
			},
		},
		pendingDeletes: map[string]recordDelete{
			"id=1": {
				identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
				},
			},
		},
	}

	// Act
	changes, err := model.buildTableChanges()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(changes.Inserts) != 1 {
		t.Fatalf("expected one insert, got %d", len(changes.Inserts))
	}
	if len(changes.Updates) != 0 {
		t.Fatalf("expected updates for deleted row to be ignored")
	}
	if len(changes.Deletes) != 1 {
		t.Fatalf("expected one delete, got %d", len(changes.Deletes))
	}
}

func TestDirtyEditCount_IncludesInsertsDeletesAndUpdates(t *testing.T) {
	// Arrange
	model := &Model{
		pendingInserts: []pendingInsertRow{{}},
		pendingDeletes: map[string]recordDelete{"id=1": {}},
		pendingUpdates: map[string]recordEdits{
			"id=2": {
				changes: map[int]stagedEdit{
					0: {Value: dto.StagedValue{Text: "x", Raw: "x"}},
					1: {Value: dto.StagedValue{Text: "y", Raw: "y"}},
				},
			},
		},
	}

	// Act
	dirty := model.dirtyEditCount()

	// Assert
	if dirty != 4 {
		t.Fatalf("expected dirty count 4, got %d", dirty)
	}
}

func TestConfirmSaveChanges_SubmitsBuiltTableChanges(t *testing.T) {
	// Arrange
	saveChanges := &spySaveChangesUseCase{}
	model := &Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
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
		tables: []dto.Table{{Name: "users"}},
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
	if len(saveChanges.lastChanges.Inserts) != 1 {
		t.Fatalf("expected one insert payload, got %d", len(saveChanges.lastChanges.Inserts))
	}
}

func TestSetTableSelection_WithDirtyStateOpensInformationalSwitchTablePopup(t *testing.T) {
	// Arrange
	model := &Model{
		tables:            []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable:     0,
		pendingInserts:    []pendingInsertRow{{}},
		pendingUpdates:    map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
		pendingDeletes:    map[string]recordDelete{"id=2": {}},
		pendingTableIndex: -1,
	}

	// Act
	model.setTableSelection(1)

	// Assert
	if !model.confirmPopup.active {
		t.Fatalf("expected discard confirmation popup")
	}
	if !model.confirmPopup.modal {
		t.Fatalf("expected table switch popup to be modal")
	}
	if model.confirmPopup.title != "Switch Table" {
		t.Fatalf("expected switch table title, got %q", model.confirmPopup.title)
	}
	if !strings.Contains(model.confirmPopup.message, "Switching tables will cause loss of unsaved data (3 changes).") {
		t.Fatalf("expected message with unsaved changes count, got %q", model.confirmPopup.message)
	}
	if !strings.Contains(model.confirmPopup.message, "Are you sure you want to discard unsaved data?") {
		t.Fatalf("expected discard confirmation question, got %q", model.confirmPopup.message)
	}
	if len(model.confirmPopup.options) != 2 {
		t.Fatalf("expected two explicit options, got %d", len(model.confirmPopup.options))
	}
	if model.confirmPopup.options[0].label != "(y) Yes, discard changes and switch table" {
		t.Fatalf("expected explicit yes option, got %q", model.confirmPopup.options[0].label)
	}
	if model.confirmPopup.options[1].label != "(n) No, continue editing" {
		t.Fatalf("expected explicit no option, got %q", model.confirmPopup.options[1].label)
	}
	if model.selectedTable != 0 {
		t.Fatalf("expected table selection not to change before confirmation")
	}
}

func TestSetTableSelection_WithDirtyStateYesOptionClearsStagingAndSwitches(t *testing.T) {
	// Arrange
	model := &Model{
		tables:            []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable:     0,
		pendingInserts:    []pendingInsertRow{{}},
		pendingUpdates:    map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
		pendingDeletes:    map[string]recordDelete{"id=2": {}},
		pendingTableIndex: -1,
	}

	// Act
	model.setTableSelection(1)
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.selectedTable != 1 {
		t.Fatalf("expected table switch after selecting yes")
	}
	if model.hasDirtyEdits() {
		t.Fatalf("expected staged state to be cleared after discard")
	}
}

func TestSetTableSelection_WithDirtyStateNoOptionPreservesStagingAndSelection(t *testing.T) {
	// Arrange
	model := &Model{
		tables:            []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable:     0,
		pendingInserts:    []pendingInsertRow{{}},
		pendingUpdates:    map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
		pendingDeletes:    map[string]recordDelete{"id=2": {}},
		pendingTableIndex: -1,
	}

	// Act
	model.setTableSelection(1)
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.selectedTable != 0 {
		t.Fatalf("expected table selection to stay unchanged after selecting no")
	}
	if !model.hasDirtyEdits() {
		t.Fatalf("expected staged state to remain after selecting no")
	}
	if model.pendingTableIndex != -1 {
		t.Fatalf("expected pending table index reset after selecting no, got %d", model.pendingTableIndex)
	}
}

func TestSetTableSelection_WithDirtyStateNoKeyPreservesStagingAndSelection(t *testing.T) {
	// Arrange
	model := &Model{
		tables:            []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable:     0,
		pendingInserts:    []pendingInsertRow{{}},
		pendingUpdates:    map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
		pendingDeletes:    map[string]recordDelete{"id=2": {}},
		pendingTableIndex: -1,
	}

	// Act
	model.setTableSelection(1)
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Assert
	if model.selectedTable != 0 {
		t.Fatalf("expected table selection to stay unchanged after no key")
	}
	if !model.hasDirtyEdits() {
		t.Fatalf("expected staged state to remain after no key")
	}
	if model.pendingTableIndex != -1 {
		t.Fatalf("expected pending table index reset after no key, got %d", model.pendingTableIndex)
	}
}

func TestSetTableSelection_ClearsSortOnTableSwitch(t *testing.T) {
	// Arrange
	model := &Model{
		tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable: 0,
		currentSort: &dto.Sort{
			Column:    "name",
			Direction: dto.SortDirectionAsc,
		},
	}

	// Act
	model.setTableSelection(1)

	// Assert
	if model.selectedTable != 1 {
		t.Fatalf("expected selected table to switch, got %d", model.selectedTable)
	}
	if model.currentSort != nil {
		t.Fatalf("expected sort to reset on table switch, got %+v", model.currentSort)
	}
}

func TestHandleKey_UndoRedo_RevertsAndReappliesInsertAdd(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	if len(model.pendingInserts) != 1 {
		t.Fatalf("expected one pending insert after add, got %d", len(model.pendingInserts))
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.pendingInserts) != 0 {
		t.Fatalf("expected insert to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.pendingInserts) != 1 {
		t.Fatalf("expected insert to be redone")
	}
}

func TestHandleKey_NewActionClearsRedoStack(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	if len(model.future) != 0 {
		t.Fatalf("expected redo stack to be cleared by new staged action")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.pendingInserts) != 1 {
		t.Fatalf("expected redo to have no effect after redo stack clear, got %d inserts", len(model.pendingInserts))
	}
}

func TestHandleKey_UndoRedo_RevertsAndReappliesPersistedCellEdit(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "name", Type: "TEXT"},
			},
		},
	}
	if err := model.stageEdit(0, 1, dto.StagedValue{Text: "bob", Raw: "bob"}); err != nil {
		t.Fatalf("expected staged edit, got error %v", err)
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.pendingUpdates) != 0 {
		t.Fatalf("expected persisted edit to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	edits := model.pendingUpdates["id=1"]
	change, ok := edits.changes[1]
	if !ok {
		t.Fatalf("expected persisted edit to be restored by redo")
	}
	if got := displayValue(change.Value); got != "bob" {
		t.Fatalf("expected restored value bob, got %q", got)
	}
}

func TestHandleKey_UndoRedo_RevertsAndReappliesDeleteToggle(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1", "alice"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.pendingDeletes) != 1 {
		t.Fatalf("expected delete toggle to be staged")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	// Assert
	if len(model.pendingDeletes) != 0 {
		t.Fatalf("expected delete toggle to be undone")
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Assert
	if len(model.pendingDeletes) != 1 {
		t.Fatalf("expected delete toggle to be redone")
	}
}

func TestHandleConfirmPopupKey_DirtyConfigSaveStartsSaveFlow(t *testing.T) {
	// Arrange
	saveChanges := &spySaveChangesUseCase{}
	model := &Model{
		viewMode:      ViewRecords,
		focus:         FocusContent,
		saveChanges:   saveChanges,
		tables:        []dto.Table{{Name: "users"}},
		selectedTable: 0,
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
		pendingInserts: []pendingInsertRow{
			{
				values: map[int]stagedEdit{
					0: {Value: dto.StagedValue{Text: "1", Raw: "1"}},
					1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
		confirmPopup: confirmPopup{
			active: true,
			action: confirmConfigSaveAndOpen,
		},
	}

	// Act
	_, cmd := model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd == nil {
		t.Fatal("expected save command to be returned")
	}
	if !model.pendingConfigOpen {
		t.Fatal("expected pending config open flag to be set for save-and-open flow")
	}
}
