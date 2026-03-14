package usecase_test

import (
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestStagingPolicy_InitialInsertValue_UsesColumnDefault(t *testing.T) {
	// Arrange
	policy := usecase.NewStagingPolicy()
	defaultValue := "guest"
	column := dto.SchemaColumn{
		Name:         "name",
		Type:         "TEXT",
		DefaultValue: &defaultValue,
		Nullable:     false,
	}

	// Act
	value := policy.InitialInsertValue(column)

	// Assert
	if value.IsNull {
		t.Fatal("expected non-null value")
	}
	if value.Text != "guest" {
		t.Fatalf("expected default text guest, got %q", value.Text)
	}
	if raw, ok := value.Raw.(string); !ok || raw != "guest" {
		t.Fatalf("expected raw guest string, got %#v", value.Raw)
	}
}

func TestStagingPolicy_InitialInsertValue_UsesNullForNullableColumnsWithoutDefault(t *testing.T) {
	// Arrange
	policy := usecase.NewStagingPolicy()
	column := dto.SchemaColumn{Name: "nickname", Type: "TEXT", Nullable: true}

	// Act
	value := policy.InitialInsertValue(column)

	// Assert
	if !value.IsNull {
		t.Fatal("expected null staged value")
	}
	if value.Text != "NULL" {
		t.Fatalf("expected NULL text, got %q", value.Text)
	}
}

func TestStagingPolicy_InitialInsertValue_UsesEmptyValueForRequiredColumnsWithoutDefault(t *testing.T) {
	// Arrange
	policy := usecase.NewStagingPolicy()
	column := dto.SchemaColumn{Name: "name", Type: "TEXT", Nullable: false}

	// Act
	value := policy.InitialInsertValue(column)

	// Assert
	if value.IsNull {
		t.Fatal("expected non-null staged value")
	}
	if value.Text != "" {
		t.Fatalf("expected empty text, got %q", value.Text)
	}
	if raw, ok := value.Raw.(string); !ok || raw != "" {
		t.Fatalf("expected raw empty string, got %#v", value.Raw)
	}
}

func TestStagingPolicy_DirtyEditCount_CountsOneRowForMultipleEditedColumns(t *testing.T) {
	// Arrange
	policy := usecase.NewStagingPolicy()
	pendingUpdates := map[string]dto.PendingRecordEdits{
		"id=1": {
			Changes: map[int]dto.StagedEdit{
				0: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
				1: {Value: dto.StagedValue{Text: "bob", Raw: "bob"}},
			},
		},
	}

	// Act
	count := policy.DirtyEditCount(nil, pendingUpdates, nil)

	// Assert
	if count != 1 {
		t.Fatalf("expected dirty count 1, got %d", count)
	}
}

func TestStagingPolicy_DirtyEditCount_CountsOneRowPerInsert(t *testing.T) {
	// Arrange
	policy := usecase.NewStagingPolicy()
	pendingInserts := []dto.PendingInsertRow{{}}

	// Act
	count := policy.DirtyEditCount(pendingInserts, nil, nil)

	// Assert
	if count != 1 {
		t.Fatalf("expected dirty count 1, got %d", count)
	}
}

func TestStagingPolicy_DirtyEditCount_CountsOneRowPerDelete(t *testing.T) {
	// Arrange
	policy := usecase.NewStagingPolicy()
	pendingDeletes := map[string]dto.PendingRecordDelete{
		"id=2": {},
	}

	// Act
	count := policy.DirtyEditCount(nil, nil, pendingDeletes)

	// Assert
	if count != 1 {
		t.Fatalf("expected dirty count 1, got %d", count)
	}
}

func TestStagingPolicy_DirtyEditCount_DeduplicatesEditedThenDeletedPersistedRow(t *testing.T) {
	// Arrange
	policy := usecase.NewStagingPolicy()
	pendingUpdates := map[string]dto.PendingRecordEdits{
		"id=1": {
			Changes: map[int]dto.StagedEdit{
				0: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
			},
		},
	}
	pendingDeletes := map[string]dto.PendingRecordDelete{
		"id=1": {},
	}

	// Act
	count := policy.DirtyEditCount(nil, pendingUpdates, pendingDeletes)

	// Assert
	if count != 1 {
		t.Fatalf("expected dirty count 1, got %d", count)
	}
}

func TestStagingPolicy_DirtyEditCount_ReturnsZeroWhenRemovedInsertNoLongerExists(t *testing.T) {
	// Arrange
	policy := usecase.NewStagingPolicy()

	// Act
	count := policy.DirtyEditCount(nil, nil, nil)

	// Assert
	if count != 0 {
		t.Fatalf("expected dirty count 0, got %d", count)
	}
}
