package usecase_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/domain/model"
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
				Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
			},
			Changes: map[int]dto.StagedEdit{
				1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
			},
		},
	}
	pendingDeletes := map[string]dto.PendingRecordDelete{
		"id=1": {
			Identity: dto.RecordIdentity{
				Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
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

func TestStagedChangesTranslator_BuildRecordIdentity_ReturnsErrorWhenNoPrimaryKey(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "name", Type: "TEXT", Nullable: true},
		},
	}
	row := dto.RecordRow{Values: []string{"alice"}}

	// Act
	_, _, err := translator.BuildRecordIdentity(schema, row)

	// Assert
	if err == nil {
		t.Fatal("expected error for schema without primary key")
	}
	if !strings.Contains(err.Error(), "table has no primary key") {
		t.Fatalf("expected missing primary key error, got %v", err)
	}
}

func TestStagedChangesTranslator_BuildRecordIdentity_ReturnsErrorWhenPrimaryKeyIndexOutOfRange(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
		},
	}
	row := dto.RecordRow{Values: []string{}}

	// Act
	_, _, err := translator.BuildRecordIdentity(schema, row)

	// Assert
	if err == nil {
		t.Fatal("expected out-of-range error")
	}
	if !strings.Contains(err.Error(), "primary key index out of range") {
		t.Fatalf("expected primary-key-index error, got %v", err)
	}
}

func TestStagedChangesTranslator_BuildRecordIdentity_ReturnsErrorForInvalidPrimaryKeyParse(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
		},
	}
	row := dto.RecordRow{Values: []string{"abc"}}

	// Act
	_, _, err := translator.BuildRecordIdentity(schema, row)

	// Assert
	if err == nil {
		t.Fatal("expected parse error for invalid integer primary key")
	}
}

func TestStagedChangesTranslator_BuildTableChanges_ReturnsErrorWhenUpdateIdentityMissing(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}

	// Act
	_, err := translator.BuildTableChanges(
		schema,
		nil,
		map[string]dto.PendingRecordEdits{
			"missing": {
				Identity: dto.RecordIdentity{},
				Changes: map[int]dto.StagedEdit{
					1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
				},
			},
		},
		nil,
	)

	// Assert
	if err == nil {
		t.Fatal("expected missing identity error")
	}
	if !strings.Contains(err.Error(), "record identity missing") {
		t.Fatalf("expected missing identity error, got %v", err)
	}
}

func TestStagedChangesTranslator_BuildTableChanges_ReturnsErrorWhenUpdateColumnIndexOutOfRange(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}

	// Act
	_, err := translator.BuildTableChanges(
		schema,
		nil,
		map[string]dto.PendingRecordEdits{
			"id=1": {
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
					},
				},
				Changes: map[int]dto.StagedEdit{
					3: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
				},
			},
		},
		nil,
	)

	// Assert
	if err == nil {
		t.Fatal("expected out-of-range update column error")
	}
	if !strings.Contains(err.Error(), "column index out of range") {
		t.Fatalf("expected update column range error, got %v", err)
	}
}

func TestStagedChangesTranslator_BuildTableChanges_ReturnsErrorWhenRequiredInsertValueMissing(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}

	// Act
	_, err := translator.BuildTableChanges(
		schema,
		[]dto.PendingInsertRow{
			{
				Values: map[int]dto.StagedEdit{
					0: {Value: dto.StagedValue{Text: "", Raw: ""}},
				},
				ExplicitAuto: map[int]bool{},
			},
		},
		nil,
		nil,
	)

	// Assert
	if err == nil {
		t.Fatal("expected required value validation error")
	}
	if !strings.Contains(err.Error(), "value for column \"name\" is required") {
		t.Fatalf("expected required value error, got %v", err)
	}
}

func TestStagedChangesTranslator_BuildTableChanges_ReturnsMissingInsertValuesError(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true, Nullable: false},
		},
	}

	// Act
	_, err := translator.BuildTableChanges(
		schema,
		[]dto.PendingInsertRow{
			{
				Values:       map[int]dto.StagedEdit{},
				ExplicitAuto: map[int]bool{},
			},
		},
		nil,
		nil,
	)

	// Assert
	if !errors.Is(err, model.ErrMissingInsertValues) {
		t.Fatalf("expected error %v, got %v", model.ErrMissingInsertValues, err)
	}
}

func TestStagedChangesTranslator_BuildTableChanges_TracksExplicitAutoIncrementValues(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}

	// Act
	changes, err := translator.BuildTableChanges(
		schema,
		[]dto.PendingInsertRow{
			{
				Values: map[int]dto.StagedEdit{
					0: {Value: dto.StagedValue{Text: "42", Raw: int64(42)}},
					1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
				},
				ExplicitAuto: map[int]bool{0: true},
			},
		},
		nil,
		nil,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(changes.Inserts) != 1 {
		t.Fatalf("expected one insert, got %d", len(changes.Inserts))
	}
	insert := changes.Inserts[0]
	if len(insert.Values) != 1 || insert.Values[0].Column != "name" {
		t.Fatalf("expected only non-auto value in insert.Values, got %#v", insert.Values)
	}
	if len(insert.ExplicitAutoValues) != 1 || insert.ExplicitAutoValues[0].Column != "id" {
		t.Fatalf("expected explicit auto value for id, got %#v", insert.ExplicitAutoValues)
	}
}

func TestStagedChangesTranslator_BuildTableChanges_ExcludesImplicitAutoIncrementValues(t *testing.T) {
	// Arrange
	translator := usecase.NewStagedChangesTranslator()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}

	// Act
	changes, err := translator.BuildTableChanges(
		schema,
		[]dto.PendingInsertRow{
			{
				Values: map[int]dto.StagedEdit{
					0: {Value: dto.StagedValue{Text: "7", Raw: int64(7)}},
					1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
				},
				ExplicitAuto: map[int]bool{},
			},
		},
		nil,
		nil,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(changes.Inserts) != 1 {
		t.Fatalf("expected one insert, got %d", len(changes.Inserts))
	}
	insert := changes.Inserts[0]
	if len(insert.Values) != 1 || insert.Values[0].Column != "name" {
		t.Fatalf("expected only non-auto value in insert.Values, got %#v", insert.Values)
	}
	if len(insert.ExplicitAutoValues) != 0 {
		t.Fatalf("expected no explicit auto values, got %#v", insert.ExplicitAutoValues)
	}
}
