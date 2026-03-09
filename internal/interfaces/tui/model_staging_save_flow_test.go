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
		tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable: 0,
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
			pendingUpdates: map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
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
		tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable: 0,
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
			pendingUpdates: map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
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
		tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable: 0,
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
			pendingUpdates: map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
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
		tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable: 0,
		staging: stagingState{
			pendingInserts: []pendingInsertRow{{}},
			pendingUpdates: map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}}},
			pendingDeletes: map[string]recordDelete{"id=2": {}},
		},
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
