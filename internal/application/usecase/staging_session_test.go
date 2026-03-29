package usecase_test

import (
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestStagingSession_AddInsert_SeedsDefaultsFromSchema(t *testing.T) {
	// Arrange
	defaultValue := "guest"
	session := usecase.NewStagingSession(nil, nil)
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
			{Name: "name", Type: "TEXT", Nullable: false, DefaultValue: &defaultValue},
			{Name: "note", Type: "TEXT", Nullable: true},
			{Name: "age", Type: "INTEGER", Nullable: false},
		},
	}

	// Act
	insertID, err := session.AddInsert(schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	snapshot := session.Snapshot()
	if len(snapshot.PendingInserts) != 1 {
		t.Fatalf("expected one pending insert, got %d", len(snapshot.PendingInserts))
	}
	row := snapshot.PendingInserts[0]
	if row.ID != insertID {
		t.Fatalf("expected stable insert ID %q, got %q", insertID, row.ID)
	}
	if got := displayValueForTest(row.Values[1].Value); got != "guest" {
		t.Fatalf("expected default value guest, got %q", got)
	}
	if !row.Values[2].Value.IsNull {
		t.Fatal("expected nullable column to seed NULL")
	}
	if got := displayValueForTest(row.Values[3].Value); got != "" {
		t.Fatalf("expected required no-default column to seed empty value, got %q", got)
	}
}

func TestStagingSession_StagePersistedEdit_RemovesEditWhenValueReturnsToOriginal(t *testing.T) {
	// Arrange
	session := usecase.NewStagingSession(nil, nil)
	identity := dto.RecordIdentity{
		Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
	}

	// Act
	err := session.StagePersistedEdit(
		"id=1",
		identity,
		1,
		"alice",
		dto.StagedValue{Text: "alice", Raw: "alice"},
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(session.Snapshot().PendingUpdates) != 0 {
		t.Fatal("expected edit to be removed when value returns to original")
	}
}

func TestStagingSession_DirtyEditCount_DeduplicatesUpdatedThenDeletedPersistedRow(t *testing.T) {
	// Arrange
	session := usecase.NewStagingSession(nil, nil)
	identity := dto.RecordIdentity{
		Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
	}
	if err := session.StagePersistedEdit(
		"id=1",
		identity,
		1,
		"alice",
		dto.StagedValue{Text: "bob", Raw: "bob"},
	); err != nil {
		t.Fatalf("expected edit to stage, got %v", err)
	}
	if err := session.SetDeleteMark("id=1", identity, true); err != nil {
		t.Fatalf("expected delete mark to stage, got %v", err)
	}

	// Act
	count := session.DirtyEditCount()

	// Assert
	if count != 1 {
		t.Fatalf("expected dirty count 1, got %d", count)
	}
}

func TestStagingSession_UndoRedo_PreservesInsertEditAndDeleteSemantics(t *testing.T) {
	// Arrange
	session := usecase.NewStagingSession(nil, nil)
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}
	identity := dto.RecordIdentity{
		Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
	}
	insertID, err := session.AddInsert(schema)
	if err != nil {
		t.Fatalf("expected insert add, got %v", err)
	}
	if err := session.StageInsertEdit(insertID, 1, dto.StagedValue{Text: "new", Raw: "new"}); err != nil {
		t.Fatalf("expected insert edit, got %v", err)
	}
	if err := session.StagePersistedEdit(
		"id=1",
		identity,
		1,
		"alice",
		dto.StagedValue{Text: "bob", Raw: "bob"},
	); err != nil {
		t.Fatalf("expected persisted edit, got %v", err)
	}
	if err := session.SetDeleteMark("id=1", identity, true); err != nil {
		t.Fatalf("expected delete mark, got %v", err)
	}

	// Act
	if err := session.Undo(); err != nil {
		t.Fatalf("expected first undo to succeed, got %v", err)
	}
	if err := session.Undo(); err != nil {
		t.Fatalf("expected second undo to succeed, got %v", err)
	}
	if err := session.Undo(); err != nil {
		t.Fatalf("expected third undo to succeed, got %v", err)
	}
	if err := session.Undo(); err != nil {
		t.Fatalf("expected fourth undo to succeed, got %v", err)
	}
	if err := session.Redo(); err != nil {
		t.Fatalf("expected first redo to succeed, got %v", err)
	}
	if err := session.Redo(); err != nil {
		t.Fatalf("expected second redo to succeed, got %v", err)
	}
	if err := session.Redo(); err != nil {
		t.Fatalf("expected third redo to succeed, got %v", err)
	}
	if err := session.Redo(); err != nil {
		t.Fatalf("expected fourth redo to succeed, got %v", err)
	}

	// Assert
	snapshot := session.Snapshot()
	if len(snapshot.PendingInserts) != 1 {
		t.Fatalf("expected one pending insert after redo, got %d", len(snapshot.PendingInserts))
	}
	if got := displayValueForTest(snapshot.PendingInserts[0].Values[1].Value); got != "new" {
		t.Fatalf("expected redone insert edit value new, got %q", got)
	}
	updated := snapshot.PendingUpdates["id=1"]
	if got := displayValueForTest(updated.Changes[1].Value); got != "bob" {
		t.Fatalf("expected redone persisted edit value bob, got %q", got)
	}
	if _, ok := snapshot.PendingDeletes["id=1"]; !ok {
		t.Fatal("expected delete mark to be restored by redo")
	}
}

