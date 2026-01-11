package engine

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mgierok/dbc/internal/domain/model"
)

func TestSQLiteEngine_ApplyRecordUpdates_UpdatesByPrimaryKey(t *testing.T) {
	// Arrange
	db := setupSQLiteUpdateDB(t)
	engine := NewSQLiteEngine(db)
	updates := []model.RecordUpdate{
		{
			Identity: model.RecordIdentity{
				Keys: []model.ColumnValue{
					{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}},
				},
			},
			Changes: []model.ColumnValue{
				{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}},
				{Column: "active", Value: model.Value{Text: "false", Raw: int64(0)}},
				{Column: "score", Value: model.Value{Text: "2.5", Raw: 2.5}},
				{Column: "note", Value: model.Value{IsNull: true}},
				{Column: "data", Value: model.Value{Text: "0xFF", Raw: []byte{0xff}}},
			},
		},
	}

	// Act
	err := engine.ApplyRecordUpdates(context.Background(), "users", updates)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	var (
		name   string
		active int64
		score  float64
		note   sql.NullString
		data   []byte
	)
	row := db.QueryRow("SELECT name, active, score, note, data FROM users WHERE id = 1")
	if err := row.Scan(&name, &active, &score, &note, &data); err != nil {
		t.Fatalf("failed to scan row: %v", err)
	}
	if name != "bob" {
		t.Fatalf("expected name %q, got %q", "bob", name)
	}
	if active != 0 {
		t.Fatalf("expected active 0, got %d", active)
	}
	if score != 2.5 {
		t.Fatalf("expected score 2.5, got %v", score)
	}
	if note.Valid {
		t.Fatalf("expected note to be null, got %q", note.String)
	}
	if len(data) != 1 || data[0] != 0xff {
		t.Fatalf("expected data %v, got %v", []byte{0xff}, data)
	}
}

func TestSQLiteEngine_ApplyRecordUpdates_RollsBackOnError(t *testing.T) {
	// Arrange
	db := setupSQLiteUpdateDB(t)
	engine := NewSQLiteEngine(db)
	updates := []model.RecordUpdate{
		{
			Identity: model.RecordIdentity{
				Keys: []model.ColumnValue{
					{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}},
				},
			},
			Changes: []model.ColumnValue{
				{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}},
			},
		},
		{
			Identity: model.RecordIdentity{
				Keys: []model.ColumnValue{
					{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}},
				},
			},
			Changes: []model.ColumnValue{
				{Column: "missing_column", Value: model.Value{Text: "oops", Raw: "oops"}},
			},
		},
	}

	// Act
	err := engine.ApplyRecordUpdates(context.Background(), "users", updates)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var name string
	row := db.QueryRow("SELECT name FROM users WHERE id = 1")
	if err := row.Scan(&name); err != nil {
		t.Fatalf("failed to scan row: %v", err)
	}
	if name != "alice" {
		t.Fatalf("expected rollback to keep name %q, got %q", "alice", name)
	}
}

func setupSQLiteUpdateDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:updates?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
	})
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			active BOOLEAN NOT NULL,
			score REAL NOT NULL,
			note TEXT,
			data BLOB
		);
		INSERT INTO users (id, name, active, score, note, data)
		VALUES (1, 'alice', 1, 1.25, 'hello', x'0102');
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}
	return db
}
