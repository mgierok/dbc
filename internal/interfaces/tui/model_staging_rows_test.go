package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func newDeleteSelectionTestModel(records []dto.RecordRow, schema dto.Schema) *Model {
	return &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			records:  records,
			schema:   schema,
		},
	}
}

func TestHandleKey_InsertCreatesPendingRowAtTop(t *testing.T) {
	// Arrange
	defaultName := "guest"
	model := &Model{
		read: runtimeReadState{
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
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Assert
	snapshot := model.currentStagingSnapshot()
	if len(snapshot.PendingInserts) != 1 {
		t.Fatalf("expected one pending insert, got %d", len(snapshot.PendingInserts))
	}
	if model.read.recordSelection != 0 {
		t.Fatalf("expected selection at top pending row, got %d", model.read.recordSelection)
	}
	if model.read.recordColumn != 1 {
		t.Fatalf("expected first editable column to skip auto field, got %d", model.read.recordColumn)
	}
	row := snapshot.PendingInserts[0]
	if got := displayValue(row.Values[1].Value); got != "guest" {
		t.Fatalf("expected default value guest, got %q", got)
	}
	if !row.Values[2].Value.IsNull {
		t.Fatalf("expected nullable column to default to NULL")
	}
	if got := displayValue(row.Values[3].Value); got != "" {
		t.Fatalf("expected required no-default column to be empty, got %q", got)
	}
}

func TestHandleKey_DeleteTogglesPersistedRow(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			records: []dto.RecordRow{{
				Values: []string{"not-an-integer", "alice"},
				RowKey: "id=1",
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
					},
				},
			}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.currentStagingSnapshot().PendingDeletes) != 1 {
		t.Fatalf("expected one pending delete, got %d", len(model.currentStagingSnapshot().PendingDeletes))
	}
	deleteChange, ok := model.currentStagingSnapshot().PendingDeletes["id=1"]
	if !ok {
		t.Fatal("expected pending delete to use resolved row key")
	}
	if len(deleteChange.Identity.Keys) != 1 || deleteChange.Identity.Keys[0].Column != "id" {
		t.Fatalf("expected resolved delete identity, got %+v", deleteChange.Identity)
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.currentStagingSnapshot().PendingDeletes) != 0 {
		t.Fatalf("expected pending delete to toggle off")
	}
}

func TestHandleKey_DeleteBlocksUnremovableRows(t *testing.T) {
	tests := []struct {
		name    string
		records []dto.RecordRow
		schema  dto.Schema
		wantErr string
	}{
		{
			name: "identity unavailable",
			records: []dto.RecordRow{
				{
					Values:              []string{"<truncated 262145 bytes>", "alice"},
					IdentityUnavailable: true,
				},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "TEXT", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
			wantErr: "Error: selected record identity exceeds safe browse limit",
		},
		{
			name: "missing primary key",
			records: []dto.RecordRow{
				{Values: []string{"alice"}},
			},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "name", Type: "TEXT"},
				},
			},
			wantErr: "Error: table has no primary key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			model := newDeleteSelectionTestModel(tt.records, tt.schema)

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

			// Assert
			if len(model.currentStagingSnapshot().PendingDeletes) != 0 {
				t.Fatalf("expected no pending deletes, got %d", len(model.currentStagingSnapshot().PendingDeletes))
			}
			if model.ui.statusMessage != tt.wantErr {
				t.Fatalf("expected status %q, got %q", tt.wantErr, model.ui.statusMessage)
			}
		})
	}
}

func TestHandleKey_DeleteRemovesPendingInsert(t *testing.T) {
	// Arrange
	model := withTestStaging(&Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
		},
	}, stagingState{
		pendingInserts: []pendingInsertRow{{values: map[int]stagedEdit{}}},
	})

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	// Assert
	if len(model.currentStagingSnapshot().PendingInserts) != 0 {
		t.Fatalf("expected pending insert to be removed")
	}
}