func TestStagingSession_BuildTableChanges_MatchesSavePayloadSemantics(t *testing.T) {
	// Arrange
	session := usecase.NewStagingSession(nil, nil)
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true, Nullable: false},
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}
	identity := dto.RecordIdentity{
		Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}}},
	}
	insertID, err := session.AddInsert(schema)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := session.StageInsertEdit(insertID, 0, dto.StagedValue{Text: "7", Raw: int64(7)}); err != nil {
		t.Fatalf("expected auto value to stage, got %v", err)
	}
	if err := session.StageInsertEdit(insertID, 1, dto.StagedValue{Text: "new", Raw: "new"}); err != nil {
		t.Fatalf("expected insert name to stage, got %v", err)
	}
	if err := session.StagePersistedEdit(
		"id=1",
		identity,
		1,
		"alice",
		dto.StagedValue{Text: "bob", Raw: "bob"},
	); err != nil {
		t.Fatalf("expected persisted edit, got %v", err)
	}
	if err := session.SetDeleteMark("id=1", identity, true); err != nil {
		t.Fatalf("expected delete mark, got %v", err)
	}

	// Act
	changes, err := session.BuildTableChanges(schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(changes.Inserts) != 1 {
		t.Fatalf("expected one insert, got %d", len(changes.Inserts))
	}
	if len(changes.Inserts[0].Values) != 1 || changes.Inserts[0].Values[0].Column != "name" {
		t.Fatalf("expected insert payload to include only non-auto values, got %+v", changes.Inserts[0].Values)
	}
	if len(changes.Inserts[0].ExplicitAutoValues) != 1 || changes.Inserts[0].ExplicitAutoValues[0].Column != "id" {
		t.Fatalf("expected explicit auto payload for id, got %+v", changes.Inserts[0].ExplicitAutoValues)
	}
	if len(changes.Updates) != 0 {
		t.Fatalf("expected updates for deleted rows to be ignored, got %+v", changes.Updates)
	}
	if len(changes.Deletes) != 1 {
		t.Fatalf("expected one delete, got %d", len(changes.Deletes))
	}
}

func TestStagingSession_Reset_ClearsStateHistoryAndFuture(t *testing.T) {
	// Arrange
	session := usecase.NewStagingSession(nil, nil)
	schema := dto.Schema{
		Columns: []dto.SchemaColumn{
			{Name: "name", Type: "TEXT", Nullable: false},
		},
	}
	insertID, err := session.AddInsert(schema)
	if err != nil {
		t.Fatalf("expected insert add, got %v", err)
	}
	if err := session.StageInsertEdit(insertID, 0, dto.StagedValue{Text: "new", Raw: "new"}); err != nil {
		t.Fatalf("expected insert edit, got %v", err)
	}
	if err := session.Undo(); err != nil {
		t.Fatalf("expected undo to succeed, got %v", err)
	}

	// Act
	session.Reset()

	// Assert
	snapshot := session.Snapshot()
	if len(snapshot.PendingInserts) != 0 {
		t.Fatalf("expected no inserts after reset, got %d", len(snapshot.PendingInserts))
	}
	if len(snapshot.PendingUpdates) != 0 {
		t.Fatalf("expected no updates after reset, got %d", len(snapshot.PendingUpdates))
	}
	if len(snapshot.PendingDeletes) != 0 {
		t.Fatalf("expected no deletes after reset, got %d", len(snapshot.PendingDeletes))
	}
	if session.DirtyEditCount() != 0 {
		t.Fatalf("expected zero dirty count after reset, got %d", session.DirtyEditCount())
	}
	if err := session.Redo(); err != nil {
		t.Fatalf("expected redo after reset to be a no-op, got %v", err)
	}
	if len(session.Snapshot().PendingInserts) != 0 {
		t.Fatal("expected reset to clear redo history")
	}
}

func displayValueForTest(value dto.StagedValue) string {
	if value.IsNull {
		return "NULL"
	}
	if value.Text != "" {
		return value.Text
	}
	if raw, ok := value.Raw.(string); ok {
		return raw
	}
	return ""
}
