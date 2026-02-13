package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	domainmodel "github.com/mgierok/dbc/internal/domain/model"
)

func TestHandleKey_EnterSwitchesToRecords(t *testing.T) {
	// Arrange
	model := &Model{viewMode: ViewSchema}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.viewMode != ViewRecords {
		t.Fatalf("expected view mode to switch to records, got %v", model.viewMode)
	}
}

func TestHandleKey_EnterInRecordsEnablesFieldFocus(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		records:  []dto.RecordRow{{Values: []string{"1"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.recordFieldFocus {
		t.Fatalf("expected record field focus to be enabled")
	}
}

func TestHandleKey_EnterInFieldFocusOpensEditPopup(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
		records:          []dto.RecordRow{{Values: []string{"1"}}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if !model.editPopup.active {
		t.Fatalf("expected edit popup to be active")
	}
}

func TestHandleKey_EscClearsFieldFocus(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if model.recordFieldFocus {
		t.Fatalf("expected record field focus to be disabled")
	}
}

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
					0: {Value: domainmodel.Value{Text: "", Raw: ""}},
					1: {Value: domainmodel.Value{Text: "new", Raw: "new"}},
				},
				explicitAuto: map[int]bool{},
			},
		},
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				identity: domainmodel.RecordIdentity{
					Keys: []domainmodel.ColumnValue{{Column: "id", Value: domainmodel.Value{Text: "1", Raw: int64(1)}}},
				},
				changes: map[int]stagedEdit{
					1: {Value: domainmodel.Value{Text: "bob", Raw: "bob"}},
				},
			},
		},
		pendingDeletes: map[string]recordDelete{
			"id=1": {
				identity: domainmodel.RecordIdentity{
					Keys: []domainmodel.ColumnValue{{Column: "id", Value: domainmodel.Value{Text: "1", Raw: int64(1)}}},
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
					0: {Value: domainmodel.Value{Text: "x", Raw: "x"}},
					1: {Value: domainmodel.Value{Text: "y", Raw: "y"}},
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
	engine := &tuiSpyEngine{}
	saveChanges := usecase.NewSaveTableChanges(engine)
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
					0: {Value: domainmodel.Value{Text: "", Raw: ""}},
					1: {Value: domainmodel.Value{Text: "new", Raw: "new"}},
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
	if len(engine.lastChanges.Inserts) != 1 {
		t.Fatalf("expected one insert payload, got %d", len(engine.lastChanges.Inserts))
	}
}

func TestSetTableSelection_WithDirtyStatePromptsAndDiscardClearsStaging(t *testing.T) {
	// Arrange
	model := &Model{
		tables:            []dto.Table{{Name: "users"}, {Name: "orders"}},
		selectedTable:     0,
		pendingInserts:    []pendingInsertRow{{}},
		pendingUpdates:    map[string]recordEdits{"id=1": {changes: map[int]stagedEdit{0: {Value: domainmodel.Value{Text: "x", Raw: "x"}}}}},
		pendingDeletes:    map[string]recordDelete{"id=2": {}},
		pendingTableIndex: -1,
	}

	// Act
	model.setTableSelection(1)

	// Assert
	if !model.confirmPopup.active || model.confirmPopup.action != confirmDiscardTable {
		t.Fatalf("expected discard confirmation popup")
	}
	if model.selectedTable != 0 {
		t.Fatalf("expected table selection not to change before confirmation")
	}

	// Act
	model.handleConfirmPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if model.selectedTable != 1 {
		t.Fatalf("expected table switch after confirmation")
	}
	if model.hasDirtyEdits() {
		t.Fatalf("expected staged state to be cleared after discard")
	}
}

type tuiSpyEngine struct {
	lastChanges domainmodel.TableChanges
}

func (s *tuiSpyEngine) ListTables(ctx context.Context) ([]domainmodel.Table, error) {
	return nil, nil
}

func (s *tuiSpyEngine) GetSchema(ctx context.Context, tableName string) (domainmodel.Schema, error) {
	return domainmodel.Schema{}, nil
}

func (s *tuiSpyEngine) ListRecords(ctx context.Context, tableName string, offset, limit int, filter *domainmodel.Filter) (domainmodel.RecordPage, error) {
	return domainmodel.RecordPage{}, nil
}

func (s *tuiSpyEngine) ListOperators(ctx context.Context, columnType string) ([]domainmodel.Operator, error) {
	return nil, nil
}

func (s *tuiSpyEngine) ApplyRecordChanges(ctx context.Context, tableName string, changes domainmodel.TableChanges) error {
	s.lastChanges = changes
	return nil
}
