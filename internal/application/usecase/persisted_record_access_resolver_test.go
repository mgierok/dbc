package usecase_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestPersistedRecordAccessResolver_ResolveForDelete_ReturnsPrecomputedRef(t *testing.T) {
	// Arrange
	resolver := usecase.NewPersistedRecordAccessResolver()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
		},
	}
	row := dto.RecordRow{
		Values: []string{"not-an-integer"},
		RowKey: "id=7",
		Identity: dto.RecordIdentity{
			Keys: []dto.RecordIdentityKey{
				{Column: "id", Value: dto.StagedValue{Text: "7", Raw: int64(7)}},
			},
		},
	}

	// Act
	recordRef, err := resolver.ResolveForDelete(schema, row)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if recordRef.RowKey != "id=7" {
		t.Fatalf("expected precomputed key, got %q", recordRef.RowKey)
	}
	if len(recordRef.Identity.Keys) != 1 || recordRef.Identity.Keys[0].Column != "id" {
		t.Fatalf("expected precomputed identity, got %+v", recordRef.Identity)
	}
}

func TestPersistedRecordAccessResolver_ResolveForDelete_FallsBackToPrimaryKeyWhenPrecomputedIdentityMissing(t *testing.T) {
	// Arrange
	resolver := usecase.NewPersistedRecordAccessResolver()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: true},
		},
	}
	row := dto.RecordRow{Values: []string{"1", "alice"}}

	// Act
	recordRef, err := resolver.ResolveForDelete(schema, row)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if recordRef.RowKey != "id=1" {
		t.Fatalf("expected key id=1, got %q", recordRef.RowKey)
	}
	if len(recordRef.Identity.Keys) != 1 {
		t.Fatalf("expected one identity key, got %d", len(recordRef.Identity.Keys))
	}
	if recordRef.Identity.Keys[0].Column != "id" {
		t.Fatalf("expected key column id, got %q", recordRef.Identity.Keys[0].Column)
	}
}

func TestPersistedRecordAccessResolver_ResolveForDelete_ReturnsErrorWhenIdentityUnavailable(t *testing.T) {
	// Arrange
	resolver := usecase.NewPersistedRecordAccessResolver()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
		},
	}
	row := dto.RecordRow{
		Values:              []string{"1"},
		IdentityUnavailable: true,
	}

	// Act
	_, err := resolver.ResolveForDelete(schema, row)

	// Assert
	if !errors.Is(err, usecase.ErrSelectedRecordIdentityExceedsSafeBrowseLimit) {
		t.Fatalf("expected oversized-identity error, got %v", err)
	}
}

func TestPersistedRecordAccessResolver_ResolveForDelete_ReturnsErrorWhenTableHasNoPrimaryKey(t *testing.T) {
	// Arrange
	resolver := usecase.NewPersistedRecordAccessResolver()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "name", Type: "TEXT", Nullable: true},
		},
	}
	row := dto.RecordRow{Values: []string{"alice"}}

	// Act
	_, err := resolver.ResolveForDelete(schema, row)

	// Assert
	if err == nil {
		t.Fatal("expected error for schema without primary key")
	}
	if !strings.Contains(err.Error(), "table has no primary key") {
		t.Fatalf("expected missing primary key error, got %v", err)
	}
}

func TestPersistedRecordAccessResolver_ResolveForDelete_ReturnsErrorWhenContractIsInconsistent(t *testing.T) {
	tests := []struct {
		name string
		row  dto.RecordRow
	}{
		{
			name: "row key without identity",
			row: dto.RecordRow{
				Values:   []string{"1"},
				RowKey:   "id=1",
				Identity: dto.RecordIdentity{},
			},
		},
		{
			name: "identity without row key",
			row: dto.RecordRow{
				Values: []string{"1"},
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resolver := usecase.NewPersistedRecordAccessResolver()
			schema := dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
				},
			}

			// Act
			_, err := resolver.ResolveForDelete(schema, tt.row)

			// Assert
			if err == nil {
				t.Fatal("expected inconsistent contract error")
			}
			if !strings.Contains(err.Error(), "record identity missing") {
				t.Fatalf("expected missing identity error, got %v", err)
			}
		})
	}
}

func TestPersistedRecordAccessResolver_ResolveForEdit_ReturnsPersistedRecordRefWhenCellIsEditableFromBrowse(t *testing.T) {
	// Arrange
	resolver := usecase.NewPersistedRecordAccessResolver()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: true},
		},
	}
	row := dto.RecordRow{
		Values:             []string{"1", "alice"},
		EditableFromBrowse: []bool{true, true},
	}

	// Act
	recordRef, err := resolver.ResolveForEdit(schema, row, 1)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if recordRef.RowKey != "id=1" {
		t.Fatalf("expected key id=1, got %q", recordRef.RowKey)
	}
}

func TestPersistedRecordAccessResolver_ResolveForEdit_ReturnsErrorWhenCellHasNoSafeEditableSource(t *testing.T) {
	// Arrange
	resolver := usecase.NewPersistedRecordAccessResolver()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
			{Name: "payload", Type: "BLOB", Nullable: true},
		},
	}
	row := dto.RecordRow{
		Values:             []string{"1", "<blob 2 bytes>"},
		EditableFromBrowse: []bool{true, false},
	}

	// Act
	_, err := resolver.ResolveForEdit(schema, row, 1)

	// Assert
	if !errors.Is(err, usecase.ErrSelectedCellHasNoSafeEditableSource) {
		t.Fatalf("expected browse-edit safety error, got %v", err)
	}
}

func TestPersistedRecordAccessResolver_ResolveForEdit_ReturnsErrorWhenColumnIndexOutOfRange(t *testing.T) {
	// Arrange
	resolver := usecase.NewPersistedRecordAccessResolver()
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
		},
	}
	row := dto.RecordRow{Values: []string{"1"}}

	// Act
	_, err := resolver.ResolveForEdit(schema, row, 1)

	// Assert
	if err == nil {
		t.Fatal("expected out-of-range error")
	}
	if !strings.Contains(err.Error(), "column index out of range") {
		t.Fatalf("expected column range error, got %v", err)
	}
}
