package usecase_test

import (
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestStagedChangesTranslator_ParseStagedValue_ReturnsParsedInteger(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	column := dto.SchemaColumn{Name: "id", Type: "INTEGER", Nullable: false}

	// Act
	value, err := translator.ParseStagedValue(column, "42", false)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value.Text != "42" {
		t.Fatalf("expected text 42, got %q", value.Text)
	}
	if typed, ok := value.Raw.(int64); !ok || typed != 42 {
		t.Fatalf("expected raw int64(42), got %#v", value.Raw)
	}
}

func TestStagedChangesTranslator_ParseStagedValue_RejectsNullWhenNotNullable(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	column := dto.SchemaColumn{Name: "name", Type: "TEXT", Nullable: false}

	// Act
	_, err := translator.ParseStagedValue(column, "", true)

	// Assert
	if err == nil {
		t.Fatal("expected error for null in non-nullable column")
	}
}

func TestStagedChangesTranslator_BuildRecordIdentity_ReturnsPrimaryKeyIdentity(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: true},
		},
	}
	row := dto.RecordRow{Values: []string{"1", "alice"}}

	// Act
	key, identity, err := translator.BuildRecordIdentity(schema, row)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if key != "id=1" {
		t.Fatalf("expected key id=1, got %q", key)
	}
	if len(identity.Keys) != 1 {
		t.Fatalf("expected one identity key, got %d", len(identity.Keys))
	}
	if identity.Keys[0].Column != "id" {
		t.Fatalf("expected key column id, got %q", identity.Keys[0].Column)
	}
}

func TestStagedChangesTranslator_BuildTableChanges_IgnoresUpdatesForDeletedRows(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}
	pendingInserts := []dto.PendingInsertRow{
		{
			Values: map[int]dto.StagedEdit{
				0: {Value: dto.StagedValue{Text: "", Raw: ""}},
				1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
			},
			ExplicitAuto: map[int]bool{},
		},
	}
	pendingUpdates := map[string]dto.PendingRecordEdits{
		"id=1": {
			Identity: dto.RecordIdentity{
				Keys: []dto.ColumnValue{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
			},
			Changes: map[int]dto.StagedEdit{
				1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
			},
		},
	}
	pendingDeletes := map[string]dto.PendingRecordDelete{
		"id=1": {
			Identity: dto.RecordIdentity{
				Keys: []dto.ColumnValue{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
			},
		},
	}

	// Act
	changes, err := translator.BuildTableChanges(schema, pendingInserts, pendingUpdates, pendingDeletes)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(changes.Inserts) != 1 {
		t.Fatalf("expected one insert, got %d", len(changes.Inserts))
	}
	if len(changes.Updates) != 0 {
		t.Fatalf("expected updates for deleted rows to be ignored")
	}
	if len(changes.Deletes) != 1 {
		t.Fatalf("expected one delete, got %d", len(changes.Deletes))
	}
}
